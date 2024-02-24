package lang

type Err struct {
	Message           string
	Line              int
	Column            int
	FromGeneratedCode bool
}
