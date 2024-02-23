package lang

import "fmt"

type Parser struct {
	tokens         []Token
	nextTokenIndex int
	errs           []Err
}

func (p *Parser) Parse(tokens []Token) ([]Statement, []Err) {
	p.tokens = tokens
	p.nextTokenIndex = 0
	statements := make([]Statement, 0)
	for p.nextTokenIndex < len(tokens) &&
		tokens[p.nextTokenIndex].Type != EOFTokenType {
		statement := p.scanDeclaration()
		if statement == nil {
			continue
		}

		statements = append(statements, *statement)
	}

	return statements, p.errs
}

func (p *Parser) scanDeclaration() *Statement {
	if p.matchTokenType([]TokenType{LetKeywordTokenType}) {
		letToken := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		statement, err := p.scanLetDeclaration(letToken.Line, letToken.Column)
		if err != nil {
			p.synchronize()
			return nil
		}

		return &statement
	}

	statement, err := p.scanStatement()
	if err != nil {
		p.synchronize()
		return nil
	}

	return &statement
}

func (p *Parser) scanLetDeclaration(line int, column int) (Statement, *Err) {
	if p.matchTokenType([]TokenType{IdentifierTokenType}) {
		identifierToken := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++

		var initializer *Expression
		if p.matchTokenType([]TokenType{AssignTokenType}) {
			p.nextTokenIndex++
			tmpInitializer, err := p.scanExpressionList()
			if err != nil {
				return Statement{}, err
			}

			initializer = &tmpInitializer
		}

		if !p.matchTokenType([]TokenType{SemicolonTokenType}) {
			token := p.tokens[p.nextTokenIndex]
			return Statement{}, &Err{
				Message: fmt.Sprintf("expect ';', got %v", token.Lexeme),
				Line:    token.Line,
				Column:  token.Column,
			}
		}

		p.nextTokenIndex++

		return Statement{
			Type:                     LetStatementType,
			LetIdentifier:            &identifierToken,
			LetInitializerExpression: initializer,
			Line:                     line,
			Column:                   column,
		}, nil
	}

	return Statement{}, &Err{
		Message: "expect identifier",
		Line:    line,
		Column:  column,
	}
}

func (p *Parser) scanStatement() (Statement, *Err) {
	if p.matchTokenType([]TokenType{PrintKeywordTokenType}) {
		token := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		return p.scanPrintStatement(token.Line, token.Column)
	}

	if p.matchTokenType([]TokenType{LeftBraceTokenType}) {
		token := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		return p.scanBlockStatement(token.Line, token.Column)
	}

	return p.scanExpressionStatement()
}

func (p *Parser) scanBlockStatement(line int, column int) (Statement, *Err) {
	var statements []Statement
	for p.nextTokenIndex < len(p.tokens) &&
		!p.matchTokenType([]TokenType{RightBraceTokenType, EOFTokenType}) {
		statement := p.scanDeclaration()
		if statement == nil {
			continue
		}

		statements = append(statements, *statement)
	}

	if p.nextTokenIndex >= len(p.tokens) {
		lastToken := p.tokens[len(p.tokens)-1]
		return Statement{}, &Err{
			Message: "'}' not found",
			Line:    lastToken.Line,
			Column:  lastToken.Column,
		}
	}

	if !p.matchTokenType([]TokenType{RightBraceTokenType}) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message: fmt.Sprintf("expect '}', got '%v'", token.Lexeme),
			Line:    token.Line,
			Column:  token.Column,
		}
	}

	p.nextTokenIndex++
	return Statement{
		Type:                 BlockStatementType,
		BlockInnerStatements: statements,
		Line:                 line,
		Column:               column,
	}, nil
}

func (p *Parser) scanPrintStatement(line int, column int) (Statement, *Err) {
	if !p.matchTokenType([]TokenType{LeftParenthesisTokenType}) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message: fmt.Sprintf("expect '(', but got %v", token.Lexeme),
			Line:    token.Line,
			Column:  token.Column,
		}
	}

	p.nextTokenIndex++
	expr, err := p.scanExpressionList()
	if err != nil {
		return Statement{}, err
	}

	if !p.matchTokenType([]TokenType{RightParenthesisTokenType}) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message: fmt.Sprintf("expect ')', but got %v", token.Lexeme),
			Line:    token.Line,
			Column:  token.Column,
		}
	}

	p.nextTokenIndex++

	if !p.matchTokenType([]TokenType{SemicolonTokenType}) {
		token := p.tokens[p.nextTokenIndex]
		err = &Err{
			Message: fmt.Sprintf("expect ';', but got %v", token.Lexeme),
			Line:    token.Line,
			Column:  token.Column,
		}
		p.errs = append(p.errs, *err)
		return Statement{}, err
	}

	p.nextTokenIndex++

	return Statement{
		Type:               PrintStatementType,
		PrintArgExpression: &expr,
		Line:               line,
		Column:             column,
	}, nil
}

func (p *Parser) scanExpressionStatement() (Statement, *Err) {
	expr, err := p.scanExpressionList()
	if err != nil {
		return Statement{}, err
	}

	if !p.matchTokenType([]TokenType{SemicolonTokenType}) {
		token := p.tokens[p.nextTokenIndex]
		err = &Err{
			Message: fmt.Sprintf("expect ';', but got %v", token.Lexeme),
			Line:    token.Line,
			Column:  token.Column,
		}
		p.errs = append(p.errs, *err)
		return Statement{}, err
	}

	p.nextTokenIndex++
	return Statement{
		Type:                ExpressionStatementType,
		StatementExpression: &expr,
		Line:                expr.Line,
		Column:              expr.Column,
	}, nil
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
	return p.scanAssignment()
}

func (p *Parser) scanAssignment() (Expression, *Err) {
	expr, err := p.scanConditional()
	if err != nil {
		return Expression{}, err
	}

	if p.matchTokenType([]TokenType{AssignTokenType}) {
		assignToken := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++

		if expr.Type == IdentifierExpressionType {
			valueExpr, err := p.scanAssignment()
			if err != nil {
				return Expression{}, err
			}

			return Expression{
				Type:                      AssignmentExpressionType,
				Line:                      expr.Line,
				Column:                    expr.Column,
				Identifier:                expr.Identifier,
				AssignmentValueExpression: &valueExpr,
			}, nil
		}

		return Expression{}, &Err{
			Message: fmt.Sprintf("assignment target must be identifier"),
			Line:    assignToken.Line,
			Column:  assignToken.Column,
		}
	}

	return expr, nil
}

func (p *Parser) scanConditional() (Expression, *Err) {
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
		falseExpression, err := p.scanConditional()
		if err != nil {
			return Expression{}, err
		}

		return Expression{
			Type:                       TernaryExpressionType,
			TernaryConditionExpression: &tempConditionExpr,
			TernaryTrueExpression:      &trueExpression,
			TernaryFalseExpression:     &falseExpression,
			Line:                       conditionExpr.Line,
			Column:                     conditionExpr.Column,
		}, nil
	}

	return conditionExpr, nil
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

	if p.matchTokenType([]TokenType{IdentifierTokenType}) {
		identifierToken := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		return Expression{
			Type:       IdentifierExpressionType,
			Identifier: identifierToken,
			Line:       identifierToken.Line,
			Column:     identifierToken.Column,
		}, nil
	}

	token := p.tokens[p.nextTokenIndex]
	p.nextTokenIndex++
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

func NewParser() *Parser {
	return &Parser{}
}
