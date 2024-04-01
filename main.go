package main

import (
	"fmt"
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

func main() {
	rootCmd.Flags().StringP("static", "s", "", "Path to the static files directory")
	rootCmd.Flags().BoolP("refresh", "r", false, "ignore the cache and force a refresh of the schema file")
	rootCmd.Flags().StringP("db-path", "d", "./db", "Path to the database directory. It will be created if it doesn't exist")
	rootCmd.Flags().StringP("out-dump", "o", "", "Path to the dump file. the output will be a json file")
	rootCmd.Flags().StringP("ingest", "i", "", "Path to the ingest file. It should be a json file. If schema is provided, it will be used to validate the data")
	rootCmd.Flags().BoolP("memmory", "m", false, "Run the server in memmory mode. No data will be persisted")
	rootCmd.Flags().IntP("port", "p", 3000, "Port to listen on")
	rootCmd.Flags().BoolP("verbose", "v", false, "Verbose mode")
	rootCmd.Flags().StringP("log", "l", "serveur.log.txt", "write logs to a specific file")

	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(genCmd)
	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
