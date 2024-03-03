package lang

import "C"

const constructorMethodName = "constructor"

var thisIdentifier = Token{
	Type:        IdentifierTokenType,
	Lexeme:      "this",
	Value:       "this",
	Line:        0,
	Column:      0,
	IsGenerated: true,
}

type Class struct {
	name            string
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
	return method, ok
}

func (c *Class) GetInstanceGetter(name string) (Callable, bool) {
	getter, ok := c.instanceGetters[name]
	return getter, ok
}

func (c *Class) GetInstanceSetter(name string) (Callable, bool) {
	setter, ok := c.instanceSetters[name]
	return setter, ok
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
	instanceMethods map[string]Callable,
	instanceGetters map[string]Callable,
	instanceSetters map[string]Callable,
	staticMethods map[string]Callable,
	staticGetters map[string]Callable,
	staticSetters map[string]Callable,
	isGenerated bool,
) *Class {
	return &Class{
		name:            name,
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
		constructor, ok := class.GetInstanceMethod(constructorMethodName)
		if ok {
			constructor = bindMethodToInstance(instance, constructor)
			constructor.Execute(&constructor, constructorArgs...)
		}
	}

	return instance
}

func bindMethodToInstance(instance *Instance, method Callable) Callable {
	environment := method.Closure.NewInnerEnvironment()
	environment.DefineWithInitializer(thisIdentifier, instance)
	method = method.Copy()
	method.Closure = environment
	return method
}
