package server

import (
	"crypto/tls"
	"errors"
	"net/http"

	"github.com/caddyserver/certmagic"
	"github.com/go-acme/lego/v3/providers/dns/cloudflare"
	"go.uber.org/zap"

	"github.com/gofiber/fiber"

	"github.com/matoous/mailback/internal/cfg"
	"github.com/matoous/mailback/internal/models"
	"github.com/matoous/mailback/internal/store"
)

// Store is storage of entries that can delete an entry by the id from unsubscribe link.
type Store interface {
	Delete(e *models.Entry) error
}

// Server is web server.
type Server struct {
	store     Store
	log       *zap.Logger
	router    *fiber.App
	tlsConfig *tls.Config
	port      string
}

func (s *Server) handleIndex(ctx *fiber.Ctx) {
	if err := ctx.Render("index.html", map[string]interface{}{}); err != nil {
		s.log.Error("server.index.render", zap.Error(err))
	}
}

func (s *Server) handleUnsubscribe(ctx *fiber.Ctx) {
	err := s.store.Delete(&models.Entry{ID: ctx.Params("id")})
	switch {
	case errors.Is(err, store.ErrNotFound):
		ctx.Status(http.StatusNotFound)
		ctx.SendString("Didn't find subscription with this ID")
	case err != nil:
		ctx.Status(http.StatusInternalServerError)
	default:
		ctx.SendString("Unsubscribed!")
	}
}

// New creates new server that can handle clicks on unsubscribe links.
func New(s Store, l *zap.Logger, config cfg.WebServerConfig) (*Server, error) {
	srv := &Server{
		store: s,
		log:   l,
		port:  config.Port,
	}

	if config.Host != "localhost" {
		provider, err := cloudflare.NewDNSProvider()
		if err != nil {
			return nil, err
		}
		certmagic.DefaultACME.DNSProvider = provider
		tlsConfig, err := certmagic.TLS([]string{config.Host})
		if err != nil {
			return nil, err
		}
		srv.tlsConfig = tlsConfig
	}

	router := fiber.New(&fiber.Settings{
		ServerHeader:   "Sendback.mail",
		TemplateFolder: "./templates",
		TemplateEngine: "html",
	})
	router.Get("/unsubscribe/:id", srv.handleUnsubscribe)
	router.Get("/", srv.handleIndex)
	router.Static("/", "./public")

	srv.router = router
	return srv, nil
}

// Run runs the server.
func (s *Server) Run() error {
	if s.tlsConfig != nil {
		return s.router.Listen(s.port, s.tlsConfig)
	}
	return s.router.Listen(s.port)
}
