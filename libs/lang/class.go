package lang

import "fmt"

const ConstructorMethodName = "constructor"

var thisIdentifier = Token{
	Type:        IdentifierTokenType,
	Lexeme:      "this",
	Value:       "this",
	IsGenerated: true,
}

var superIdentifier = Token{
	Type:        IdentifierTokenType,
	Lexeme:      "super",
	Value:       "super",
	IsGenerated: true,
}

type Class struct {
	name            string
	superClass      *Class
	instanceMethods map[string]Callable
	instanceGetters map[string]Callable
	instanceSetters map[string]Callable
	staticFields    map[string]any
	staticMethods   map[string]Callable
	staticGetters   map[string]Callable
	staticSetters   map[string]Callable
	isGenerated     bool
}

func (c *Class) GetInstanceMethod(methodName string) (Callable, bool) {
	method, ok := c.instanceMethods[methodName]
	if ok {
		return method, ok
	}

	if c.superClass != nil {
		return c.superClass.GetInstanceMethod(methodName)
	}

	return Callable{}, false
}

func (c *Class) GetInstanceGetter(name string) (Callable, bool) {
	getter, ok := c.instanceGetters[name]
	if ok {
		return getter, ok
	}

	if c.superClass != nil {
		return c.superClass.GetInstanceGetter(name)
	}

	return Callable{}, false
}

func (c *Class) GetInstanceSetter(name string) (Callable, bool) {
	setter, ok := c.instanceSetters[name]
	if ok {
		return setter, ok
	}

	if c.superClass != nil {
		return c.superClass.GetInstanceSetter(name)
	}

	return Callable{}, false
}

func (c *Class) GetStaticMethod(methodName string) (Callable, bool) {
	method, ok := c.staticMethods[methodName]
	return method, ok
}

func (c *Class) GetStaticGetter(name string) (Callable, bool) {
	getter, ok := c.staticGetters[name]
	return getter, ok
}

func (c *Class) GetStatic(fieldName string) (any, bool, *Err) {
	props, ok := c.staticFields[fieldName]
	if ok {
		return props, ok, nil
	}

	getter, ok := c.staticGetters[fieldName]
	if ok {
		value, err := getter.Execute(&getter)
		return value, ok, err
	}

	method, ok := c.staticMethods[fieldName]
	return method, ok, nil
}

func (c *Class) SetStatic(fieldName string, value any) *Err {
	setter, ok := c.staticSetters[fieldName]
	if ok {
		_, err := setter.Execute(&setter, value)
		return err
	}

	c.staticFields[fieldName] = value
	return nil
}

func NewClass(
	name string,
	superClass *Class,
	instanceMethods map[string]Callable,
	instanceGetters map[string]Callable,
	instanceSetters map[string]Callable,
	staticMethods map[string]Callable,
	staticGetters map[string]Callable,
	staticSetters map[string]Callable,
	isGenerated bool,
) *Class {
	if instanceMethods == nil {
		instanceMethods = make(map[string]Callable)
	}

	if instanceGetters == nil {
		instanceGetters = make(map[string]Callable)
	}

	if instanceSetters == nil {
		instanceSetters = make(map[string]Callable)
	}

	if staticMethods == nil {
		staticMethods = make(map[string]Callable)
	}

	if staticGetters == nil {
		staticGetters = make(map[string]Callable)
	}

	if staticSetters == nil {
		staticSetters = make(map[string]Callable)
	}

	return &Class{
		name:            name,
		superClass:      superClass,
		instanceMethods: instanceMethods,
		instanceGetters: instanceGetters,
		instanceSetters: instanceSetters,
		staticFields:    make(map[string]any),
		staticMethods:   staticMethods,
		staticGetters:   staticGetters,
		staticSetters:   staticSetters,
		isGenerated:     isGenerated,
	}
}

type Instance struct {
	class       *Class
	fields      map[string]any
	isGenerated bool
}

func (i *Instance) Get(fieldName string) (any, bool, *Err) {
	props, ok := i.fields[fieldName]
	if ok {
		return props, ok, nil
	}

	if i.class == nil {
		return nil, false, nil
	}

	getter, ok := i.class.GetInstanceGetter(fieldName)
	if ok {
		getter = bindMethodToInstance(i, getter)
		value, err := getter.Execute(&getter)
		return value, ok, err
	}

	method, ok := i.class.GetInstanceMethod(fieldName)
	if ok {
		method = bindMethodToInstance(i, method)
	}

	return method, ok, nil
}

func (i *Instance) Set(fieldName string, value any) *Err {
	if i.class != nil {
		setter, ok := i.class.GetInstanceSetter(fieldName)
		if ok {
			setter = bindMethodToInstance(i, setter)
			_, err := setter.Execute(&setter, value)
			return err
		}
	}

	i.fields[fieldName] = value
	return nil
}

func NewInstance(class *Class, constructorArgs []any, isGenerated bool) *Instance {
	instance := &Instance{
		class:       class,
		fields:      make(map[string]any),
		isGenerated: isGenerated,
	}

	if class != nil {
		constructor, ok := class.GetInstanceMethod(ConstructorMethodName)
		if ok {
			constructor = bindMethodToInstance(instance, constructor)
			constructor.Execute(&constructor, constructorArgs...)
		}
	}

	return instance
}

func NewConstructorWithFields(fields map[string]any) Callable {
	return Callable{
		Name:          ConstructorMethodName,
		IsConstructor: true,
		Arity:         0,
		Execute: func(callable *Callable, args ...any) (any, *Err) {
			instanceVal, err := callable.Closure.Get(thisIdentifier)
			if err != nil {
				return nil, err
			}

			instance, ok := instanceVal.(*Instance)
			if !ok {
				return nil, &Err{
					Message:           fmt.Sprintf("Internal error: '%s' is not an instance", thisIdentifier.Value),
					FromGeneratedCode: true,
				}
			}

			instance.fields = make(map[string]any)
			for fieldName, fieldValue := range fields {
				instance.fields[fieldName] = ToInternalValue(fieldValue)
			}

			return nil, nil
		},
		IsGenerated: true,
	}
}

func bindMethodToInstance(instance *Instance, method Callable) Callable {
	var environment *Environment
	if method.Closure != nil {
		environment = method.Closure.NewInnerEnvironment()
	} else {
		environment = NewEnvironment()
	}

	environment.DefineWithInitializer(thisIdentifier, instance)
	method = method.Copy()
	method.Closure = environment
	return method
}

func newThisIdentifier(line int, column int, isGenerated bool) Token {
	return Token{
		Type:        thisIdentifier.Type,
		Lexeme:      thisIdentifier.Lexeme,
		Value:       thisIdentifier.Value,
		Line:        line,
		Column:      column,
		IsGenerated: isGenerated,
	}
}

func newSuperIdentifier(line int, column int, isGenerated bool) Token {
	return Token{
		Type:        superIdentifier.Type,
		Lexeme:      superIdentifier.Lexeme,
		Value:       superIdentifier.Value,
		Line:        line,
		Column:      column,
		IsGenerated: isGenerated,
	}
}
