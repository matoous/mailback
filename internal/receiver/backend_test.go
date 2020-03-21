package receiver

import (
	"testing"

	"github.com/emersion/go-smtp"
	"github.com/stretchr/testify/assert"
)

func TestBackend_Login(t *testing.T) {
	be := &Receiver{}
	_, err := be.Login(&smtp.ConnectionState{}, "", "")
	assert.Error(t, err, "should return error")
	assert.Equal(t, err, smtp.ErrAuthUnsupported, "should return ErrAuthUnsupported")
}
