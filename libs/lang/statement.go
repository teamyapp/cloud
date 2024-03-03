package lang

import (
	"fmt"
	"strings"
)

type StatementType string

const (
	ExpressionStatementType StatementType = "Expression"
	LetStatementType        StatementType = "Let"
	BlockStatementType      StatementType = "Block"
	IfStatementType         StatementType = "If"
	WhileStatementType      StatementType = "While"
	BreakStatementType      StatementType = "Break"
	ContinueStatementType   StatementType = "Continue"
	CallableStatementType   StatementType = "Callable"
	ReturnStatementType     StatementType = "Return"
	ClassStatementType      StatementType = "Class"
)

type Statement struct {
	NodeID                          uint64
	Type                            StatementType
	Line                            int
	Column                          int
	IsGenerated                     bool
	PrintArgExpression              *Expression
	StatementExpression             *Expression
	LetIdentifier                   *Token
	LetInitializerExpression        *Expression
	BlockInnerStatements            []Statement
	IfConditionExpression           *Expression
	IfTrueBranchStatement           *Statement
	IfFalseBranchStatement          *Statement
	WhileConditionExpression        *Expression
	WhileBodyStatement              *Statement
	CallableName                    *Token
	CallableParameters              []Token
	CallableBody                    *Statement
	ReturnValueExpression           *Expression
	ClassIdentifier                 *Token
	ClassInstanceMethodDeclarations []Statement
	ClassStaticMethodDeclarations   []Statement
	ClassInstanceGetterDeclarations []Statement
	ClassInstanceSetterDeclarations []Statement
	ClassStaticGetterDeclarations   []Statement
	ClassStaticSetterDeclarations   []Statement
}

var _ fmt.Stringer = (*Statement)(nil)

func (s Statement) String() string {
	switch s.Type {
	case ExpressionStatementType:
		return s.StatementExpression.String()
	case LetStatementType:
		return fmt.Sprintf("(let %v %s)", s.LetIdentifier.Lexeme, s.LetInitializerExpression)
	case BlockStatementType:
		var statements []string
		for _, innerStatement := range s.BlockInnerStatements {
			statements = append(statements, innerStatement.String())
		}

		return fmt.Sprintf("{%s}", strings.Join(statements, " "))
	case IfStatementType:
		if s.IfFalseBranchStatement == nil {
			return fmt.Sprintf("(if %s %s)", s.IfConditionExpression, s.IfTrueBranchStatement)
		}

		return fmt.Sprintf("(if %s %s else %s)", s.IfConditionExpression, s.IfTrueBranchStatement, s.IfFalseBranchStatement)
	case WhileStatementType:
		return fmt.Sprintf("(while %s %s)", s.WhileConditionExpression, s.WhileBodyStatement)
	case BreakStatementType:
		return "(break)"
	case ContinueStatementType:
		return "(continue)"
	case CallableStatementType:
		var parameters []string
		for _, parameter := range s.CallableParameters {
			parameters = append(parameters, parameter.Lexeme)
		}

		return fmt.Sprintf("(func %v [%s] %s)", s.CallableName.Lexeme, strings.Join(parameters, " "), s.CallableBody)
	case ReturnStatementType:
		if s.ReturnValueExpression == nil {
			return "(return)"
		}

		return fmt.Sprintf("(return %s)", s.ReturnValueExpression)
	case ClassStatementType:
		var instanceMethods []string
		for _, methodDeclaration := range s.ClassInstanceMethodDeclarations {
			instanceMethods = append(instanceMethods, methodDeclaration.String())
		}

		var staticMethods []string
		for _, methodDeclaration := range s.ClassStaticMethodDeclarations {
			staticMethods = append(staticMethods, methodDeclaration.String())
		}

		var instanceGetters []string
		for _, getterDeclaration := range s.ClassInstanceGetterDeclarations {
			instanceGetters = append(instanceGetters, getterDeclaration.String())
		}

		var instanceSetters []string
		for _, setterDeclaration := range s.ClassInstanceSetterDeclarations {
			instanceSetters = append(instanceSetters, setterDeclaration.String())
		}

		var staticGetters []string
		for _, getterDeclaration := range s.ClassStaticGetterDeclarations {
			staticGetters = append(staticGetters, getterDeclaration.String())
		}

		var staticSetters []string
		for _, setterDeclaration := range s.ClassStaticSetterDeclarations {
			staticSetters = append(staticSetters, setterDeclaration.String())
		}

		return fmt.Sprintf("(class %s [instance (methods %s) (getters %s) (setters %s)] [static (methods %s) (getters %s) (setters %s)])",
			s.ClassIdentifier.Lexeme,
			strings.Join(instanceMethods, " "),
			strings.Join(instanceGetters, " "),
			strings.Join(instanceSetters, " "),
			strings.Join(staticMethods, " "),
			strings.Join(staticGetters, " "),
			strings.Join(staticSetters, " "))
	}

	return fmt.Sprintf("[unknown statement type: %v]", s.Type)
}
