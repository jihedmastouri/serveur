package main

import (
	"fmt"
	"log"
	"net/http"
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
			ErrExit("Couldn't get the memmory flag", err)
		}

		dbPath, err := cmd.Flags().GetString("db-path")
		if err != nil {
			ErrExit("Couldn't get the db-path flag", err)
		}

		db := NewDB(isInMemory, dbPath)
		defer db.Close()

		staticPath, err := cmd.Flags().GetString("static")
		if err != nil {
			ErrExit("Couldn't get the static path", err)
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
			ErrExit("Couldn't parse the schema file", err)
		}

		isForceRefresh, err := cmd.Flags().GetBool("refresh")
		if err != nil {
			ErrExit("Couldn't get the refresh flag", err)
		}

		prevSchema := db.getSchema()
		isPrevSchemaValid := ValidateSchema(entities, prevSchema)
		if !isPrevSchemaValid || isForceRefresh {
			db.storeSchema(entities)
			FillDatabase(entities, db)
		}

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
					ErrExit("Couldn't parse the schema file", err)
				}

				prevSchema := db.getSchema()
				isPrevSchemaValid := ValidateSchema(entities, prevSchema)
				if !isPrevSchemaValid || isForceRefresh {
					// Close the previous db and create a new one
					db.Close()
					db = NewDB(isInMemory, dbPath)
					FillDatabase(entities, db)
					db.storeSchema(entities)
				}

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

var genCmd = &cobra.Command{
	Use:       "gen",
	Short:     "Generate fake data from a schema file",
	Long:      ``,
	Example:   "gen ./schema.json",
	ValidArgs: []string{
		// file or url
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		// logic to make sure that the flags are valid
		// and that we don't need flag for providing file and a url
	},
	Run: func(cmd *cobra.Command, args []string) {},
}

var checkCmd = &cobra.Command{
	Use:       "check",
	Short:     "Validate a data file against a schema file",
	Long:      ``,
	Example:   "check ./schema.json ./db.json",
	ValidArgs: []string{
		// file or url
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		// logic to make sure that the flags are valid
		// and that we don't need flag for providing file and a url
	},
	Run: func(cmd *cobra.Command, args []string) {},
}
