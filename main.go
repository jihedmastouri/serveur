package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "serveur",
	Short: "A mock server with auto-generated data.",
	Long: `Serveur is a mock server for testing purposes.
Provided with a schema file, Serveur will generate fake data and spin-up a http server.
the schema file should be a JSON with the following structure:
	{
		"entity1": {
			"count": 10, // number of records to generate
			"fileds": {
				"field1": "type",
				"field2": "type",
				...
			}
		},
		"entity2": {
			"count": 10,
			"fileds": {
				"field1": "type",
				"field2": "type",
				...
			}
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
	- ref
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
		isInMemory, err := cmd.Flags().GetBool("memmory")
		if err != nil {
			log.Fatal("Couldn't get the memmory flag")
		}

		db := NewDB(isInMemory)
		defer db.Close()

		staticPath, err := cmd.Flags().GetString("static")
		if err != nil {
			ErrExit("Couldn't get the static path:", err)
		}

		schemaPath := "./schema.json"
		if len(args) != 0 && args[0] != "" {
			schemaPath = args[0]
		}

		// Download the schema file if it's a url
		// Watch the schema file for changes
		path, watcher := initFile(schemaPath, "json")
		defer watcher.Close()

		entities, err := ParseFile(path)
		if err != nil {
			ErrExit("Couldn't parse the schema file:", err)
		}

		FillDatabase(entities, db)

		// Initialize the server
		server := NewRestServer(
			db,
			entities,
			AddLogger(),
			AddHomePage(schemaPath),
			AddStaticFiles(staticPath),
		)
		server.InitRouter()
		srv := &http.Server{
			Addr:    fmt.Sprintf(":%d", 3000),
			Handler: server.mux,
		}

		// Start the server
		go func() {
			log.Fatal(srv.ListenAndServe())
		}()

		for {
			event := <-watcher.Events

			if event.Has(fsnotify.Rename) {
				// HACK: The only way I found to makr sure I keep watching the file :(
				watcher.Remove(event.Name)
				watcher.Add(event.Name)
			}

			// To not spam the server with multiple events
			time.Sleep(1 * time.Second)

			// If the file is written to or renamed (which is the case when the file is saved in an editor)
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Rename) {
				entities, err := ParseFile(event.Name)
				if err != nil {
					log.Println("Couldn't parse the schema file:", err)
					continue
				}

				// Close the previous db and create a new one
				db.Close()
				db = NewDB(isInMemory)
				FillDatabase(entities, db)

				// Close the previous server and create a new one
				server := NewRestServer(
					db,
					entities,
					AddLogger(),
					AddHomePage(schemaPath),
					AddStaticFiles(staticPath),
				)
				server.InitRouter()
				srv.Handler = server.mux
				srv.ListenAndServe()
			}
		}
	},
}

func main() {
	rootCmd.Flags().StringP("static", "s", "", "Path to the static files directory")
	rootCmd.Flags().StringP("dump", "o", "", "Path to the dump file. the output will be a json file")
	rootCmd.Flags().StringP("ingest", "i", "", "Path to the ingest file. It should be a json file. If schema is provided, it will be used to validate the data")
	rootCmd.Flags().BoolP("memmory", "m", false, "Run the server in memmory mode. No data will be persisted")
	rootCmd.Flags().IntP("port", "p", 8080, "Port to listen on")
	rootCmd.Flags().BoolP("verbose", "v", false, "Verbose mode")
	rootCmd.Flags().StringP("log", "l", "serveur.log.txt", "write logs to a specific file")
	rootCmd.Flags().StringP("ext", "e", "json", "specify the file extension for the schema file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
