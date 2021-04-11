package cmd

import (
	"errors"
	"fmt"
)

var (
	ErrNotEnoughArguments  = errors.New("not enough arguments")
	ErrUnrecognisedCommand = errors.New("unrecognised command")
	ErrWrongUsage          = errors.New("wrong usage")
)

func NewErrUnrecognisedCommand(cmd string) error {
	return fmt.Errorf("%s: %w", cmd, ErrUnrecognisedCommand)
}

func NewErrWrongUsage(correct string) error {
	return fmt.Errorf("%w: correct %s", ErrWrongUsage, correct)
}
