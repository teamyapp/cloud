package matcher

type Operator string

const (
	NoneOperator                 Operator = ""
	AndOperator                  Operator = "And"
	OrOperator                   Operator = "Or"
	NotOperator                  Operator = "Not"
	EqualToOperator              Operator = "=="
	ContainsOperator             Operator = "Contains"
	LessThanOperator             Operator = "<"
	LessThanOrEqualToOperator    Operator = "<="
	GreaterThanOperator          Operator = ">"
	GreaterThanOrEqualToOperator Operator = ">="
)
