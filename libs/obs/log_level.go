package obs

type Severity string

const (
	Fatal   Severity = "Fatal"
	Error            = "Error"
	Warning          = "Warning"
	Info             = "Info"
	Debug            = "Debug"
)

var severities = map[Severity]int{
	Debug:   0,
	Info:    1,
	Warning: 2,
	Error:   3,
	Fatal:   4,
}
