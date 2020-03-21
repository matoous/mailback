// Package cfg provides primitives for config loading.
// Internally it uses https://github.com/caarlos0/env library for parsing.
package cfg

import (
	"time"

	"github.com/caarlos0/env/v6"
)

// LoadConfigs loads environment variables into provided configStructs pointers.
// It uses https://github.com/caarlos0/env library for parsing. So provided structs must
// be provided with tags according to this library.
//
// WARNING: configStructs should be pointers!
func LoadConfigs(configStructs ...interface{}) error {
	for _, c := range configStructs {
		if err := env.Parse(c); err != nil {
			return err
		}
	}
	return nil
}

// WebServerConfig ...
type WebServerConfig struct {
	Host string `env:"HOST" envDefault:"localhost"`
	Port string `env:"SERVER_PORT" envDefault:":8080"`
}

// ReceiverConfig ...
type ReceiverConfig struct {
	Host string `env:"HOST" envDefault:"localhost"`
	Port string `env:"RECEIVER_PORT" envDefault:":25"`
}

// SenderConfig ...
type SenderConfig struct {
	Host         string        `env:"HOST" envDefault:"localhost"`
	Tick         time.Duration `env:"SENDER_TICK" envDefault:"1m"`
	Cert         string        `env:"SENDER_CERT"`
	CertSelector string        `env:"SENDER_CERT_SELECTOR" envDefault:"blahblah"`
	WorkerCount  int           `env:"SENDER_WORKER_COUNT" envDefault:"16"`
	SenderMail   string        `env:"SENDER_MAIL" envDefault:"postman"`
	SenderName   string        `env:"SENDER_NAME" envDefault:"Mailback Postman"`
}

// StorageConfig ...
type StorageConfig struct {
	Database string `env:"DATABASE" envDefault:"test.db"`
}
