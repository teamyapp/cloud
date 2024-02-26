package lang

import "fmt"

type Environment struct {
	outerEnvironment  *Environment
	identifierToValue map[string]any
	isInitialized     map[string]bool
}

func (e *Environment) Define(identifierToken Token) {
	identifier := identifierToken.Value.(string)
	e.identifierToValue[identifier] = nil
}

func (e *Environment) DefineWithInitializer(identifierToken Token, value any) {
	identifier := identifierToken.Value.(string)
	e.identifierToValue[identifier] = value
	e.isInitialized[identifier] = true
}

func (e *Environment) Get(identifierToken Token) (any, *Err) {
	identifier := identifierToken.Value.(string)
	value, ok := e.identifierToValue[identifier]
	if !ok {
		if e.outerEnvironment != nil {
			return e.outerEnvironment.Get(identifierToken)
		}

		return nil, &Err{
			Message: fmt.Sprintf("undefined identifier: %v", identifier),
			Line:    identifierToken.Line,
			Column:  identifierToken.Column,
		}
	}

	if !e.isInitialized[identifier] {
		return nil, &Err{
			Message: fmt.Sprintf("uninitialized identifier: %v", identifier),
			Line:    identifierToken.Line,
			Column:  identifierToken.Column,
		}
	}

	return value, nil
}

func (e *Environment) Assign(identifierToken Token, value any) *Err {
	identifier := identifierToken.Value.(string)
	_, ok := e.identifierToValue[identifier]
	if !ok {
		if e.outerEnvironment != nil {
			return e.outerEnvironment.Assign(identifierToken, value)
		}

		return &Err{
			Message: fmt.Sprintf("undefined identifier: %v", identifier),
			Line:    identifierToken.Line,
			Column:  identifierToken.Column,
		}
	}

	e.identifierToValue[identifier] = value
	e.isInitialized[identifier] = true
	return nil
}

func (e *Environment) NewInnerEnvironment() *Environment {
	innerEnvironment := NewEnvironment()
	innerEnvironment.outerEnvironment = e
	return innerEnvironment
}

func (e *Environment) AttachOuterEnvironment(outerEnvironment *Environment) {
	e.outerEnvironment = outerEnvironment
}

func NewEnvironment() *Environment {
	return &Environment{
		identifierToValue: make(map[string]any),
		isInitialized:     make(map[string]bool),
	}
}
