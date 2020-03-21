package cfg

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	Foo        string `env:"FOO"`
	BAR        int    `env:"BAR"`
	Secret     string `env:"SECRET"`
	SecretFile string `env:"SECRET_FILE"`
}

func TestLoadConfigs(t *testing.T) {

	tests := []struct {
		Name    string
		Env     map[string]string
		WantErr bool
		Want    TestConfig
	}{
		{
			Name: "successfully load config",
			Env: map[string]string{
				"FOO": "my_foo_value",
				"BAR": "1337",
			},
			WantErr: false,
			Want: TestConfig{
				Foo: "my_foo_value",
				BAR: 1337,
			},
		},
		{
			Name: "fail on wrong config",
			Env: map[string]string{
				"FOO": "my_foo_value",
				"BAR": "my_bar_value",
			},
			WantErr: true,
			Want: TestConfig{
				Foo: "my_foo_value",
				BAR: 0,
			},
		},
		{
			Name: "secret value",
			Env: map[string]string{
				"FOO":    "my_foo_value",
				"BAR":    "my_bar_value",
				"SECRET": "my_secret",
			},
			WantErr: true,
			Want: TestConfig{
				Foo:    "my_foo_value",
				BAR:    0,
				Secret: "my_secret",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			for k, v := range tt.Env {
				os.Setenv(k, v)
			}
			var c TestConfig
			err := LoadConfigs(&c)
			if tt.WantErr {
				assert.Error(t, err, "should return error")
			} else {
				assert.NoError(t, err, "shouldn't return error")
				assert.Equal(t, tt.Want, c, "should parse config")
			}
			for k := range tt.Env {
				os.Unsetenv(k)
			}
		})
	}
}
