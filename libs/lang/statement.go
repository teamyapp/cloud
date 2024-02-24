package lang

import (
	"fmt"
	"strings"
)

type StatementType string

const (
	ExpressionStatementType StatementType = "Expression"
	PrintStatementType      StatementType = "Print"
	LetStatementType        StatementType = "Let"
	BlockStatementType      StatementType = "Block"
	IfStatementType         StatementType = "If"
	WhileStatementType      StatementType = "While"
)

type Statement struct {
	Type                     StatementType
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
	Line                     int
	Column                   int
}

var _ fmt.Stringer = (*Statement)(nil)

func (s Statement) String() string {
	switch s.Type {
	case ExpressionStatementType:
		return s.StatementExpression.String()
	case PrintStatementType:
		return fmt.Sprintf("(print %s)", s.PrintArgExpression)
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
	}

	return "unknown statement type"
}
