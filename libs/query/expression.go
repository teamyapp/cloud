package query

type ExpressionType string

const (
	InvocationExpressionType ExpressionType = "Invocation"
	IdentifierExpressionType ExpressionType = "Identifier"
	ValueExpressionType      ExpressionType = "Value"
)

type Identifier struct {
	Name     string
	DataType DataType
}

type Expression struct {
	Type            ExpressionType
	Identifier      string
	Value           any
	ValueType       DataType
	FuncExpression  *Expression
	FuncInputValues []Expression
}
