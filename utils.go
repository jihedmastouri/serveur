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

type ResError struct {
	Error  string `json:"error"`
	Status int    `json:"status"`
}

const SuccessMessage = "{\"message\": \"ok\"}"

type handlerResponse func(*http.Request) ([]byte, *ResError)

// Helper function to return a json response
func Response(fn func(*http.Request) ([]byte, *ResError)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := fn(r)
		if err != nil {
			render.Status(r, err.Status)
			render.JSON(w, r, map[string]string{"error": err.Error})
		}
		render.JSON(w, r, data)
	}
}

func flattenBytes(twoDBytes [][]byte) []byte {
	var result []byte

	for _, b := range twoDBytes {
		result = append(result, b...)
	}

	return result
}
