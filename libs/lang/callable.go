package lang

import "fmt"

type CallableType string

const (
	FunctionCallableType CallableType = "Function"
	MethodCallableType   CallableType = "Method"
)

type Callable struct {
	Name        string
	IsAnonymous bool
	Closure     *Environment
	Arity       int
	Execute     func(closure *Environment, arguments ...any) (any, *Err)
	Line        int
	Column      int
}

var _ fmt.Stringer = (*Callable)(nil)

func (c Callable) String() string {
	return fmt.Sprintf("<func %s at (line:%d, column:%d)>",
		c.Name,
		c.Line,
		c.Column)
}
