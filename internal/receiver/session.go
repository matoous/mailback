package receiver

import (
	"errors"
	"io"
	"net"
	"strings"
	"time"
	"unicode"

	"blitiri.com.ar/go/spf"
	"github.com/DusanKasan/parsemail"
	"github.com/emersion/go-smtp"
	"go.uber.org/zap"

	"github.com/matoous/mailback/internal/cfg"
	"github.com/matoous/mailback/internal/mail"
	"github.com/matoous/mailback/internal/models"
	"github.com/matoous/mailback/internal/when"
)

// Session is spawned for each incoming smtp request and handles its lifecycle.
type Session struct {
	// TargetTime is the time (and optionally the period) that the email should be scheduled for
	TargetTime *when.Result
	// From is the sender of the email that should receive the email back eventually.
	From string
	// Content is the content of the email that will be send back.
	Content string
	// Title is the subject of the email that will be used in the reply.
	Title string
	// ToUs is true if the email is supposed to be delivered to the owner of the domain
	// instead of scheduled for delivery.
	ToUs bool

	store      Storer
	config     *cfg.ReceiverConfig
	hostname   string
	remoteAddr net.Addr
	log        *zap.Logger
}

// SPFCheck does the Sender Policy Framework check on the email sender.
func (s *Session) SPFCheck(from string) error {
	// try to verify SPF
	remote := s.remoteAddr.String()
	idx := strings.LastIndex(remote, ":")
	if idx > 0 {
		remote = remote[:idx]
	}
	remote = strings.TrimLeft(remote, "[")
	remote = strings.TrimRight(remote, "]")
	s.log.Debug(remote)
	ip := net.ParseIP(remote)
	result, err := spf.CheckHostWithSender(ip, s.hostname, from)
	if err != nil {
		// if we can't check, accept the email
		s.log.Error(
			"session.spf_check",
			zap.Error(err),
			zap.String("ip", string(ip)),
			zap.String("hostname", s.hostname),
			zap.String("from", from),
		)
		return nil
	}
	if result == spf.Fail {
		return errors.New("unauthorized")
	}
	return nil
}

// Mail handles MAIL command setting the From field for the session.
func (s *Session) Mail(from string, _ smtp.MailOptions) error {
	if s.config.Host != "localhost" {
		err := s.SPFCheck(from)
		if err != nil {
			s.log.Error("session.mail.spf_check", zap.Error(err))
			return err
		}
	}
	s.From = from
	return nil
}

// Rcpt handles the SMTP RCPT command.
func (s *Session) Rcpt(to string) error {
	target := mail.User(to)
	if target == "admin" {
		s.ToUs = true
		return nil
	}
	// map symbols to spaces, this allow addresses such as in_2_days@mailback.io
	target = strings.Map(func(r rune) rune {
		if unicode.IsSymbol(r) {
			return ' '
		}
		return r
	}, target)
	target = strings.TrimSpace(target)
	x, err := when.Parse(target, time.Now())
	if err != nil {
		s.log.Error("session.rcpt.parse", zap.Error(err), zap.String("target", target))
		return err
	}
	s.TargetTime = x
	return nil
}

// Data handles the mail data. It reads the received email, creates entry on our sade and saves it into the database.
func (s *Session) Data(r io.Reader) error {
	email, err := parsemail.Parse(r)
	if err != nil {
		s.log.Error("session.entry.parse", zap.Error(err))
		return err
	}

	s.Content = email.TextBody
	s.Title = email.Subject
	// TODO verify the DKIM

	entry, err := models.NewEntry(s.From, s.Content, s.Title, s.TargetTime)
	if err != nil {
		s.log.Error("session.entry.new", zap.Error(err))
		return err
	}

	err = s.store.Save(entry)
	if err != nil {
		s.log.Error("session.entry.save", zap.Error(err))
		return err
	}

	s.log.Info("session.entry.save")
	return nil
}

// Reset resets the session to initial state.
func (s *Session) Reset() {
	s.Title = ""
	s.Content = ""
	s.TargetTime = nil
	s.From = ""
}

// Logout handles SMTP logout command.
func (s *Session) Logout() error {
	return nil
}
