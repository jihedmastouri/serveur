package main

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"os"
	"strconv"
	"testing"
)

const testEntity = "testEntity"

var dummyData []map[string]any = []map[string]any{
	{"key": "key1", "value": "value1"},
	{"key": "key2", "value": map[string]any{"foo": []string{"bar", "baz"}}},
	{"key": "key3", "value": []int{1, 3, 4}},
}

func TestDBGet(t *testing.T) {
	testFunc := func(t *testing.T, inMemory bool) {
		db, path := setupDB(t, inMemory)
		defer cleanupDB(db, path)

		for _, data := range dummyData {
			b, err := db.Get(testEntity, data["key"].(string))
			if err != nil {
				t.Fatal("Failed to get data: ", err)
			}

			val, err := json.Marshal(data["value"])
			if err != nil {
				t.Fatal("Failed marshel value for comparison: ", err)
			}

			if !bytes.Equal(b, val) {
				t.Fatalf("Wrong Data")
			}
		}
	}

	t.Run("Get inMemory", func(t *testing.T) {
		testFunc(t, true)
	})
	t.Run("Get", func(t *testing.T) {
		testFunc(t, false)
	})
}

// TODO: all and with validators and end functions
func TestDBGetAll(t *testing.T) {}

func TestDBSet(t *testing.T) {
	testFunc := func(t *testing.T, db *DB) {
		// SET
		for _, data := range dummyData {
			val, err := json.Marshal(data["value"])
			if err != nil {
				t.Fatal("Failed to marshel value before set")
			}
			db.Set(testEntity, data["key"].(string), val)
		}

		// GET
		for _, data := range dummyData {
			b, err := db.Get(testEntity, data["key"].(string))
			if err != nil {
				t.Fatal("Failed to get data: ", err)
			}

			val, err := json.Marshal(data["value"])
			if err != nil {
				t.Fatal("Failed marshel value for comparison: ", err)
			}

			if !bytes.Equal(b, val) {
				t.Fatalf("Wrong Data")
			}
		}
	}

	t.Run("SET inMemory", func(t *testing.T) {
		db := NewDB(true, "set_test_inmemory.db")
		defer cleanupDB(db, "")
		testFunc(t, db)
	})

	t.Run("SET", func(t *testing.T) {
		path := "set_test.db"
		db := NewDB(false, path)
		defer cleanupDB(db, path)
		testFunc(t, db)
	})

	t.Run("PUT inMemory", func(t *testing.T) {
		db, _ := setupDB(t, true)
		defer cleanupDB(db, "")
		testFunc(t, db)
	})

	t.Run("PUT", func(t *testing.T) {
		db, path := setupDB(t, false)
		defer cleanupDB(db, path)
		testFunc(t, db)
	})

}

// TODO: like set
func TestDBPatch(t *testing.T) {}

func TestDBDelete(t *testing.T) {
	testFunc := func(t *testing.T, inMemory bool) {
		db, path := setupDB(t, inMemory)
		defer cleanupDB(db, path)

		for _, data := range dummyData {
			err := db.Delete(testEntity, data["key"].(string))
			if err != nil {
				t.Fatal("Failed to Delete element: ", err)
			}

			_, err = db.Get(testEntity, data["key"].(string))
			if err == nil {
				t.Fatal("Element was not deleted: ", data["key"])
			}
		}
	}

	t.Run("DEL inMemory", func(t *testing.T) {
		testFunc(t, true)
	})

	t.Run("DEL", func(t *testing.T) {
		testFunc(t, false)
	})
}

func generateDBName() string {
	// Generate A random suffix
	perm := rand.Perm(10)

	dbName := "test"
	for _, num := range perm {
		dbName += strconv.Itoa(num)
	}
	return dbName + ".db"
}

func setupDB(t *testing.T, inMemory bool) (*DB, string) {
	dbName := generateDBName()
	db := NewDB(inMemory, dbName)
	txn := db.db.NewTransaction(true)
	defer txn.Discard()

	for _, data := range dummyData {
		value, err := json.Marshal(data["value"])
		if err != nil {
			t.Error("Setup: Failed to marshal data ", err)
			return nil, ""
		}

		key := "testEntity" + "-" + data["key"].(string)

		err = txn.Set([]byte(key), []byte(value))
		if err != nil {
			t.Error("Setup: Failed to load data ", err)
			return nil, ""
		}
	}

	if err := txn.Commit(); err != nil {
		t.Error("Setup: Failed to commit data ", err)
		return nil, ""
	}

	txn.Commit()
	return db, dbName
}

func cleanupDB(db *DB, path string) {
	db.Drop()
	db.Close()
	if len(path) > 0 {
		os.RemoveAll(path)
	}
}
