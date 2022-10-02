package test

import (
	"fmt"
	"strings"
)

type ContainsMatcher struct {
	s string
}

func Contains(s string) ContainsMatcher {
	return ContainsMatcher{s}
}

func (m ContainsMatcher) Matches(x interface{}) bool {
	s, ok := x.(string)
	if !ok {
		return false
	}

	return strings.Contains(s, m.s)
}

func (m ContainsMatcher) String() string {
	return fmt.Sprintf("contains %v (%T)", m.s, m.s)
}
