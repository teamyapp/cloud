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
)

type Statement struct {
	Type                     StatementType
	PrintArgExpression       *Expression
	StatementExpression      *Expression
	LetIdentifier            *Token
	LetInitializerExpression *Expression
	BlockInnerStatements     []Statement
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

		return fmt.Sprintf("({} %s)", strings.Join(statements, " "))
	}

	return "unknown statement type"
}
