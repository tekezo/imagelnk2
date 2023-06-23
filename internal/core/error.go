package core

import "fmt"

type ErrMandatoryElementNotFound struct {
	name string
}

func (e *ErrMandatoryElementNotFound) Error() string {
	return fmt.Sprintf("mandatory element (%s) is not found", e.name)
}

func NewErrMandatoryElementNotFound(name string) *ErrMandatoryElementNotFound {
	return &ErrMandatoryElementNotFound{
		name: name,
	}
}
