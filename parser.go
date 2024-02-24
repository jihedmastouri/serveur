package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"

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

func initFile(path string, extension string) (string, *fsnotify.Watcher) {
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
		path = "./schema." + extension
		os.WriteFile(path, content, 0644)
	}
	watcher, _ := fsnotify.NewWatcher()
	watcher.Add(path)
	return path, watcher
}

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
	return entities, nil
}
