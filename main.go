package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "serveur",
	Short: "A mock server for testing purposes",
	Long: `Serveur is a mock server for testing purposes.
Provided with a schema file, Serveur will generate fake data and spin-up a http server.
the schema file should be a JSON with the following structure:
	{
		"entity1": {
			"field1": "type",
			"field2": "type",
			...
		},
		"entity2": {
			"field1": "type",
			"field2": "type",
			...
		},
		...
	}
The types can be one of the following:
	- str / string
	- num / number
	- bool
	- date
	- email
	- url
	- ip
	- uuid
	- id
	- name
	- username
	- fullname
	- address / addr
	- phone
	- paragraph / pg
The file can be provided as a local file or a url.
Each entity will be served at a different endpoint (the same way "json-server" does it).
	`,
	Example:   "serveur ./schema.json --port 8080",
	ValidArgs: []string{
		// file or url
	},
	Version: "v0.0.0",
	PreRun: func(cmd *cobra.Command, args []string) {
		// logic to make sure that the flags are valid
		// and that we don't need flag for providing file and a url
	},
	Run: func(cmd *cobra.Command, args []string) {
		server := NewRestServer(AddLogger(), AddHomePage)
		m := make(map[string]any)
		m["name"] = "str"
		m["title"] = "str"
		m2 := make(map[string]any)
		m2["name"] = "str"
		m2["title"] = "str"
		server.InitRouter([]Entity{{name: "users", fields: m}, {name: "posts", fields: m2}})
		log.Fatal(http.ListenAndServe(":3000", server.mux))
	},
}

func main() {
	// rootCmd.Flags().StringP("file", "f", "schema.json", "Path to the schema file")
	// rootCmd.Flags().StringP("url", "u", "", "Url to the schema file")
	rootCmd.Flags().StringP("static", "s", "", "Path to the static files directory")
	rootCmd.Flags().StringP("dump", "o", "", "Path to the dump file. the output will be a json file")
	rootCmd.Flags().StringP("ingest", "i", "", "Path to the ingest file. It should be a json file. If schema is provided, it will be used to validate the data")
	rootCmd.Flags().IntP("port", "p", 8080, "Port to listen on")
	rootCmd.Flags().BoolP("verbose", "v", false, "Verbose mode")
	rootCmd.Flags().StringP("log", "l", "serveur.log.txt", "write logs to a specific file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
