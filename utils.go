package main

import (
	"os"

	"github.com/fatih/color"
)

var (
	cyan *color.Color = color.New(color.FgCyan)
	red               = color.New(color.FgRed)
)

func ErrExit(exitMsg string, err error) {
	red.Fprintln(os.Stderr, exitMsg, err)
	os.Exit(1)
}
