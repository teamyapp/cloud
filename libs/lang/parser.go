package lang

import "fmt"

/*
Grammar:
	expressions -> expression ("," expression)*
	expression -> conditional
	conditional -> equality ("?" equality ":" equality)*
	equality -> logical (("=="|"!=") logical)*
	logicalOr -> logicalAnd ("||" logicalAnd)*
	logicalAnd -> comparison ("&&" comparison)*
	comparison -> bitwiseOr ((">"|"<"|">="|"<=") bitwiseOr)*
	bitwiseOr -> bitwiseXor ("|" bitwiseXor)*
	bitwiseXor -> bitwiseAnd ("^" bitwiseAnd)*
	bitwiseAnd -> term ("&" term)*
	shift -> term (("<<"|">>") term)*
	term -> factor (("+"|"-") factor)*
	factor -> unary (("*"|"/") unary)*
	unary -> ("!"|"-"|"~") unary | primary
	primary -> NUMBER | STRING | "true" | "false" | "nil" | "(" expression ")"
*/

type Parser struct {
	tokens         []Token
	nextTokenIndex int
	errs           []Err
}

func (p *Parser) parse(tokens []Token) (Expression, *Err) {
	p.tokens = tokens
	p.nextTokenIndex = 0
	return p.scanExpressionList()
}

func (p *Parser) scanExpressionList() (Expression, *Err) {
	expr, err := p.scanExpression()
	if err != nil {
		return Expression{}, err
	}

	expressionList := []Expression{expr}
	var foundComma bool
	for p.matchTokenType([]TokenType{CommaTokenType}) {
		foundComma = true
		p.nextTokenIndex++

		nextExpr, err := p.scanExpression()
		if err != nil {
			return Expression{}, err
		}

		expressionList = append(expressionList, nextExpr)
	}

	if foundComma {
		return Expression{
			Type:           ExpressionListExpressionType,
			ExpressionList: expressionList,
			Line:           expr.Line,
			Column:         expr.Column,
		}, nil
	}

	return expr, nil
}

func (p *Parser) scanExpression() (Expression, *Err) {
	return p.scanConditional()
}

func (p *Parser) scanConditional() (Expression, *Err) {
	ternaryExprDummyHead := Expression{}
	ternaryExpr := &ternaryExprDummyHead
	conditionExpr, err := p.scanEquality()
	if err != nil {
		return Expression{}, err
	}

	for p.matchTokenType([]TokenType{QuestionMarkTokenType}) {
		p.nextTokenIndex++
		trueExpression, err := p.scanEquality()
		if err != nil {
			return Expression{}, err
		}

		if !p.matchTokenType([]TokenType{ColonTokenType}) {
			token := p.tokens[p.nextTokenIndex]
			colonErr := Err{
				Message: fmt.Sprintf("expect ':' but got %s", token.Lexeme),
				Line:    token.Line,
				Column:  token.Column,
			}
			p.errs = append(p.errs, colonErr)
			return Expression{}, err
		}
		p.nextTokenIndex++

		tempConditionExpr := conditionExpr
		nextTernaryExpr := &Expression{
			Type:                       TernaryExpressionType,
			TernaryConditionExpression: &tempConditionExpr,
			TernaryTrueExpression:      &trueExpression,
			Line:                       conditionExpr.Line,
			Column:                     conditionExpr.Column,
		}
		ternaryExpr.TernaryFalseExpression = nextTernaryExpr
		ternaryExpr = nextTernaryExpr

		falseExpression, err := p.scanEquality()
		if err != nil {
			return Expression{}, err
		}

		conditionExpr = falseExpression
	}

	ternaryExpr.TernaryFalseExpression = &conditionExpr
	return *ternaryExprDummyHead.TernaryFalseExpression, nil
}

func (p *Parser) scanEquality() (Expression, *Err) {
	return p.scanBinaryExpression(p.scanLogicalOr, []TokenType{LogicalEqualTokenType, LogicalNotEqualTokenType})
}

func (p *Parser) scanLogicalOr() (Expression, *Err) {
	return p.scanBinaryExpression(p.scanLogicalAnd, []TokenType{LogicalOrTokenType})
}

func (p *Parser) scanLogicalAnd() (Expression, *Err) {
	return p.scanBinaryExpression(p.scanComparison, []TokenType{LogicalAndTokenType})
}

func (p *Parser) scanComparison() (Expression, *Err) {
	return p.scanBinaryExpression(p.scanBitwiseOr, []TokenType{
		GreaterThanTokenType,
		GreaterThanOrEqualTokenType,
		LessThanTokenType,
		LessThanOrEqualTokenType,
	})
}

func (p *Parser) scanBitwiseOr() (Expression, *Err) {
	return p.scanBinaryExpression(p.scanBitwiseXor, []TokenType{BitwiseOrTokenType})
}

