package obs

type Logger interface {
	Log(severity Severity, properties Props)
}
