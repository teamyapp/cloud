package lang

import "fmt"

type ParametersAndBody struct {
	Parameters []Token
	Body       Statement
}

type ClassBody struct {
	InstanceMethods []Statement
	InstanceGetters []Statement
	InstanceSetters []Statement
	StaticMethods   []Statement
	StaticGetters   []Statement
	StaticSetters   []Statement
}

type Parser struct {
	nextNodeID     uint64
	tokens         []Token
	nextTokenIndex int
	errs           []Err
}

func (p *Parser) Parse(tokens []Token) ([]Statement, []Err) {
	p.nextNodeID = 1
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
	if p.matchTokenType([]TokenType{LetKeywordTokenType}, 0) {
		letToken := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		statement, err := p.scanLetDeclaration(letToken)
		if err != nil {
			p.synchronize()
			return nil
		}

		return &statement
	}

	if p.matchTokenType([]TokenType{FuncKeywordTokenType}, 0) {
		if p.matchTokenType([]TokenType{IdentifierTokenType}, 1) {
			funcToken := p.tokens[p.nextTokenIndex]
			p.nextTokenIndex++
			statement, err := p.scanCallableDeclaration(funcToken)
			if err != nil {
				p.synchronize()
				return nil
			}

			return &statement
		}
	}

	if p.matchTokenType([]TokenType{ClassKeywordTokenType}, 0) {
		classToken := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		statement, err := p.scanClassDeclaration(classToken)
		if err != nil {
			p.synchronize()
			return nil
		}

		return statement
	}

	statement, err := p.scanStatement()
	if err != nil {
		p.synchronize()
		return nil
	}

	return &statement
}

