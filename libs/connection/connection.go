package connection

import (
	"github.com/teamyapp/cloud/libs/errs"
)

const ConnErr errs.ErrorCode = "Connection"

type Connection interface {
	OnErrors() <-chan errs.Error
	OnMessageReceived() <-chan []byte
	SendMessage(message []byte)
	OnClientDisconnect() <-chan bool
	Close() *errs.Error
}
