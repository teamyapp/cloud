package obs

type VisibleLevel string

const (
	Off     VisibleLevel = "Off"
	Fatal   VisibleLevel = "Fatal"
	Error   VisibleLevel = "Error"
	Warning VisibleLevel = "Warning"
	Info    VisibleLevel = "Info"
	Debug   VisibleLevel = "Debug"
)

var visibleLevelRank = map[VisibleLevel]int{
	Off:     0,
	Fatal:   1,
	Error:   2,
	Warning: 3,
	Info:    4,
	Debug:   5,
}
