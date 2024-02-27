package lang

import "fmt"

type StaticAnalyzer struct {
	scopes         []map[string]bool
	localDistances map[uint64]int
	isInCallable   bool
}

func (s *StaticAnalyzer) Analyze(statements []Statement) *Err {
	s.isInCallable = false
	s.beginScope()
	err := s.resolve(statements)
	if err != nil {
		return err
	}

	s.endScope()
	return nil
}

func (s *StaticAnalyzer) resolve(statements []Statement) *Err {
	for _, statement := range statements {
		err := s.resolveStatement(statement)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *StaticAnalyzer) GetLocalScopeDistance(nodeID uint64) (int, bool) {
	distance, ok := s.localDistances[nodeID]
	return distance, ok
}

func (s *StaticAnalyzer) resolveStatement(statement Statement) *Err {
	switch statement.Type {
	case ExpressionStatementType:
		return s.resolveExpressionStatement(statement)
	case LetStatementType:
		return s.resolveLetStatement(statement)
	case BlockStatementType:
		return s.resolveBlockStatement(statement, true)
	case IfStatementType:
		return s.resolveIfStatement(statement)
	case WhileStatementType:
		return s.resolveWhileStatement(statement)
	case BreakStatementType:
		return nil
	case ContinueStatementType:
		return nil
	case CallableStatementType:
		return s.resolveCallableStatement(statement)
	case ReturnStatementType:
		return s.resolveReturnStatement(statement)
	default:
		return &Err{
			Message:           fmt.Sprintf("Unknown statement type: %v", statement.Type),
			Line:              statement.Line,
			Column:            statement.Column,
			FromGeneratedCode: statement.IsGenerated,
		}
	}
}

func (s *StaticAnalyzer) resolveExpressionStatement(statement Statement) *Err {
	return s.resolveExpression(*statement.StatementExpression)
}

func (s *StaticAnalyzer) resolveReturnStatement(statement Statement) *Err {
	if !s.isInCallable {
		return &Err{
			Message:           "Return outside of a callable",
			Line:              statement.Line,
			Column:            statement.Column,
			FromGeneratedCode: statement.IsGenerated,
		}
	}

	if statement.ReturnValueExpression != nil {
		return s.resolveExpression(*statement.ReturnValueExpression)
	}

	return nil
}

func (s *StaticAnalyzer) resolveWhileStatement(statement Statement) *Err {
	err := s.resolveExpression(*statement.WhileConditionExpression)
	if err != nil {
		return err
	}

	return s.resolveStatement(*statement.WhileBodyStatement)
}

func (s *StaticAnalyzer) resolveIfStatement(statement Statement) *Err {
	err := s.resolveExpression(*statement.IfConditionExpression)
	if err != nil {
		return err
	}

	err = s.resolveStatement(*statement.IfTrueBranchStatement)
	if err != nil {
		return err
	}

	if statement.IfFalseBranchStatement != nil {
		return s.resolveStatement(*statement.IfFalseBranchStatement)
	}

	return nil
}

func (s *StaticAnalyzer) resolveCallableStatement(statement Statement) *Err {
	err := s.declare(*statement.CallableName)
	if err != nil {
		return err
	}

	s.define(statement.CallableName.Value.(string))

	isInCallable := s.isInCallable
	s.isInCallable = true
	s.beginScope()
	for _, param := range statement.CallableParameters {
		err = s.declare(param)
		if err != nil {
			return err
		}

		s.define(param.Value.(string))
	}

	err = s.resolveBlockStatement(*statement.CallableBody, false)
	if err != nil {
		return err
	}

	s.endScope()
	s.isInCallable = isInCallable
	return nil
}

func (s *StaticAnalyzer) resolveLetStatement(statement Statement) *Err {
	err := s.declare(*statement.LetIdentifier)
	if err != nil {
		return err
	}

	if statement.LetInitializerExpression != nil {
		err = s.resolveExpression(*statement.LetInitializerExpression)
		if err != nil {
			return err
		}
	}

	s.define(statement.LetIdentifier.Value.(string))
	return nil
}

func (s *StaticAnalyzer) resolveBlockStatement(statement Statement, createNewScope bool) *Err {
	if createNewScope {
		s.beginScope()
	}

	err := s.resolve(statement.BlockInnerStatements)
	if err != nil {
		return err
	}

	if createNewScope {
		s.endScope()
	}

	return nil
}

func (s *StaticAnalyzer) resolveExpression(expression Expression) *Err {
	switch expression.Type {
	case TernaryExpressionType:
		return s.resolveTernaryExpression(expression)
	case BinaryExpressionType:
		return s.resolveBinaryExpression(expression)
	case UnaryExpressionType:
		return s.resolveUnaryExpression(expression)
	case LiteralExpressionType:
		return nil
	case GroupingExpressionType:
		return s.resolveExpression(*expression.GroupInnerExpression)
	case ExpressionListExpressionType:
		return s.resolveExpressionList(expression)
	case IdentifierExpressionType:
		return s.resolveIdentifierExpression(expression)
	case AssignmentExpressionType:
		return s.resolveAssignmentExpression(expression)
	case CallExpressionType:
		return s.resolveCallExpression(expression)
	case LambdaExpressionType:
		return s.resolveLambdaExpression(expression)
	default:
		return &Err{
			Message:           fmt.Sprintf("Unknown expression type: %v", expression.Type),
			Line:              expression.Line,
			Column:            expression.Column,
			FromGeneratedCode: expression.IsGenerated,
		}
	}
}

func (s *StaticAnalyzer) resolveLambdaExpression(expression Expression) *Err {
	isInCallable := s.isInCallable
	s.isInCallable = true
	s.beginScope()
	for _, param := range expression.LambdaParameters {
		err := s.declare(param)
		if err != nil {
			return err
		}

		s.define(param.Value.(string))
	}

	err := s.resolveBlockStatement(expression.LambdaBody, false)
	if err != nil {
		return err
	}

	s.endScope()
	s.isInCallable = isInCallable
	return nil
}

func (s *StaticAnalyzer) resolveCallExpression(expression Expression) *Err {
	err := s.resolveExpression(*expression.CallableExpression)
	if err != nil {
		return err
	}

	for _, argExpr := range expression.CallArgumentExpressions {
		err = s.resolveExpression(argExpr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *StaticAnalyzer) resolveAssignmentExpression(expression Expression) *Err {
	err := s.resolveExpression(*expression.AssignmentValueExpression)
	if err != nil {
		return err
	}

	s.resolveLocal(expression.NodeID, expression.AssignmentIdentifier)
	return nil
}

func (s *StaticAnalyzer) resolveIdentifierExpression(expression Expression) *Err {
	name := expression.Identifier.Value.(string)
	initialized, ok := s.scopes[len(s.scopes)-1][name]
	if ok {
		if !initialized {
			return &Err{
				Message:           fmt.Sprintf("Reading uninitialized variable '%s'", name),
				Line:              expression.Identifier.Line,
				Column:            expression.Identifier.Column,
				FromGeneratedCode: expression.IsGenerated,
			}
		}
	}

	s.resolveLocal(expression.NodeID, expression.Identifier)
	return nil
}

func (s *StaticAnalyzer) resolveExpressionList(expression Expression) *Err {
	for _, expr := range expression.ExpressionList {
		err := s.resolveExpression(expr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *StaticAnalyzer) resolveUnaryExpression(expression Expression) *Err {
	return s.resolveExpression(*expression.UnaryExpression)
}

func (s *StaticAnalyzer) resolveBinaryExpression(expression Expression) *Err {
	err := s.resolveExpression(*expression.BinaryLeftExpression)
	if err != nil {
		return err
	}

	return s.resolveExpression(*expression.BinaryRightExpression)
}

func (s *StaticAnalyzer) resolveTernaryExpression(expression Expression) *Err {
	err := s.resolveExpression(*expression.TernaryConditionExpression)
	if err != nil {
		return err
	}

	err = s.resolveExpression(*expression.TernaryTrueExpression)
	if err != nil {
		return err
	}

	return s.resolveExpression(*expression.TernaryFalseExpression)
}

func (s *StaticAnalyzer) resolveLocal(nodeID uint64, token Token) {
	for index := len(s.scopes) - 1; index >= 0; index-- {
		scope := s.scopes[index]
		initialized, ok := scope[token.Value.(string)]
		if ok && initialized {
			s.localDistances[nodeID] = len(s.scopes) - 1 - index
			return
		}
	}
}

func (s *StaticAnalyzer) beginScope() {
	s.scopes = append(s.scopes, make(map[string]bool))
}

func (s *StaticAnalyzer) endScope() {
	s.scopes = s.scopes[:len(s.scopes)-1]
}

func (s *StaticAnalyzer) declare(identifier Token) *Err {
	scope := s.scopes[len(s.scopes)-1]
	name := identifier.Value.(string)
	if _, ok := scope[name]; ok {
		return &Err{
			Message:           fmt.Sprintf("name '%s' already declared in this scope", name),
			Line:              identifier.Line,
			Column:            identifier.Column,
			FromGeneratedCode: identifier.IsGenerated,
		}
	}

	scope[name] = false
	return nil
}

func (s *StaticAnalyzer) define(name string) {
	scope := s.scopes[len(s.scopes)-1]
	scope[name] = true
}

func NewStaticAnalyzer() *StaticAnalyzer {
	return &StaticAnalyzer{
		scopes:         []map[string]bool{},
		localDistances: map[uint64]int{},
	}
}
