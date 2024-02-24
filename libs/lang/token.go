package lang

type TokenType string

const (
	IdentifierTokenType TokenType = "Identifier"

	IfKeywordTokenType     TokenType = "IfKeyword"
	ElseKeywordTokenType   TokenType = "ElseKeyword"
	ForKeywordTokenType    TokenType = "ForKeyword"
	WhileKeywordTokenType  TokenType = "WhileKeyword"
	BreakKeywordTokenType  TokenType = "BreakKeyword"
	SwitchKeywordTokenType TokenType = "SwitchKeyword"
	CaseKeywordTokenType   TokenType = "CaseKeyword"
	ReturnKeywordTokenType TokenType = "ReturnKeyword"
	FuncKeywordTokenType   TokenType = "FuncKeyword"
	SuperKeywordTokenType  TokenType = "SuperKeyword"
	ThisKeywordTokenType   TokenType = "ThisKeyword"
	LetKeywordTokenType    TokenType = "LetKeyword"
	ClassKeywordTokenType  TokenType = "ClassKeyword"
	PrintKeywordTokenType  TokenType = "PrintKeyword"

	LeftParenthesisTokenType  TokenType = "LeftParenthesis"
	RightParenthesisTokenType TokenType = "RightParenthesis"
	LeftBraceTokenType        TokenType = "LeftBrace"
	RightBraceTokenType       TokenType = "RightBrace"
	LeftBracketTokenType      TokenType = "LeftBracket"
	RightBracketTokenType     TokenType = "RightBracket"
	CommaTokenType            TokenType = "Comma"
	SemicolonTokenType        TokenType = "Semicolon"
	ColonTokenType            TokenType = "Colon"
	DotTokenType              TokenType = "Dot"
	StarTokenType             TokenType = "Star"
	QuestionMarkTokenType     TokenType = "QuestionMark"

	AssignTokenType             TokenType = "Assign"
	AddTokenType                TokenType = "Add"
	AddAssignTokenType          TokenType = "AddAssign"
	MinusTokenType              TokenType = "Minus"
	SubtractAssignTokenType     TokenType = "SubtractAssign"
	MultiplyAssignTokenType     TokenType = "MultiplyAssign"
	DivideTokenType             TokenType = "Divide"
	DivideAssignTokenType       TokenType = "DivideAssign"
	ModuloTokenType             TokenType = "Modulo"
	ModuloAssignTokenType       TokenType = "ModuloAssign"
	LogicalEqualTokenType       TokenType = "LogicalEqual"
	LogicalAndTokenType         TokenType = "LogicalAnd"
	LogicalOrTokenType          TokenType = "LogicalOr"
	LogicalNotTokenType         TokenType = "LogicalNot"
	LogicalNotEqualTokenType    TokenType = "LogicalNotEqual"
	GreaterThanTokenType        TokenType = "GreaterThan"
	GreaterThanOrEqualTokenType TokenType = "GreaterThanOrEqual"
	LessThanTokenType           TokenType = "LessThan"
	LessThanOrEqualTokenType    TokenType = "LessThanOrEqual"
	BitwiseAndTokenType         TokenType = "BitwiseAnd"
	BitwiseOrTokenType          TokenType = "BitwiseOr"
	BitwiseXorTokenType         TokenType = "BitwiseXor"
	BitwiseNotTokenType         TokenType = "BitwiseNot"
	BitwiseAndAssignTokenType   TokenType = "BitwiseAndAssign"
	BitwiseOrAssignTokenType    TokenType = "BitwiseOrAssign"
	BitwiseXorAssignTokenType   TokenType = "BitwiseXorAssign"
	BitwiseNotAssignTokenType   TokenType = "BitwiseNotAssign"
	BitwiseLeftShiftTokenType   TokenType = "BitwiseLeftShift"
	BitwiseRightShiftTokenType  TokenType = "BitwiseRightShift"

	IntTokenType      TokenType = "IntToken"
	DecimalTokenType  TokenType = "DecimalToken"
	BoolTokenType     TokenType = "BoolToken"
	StringTokenType   TokenType = "StringToken"
	DatetimeTokenType TokenType = "DatetimeToken"
	NilTokenType      TokenType = "NilToken"

	WhitespaceTokenType TokenType = "Whitespace"
	CommentTokenType    TokenType = "Comment"
	EOFTokenType        TokenType = "EOF"
)

var keywords = map[string]TokenType{
	"if":     IfKeywordTokenType,
	"else":   ElseKeywordTokenType,
	"for":    ForKeywordTokenType,
	"while":  WhileKeywordTokenType,
	"break":  BreakKeywordTokenType,
	"switch": SwitchKeywordTokenType,
	"case":   CaseKeywordTokenType,
	"return": ReturnKeywordTokenType,
	"func":   FuncKeywordTokenType,
	"super":  SuperKeywordTokenType,
	"this":   ThisKeywordTokenType,
	"let":    LetKeywordTokenType,
	"class":  ClassKeywordTokenType,
	"print":  PrintKeywordTokenType,
}

type Token struct {
	Type        TokenType
	Lexeme      string
	Value       any
	Line        int
	Column      int
	IsGenerated bool
}
