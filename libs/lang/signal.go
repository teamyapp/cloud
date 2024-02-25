package lang

type SignalType string

const (
	BreakSignalType  SignalType = "Break"
	ReturnSignalType SignalType = "Return"
)

type Signal struct {
	Type        SignalType
	Value       any
	Line        int
	Column      int
	IsGenerated bool
}
