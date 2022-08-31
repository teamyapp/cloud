package di

import (
	"path"
	"reflect"
)

type Factory func() interface{}

type Injector struct {
	deps map[string]Factory
}

func (i Injector) BindType(tp interface{}, factory Factory) {
	typeOf := reflect.TypeOf(tp)
	i.deps[getTypeKey(typeOf)] = factory
}

func (i Injector) Get(tp interface{}) interface{} {
	typeOf := reflect.TypeOf(tp)
	return i.deps[getTypeKey(typeOf)]()
}

func NewInjector() Injector {
	return Injector{
		deps: map[string]Factory{},
	}
}

func getTypeKey(tp reflect.Type) string {
	return path.Join(tp.PkgPath(), tp.Name())
}
