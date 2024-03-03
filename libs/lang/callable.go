package lang

import "fmt"

type Callable struct {
	Name          string
	IsAnonymous   bool
	IsConstructor bool
	Closure       *Environment
	Arity         int
	Execute       func(callable *Callable, arguments ...any) (any, *Err)
	Line          int
	Column        int
	IsGenerated   bool
}

var _ fmt.Stringer = (*Callable)(nil)

func (c Callable) String() string {
	return fmt.Sprintf("<func %s at (line:%d, column:%d)>",
		c.Name,
		c.Line,
		c.Column)
}

func (c Callable) Copy() Callable {
	return Callable{
		Name:          c.Name,
		IsAnonymous:   c.IsAnonymous,
		IsConstructor: c.IsConstructor,
		Closure:       c.Closure,
		Arity:         c.Arity,
		Execute:       c.Execute,
		Line:          c.Line,
		Column:        c.Column,
		IsGenerated:   c.IsGenerated,
	}
}
