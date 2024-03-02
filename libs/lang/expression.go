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
	CallExpressionType           ExpressionType = "Call"
	LambdaExpressionType         ExpressionType = "Lambda"
	NewInstanceExpressionType    ExpressionType = "NewInstance"
)

type Expression struct {
	NodeID                                    uint64
	Type                                      ExpressionType
	Line                                      int
	Column                                    int
	IsGenerated                               bool
	Literal                                   Token
	Operator                                  Token
	Identifier                                Token
	UnaryExpression                           *Expression
	BinaryLeftExpression                      *Expression
	BinaryRightExpression                     *Expression
	TernaryConditionExpression                *Expression
	TernaryTrueExpression                     *Expression
	TernaryFalseExpression                    *Expression
	GroupInnerExpression                      *Expression
	ExpressionList                            []Expression
	AssignmentIdentifier                      Token
	AssignmentValueExpression                 *Expression
	CallableExpression                        *Expression
	CallArgumentExpressions                   []Expression
	LambdaParameters                          []Token
	LambdaBody                                Statement
	NewInstanceClassIdentifier                Token
	NewInstanceConstructorArgumentExpressions []Expression
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
		return fmt.Sprintf("(= %s %s)", e.AssignmentIdentifier.Lexeme, e.AssignmentValueExpression)
	case CallExpressionType:
		args := make([]string, 0)
		for _, argExpr := range e.CallArgumentExpressions {
			args = append(args, argExpr.String())
		}

		return fmt.Sprintf("(call %s [%v])", e.CallableExpression, strings.Join(args, " "))
	case LambdaExpressionType:
		parameters := make([]string, 0)
		for _, param := range e.LambdaParameters {
			parameters = append(parameters, param.Lexeme)
		}

		return fmt.Sprintf("(lambda [%v] %s)", strings.Join(parameters, " "), e.LambdaBody)
	case NewInstanceExpressionType:
		args := make([]string, 0)
		for _, argExpr := range e.NewInstanceConstructorArgumentExpressions {
			args = append(args, argExpr.String())
		}

		return fmt.Sprintf("(new %s [%v])", e.NewInstanceClassIdentifier.Lexeme, strings.Join(args, " "))
	}

	return fmt.Sprintf("[unknown expression type: %s]", e.Type)
}