func (p *Parser) scanClassDeclaration(startToken Token) (*Statement, *Err) {
	if !p.matchTokenType([]TokenType{IdentifierTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return nil, &Err{
			Message:           fmt.Sprintf("expect identifier, got %v", token.Lexeme),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
	}

	classIdentifier := p.tokens[p.nextTokenIndex]
	p.nextTokenIndex++

	var superClassExpression *Expression
	if p.matchTokenType([]TokenType{ColonTokenType}, 0) {
		p.nextTokenIndex++

		if !p.matchTokenType([]TokenType{IdentifierTokenType}, 0) {
			token := p.tokens[p.nextTokenIndex]
			return nil, &Err{
				Message:           fmt.Sprintf("expect identifier, got %v", token.Lexeme),
				Line:              token.Line,
				Column:            token.Column,
				FromGeneratedCode: token.IsGenerated,
			}
		}

		superClassIdentifier := p.tokens[p.nextTokenIndex]
		superClassExpression = &Expression{
			NodeID:      p.newNodeID(),
			Type:        IdentifierExpressionType,
			Identifier:  superClassIdentifier,
			Line:        superClassIdentifier.Line,
			Column:      superClassIdentifier.Column,
			IsGenerated: superClassIdentifier.IsGenerated,
		}
		p.nextTokenIndex++
	}

	if !p.matchTokenType([]TokenType{LeftBraceTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return nil, &Err{
			Message:           fmt.Sprintf("expect '{', got %v", token.Lexeme),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
	}

	p.nextTokenIndex++

	classBody, err := p.scanClassBodyDeclaration()
	if err != nil {
		return nil, err
	}

	if !p.matchTokenType([]TokenType{RightBraceTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return nil, &Err{
			Message:           fmt.Sprintf("expect '}', got %v", token.Lexeme),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
	}

	p.nextTokenIndex++
	return &Statement{
		NodeID:                          p.newNodeID(),
		Type:                            ClassStatementType,
		ClassIdentifier:                 &classIdentifier,
		ClassSuperClassExpression:       superClassExpression,
		ClassInstanceMethodDeclarations: classBody.InstanceMethods,
		ClassStaticMethodDeclarations:   classBody.StaticMethods,
		ClassInstanceGetterDeclarations: classBody.InstanceGetters,
		ClassInstanceSetterDeclarations: classBody.InstanceSetters,
		ClassStaticGetterDeclarations:   classBody.StaticGetters,
		ClassStaticSetterDeclarations:   classBody.StaticSetters,
		Line:                            startToken.Line,
		Column:                          startToken.Column,
		IsGenerated:                     startToken.IsGenerated,
	}, nil
}

func (p *Parser) scanClassBodyDeclaration() (ClassBody, *Err) {
	var instanceMethods []Statement
	var staticMethods []Statement
	var instanceGetters []Statement
	var instanceSetters []Statement
	var staticGetters []Statement
	var staticSetters []Statement

	for p.nextTokenIndex < len(p.tokens) &&
		!p.matchTokenType([]TokenType{RightBraceTokenType}, 0) {
		startToken := p.tokens[p.nextTokenIndex]

		foundStatic := p.matchTokenType([]TokenType{StaticKeywordTokenType}, 0)
		if foundStatic {
			p.nextTokenIndex++
			startToken = p.tokens[p.nextTokenIndex]
		}

		if p.matchTokenType([]TokenType{GetKeywordTokenType}, 0) {
			p.nextTokenIndex++
			getterDeclaration, err := p.scanGetterDeclaration()
			if err != nil {
				return ClassBody{}, err
			}

			if foundStatic {
				staticGetters = append(staticGetters, getterDeclaration)
			} else {
				instanceGetters = append(instanceGetters, getterDeclaration)
			}

			continue
		} else if p.matchTokenType([]TokenType{SetKeywordTokenType}, 0) {
			p.nextTokenIndex++
			setterDeclaration, err := p.scanSetterDeclaration()
			if err != nil {
				return ClassBody{}, err
			}

			if foundStatic {
				staticSetters = append(staticSetters, setterDeclaration)
			} else {
				instanceSetters = append(instanceSetters, setterDeclaration)
			}

			continue
		}

		methodDeclaration, err := p.scanCallableDeclaration(startToken)
		if err != nil {
			return ClassBody{}, err
		}

		if foundStatic {
			staticMethods = append(staticMethods, methodDeclaration)
		} else {
			instanceMethods = append(instanceMethods, methodDeclaration)
		}
	}

	return ClassBody{
		InstanceMethods: instanceMethods,
		InstanceGetters: instanceGetters,
		InstanceSetters: instanceSetters,
		StaticMethods:   staticMethods,
		StaticGetters:   staticGetters,
		StaticSetters:   staticSetters,
	}, nil
}

func (p *Parser) scanGetterDeclaration() (Statement, *Err) {
	if !p.matchTokenType([]TokenType{IdentifierTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message:           fmt.Sprintf("expect identifier, got %v", token.Lexeme),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
	}

	identifier := p.tokens[p.nextTokenIndex]
	p.nextTokenIndex++

	if !p.matchTokenType([]TokenType{LeftBraceTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message:           fmt.Sprintf("expect '{', got %v", token.Lexeme),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
	}

	blockStartToken := p.tokens[p.nextTokenIndex]
	p.nextTokenIndex++
	body, err := p.scanBlockStatement(blockStartToken)
	if err != nil {
		return Statement{}, err
	}

	return Statement{
		NodeID:             p.newNodeID(),
		Type:               CallableStatementType,
		CallableName:       &identifier,
		CallableParameters: []Token{},
		CallableBody:       &body,
	}, nil
}

func (p *Parser) scanSetterDeclaration() (Statement, *Err) {
	if !p.matchTokenType([]TokenType{IdentifierTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message:           fmt.Sprintf("expect identifier, got %v", token.Lexeme),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
	}

	identifier := p.tokens[p.nextTokenIndex]
	p.nextTokenIndex++

	if !p.matchTokenType([]TokenType{LeftParenthesisTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message:           fmt.Sprintf("expect '(', got %v", token.Lexeme),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
	}

	p.nextTokenIndex++

	if !p.matchTokenType([]TokenType{IdentifierTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message:           fmt.Sprintf("expect identifier, got %v", token.Lexeme),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
	}

	newValueParameter := p.tokens[p.nextTokenIndex]
	p.nextTokenIndex++

	if !p.matchTokenType([]TokenType{RightParenthesisTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message:           fmt.Sprintf("expect ')', got %v", token.Lexeme),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
	}

	p.nextTokenIndex++

	if !p.matchTokenType([]TokenType{LeftBraceTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message:           fmt.Sprintf("expect '{', got %v", token.Lexeme),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
	}

	blockStartToken := p.tokens[p.nextTokenIndex]
	p.nextTokenIndex++
	body, err := p.scanBlockStatement(blockStartToken)
	if err != nil {
		return Statement{}, err
	}

	return Statement{
		NodeID:             p.newNodeID(),
		Type:               CallableStatementType,
		CallableName:       &identifier,
		CallableParameters: []Token{newValueParameter},
		CallableBody:       &body,
	}, nil
}

func (p *Parser) scanCallableDeclaration(startToken Token) (Statement, *Err) {
	identifier := p.tokens[p.nextTokenIndex]
	p.nextTokenIndex++

	parametersAndBody, err := p.scanParametersAndBody()
	if err != nil {
		return Statement{}, err
	}

	return Statement{
		NodeID:             p.newNodeID(),
		Type:               CallableStatementType,
		CallableName:       &identifier,
		CallableParameters: parametersAndBody.Parameters,
		CallableBody:       &parametersAndBody.Body,
		Line:               startToken.Line,
		Column:             startToken.Column,
	}, nil
}

func (p *Parser) scanParametersAndBody() (ParametersAndBody, *Err) {
	if !p.matchTokenType([]TokenType{LeftParenthesisTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return ParametersAndBody{}, &Err{
			Message:           fmt.Sprintf("expect '(', got %v", token.Lexeme),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
	}

	p.nextTokenIndex++

	parameters, err := p.scanParameters()
	if err != nil {
		return ParametersAndBody{}, err
	}

	if !p.matchTokenType([]TokenType{RightParenthesisTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return ParametersAndBody{}, &Err{
			Message:           fmt.Sprintf("expect ')' but got %s", token.Lexeme),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: false,
		}
	}

	p.nextTokenIndex++

	if !p.matchTokenType([]TokenType{LeftBraceTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return ParametersAndBody{}, &Err{
			Message:           fmt.Sprintf("expect '{' but got %s", token.Lexeme),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: false,
		}
	}

	token := p.tokens[p.nextTokenIndex]
	p.nextTokenIndex++
	body, err := p.scanBlockStatement(token)
	if err != nil {
		return ParametersAndBody{}, err
	}

	return ParametersAndBody{
		Parameters: parameters,
		Body:       body,
	}, nil
}

func (p *Parser) scanParameters() ([]Token, *Err) {
	var parameters []Token
	if !p.matchTokenType([]TokenType{RightParenthesisTokenType}, 0) {
		for {
			token := p.tokens[p.nextTokenIndex]
			p.nextTokenIndex++
			if len(parameters) >= 255 {
				err := &Err{
					Message: fmt.Sprintf("cannot have more than 255 parameters"),
					Line:    token.Line,
					Column:  token.Column,
				}
				p.errs = append(p.errs, *err)
			}

			parameters = append(parameters, token)
			if !p.matchTokenType([]TokenType{CommaTokenType}, 0) {
				break
			}

			p.nextTokenIndex++
		}
	}

	return parameters, nil
}

func (p *Parser) scanLetDeclaration(startToken Token) (Statement, *Err) {
	if p.matchTokenType([]TokenType{IdentifierTokenType}, 0) {
		identifierToken := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++

		var initializer *Expression
		if p.matchTokenType([]TokenType{AssignTokenType}, 0) {
			p.nextTokenIndex++
			tmpInitializer, err := p.scanExpressionList()
			if err != nil {
				return Statement{}, err
			}

			initializer = &tmpInitializer
		}

		if !p.matchTokenType([]TokenType{SemicolonTokenType}, 0) {
			token := p.tokens[p.nextTokenIndex]
			return Statement{}, &Err{
				Message:           fmt.Sprintf("expect ';', got %v", token.Lexeme),
				Line:              token.Line,
				Column:            token.Column,
				FromGeneratedCode: token.IsGenerated,
			}
		}

		p.nextTokenIndex++

		return Statement{
			NodeID:                   p.newNodeID(),
			Type:                     LetStatementType,
			LetIdentifier:            &identifierToken,
			LetInitializerExpression: initializer,
			Line:                     startToken.Line,
			Column:                   startToken.Column,
		}, nil
	}

	return Statement{}, &Err{
		Message:           "expect identifier",
		Line:              startToken.Line,
		Column:            startToken.Column,
		FromGeneratedCode: startToken.IsGenerated,
	}
}

func (p *Parser) scanStatement() (Statement, *Err) {
	if p.matchTokenType([]TokenType{LeftBraceTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		return p.scanBlockStatement(token)
	}

	if p.matchTokenType([]TokenType{IfKeywordTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		return p.scanIfStatement(token)
	}

	if p.matchTokenType([]TokenType{WhileKeywordTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		return p.scanWhileStatement(token)
	}

	if p.matchTokenType([]TokenType{ForKeywordTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		return p.scanForStatement(token)
	}

	if p.matchTokenType([]TokenType{BreakKeywordTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		return p.scanBreakStatement(token)
	}

	if p.matchTokenType([]TokenType{ContinueKeywordTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		return p.scanContinueStatement(token)
	}

	if p.matchTokenType([]TokenType{ReturnKeywordTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		return p.scanReturnStatement(token)
	}

	return p.scanExpressionStatement()
}

func (p *Parser) scanReturnStatement(startToken Token) (Statement, *Err) {
	var expr *Expression
	if !p.matchTokenType([]TokenType{SemicolonTokenType}, 0) {
		tmpExpr, err := p.scanExpressionList()
		if err != nil {
			return Statement{}, err
		}

		expr = &tmpExpr
	}

	if !p.matchTokenType([]TokenType{SemicolonTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message:           fmt.Sprintf("expect ';', found '%v'", token),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
	}

	p.nextTokenIndex++
	return Statement{
		NodeID:                p.newNodeID(),
		Type:                  ReturnStatementType,
		ReturnValueExpression: expr,
		Line:                  startToken.Line,
		Column:                startToken.Column,
	}, nil
}

func (p *Parser) scanContinueStatement(startToken Token) (Statement, *Err) {
	if !p.matchTokenType([]TokenType{SemicolonTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message: fmt.Sprintf("expect ';', found '%v'", token),
			Line:    token.Line,
			Column:  token.Column,
		}
	}

	p.nextTokenIndex++

	return Statement{
		NodeID: p.newNodeID(),
		Type:   ContinueStatementType,
		Line:   startToken.Line,
		Column: startToken.Column,
	}, nil
}

func (p *Parser) scanBreakStatement(startToken Token) (Statement, *Err) {
	if !p.matchTokenType([]TokenType{SemicolonTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message:           fmt.Sprintf("expect ';', found '%v'", token),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
	}

	p.nextTokenIndex++

	return Statement{
		NodeID: p.newNodeID(),
		Type:   BreakStatementType,
		Line:   startToken.Line,
		Column: startToken.Column,
	}, nil
}

func (p *Parser) scanForStatement(startToken Token) (Statement, *Err) {
	if !p.matchTokenType([]TokenType{LeftParenthesisTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message:           fmt.Sprintf("expect '(', found '%v'", token),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
	}

	p.nextTokenIndex++

	var initializer *Statement
	if p.matchTokenType([]TokenType{SemicolonTokenType}, 0) {
		p.nextTokenIndex++
	} else if p.matchTokenType([]TokenType{LetKeywordTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		tmpInitializer, err := p.scanLetDeclaration(token)
		if err != nil {
			return Statement{}, err
		}

		initializer = &tmpInitializer
	} else {
		tmpInitializer, err := p.scanExpressionStatement()
		if err != nil {
			return Statement{}, err
		}

		initializer = &tmpInitializer
	}

	var condition *Expression
	if !p.matchTokenType([]TokenType{SemicolonTokenType}, 0) {
		tmpCondition, err := p.scanExpressionList()
		if err != nil {
			return Statement{}, err
		}

		condition = &tmpCondition
	}

	if !p.matchTokenType([]TokenType{SemicolonTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message: fmt.Sprintf("expect ';', but got '%v'", token.Lexeme),
			Line:    token.Line,
			Column:  token.Column,
		}
	}

	p.nextTokenIndex++

	var increment *Expression
	if !p.matchTokenType([]TokenType{RightParenthesisTokenType}, 0) {
		tmpIncrement, err := p.scanExpressionList()
		if err != nil {
			return Statement{}, err
		}

		increment = &tmpIncrement
	}

	if !p.matchTokenType([]TokenType{RightParenthesisTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message: fmt.Sprintf("expect ')', but got '%v'", token.Lexeme),
			Line:    token.Line,
			Column:  token.Column,
		}
	}

	p.nextTokenIndex++

	body, err := p.scanStatement()
	if err != nil {
		return Statement{}, err
	}

	if increment != nil {
		body = Statement{
			NodeID: p.newNodeID(),
			Type:   BlockStatementType,
			BlockInnerStatements: []Statement{
				body,
				{
					NodeID:              p.newNodeID(),
					Type:                ExpressionStatementType,
					StatementExpression: increment,
					IsGenerated:         true,
				},
			},
			IsGenerated: true,
		}
	}

	if condition == nil {
		condition = &Expression{
			NodeID: p.newNodeID(),
			Type:   LiteralExpressionType,
			Literal: Token{
				Type:        BoolTokenType,
				Lexeme:      "true",
				Value:       true,
				IsGenerated: true,
			},
			IsGenerated: true,
		}
	}

	statement := Statement{
		NodeID:                   p.newNodeID(),
		Type:                     WhileStatementType,
		WhileConditionExpression: condition,
		WhileBodyStatement:       &body,
		Line:                     startToken.Line,
		Column:                   startToken.Column,
	}

	if initializer != nil {
		statement = Statement{
			NodeID: p.newNodeID(),
			Type:   BlockStatementType,
			BlockInnerStatements: []Statement{
				*initializer,
				statement,
			},
			IsGenerated: true,
		}
	}

	return statement, nil
}

func (p *Parser) scanWhileStatement(startToken Token) (Statement, *Err) {
	if !p.matchTokenType([]TokenType{LeftParenthesisTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message: fmt.Sprintf("expect '(', found '%v'", token),
			Line:    token.Line,
			Column:  token.Column,
		}
	}

	p.nextTokenIndex++

	conditionExpr, err := p.scanExpressionList()
	if err != nil {
		return Statement{}, err
	}

	if !p.matchTokenType([]TokenType{RightParenthesisTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message:           fmt.Sprintf("expect ')', found '%v'", token),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
	}

	p.nextTokenIndex++

	bodyStmt, err := p.scanStatement()
	if err != nil {
		return Statement{}, err
	}

	return Statement{
		NodeID:                   p.newNodeID(),
		Type:                     WhileStatementType,
		WhileConditionExpression: &conditionExpr,
		WhileBodyStatement:       &bodyStmt,
		Line:                     startToken.Line,
		Column:                   startToken.Column,
	}, nil
}

func (p *Parser) scanIfStatement(startToken Token) (Statement, *Err) {
	if !p.matchTokenType([]TokenType{LeftParenthesisTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message:           fmt.Sprintf("expect '(', found '%v'", token),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
	}

	p.nextTokenIndex++

	conditionExpr, err := p.scanExpressionList()
	if err != nil {
		return Statement{}, err
	}

	if !p.matchTokenType([]TokenType{RightParenthesisTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message:           fmt.Sprintf("expect ')', found '%v'", token),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
	}

	p.nextTokenIndex++

	trueBranchStmt, err := p.scanStatement()
	if err != nil {
		return Statement{}, err
	}

	var falseBranchStmt *Statement
	if p.matchTokenType([]TokenType{ElseKeywordTokenType}, 0) {
		p.nextTokenIndex++
		falseTmpBranchStmt, err := p.scanStatement()
		if err != nil {
			return Statement{}, err
		}

		falseBranchStmt = &falseTmpBranchStmt
	}

	return Statement{
		NodeID:                 p.newNodeID(),
		Type:                   IfStatementType,
		IfConditionExpression:  &conditionExpr,
		IfTrueBranchStatement:  &trueBranchStmt,
		IfFalseBranchStatement: falseBranchStmt,
		Line:                   startToken.Line,
		Column:                 startToken.Column,
	}, nil
}

func (p *Parser) scanBlockStatement(startToken Token) (Statement, *Err) {
	var statements []Statement
	for p.nextTokenIndex < len(p.tokens) &&
		!p.matchTokenType([]TokenType{RightBraceTokenType, EOFTokenType}, 0) {
		statement := p.scanDeclaration()
		if statement == nil {
			continue
		}

		statements = append(statements, *statement)
	}

	if p.nextTokenIndex >= len(p.tokens) {
		lastToken := p.tokens[len(p.tokens)-1]
		return Statement{}, &Err{
			Message:           "'}' not found",
			Line:              lastToken.Line,
			Column:            lastToken.Column,
			FromGeneratedCode: lastToken.IsGenerated,
		}
	}

	if !p.matchTokenType([]TokenType{RightBraceTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		return Statement{}, &Err{
			Message:           fmt.Sprintf("expect '}', got '%v'", token.Lexeme),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
	}

	p.nextTokenIndex++
	return Statement{
		NodeID:               p.newNodeID(),
		Type:                 BlockStatementType,
		BlockInnerStatements: statements,
		Line:                 startToken.Line,
		Column:               startToken.Column,
	}, nil
}

func (p *Parser) scanExpressionStatement() (Statement, *Err) {
	expr, err := p.scanExpressionList()
	if err != nil {
		return Statement{}, err
	}

	if !p.matchTokenType([]TokenType{SemicolonTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		err = &Err{
			Message:           fmt.Sprintf("expect ';', but got %v", token.Lexeme),
			Line:              token.Line,
			Column:            token.Column,
			FromGeneratedCode: token.IsGenerated,
		}
		p.errs = append(p.errs, *err)
		return Statement{}, err
	}

	p.nextTokenIndex++
	return Statement{
		NodeID:              p.newNodeID(),
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
	for p.matchTokenType([]TokenType{CommaTokenType}, 0) {
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
			NodeID:         p.newNodeID(),
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

	if p.matchTokenType([]TokenType{AssignTokenType}, 0) {
		assignToken := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++

		switch expr.Type {
		case IdentifierExpressionType:
			valueExpr, err := p.scanAssignment()
			if err != nil {
				return Expression{}, err
			}

			return Expression{
				NodeID:                    p.newNodeID(),
				Type:                      AssignmentExpressionType,
				Line:                      expr.Line,
				Column:                    expr.Column,
				AssignmentIdentifier:      expr.Identifier,
				AssignmentValueExpression: &valueExpr,
			}, nil
		case GetExpressionType:
			valueExpr, err := p.scanAssignment()
			if err != nil {
				return Expression{}, err
			}

			return Expression{
				NodeID:              p.newNodeID(),
				Type:                SetExpressionType,
				Line:                expr.Line,
				Column:              expr.Column,
				SetObjectExpression: expr.GetObjectExpression,
				SetFieldName:        expr.GetFieldName,
				SetValueExpression:  &valueExpr,
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
	conditionExpr, err := p.scanLogicalOr()
	if err != nil {
		return Expression{}, err
	}

	for p.matchTokenType([]TokenType{QuestionMarkTokenType}, 0) {
		p.nextTokenIndex++
		trueExpression, err := p.scanLogicalOr()
		if err != nil {
			return Expression{}, err
		}

		if !p.matchTokenType([]TokenType{ColonTokenType}, 0) {
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
			NodeID:                     p.newNodeID(),
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

func (p *Parser) scanLogicalOr() (Expression, *Err) {
	return p.scanBinaryExpression(p.scanLogicalAnd, []TokenType{LogicalOrTokenType})
}

func (p *Parser) scanLogicalAnd() (Expression, *Err) {
	return p.scanBinaryExpression(p.scanEquality, []TokenType{LogicalAndTokenType})
}

func (p *Parser) scanEquality() (Expression, *Err) {
	return p.scanBinaryExpression(p.scanComparison, []TokenType{LogicalEqualTokenType, LogicalNotEqualTokenType})
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
	}, 0) {
		operator := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++

		unaryExpr, err := p.scanUnary()
		if err != nil {
			return Expression{}, err
		}

		return Expression{
			NodeID:          p.newNodeID(),
			Type:            UnaryExpressionType,
			Operator:        operator,
			UnaryExpression: &unaryExpr,
			Line:            operator.Line,
			Column:          operator.Column,
		}, nil
	}

	return p.scanNewInstance()
}

func (p *Parser) scanNewInstance() (Expression, *Err) {
	if p.matchTokenType([]TokenType{NewKeywordTokenType}, 0) {
		p.nextTokenIndex++

		expr, err := p.scanPrimary()
		if err != nil {
			return Expression{}, err
		}

		if !p.matchTokenType([]TokenType{LeftParenthesisTokenType}, 0) {
			token := p.tokens[p.nextTokenIndex]
			return Expression{}, &Err{
				Message:           fmt.Sprintf("expect '(', got '%v'", token.Lexeme),
				Line:              token.Line,
				Column:            token.Column,
				FromGeneratedCode: token.IsGenerated,
			}
		}

		p.nextTokenIndex++

		args, err := p.scanArguments()
		if err != nil {
			return Expression{}, err
		}

		if !p.matchTokenType([]TokenType{RightParenthesisTokenType}, 0) {
			token := p.tokens[p.nextTokenIndex]
			return Expression{}, &Err{
				Message:           fmt.Sprintf("expect ')', got '%v'", token.Lexeme),
				Line:              token.Line,
				Column:            token.Column,
				FromGeneratedCode: token.IsGenerated,
			}
		}

		p.nextTokenIndex++
		return Expression{
			NodeID:                     p.newNodeID(),
			Type:                       NewInstanceExpressionType,
			NewInstanceClassExpression: &expr,
			NewInstanceConstructorArgumentExpressions: args,
		}, nil
	}

	return p.scanCallAndGet()
}

func (p *Parser) scanCallAndGet() (Expression, *Err) {
	expr, err := p.scanPrimary()
	if err != nil {
		return Expression{}, err
	}

	for {
		if p.matchTokenType([]TokenType{LeftParenthesisTokenType}, 0) {
			startToken := p.tokens[p.nextTokenIndex]
			p.nextTokenIndex++
			args, err := p.scanArguments()
			if err != nil {
				return Expression{}, err
			}

			if !p.matchTokenType([]TokenType{RightParenthesisTokenType}, 0) {
				token := p.tokens[p.nextTokenIndex]
				return Expression{}, &Err{
					Message:           fmt.Sprintf("expect ')' but got %s", token.Lexeme),
					Line:              token.Line,
					Column:            token.Column,
					FromGeneratedCode: false,
				}
			}

			p.nextTokenIndex++
			tmpExpr := expr
			expr = Expression{
				NodeID:                  p.newNodeID(),
				Type:                    CallExpressionType,
				Line:                    startToken.Line,
				Column:                  startToken.Column,
				CallableExpression:      &tmpExpr,
				CallArgumentExpressions: args,
			}
		} else if p.matchTokenType([]TokenType{DotTokenType}, 0) {
			startToken := p.tokens[p.nextTokenIndex]
			p.nextTokenIndex++
			if !p.matchTokenType([]TokenType{IdentifierTokenType}, 0) {
				token := p.tokens[p.nextTokenIndex]
				return Expression{}, &Err{
					Message: fmt.Sprintf("expect identifier, got %v", token.Lexeme),
					Line:    token.Line,
					Column:  token.Column,
				}
			}

			identifierToken := p.tokens[p.nextTokenIndex]
			p.nextTokenIndex++
			tmpExpr := expr
			expr = Expression{
				NodeID:              p.newNodeID(),
				Type:                GetExpressionType,
				Line:                startToken.Line,
				Column:              startToken.Column,
				GetObjectExpression: &tmpExpr,
				GetFieldName:        identifierToken,
			}
		} else {
			break
		}
	}

	return expr, nil
}

func (p *Parser) scanArguments() ([]Expression, *Err) {
	var args []Expression
	if !p.matchTokenType([]TokenType{RightParenthesisTokenType}, 0) {
		for {
			if len(args) >= 255 {
				token := p.tokens[p.nextTokenIndex]
				err := &Err{
					Message:           fmt.Sprintf("cannot have more than 255 arguments"),
					Line:              token.Line,
					Column:            token.Column,
					FromGeneratedCode: token.IsGenerated,
				}
				p.errs = append(p.errs, *err)
			}

			arg, err := p.scanExpression()
			if err != nil {
				return nil, err
			}

			args = append(args, arg)
			if !p.matchTokenType([]TokenType{CommaTokenType}, 0) {
				break
			}

			p.nextTokenIndex++
		}
	}

	return args, nil
}

func (p *Parser) scanPrimary() (Expression, *Err) {
	if p.matchTokenType([]TokenType{
		BoolTokenType,
		NilTokenType,
		StringTokenType,
		IntTokenType,
		DecimalTokenType,
	}, 0) {
		token := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		return Expression{
			NodeID:      p.newNodeID(),
			Type:        LiteralExpressionType,
			Literal:     token,
			Line:        token.Line,
			Column:      token.Column,
			IsGenerated: token.IsGenerated,
		}, nil
	}

	if p.matchTokenType([]TokenType{LeftParenthesisTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		innerExpr, err := p.scanExpression()
		if err != nil {
			return Expression{}, err
		}

		if !p.matchTokenType([]TokenType{RightParenthesisTokenType}, 0) {
			token = p.tokens[p.nextTokenIndex]
			p.errs = append(p.errs, Err{
				Message:           fmt.Sprintf("expect ')' but got %s", token.Lexeme),
				Line:              token.Line,
				Column:            token.Column,
				FromGeneratedCode: token.IsGenerated,
			})
			return Expression{
				Type:                 GroupingExpressionType,
				GroupInnerExpression: &innerExpr,
			}, nil
		}

		p.nextTokenIndex++
		return Expression{
			NodeID:               p.newNodeID(),
			Type:                 GroupingExpressionType,
			GroupInnerExpression: &innerExpr,
			Line:                 token.Line,
			Column:               token.Column,
			IsGenerated:          token.IsGenerated,
		}, nil
	}

	if p.matchTokenType([]TokenType{ThisKeywordTokenType}, 0) {
		token := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		return Expression{
			NodeID:      p.newNodeID(),
			Type:        ThisExpressionType,
			ThisKeyword: token,
			Line:        token.Line,
			Column:      token.Column,
			IsGenerated: token.IsGenerated,
		}, nil
	}

	if p.matchTokenType([]TokenType{SuperKeywordTokenType}, 0) {
		superToken := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++

		if !p.matchTokenType([]TokenType{DotTokenType}, 0) {
			dotToken := p.tokens[p.nextTokenIndex]
			return Expression{}, &Err{
				Message:           fmt.Sprintf("expect '.', got '%v'", dotToken.Lexeme),
				Line:              dotToken.Line,
				Column:            dotToken.Column,
				FromGeneratedCode: dotToken.IsGenerated,
			}
		}

		p.nextTokenIndex++

		if !p.matchTokenType([]TokenType{IdentifierTokenType}, 0) {
			identifierToken := p.tokens[p.nextTokenIndex]
			return Expression{}, &Err{
				Message:           fmt.Sprintf("expect identifier, got '%v'", identifierToken.Lexeme),
				Line:              identifierToken.Line,
				Column:            identifierToken.Column,
				FromGeneratedCode: identifierToken.IsGenerated,
			}
		}

		identifierToken := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		return Expression{
			NodeID:             p.newNodeID(),
			Type:               SuperExpressionType,
			SuperKeyword:       superToken,
			SuperRefIdentifier: identifierToken,
			Line:               superToken.Line,
			Column:             superToken.Column,
			IsGenerated:        superToken.IsGenerated,
		}, nil
	}

	if p.matchTokenType([]TokenType{IdentifierTokenType}, 0) {
		identifierToken := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		return Expression{
			NodeID:      p.newNodeID(),
			Type:        IdentifierExpressionType,
			Identifier:  identifierToken,
			Line:        identifierToken.Line,
			Column:      identifierToken.Column,
			IsGenerated: identifierToken.IsGenerated,
		}, nil
	}

	if p.matchTokenType([]TokenType{FuncKeywordTokenType}, 0) {
		funcToken := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++
		parametersAndBody, err := p.scanParametersAndBody()
		if err != nil {
			return Expression{}, err
		}

		return Expression{
			NodeID:           p.newNodeID(),
			Type:             LambdaExpressionType,
			LambdaParameters: parametersAndBody.Parameters,
			LambdaBody:       parametersAndBody.Body,
			Line:             funcToken.Line,
			Column:           funcToken.Column,
			IsGenerated:      funcToken.IsGenerated,
		}, nil
	}

	token := p.tokens[p.nextTokenIndex]
	p.nextTokenIndex++
	err := &Err{
		Message:           fmt.Sprintf("expect expression, got %v", token.Lexeme),
		Line:              token.Line,
		Column:            token.Column,
		FromGeneratedCode: token.IsGenerated,
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

	for p.matchTokenType(operatorTokenTypes, 0) {
		operator := p.tokens[p.nextTokenIndex]
		p.nextTokenIndex++

		rightExpr, err := scanOperand()
		if err != nil {
			return Expression{}, err
		}

		tempExpr := leftExpr
		leftExpr = Expression{
			NodeID:                p.newNodeID(),
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

func (p *Parser) matchTokenType(expectedTokenTypes []TokenType, lookAhead int) bool {
	for _, expectedTokenType := range expectedTokenTypes {
		if p.nextTokenIndex+lookAhead >= len(p.tokens) {
			return false
		}

		currToken := p.tokens[p.nextTokenIndex+lookAhead]
		if currToken.Type == expectedTokenType {
			return true
		}
	}

	return false
}

func (p *Parser) newNodeID() uint64 {
	nodeID := p.nextNodeID
	p.nextNodeID++
	return nodeID
}

func NewParser() *Parser {
	return &Parser{}
}
