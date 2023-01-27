package telemetry

type LogLevel string

const (
	Off     LogLevel = "Off"
	Fatal   LogLevel = "Fatal"
	Error   LogLevel = "Error"
	Warning LogLevel = "Warning"
	Info    LogLevel = "Info"
	Debug   LogLevel = "Debug"
)

var logLevelRank = map[LogLevel]int{
	Off:     0,
	Fatal:   1,
	Error:   2,
	Warning: 3,
	Info:    4,
	Debug:   5,
}
