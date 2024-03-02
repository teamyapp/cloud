package lang

import "fmt"

type IdentifierAndValue struct {
	Identifier    string
	Value         any
	IsInitialized bool
}

type Environment struct {
	outerEnvironment    *Environment
	IdentifierAndValues []IdentifierAndValue
}

func (e *Environment) Define(identifierToken Token) {
	identifier := identifierToken.Value.(string)
	e.IdentifierAndValues = append(e.IdentifierAndValues, IdentifierAndValue{
		Identifier:    identifier,
		IsInitialized: false,
	})
}

func (e *Environment) DefineWithInitializer(identifierToken Token, value any) {
	identifier := identifierToken.Value.(string)
	e.IdentifierAndValues = append(e.IdentifierAndValues, IdentifierAndValue{
		Identifier:    identifier,
		Value:         value,
		IsInitialized: true,
	})
}

func (e *Environment) Get(identifierToken Token) (any, *Err) {
	identifier := identifierToken.Value.(string)
	for index := len(e.IdentifierAndValues) - 1; index >= 0; index-- {
		identifierAndValue := e.IdentifierAndValues[index]
		if identifierAndValue.Identifier == identifier {
			if !identifierAndValue.IsInitialized {
				return nil, &Err{
					Message:           fmt.Sprintf("uninitialized identifier: %v", identifier),
					Line:              identifierToken.Line,
					Column:            identifierToken.Column,
					FromGeneratedCode: identifierToken.IsGenerated,
				}
			}

			return identifierAndValue.Value, nil
		}
	}

	if e.outerEnvironment != nil {
		return e.outerEnvironment.Get(identifierToken)
	}

	return nil, &Err{
		Message:           fmt.Sprintf("undefined identifier: %v", identifier),
		Line:              identifierToken.Line,
		Column:            identifierToken.Column,
		FromGeneratedCode: identifierToken.IsGenerated,
	}
}

func (e *Environment) GetAt(identifier Token, ref *Reference) (any, *Err) {
	ancestor := e.ancestorAt(ref.EnvironmentDistance)
	if ancestor == nil {
		return nil, nil
	}

	return ancestor.getAtStackIndex(identifier, ref.StackIndex)
}

func (e *Environment) Assign(identifierToken Token, value any) *Err {
	identifier := identifierToken.Value.(string)
	for index := len(e.IdentifierAndValues) - 1; index >= 0; index-- {
		identifierAndValue := e.IdentifierAndValues[index]
		if identifierAndValue.Identifier == identifier {
			if !identifierAndValue.IsInitialized {
				return &Err{
					Message:           fmt.Sprintf("uninitialized identifier: %v", identifier),
					Line:              identifierToken.Line,
					Column:            identifierToken.Column,
					FromGeneratedCode: identifierToken.IsGenerated,
				}
			}

			identifierAndValue.Value = value
			e.IdentifierAndValues[index] = identifierAndValue
			return nil
		}
	}

	if e.outerEnvironment != nil {
		return e.outerEnvironment.Assign(identifierToken, value)
	}

	return &Err{
		Message:           fmt.Sprintf("undefined identifier: %v", identifier),
		Line:              identifierToken.Line,
		Column:            identifierToken.Column,
		FromGeneratedCode: identifierToken.IsGenerated,
	}
}

func (e *Environment) AssignAt(identifierToken Token, ref *Reference, value any) *Err {
	ancestor := e.ancestorAt(ref.EnvironmentDistance)
	if ancestor == nil {
		return nil
	}

	return ancestor.assignAtStackIndex(identifierToken, ref.StackIndex, value)
}

func (e *Environment) NewInnerEnvironment() *Environment {
	innerEnvironment := NewEnvironment()
	innerEnvironment.outerEnvironment = e
	return innerEnvironment
}

func (e *Environment) AttachOuterEnvironment(outerEnvironment *Environment) {
	e.outerEnvironment = outerEnvironment
}

func (e *Environment) getAtStackIndex(identifier Token, stackIndex int) (any, *Err) {
	if stackIndex < 0 || stackIndex >= len(e.IdentifierAndValues) {
		return nil, &Err{
			Message:           "invalid identifier index",
			Line:              identifier.Line,
			Column:            identifier.Column,
			FromGeneratedCode: identifier.IsGenerated,
		}
	}

	identifierAndValue := e.IdentifierAndValues[stackIndex]
	if !identifierAndValue.IsInitialized {
		return nil, &Err{
			Message:           fmt.Sprintf("uninitialized identifier: %v", identifierAndValue.Identifier),
			Line:              identifier.Line,
			Column:            identifier.Column,
			FromGeneratedCode: identifier.IsGenerated,
		}
	}

	return identifierAndValue.Value, nil
}

func (e *Environment) ancestorAt(environmentDistance int) *Environment {
	environment := e
	for index := 0; index < environmentDistance; index++ {
		environment = environment.outerEnvironment
	}

	return environment
}

func (e *Environment) assignAtStackIndex(identifier Token, stackIndex int, value any) *Err {
	if stackIndex < 0 || stackIndex >= len(e.IdentifierAndValues) {
		return &Err{
			Message:           "invalid identifier index",
			Line:              identifier.Line,
			Column:            identifier.Column,
			FromGeneratedCode: identifier.IsGenerated,
		}
	}

	identifierAndValue := e.IdentifierAndValues[stackIndex]
	identifierAndValue.Value = value
	identifierAndValue.IsInitialized = true
	e.IdentifierAndValues[stackIndex] = identifierAndValue
	return nil
}

func NewEnvironment() *Environment {
	return &Environment{
		IdentifierAndValues: []IdentifierAndValue{},
	}
}
