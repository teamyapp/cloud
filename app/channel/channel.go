package channel

type Channel interface {
	SendMessage(message string) error
	OnMessageReceived() chan []byte
	Disconnect()
	Listen()
}
