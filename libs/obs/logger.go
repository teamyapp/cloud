package obs

type Logger interface {
	Log(logLevel LogLevel, properties Props)
}
