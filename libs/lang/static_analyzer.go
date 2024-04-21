package lang

import "fmt"

type IdentifierStatus string

const (
	DeclaredIdentifierStatus IdentifierStatus = "declared"
	DefinedIdentifierStatus  IdentifierStatus = "defined"
	UsedIdentifierStatus     IdentifierStatus = "used"
)

type CallableStatus string

const (
	GlobalCallableStatus        CallableStatus = "Global"
	InFunctionCallableStatus    CallableStatus = "InFunction"
	InMethodCallableStatus      CallableStatus = "InMethod"
	InConstructorCallableStatus CallableStatus = "InConstructor"
)

type ClassStatus string

const (
	GlobalClassStatus     ClassStatus = "Global"
	InClassClassStatus    ClassStatus = "InClass"
	InSubClassClassStatus ClassStatus = "InSubClass"
)

type Declaration struct {
	Identifier  Token
	Status      IdentifierStatus
	CheckUnused bool
}

type StaticAnalyzer struct {
	scopes                [][]*Declaration
	localRefs             map[uint64]*Reference
	currentCallableStatus CallableStatus
	currentClassStatus    ClassStatus
}

func (s *StaticAnalyzer) Analyze(statements []Statement) *Err {
	s.currentCallableStatus = GlobalCallableStatus
	s.currentClassStatus = GlobalClassStatus
	s.beginScope()
	err := s.resolve(statements)
	if err != nil {
		return err
	}

	return s.endScope()
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

func (s *StaticAnalyzer) GetLocalRef(nodeID uint64) (*Reference, bool) {
	ref, ok := s.localRefs[nodeID]
	return ref, ok
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
		return s.resolveCallableStatement(statement, InFunctionCallableStatus)
	case ReturnStatementType:
		return s.resolveReturnStatement(statement)
	case ClassStatementType:
		return s.resolveClassStatement(statement)
	default:
		return &Err{
			Message:           fmt.Sprintf("Unknown statement type: %v", statement.Type),
			Line:              statement.Line,
			Column:            statement.Column,
			FromGeneratedCode: statement.IsGenerated,
		}
	}
}

