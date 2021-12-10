package channel

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocketOriginChecker func(r *http.Request) bool

var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocket struct {
	connection          *websocket.Conn
	onMessageReceivedCh chan []byte
}

var _ Channel = (*WebSocket)(nil)

func (w WebSocket) Listen() {
	go func() {
		for {
			_, data, err := w.connection.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				break
			}
			w.onMessageReceivedCh <- data
		}
	}()
}

func (w WebSocket) Disconnect() {
	w.connection.Close()
}

func (w WebSocket) SendMessage(message string) error {
	writer, err := w.connection.NextWriter(websocket.TextMessage)
	if err != nil {
		log.Fatal(err)
		return err
	}

	writer.Write([]byte(message))

	if err = writer.Close(); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (w WebSocket) OnMessageReceived() chan []byte {
	return w.onMessageReceivedCh
}

func NewWebSocket(originChecker WebSocketOriginChecker, w http.ResponseWriter, r *http.Request) (*WebSocket, error) {
	upGrader.CheckOrigin = originChecker
	conn, err := upGrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	socket := &WebSocket{
		connection:          conn,
		onMessageReceivedCh: make(chan []byte, 0),
	}
	return socket, nil
}
