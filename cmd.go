package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:       "gen",
	Short:     "Generate fake data from a schema file",
	Example:   "gen ./schema.json",
	ValidArgs: []string{"schema-file", "data-file"},
	Run: func(cmd *cobra.Command, args []string) {
		schemaPath := "./schema.json"
		if len(args) != 0 && args[0] != "" {
			schemaPath = args[0]
		}
		downloadFile(schemaPath)

		dataPath := "./db.json"
		if len(args) != 0 && args[1] != "" {
			dataPath = args[1]
		}

		// TODO: Below code must be extracted to a function
		file, err := os.Create(dataPath)
		if err != nil {
			ErrExit("Couldn't create the output file", err)
		}

		encoder := json.NewEncoder(file)

		entities, err := ParseFile(schemaPath)
		if err != nil {
			ErrExit("Couldn't parse the schema file", err)
		}

		for _, entity := range entities {
			go func(entity Entity) {
				data, err := GenerateFakeData(entity.Schema)
				if err != nil {
					ErrExit("Couldn't generate fake data", err)
				}
				err = encoder.Encode(data)
				if err != nil {
					ErrExit("Couldn't write to the output file", err)
				}
			}(entity)
		}

	},
}

var checkCmd = &cobra.Command{
	Use:     "check",
	Short:   "Validate a data file against a schema file",
	Example: "check ./schema.json ./db.json",
	Run:     func(cmd *cobra.Command, args []string) {},
}

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "Initialize a new schema file",
	Example: "init ./schema.json",
	Run: func(cmd *cobra.Command, args []string) {
		schemaPath := "./schema.json"
		if len(args) != 0 && args[0] != "" {
			schemaPath = args[0]
		}

		file, err := os.Create(schemaPath)
		if err != nil {
			ErrExit("Couldn't create the schema file", err)
		}

		encoder := json.NewEncoder(file)
		encoder.Encode([]Entity{})
	},
}

var rootCmd = &cobra.Command{
	Use:   "serveur",
	Short: "A mock server with auto-generated data.",
	Long: `Serveur is a mock server for testing purposes.
Provided with a schema file, Serveur will generate fake data and spin-up a http server.

Each entity will be served at a different endpoint:

- GET ` + "`/entityName`" + ` (all)
- POST ` + "`/entityName`" + `

- GET ` + "`/entityName/:id`" + `
- PUT ` + "`/entityName/:id`" + `
- Patch ` + "`/entityName/:id`" + `
- DELETE ` + "`/entityName/:id`" + `

The server will also provide a home page at ` + "`/`" + ` where you can test the endpoints.
Something in the lines of Swagger UI.

You can also provide a static directory to serve static files.

The schema file can be provided as a local file or a url.
It should be a JSON with the following structure:

{
  "entity1": {
    "count": 10, // number of records to generate
    "fileds": {
      "field1": "<type>",
      "field2": "<type>",
      ...
	}
  },
  "entity2": {
	...
  },
  ...
}

A field can be one of these types:

- string/str
- number/num
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
- addr
- phone
- paragraph/pg
- ref
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

		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			ErrExit("Couldn't get the port flag", err)
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
		path, watcher := initFile(schemaPath)
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
			Addr:    fmt.Sprintf(":%d", port),
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
				srv.Close()
				srv.Handler = server.mux

				go func() {
					log.Fatal(srv.ListenAndServe())
				}()
			}
		}
	},
}
