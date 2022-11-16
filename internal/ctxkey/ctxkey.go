package ctxkey

import (
	"fmt"
)

var (
	Logger = New("Logger")
)

type Key interface {
	fmt.Stringer
}

type key struct {
	name string
}

func New(name string) Key {
	return &key{name}
}

func (k *key) String() string {
	return "context value " + k.name
}
