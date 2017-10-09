package scheme

import (
	"fmt"

	"errors"

	"github.com/travisjeffery/jocko/server/machinery"
)

var (
	ErrFuncNotImplemented = errors.New("api function not implemented")
	ErrFuncIncompatible   = errors.New("api function does not support requested version")

	blankFunc = APIFunction{}
)

type Registry struct {
	reg map[int16]APIFunction
}

func (ar Registry) Lookup(key int16, version int16) (APIFunction, error) {
	a, ok := ar.reg[key]
	if !ok {
		return blankFunc, ErrFuncNotImplemented
	}

	if !a.Version.Matches(version) {
		return blankFunc, ErrFuncIncompatible
	}

	return a, nil
}

func (ar Registry) AddFunction(key int16, version APIVersion, handler machinery.Handler) {
	ar.Add(key, NewFunction(key, version, handler))
}

func (ar Registry) Add(key int16, a APIFunction) {
	if _, exists := ar.reg[key]; exists {
		panic(fmt.Sprintf("API with key %s already registered", key))
	}
	ar.reg[key] = a
}

func (ar Registry) Len() int {
	return len(ar.reg)
}

func (ar Registry) Map() map[int16]APIFunction {
	return ar.reg[:]
}
