package obs

type LogLevel string

const (
	Fatal   LogLevel = "FATAL"
	Error            = "ERROR"
	Warning          = "WARNING"
	Info             = "INFO"
	Debug            = "DEBUG"
)

var severity = map[LogLevel]int{
	Debug:   0,
	Info:    1,
	Warning: 2,
	Error:   3,
	Fatal:   4,
}
