package sender

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/emersion/go-msgauth/dkim"
	"github.com/jpillora/backoff"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/matoous/mailback/internal/cfg"
	"github.com/matoous/mailback/internal/mail"
	"github.com/matoous/mailback/internal/models"
)

const mailTemplate = `
To: {{.to}}\r\n
From: {{.from}}\r\n
Subject: {{.subject}}\r\n
\r\n
{{.content}}

------
{{.banner}}
\r\n
`

// Storer is storage that can list the entries pending for being send, update and delete them.
type Storer interface {
	Update(e *models.Entry) error
	Delete(e *models.Entry) error
	PendingEntries() ([]models.Entry, error)
}

// Sender sends mails back to the users when the time comes.
type Sender struct {
	db       Storer
	log      *zap.Logger
	dkimOpts *dkim.SignOptions
	config   *cfg.SenderConfig
}

func loadPrivateKey(path string) (crypto.Signer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	if err := f.Close(); err != nil {
		return nil, err
	}

	block, _ := pem.Decode(b)
	if block == nil {
		return nil, fmt.Errorf("no PEM data found")
	}

	switch strings.ToUpper(block.Type) {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "EDDSA PRIVATE KEY":
		if len(block.Bytes) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("invalid Ed25519 private key size")
		}
		return ed25519.PrivateKey(block.Bytes), nil
	default:
		return nil, fmt.Errorf("unknown private key type: '%v'", block.Type)
	}
}

// New creates new un-started sender.
func New(storage Storer, log *zap.Logger, config *cfg.SenderConfig) *Sender {
	sender := &Sender{
		db:     storage,
		log:    log,
		config: config,
	}
	if config.Cert != "" {
		signer, err := loadPrivateKey(config.Cert)
		if err != nil {
			panic(err)
		}
		sender.dkimOpts = &dkim.SignOptions{
			Domain:   config.Host,
			Selector: config.CertSelector,
			Signer:   signer,
		}
	}
	return sender
}

// PrepareMail prepares the email body.
func (s *Sender) PrepareMail(e *models.Entry) ([]byte, error) {
	banner := ""
	if e.Period != nil {
		var unsubscribeLink string
		if s.config.Host == "localhost" {
			unsubscribeLink = fmt.Sprintf("http://%s/unsubscribe/%s", s.config.Host, e.ID)
		} else {
			unsubscribeLink = fmt.Sprintf("https://%s/unsubscribe/%s", s.config.Host, e.ID)
		}
		banner = fmt.Sprintf("\n\n---\nThis is a periodic email that you will receive every %s\n"+
			"To unsubscribe click here: %s\n", e.Period.Format(), unsubscribeLink)
	}
	msg := fmt.Sprintf("To: %s\r\n"+
		"From: %s <%s@%s>\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s%s\r\n", e.Mail, s.config.SenderName, s.config.SenderMail, s.config.Host, e.Title, e.Data, banner)

	if s.dkimOpts == nil {
		return []byte(msg), nil
	}

	var res bytes.Buffer
	err := dkim.Sign(&res, strings.NewReader(msg), s.dkimOpts)
	if err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}

// Send sends the entry to SMTP server of the email receiver.
func (s *Sender) Send(host string, e *models.Entry) error {
	msg, err := s.PrepareMail(e)
	if err != nil {
		return err
	}

	serverName := fmt.Sprintf("%s:%d", host, 25)
	return smtp.SendMail(serverName, nil, "testing@mailback.io", []string{e.Mail}, msg)
}

// ProcessEntry processes single entry. This means sending the scheduled entry back to the user and in case
// of periodical entry rescheduling it for next time. This process can be run concurrently on all entries
// that need to be processed.
func (s *Sender) ProcessEntry(e *models.Entry) error {
	// sanity check
	if time.Now().Before(e.ScheduledFor) {
		return nil
	}

	domain := mail.Host(e.Mail)

	host, err := mail.MXRecordForHost(domain)
	if err != nil {
		s.log.Error("sender.process_entry.mx_records", zap.Error(err))
		return err
	}

	err = s.Send(host, e)
	if err != nil {
		e.Fails++
		if e.Fails >= 3 {
			// too many failures, give up
			s.log.Error("sender.process_entry.send", zap.String("reason", "too many failures"), zap.String("to", e.Mail))
			return s.db.Delete(e)
		}
		bo := backoff.Backoff{
			Min:    5 * time.Minute,
			Max:    time.Hour,
			Factor: 2,
			Jitter: true,
		}
		e.ScheduledFor = e.ScheduledFor.Add(bo.ForAttempt(float64(e.Fails)))
		updateErr := s.db.Update(e)
		if updateErr != nil {
			s.log.Error("sender.process_entry.reschedule", zap.Error(err))
			return updateErr
		}
		return err
	}

	if e.Period != nil {
		// reschedule
		e.ScheduledFor, _ = e.Period.AddTo(e.ScheduledFor)
		e.Fails = 0 // reset the failures
		return s.db.Update(e)
	}
	// delete
	return s.db.Delete(e)
}

// SendMails attempts to send all emails that are due their scheduled for date back to their originators.
func (s *Sender) SendMails(ctx context.Context) error {
	entries, err := s.db.PendingEntries()
	if err != nil {
		return err
	}

	g, gCtx := errgroup.WithContext(ctx)

	entriesChan := make(chan models.Entry)

	g.Go(func() error {
		defer close(entriesChan)
		for i := range entries {
			select {
			case entriesChan <- entries[i]:
			case <-gCtx.Done():
				return gCtx.Err()
			}
		}
		return nil
	})

	for i := 0; i < s.config.WorkerCount; i++ {
		g.Go(func() error {
			for entry := range entriesChan {
				entry := entry
				err := s.ProcessEntry(&entry)
				if err != nil {
					return err
				}
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	s.log.Info("sender.send_mails", zap.Int("mails_send", len(entries)))
	return nil
}

// Run runs the sender, periodically querying for emails that should be send, stopping only when the passed
// context is canceled.
func (s *Sender) Run(ctx context.Context) {
	s.log.Info("sender.start")
	t := time.NewTicker(s.config.Tick)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.log.Info("sender.tick")
			err := s.SendMails(ctx)
			if err != nil {
				s.log.Error("sender.send_mails", zap.Error(err))
			}
		}
	}
}