func (p *Parser) scanBitwiseXor() (Expression, *Err) {
	return p.scanBinaryExpression(p.scanBitwiseAnd, []TokenType{BitwiseXorTokenType})
}

func (p *Parser) scanBitwiseAnd() (Expression, *Err) {
	return p.scanBinaryExpression(p.scanShift, []TokenType{BitwiseAndTokenType})
}

func (p *Parser) scanShift() (Expression, *Err) {
	return p.scanBinaryExpression(p.scanTerm, []TokenType{BitwiseLeftShiftTokenType, BitwiseRightShiftTokenType})
}

func (p *Parser) scanTerm() (Expression, *Err) {
	return p.scanBinaryExpression(p.scanFactor, []TokenType{AddTokenType, MinusTokenType})
}

func (p *Parser) scanFactor() (Expression, *Err) {
	return p.scanBinaryExpression(p.scanUnary, []TokenType{StarTokenType, DivideTokenType, ModuloTokenType})
}

func (p *Parser) scanUnary() (Expression, *Err) {
	if p.matchTokenType([]TokenType{
		LogicalNotTokenType,
		MinusTokenType,
		BitwiseNotTokenType,
	}) {
		operator := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++

		unaryExpr, err := p.scanUnary()
		if err != nil {
			return Expression{}, err
		}

		return Expression{
			Type:            UnaryExpressionType,
			Operator:        operator,
			UnaryExpression: &unaryExpr,
			Line:            operator.Line,
			Column:          operator.Column,
		}, nil
	}

	return p.scanPrimary()
}

func (p *Parser) scanPrimary() (Expression, *Err) {
	if p.matchTokenType([]TokenType{
		BoolTokenType,
		NilTokenType,
		StringTokenType,
		IntTokenType,
		DecimalTokenType,
	}) {
		token := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		return Expression{
			Type:    LiteralExpressionType,
			Literal: token,
			Line:    token.Line,
			Column:  token.Column,
		}, nil
	}

	if p.matchTokenType([]TokenType{LeftParenthesisTokenType}) {
		token := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		innerExpr, err := p.scanExpression()
		if err != nil {
			return Expression{}, err
		}

		if !p.matchTokenType([]TokenType{RightParenthesisTokenType}) {
			token := p.tokens[p.nextTokenIndex]
			p.errs = append(p.errs, Err{
				Message: fmt.Sprintf("expect ')' but got %s", token.Lexeme),
				Line:    token.Line,
				Column:  token.Column,
			})
			return Expression{
				Type:                 GroupingExpressionType,
				GroupInnerExpression: &innerExpr,
			}, nil
		}

		p.nextTokenIndex++
		return Expression{
			Type:                 GroupingExpressionType,
			GroupInnerExpression: &innerExpr,
			Line:                 token.Line,
			Column:               token.Column,
		}, nil
	}

	token := p.tokens[p.nextTokenIndex]
	err := &Err{
		Message: "expect expression",
		Line:    token.Line,
		Column:  token.Column,
	}
	p.errs = append(p.errs, *err)
	return Expression{}, err
}

func (p *Parser) synchronize() {
	p.nextTokenIndex++

	for p.nextTokenIndex < len(p.tokens) {
		if p.tokens[p.nextTokenIndex-1].Type == SemicolonTokenType {
			return
		}

		switch p.tokens[p.nextTokenIndex].Type {
		case ClassKeywordTokenType,
			FuncKeywordTokenType,
			LetKeywordTokenType,
			ForKeywordTokenType,
			IfKeywordTokenType,
			WhileKeywordTokenType,
			ReturnKeywordTokenType:
			return
		}

		p.nextTokenIndex++
	}
}

func (p *Parser) scanBinaryExpression(
	scanOperand func() (Expression, *Err),
	operatorTokenTypes []TokenType,
) (Expression, *Err) {
	leftExpr, err := scanOperand()
	if err != nil {
		return Expression{}, err
	}

	for p.matchTokenType(operatorTokenTypes) {
		operator := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++

		rightExpr, err := scanOperand()
		if err != nil {
			return Expression{}, err
		}

		tempExpr := leftExpr
		leftExpr = Expression{
			Type:                  BinaryExpressionType,
			Operator:              operator,
			BinaryLeftExpression:  &tempExpr,
			BinaryRightExpression: &rightExpr,
			Line:                  tempExpr.Line,
			Column:                tempExpr.Column,
		}
	}

	return leftExpr, nil
}

func (p *Parser) matchTokenType(expectedTokenTypes []TokenType) bool {
	for _, expectedTokenType := range expectedTokenTypes {
		currToken := p.tokens[p.nextTokenIndex]
		if currToken.Type == expectedTokenType {
			return true
		}
	}

	return false
}

func (p *Parser) GetErrors() []Err {
	return p.errs
}

func NewParser() *Parser {
	return &Parser{}
}
