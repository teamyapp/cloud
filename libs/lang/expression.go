package lang

import (
	"fmt"
	"strings"
)

type ExpressionType string

const (
	TernaryExpressionType        ExpressionType = "Ternary"
	BinaryExpressionType         ExpressionType = "Binary"
	UnaryExpressionType          ExpressionType = "Unary"
	LiteralExpressionType        ExpressionType = "Literal"
	GroupingExpressionType       ExpressionType = "Grouping"
	ExpressionListExpressionType ExpressionType = "ExpressionList"
	IdentifierExpressionType     ExpressionType = "Identifier"
	AssignmentExpressionType     ExpressionType = "Assignment"
)

type Expression struct {
	Type                       ExpressionType
	Line                       int
	Column                     int
	Literal                    Token
	Operator                   Token
	UnaryExpression            *Expression
	BinaryLeftExpression       *Expression
	BinaryRightExpression      *Expression
	TernaryConditionExpression *Expression
	TernaryTrueExpression      *Expression
	TernaryFalseExpression     *Expression
	GroupInnerExpression       *Expression
	ExpressionList             []Expression
	Identifier                 Token
	AssignmentValueExpression  *Expression
}

var _ fmt.Stringer = (*Expression)(nil)

func (e Expression) String() string {
	switch e.Type {
	case LiteralExpressionType:
		return fmt.Sprintf("%s", e.Literal.Lexeme)
	case UnaryExpressionType:
		return fmt.Sprintf("(%s %s)", e.Operator.Lexeme, e.UnaryExpression)
	case BinaryExpressionType:
		return fmt.Sprintf("(%s %s %s)", e.Operator.Lexeme, e.BinaryLeftExpression, e.BinaryRightExpression)
	case GroupingExpressionType:
		return fmt.Sprintf("(group %s)", e.GroupInnerExpression)
	case TernaryExpressionType:
		return fmt.Sprintf("(?: %s %s %s)", e.TernaryConditionExpression, e.TernaryTrueExpression, e.TernaryFalseExpression)
	case ExpressionListExpressionType:
		expressions := make([]string, 0)
		for _, expr := range e.ExpressionList {
			expressions = append(expressions, expr.String())
		}

		return fmt.Sprintf("(expressionList %s)", strings.Join(expressions, " "))
	case IdentifierExpressionType:
		return fmt.Sprintf("%s", e.Identifier.Lexeme)
	case AssignmentExpressionType:
		return fmt.Sprintf("(= %s %s)", e.Identifier.Lexeme, e.AssignmentValueExpression)
	default:
		return "[unknown expression type]"
	}
}
