package query

type DataType string

const (
	IntDataType      DataType = "Int"
	DecimalDataType  DataType = "Decimal"
	BoolDataType     DataType = "Bool"
	StringDataType   DataType = "String"
	DatetimeDataType DataType = "Datetime"
	FunctionDataType DataType = "Function"
	StructDataType   DataType = "Struct"
)
