package connection

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
)

var WebSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocket struct {
	receiveMessageCh chan []byte
	sendMessageChan  chan []byte
	errorCh          chan errs.Error
	disconnectCh     chan bool
	conn             *websocket.Conn
}

var _ Connection = (*WebSocket)(nil)

func (w WebSocket) OnErrors() <-chan errs.Error {
	return w.errorCh
}

func (w WebSocket) OnMessageReceived() <-chan []byte {
	return w.receiveMessageCh
}

func (w WebSocket) SendMessage(message []byte) {
	w.sendMessageChan <- message
}

func (w WebSocket) OnClientDisconnect() <-chan bool {
	return w.disconnectCh
}

func (w WebSocket) Close() *errs.Error {
	err := w.conn.Close()
	if err == nil {
		return nil
	}

	return &errs.Error{
		Code:     ConnErr,
		EmbedErr: err,
	}
}

func NewWebSocket(dataCollector telemetry.DataCollector, conn *websocket.Conn) WebSocket {
	receiveMessageCh := make(chan []byte)
	sendMessageCh := make(chan []byte, 500)
	errorCh := make(chan errs.Error)
	disconnectCh := make(chan bool)
	conn.SetCloseHandler(func(code int, text string) error {
		disconnectCh <- true
		return nil
	})
	go func() {
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				internalErr := errs.Error{
					Code:     errs.IO,
					EmbedErr: err,
				}
				select {
				case errorCh <- internalErr:
				default:
				}
				return
			}

			if mt != websocket.TextMessage && mt != websocket.BinaryMessage {
				continue
			}

			select {
			case receiveMessageCh <- message:
			default:
			}
		}
	}()
	go func() {
		for message := range sendMessageCh {
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				internalErr := errs.Error{
					Code:     errs.IO,
					EmbedErr: err,
				}
				dataCollector.Logger.Log(telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
				select {
				case errorCh <- internalErr:
				default:
				}
				return
			}
		}
	}()
	return WebSocket{
		conn:             conn,
		receiveMessageCh: receiveMessageCh,
		sendMessageChan:  sendMessageCh,
		errorCh:          errorCh,
		disconnectCh:     disconnectCh,
	}
}
