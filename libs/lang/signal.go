package lang

type SignalType string

const (
	BreakSignalType    SignalType = "Break"
	ContinueSignalType SignalType = "Continue"
	ReturnSignalType   SignalType = "Return"
)

type Signal struct {
	Type        SignalType
	Value       any
	Line        int
	Column      int
	IsGenerated bool
}
