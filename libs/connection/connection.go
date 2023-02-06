package connection

import (
	"github.com/teamyapp/cloud/libs/errs"
)

type Connection interface {
	OnErrors() <-chan errs.Error
	OnMessageReceived() <-chan []byte
	SendMessage(message []byte)
	OnClientDisconnect() <-chan bool
	Close() *errs.Error
}
