package receiver

import (
	"fmt"
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/emersion/go-smtp"
	"github.com/go-acme/lego/v3/providers/dns/cloudflare"
	"go.uber.org/zap"

	"github.com/matoous/mailback/internal/cfg"
	"github.com/matoous/mailback/internal/models"
)

// Storer can save entries into some kind of storage that allows their retrieval later on.
type Storer interface {
	Save(e *models.Entry) error
}

// Receiver implements `smtp.Receiver` and is used to handle all incoming smtp connection.
// Receiver spawns session for individual requests and handles authorization -
// in our case accepts only unauthorized requests.
type Receiver struct {
	storer Storer
	log    *zap.Logger
	srv    *smtp.Server
	config cfg.ReceiverConfig
}

// New creates new receiver.
func New(s Storer, log *zap.Logger, config cfg.ReceiverConfig) (*Receiver, error) {
	rc := &Receiver{
		storer: s,
		log:    log,
	}

	srv := smtp.NewServer(rc)

	if config.Host != "localhost" {
		provider, err := cloudflare.NewDNSProvider()
		if err != nil {
			return nil, fmt.Errorf("cert dns provider: %w", err)
		}

		certmagic.DefaultACME.DNSProvider = provider
		tlsConfig, err := certmagic.TLS([]string{config.Host})
		if err != nil {
			return nil, fmt.Errorf("get tls certificate: %w", err)
		}
		srv.TLSConfig = tlsConfig
	}

	srv.Addr = config.Port
	srv.Domain = config.Host
	srv.ReadTimeout = 10 * time.Second
	srv.WriteTimeout = 10 * time.Second
	srv.MaxMessageBytes = 1024 * 1024
	srv.MaxRecipients = 16
	rc.srv = srv

	return rc, nil
}

// Login implements `smtp.Receiver` interface function `Login` that should be used to authorize the incoming message.
// In our case authorization is disabled, we do not rely any emails, we just accept the ones for us.
func (be *Receiver) Login(_ *smtp.ConnectionState, _, _ string) (smtp.Session, error) {
	return nil, smtp.ErrAuthUnsupported
}

// AnonymousLogin implements `smtp.Receiver` interface function `AnonymousLogin` that is used for anonymous requests.
// This returns fresh new session.
func (be *Receiver) AnonymousLogin(c *smtp.ConnectionState) (smtp.Session, error) {
	return &Session{
		config:     &be.config,
		store:      be.storer,
		hostname:   c.Hostname,
		remoteAddr: c.RemoteAddr,
		log:        be.log,
	}, nil
}

// Run runs the receiver.
func (be *Receiver) Run() error {
	be.log.Info("receiver.run", zap.String("port", be.config.Port), zap.Bool("tls", false))
	return be.srv.ListenAndServe()
}