func (s *StaticAnalyzer) resolveClassStatement(statement Statement) *Err {
	classStatus := s.currentClassStatus
	s.currentClassStatus = InClassClassStatus

	err := s.declare(*statement.ClassIdentifier, false)
	if err != nil {
		return err
	}

	s.define(*statement.ClassIdentifier)

	var superToken *Token
	if statement.ClassSuperClassExpression != nil {
		classIdentifier := statement.ClassIdentifier.Value.(string)
		superToken = &superIdentifier

		if statement.ClassSuperClassExpression.Type != IdentifierExpressionType {
			return &Err{
				Message:           "SuperClass must be an identifier",
				Line:              statement.ClassSuperClassExpression.Line,
				Column:            statement.ClassSuperClassExpression.Column,
				FromGeneratedCode: statement.ClassSuperClassExpression.IsGenerated,
			}
		}

		superClassIdentifier := statement.ClassSuperClassExpression.Identifier.Value.(string)
		if classIdentifier == superClassIdentifier {
			return &Err{
				Message:           "A class cannot inherit from itself",
				Line:              statement.ClassSuperClassExpression.Line,
				Column:            statement.ClassSuperClassExpression.Column,
				FromGeneratedCode: statement.ClassSuperClassExpression.IsGenerated,
			}
		}

		s.currentClassStatus = InSubClassClassStatus
		err = s.resolveExpression(*statement.ClassSuperClassExpression)
		if err != nil {
			return err
		}
	}

	err = s.withIdentifierInNewScope(superToken, func() *Err {
		for _, classMethodDeclaration := range statement.ClassInstanceMethodDeclarations {
			err = s.withIdentifierInNewScope(&thisIdentifier, func() *Err {
				callableStatus := InMethodCallableStatus
				if classMethodDeclaration.CallableName.Value.(string) == ConstructorMethodName {
					callableStatus = InConstructorCallableStatus
				}

				return s.resolveCallableStatement(classMethodDeclaration, callableStatus)
			})
			if err != nil {
				return err
			}
		}

		for _, classGetterDeclaration := range statement.ClassInstanceGetterDeclarations {
			err = s.withIdentifierInNewScope(&thisIdentifier, func() *Err {
				return s.resolveCallableStatement(classGetterDeclaration, InMethodCallableStatus)
			})
			if err != nil {
				return err
			}
		}

		for _, classSetterDeclaration := range statement.ClassInstanceSetterDeclarations {
			err = s.withIdentifierInNewScope(&thisIdentifier, func() *Err {
				return s.resolveCallableStatement(classSetterDeclaration, InMethodCallableStatus)
			})
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	for _, classMethodDeclaration := range statement.ClassStaticMethodDeclarations {
		err = s.resolveCallableStatement(classMethodDeclaration, InMethodCallableStatus)
		if err != nil {
			return err
		}
	}

	for _, classGetterDeclaration := range statement.ClassStaticGetterDeclarations {
		err = s.resolveCallableStatement(classGetterDeclaration, InMethodCallableStatus)
		if err != nil {
			return err
		}
	}

	for _, classSetterDeclaration := range statement.ClassStaticSetterDeclarations {
		err = s.resolveCallableStatement(classSetterDeclaration, InMethodCallableStatus)
		if err != nil {
			return err
		}
	}

	s.currentClassStatus = classStatus
	return nil
}

func (s *StaticAnalyzer) withIdentifierInNewScope(identifier *Token, resolve func() *Err) *Err {
	if identifier != nil {
		s.beginScope()
		err := s.declare(*identifier, false)
		if err != nil {
			return err
		}

		s.define(*identifier)
	}

	resolveErr := resolve()
	if resolveErr != nil {
		return resolveErr
	}

	if identifier != nil {
		err := s.endScope()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *StaticAnalyzer) resolveExpressionStatement(statement Statement) *Err {
	return s.resolveExpression(*statement.StatementExpression)
}

func (s *StaticAnalyzer) resolveReturnStatement(statement Statement) *Err {
	if s.currentCallableStatus == GlobalCallableStatus {
		return &Err{
			Message:           "Return outside of a callable",
			Line:              statement.Line,
			Column:            statement.Column,
			FromGeneratedCode: statement.IsGenerated,
		}
	}

	if statement.ReturnValueExpression != nil {
		if s.currentCallableStatus == InConstructorCallableStatus {
			return &Err{
				Message:           "Cannot return a value from a constructor",
				Line:              statement.Line,
				Column:            statement.Column,
				FromGeneratedCode: statement.IsGenerated,
			}
		}

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

func (s *StaticAnalyzer) resolveCallableStatement(statement Statement, callableStatus CallableStatus) *Err {
	err := s.declare(*statement.CallableName, false)
	if err != nil {
		return err
	}

	s.define(*statement.CallableName)

	currentCallableStatus := s.currentCallableStatus
	s.currentCallableStatus = callableStatus
	s.beginScope()
	for _, param := range statement.CallableParameters {
		err = s.declare(param, false)
		if err != nil {
			return err
		}

		s.define(param)
	}

	err = s.resolveBlockStatement(*statement.CallableBody, false)
	if err != nil {
		return err
	}

	err = s.endScope()
	if err != nil {
		return err
	}

	s.currentCallableStatus = currentCallableStatus
	return nil
}

func (s *StaticAnalyzer) resolveLetStatement(statement Statement) *Err {
	err := s.declare(*statement.LetIdentifier, true)
	if err != nil {
		return err
	}

	if statement.LetInitializerExpression != nil {
		err = s.resolveExpression(*statement.LetInitializerExpression)
		if err != nil {
			return err
		}
	}

	s.define(*statement.LetIdentifier)
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
		return s.endScope()
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
	case NewInstanceExpressionType:
		return s.resolveNewInstanceExpression(expression)
	case GetExpressionType:
		return s.resolveGetExpression(expression)
	case SetExpressionType:
		return s.resolveSetExpression(expression)
	case ThisExpressionType:
		return s.resolveThisExpression(expression)
	case SuperExpressionType:
		return s.resolveSuperExpression(expression)
	default:
		return &Err{
			Message:           fmt.Sprintf("Unknown expression type: %v", expression.Type),
			Line:              expression.Line,
			Column:            expression.Column,
			FromGeneratedCode: expression.IsGenerated,
		}
	}
}

func (s *StaticAnalyzer) resolveSuperExpression(expression Expression) *Err {
	if s.currentClassStatus != InSubClassClassStatus {
		return &Err{
			Message:           "Cannot use 'super' outside of a subClass",
			Line:              expression.Line,
			Column:            expression.Column,
			FromGeneratedCode: expression.IsGenerated,
		}
	}

	s.resolveLocal(expression.NodeID, newSuperIdentifier(expression.Line, expression.Column, expression.IsGenerated))
	return nil
}

func (s *StaticAnalyzer) resolveThisExpression(expression Expression) *Err {
	if s.currentClassStatus != InClassClassStatus &&
		s.currentClassStatus != InSubClassClassStatus {
		return &Err{
			Message:           "Cannot use 'this' outside of a class",
			Line:              expression.Line,
			Column:            expression.Column,
			FromGeneratedCode: expression.IsGenerated,
		}
	}

	s.resolveLocal(expression.NodeID, newThisIdentifier(expression.Line, expression.Column, expression.IsGenerated))
	return nil
}

func (s *StaticAnalyzer) resolveSetExpression(expression Expression) *Err {
	err := s.resolveExpression(*expression.SetObjectExpression)
	if err != nil {
		return err
	}

	return s.resolveExpression(*expression.SetValueExpression)
}

func (s *StaticAnalyzer) resolveGetExpression(expression Expression) *Err {
	return s.resolveExpression(*expression.GetObjectExpression)
}

func (s *StaticAnalyzer) resolveNewInstanceExpression(expression Expression) *Err {
	err := s.resolveExpression(*expression.NewInstanceClassExpression)
	if err != nil {
		return err
	}

	for _, argExpr := range expression.NewInstanceConstructorArgumentExpressions {
		err = s.resolveExpression(argExpr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *StaticAnalyzer) resolveLambdaExpression(expression Expression) *Err {
	currentCallableStatus := s.currentCallableStatus
	s.currentCallableStatus = InFunctionCallableStatus
	s.beginScope()
	for _, param := range expression.LambdaParameters {
		err := s.declare(param, false)
		if err != nil {
			return err
		}

		s.define(param)
	}

	err := s.resolveBlockStatement(expression.LambdaBody, false)
	if err != nil {
		return err
	}

	err = s.endScope()
	if err != nil {
		return err
	}

	s.currentCallableStatus = currentCallableStatus
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

	return s.resolveLocal(expression.NodeID, expression.AssignmentIdentifier)
}

func (s *StaticAnalyzer) resolveIdentifierExpression(expression Expression) *Err {
	return s.resolveLocal(expression.NodeID, expression.Identifier)
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

func (s *StaticAnalyzer) resolveLocal(nodeID uint64, token Token) *Err {
	name := token.Value.(string)
	for index := len(s.scopes) - 1; index >= 0; index-- {
		scope := s.scopes[index]
		declaration, stackIndex, ok := findDeclaration(scope, token)
		if ok {
			if declaration.Status == DeclaredIdentifierStatus {
				return &Err{
					Message:           fmt.Sprintf("Reading uninitialized variable '%s'", name),
					Line:              declaration.Identifier.Line,
					Column:            declaration.Identifier.Column,
					FromGeneratedCode: declaration.Identifier.IsGenerated,
				}
			}

			declaration.Status = UsedIdentifierStatus
			s.localRefs[nodeID] = &Reference{
				EnvironmentDistance: len(s.scopes) - 1 - index,
				StackIndex:          stackIndex,
			}
			return nil
		}
	}

	return nil
}

func (s *StaticAnalyzer) beginScope() {
	s.scopes = append(s.scopes, make([]*Declaration, 0))
}

func (s *StaticAnalyzer) endScope() *Err {
	scope := s.scopes[len(s.scopes)-1]
	for _, declaration := range scope {
		if declaration.CheckUnused && declaration.Status != UsedIdentifierStatus {
			return &Err{
				Message:           fmt.Sprintf("Local variable '%s' is never used", declaration.Identifier.Value.(string)),
				Line:              declaration.Identifier.Line,
				Column:            declaration.Identifier.Column,
				FromGeneratedCode: declaration.Identifier.IsGenerated,
			}
		}
	}

	s.scopes = s.scopes[:len(s.scopes)-1]
	return nil
}

func (s *StaticAnalyzer) declare(identifier Token, checkUnused bool) *Err {
	scope := s.scopes[len(s.scopes)-1]
	_, _, ok := findDeclaration(scope, identifier)
	if ok {
		return &Err{
			Message:           fmt.Sprintf("name '%s' already declared in this scope", identifier.Value.(string)),
			Line:              identifier.Line,
			Column:            identifier.Column,
			FromGeneratedCode: identifier.IsGenerated,
		}
	}

	declaration := &Declaration{
		Identifier:  identifier,
		Status:      DeclaredIdentifierStatus,
		CheckUnused: checkUnused,
	}
	scope = append(scope, declaration)
	s.scopes[len(s.scopes)-1] = scope
	return nil
}

func (s *StaticAnalyzer) define(identifier Token) {
	scope := s.scopes[len(s.scopes)-1]
	_, index, ok := findDeclaration(scope, identifier)
	if ok {
		scope[index].Status = DefinedIdentifierStatus
	}
}

func NewStaticAnalyzer() *StaticAnalyzer {
	return &StaticAnalyzer{
		scopes:    [][]*Declaration{},
		localRefs: map[uint64]*Reference{},
	}
}

func findDeclaration(scope []*Declaration, token Token) (*Declaration, int, bool) {
	name := token.Value.(string)
	for stackIndex, declaration := range scope {
		if declaration.Identifier.Value.(string) == name {
			return declaration, stackIndex, true
		}
	}

	return nil, 0, false
}
