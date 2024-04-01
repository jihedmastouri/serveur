package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"slices"

	"github.com/fsnotify/fsnotify"
)

type Field struct {
	Name    string         `json:"name"`
	Kind    FieldType      `json:"type"`
	Options map[string]any `json:"options"`
}

type Entity struct {
	Name   string  `json:"name"`
	Count  int     `json:"count"`
	Schema []Field `json:"schema"`
}

type FieldType string

const (
	StringType    FieldType = "string"
	NumberType              = "number"
	BooleanType             = "bool"
	NameType                = "name"
	UsernameType            = "username"
	FullnameType            = "fullname"
	EmailType               = "email"
	DateType                = "date"
	UrlType                 = "url"
	IpType                  = "ip"
	UuidType                = "uuid"
	IdType                  = "id"
	AddressType             = "address,addr"
	PhoneType               = "phone"
	ParagraphType           = "paragraph,pg"
)

// Initializes a file watcher and returns the path to the file and the watcher
// If the path is a url, it downloads the file and returns the path to the downloaded file
func initFile(path string) (string, *fsnotify.Watcher) {
	downloadFile(path)
	watcher, _ := fsnotify.NewWatcher()
	watcher.Add(path)
	return path, watcher
}

func downloadFile(path string) {
	u, err := url.ParseRequestURI(path)
	if err == nil {
		resp, err := http.Get(u.String())
		if err != nil {
			ErrExit("Couldn't Download File", err)
		}
		defer resp.Body.Close()
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			ErrExit("Couldn't Read Remote File", err)
		}
		os.WriteFile(path, content, 0644)
	}
}

// Parses a json file and returns a slice of entities
func ParseFile(path string) ([]Entity, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var entities []Entity
	err = decoder.Decode(&entities)
	if err != nil {
		return nil, err
	}
	for i, entity := range entities {
		entities[i].Count = 1
		for j, field := range entity.Schema {
			if field.Kind == "" {
				entities[i].Schema[j].Kind = StringType
			}
			if reflect.TypeOf(field.Kind).Name() != "FieldType" {
				fmt.Println("Invalid Type: ", field.Kind)
			}
		}
	}
	return entities, nil
}

func ValidateSchema(entities []Entity, prevSchema []Entity) bool {
	if len(entities) != len(prevSchema) {
		return false
	}
	for _, entity := range entities {
		index := slices.IndexFunc(prevSchema, func(e Entity) bool {
			return e.Name == entity.Name
		})

		if index == -1 {
			return false
		}

		if !(reflect.DeepEqual(entity.Schema, prevSchema[index].Schema) &&
			entity.Count == prevSchema[index].Count) {
			return false
		}
	}

	return true
}
