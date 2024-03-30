package main

import (
	"net/http"
	"os"

	"github.com/fatih/color"
	"github.com/go-chi/render"
)

var (
	cyan *color.Color = color.New(color.FgCyan)
	red               = color.New(color.FgRed)
)

func ErrExit(exitMsg string, err error) {
	red.Fprintln(os.Stderr, exitMsg, err)
	os.Exit(1)
}

// Helper function to return a json response
func Response(fn func(*http.Request) ([]byte, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := fn(r)
		if err != nil {
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": err.Error()})
		}
		render.JSON(w, r, data)
	}
}
