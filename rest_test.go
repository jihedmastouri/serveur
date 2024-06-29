package main_test

import (
	"testing"
)

const url = "http://localhost:3000"

func TestRest(t *testing.T) {
}

/* FAKE STORE */

type fakeStore struct {
	data map[string]string
}

func (f *fakeStore) Get(key string) (string, error) {
	return f.data[key], nil
}

func (f *fakeStore) Set(key string, value string) error {
	f.data[key] = value
	return nil
}

func (f *fakeStore) Delete(key string) error {
	delete(f.data, key)
	return nil
}

func (f *fakeStore) GetAll() map[string]string {
	return f.data
}

func (f *fakeStore) Patch() map[string]string {
	return f.data[key] = {
		...
	}
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		data: make(map[string]string),
	}
}
