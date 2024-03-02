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
	NodeID                   uint64
	Type                     StatementType
	Line                     int
	Column                   int
	IsGenerated              bool
	PrintArgExpression       *Expression
	StatementExpression      *Expression
	LetIdentifier            *Token
	LetInitializerExpression *Expression
	BlockInnerStatements     []Statement
	IfConditionExpression    *Expression
	IfTrueBranchStatement    *Statement
	IfFalseBranchStatement   *Statement
	WhileConditionExpression *Expression
	WhileBodyStatement       *Statement
	CallableName             *Token
	CallableParameters       []Token
	CallableBody             *Statement
	ReturnValueExpression    *Expression
	ClassIdentifier          *Token
	ClassMethodDeclarations  []Statement
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
		var methodDeclarations []string
		for _, methodDeclaration := range s.ClassMethodDeclarations {
			methodDeclarations = append(methodDeclarations, methodDeclaration.String())
		}

		return fmt.Sprintf("(class %v %s)", s.ClassIdentifier.Value, strings.Join(methodDeclarations, " "))
	}

	return fmt.Sprintf("[unknown statement type: %v]", s.Type)
}
