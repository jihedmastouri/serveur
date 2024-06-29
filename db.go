package main

import (
	"encoding/json"
	"log"

	badger "github.com/dgraph-io/badger/v4"
)

type Store interface {
	GetAll(entityname string, validator *Validtor) ([][]byte, error)
	Get(entityname, key string) ([]byte, error)
	Set(entityname, key string, value []byte) error
	Delete(entityname, key string) error
	Patch(entityname, key string, value []byte) error
}

type DB struct {
	db *badger.DB
}

// Validtor is a struct that contains two slices of functions.
// validate functions are used to apply a filter on the result.
// terminate functions are used to stop the iteration over the data.
type Validtor struct {
	validate  []func([]byte) bool
	terminate []func([][]byte) bool
}

const privateSchema = "__schema__used"

func NewDB(isInMemory bool, dbPath string) *DB {
	opt := badger.DefaultOptions(dbPath)
	if isInMemory {
		opt = badger.DefaultOptions("").WithInMemory(true)
	}
	db, err := badger.Open(opt.WithLogger(nil))
	if err != nil {
		log.Fatal("Couldn't Open Database")
	}
	return &DB{db}
}

func (db *DB) Close() {
	db.db.Close()
}

func (db *DB) Clear() {
	db.db.DropAll()
}

func (db *DB) GetAll(entityname string, valid *Validtor) ([][]byte, error) {
	result := make([][]byte, 0)
	var err error
	prefix := []byte(entityname)
	db.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
	seeker:
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err = item.Value(func(v []byte) error {
				if valid != nil {
					for _, f := range valid.validate {
						if !f(v) {
							return nil
						}
					}
				}
				result = append(result, v)
				return nil
			})
			if err != nil {
				return err
			}
			if valid != nil {
				for _, f := range valid.terminate {
					if f(result) {
						break seeker
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *DB) Get(entityname string, key string) ([]byte, error) {
	var result []byte
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(append([]byte(entityname+"-"), key...))
		if err != nil {
			return err
		}
		err = item.Value(func(v []byte) error {
			result = v
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *DB) Set(entityname string, key string, value []byte) error {
	err := db.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(append([]byte(entityname+"-"), key...), value)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) Delete(entityname string, key string) error {
	err := db.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(append([]byte(entityname+"-"),
			key...))
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) Patch(entityname string, key string, value []byte) error {
	err := db.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(append([]byte(entityname+"-"), key...))
		if err != nil {
			return err
		}
		var val []byte
		err = item.Value(func(v []byte) error {
			val = v
			return nil
		})
		if err != nil {
			return err
		}
		value = append(value, value...)
		err = txn.Set(append([]byte(entityname+"-"), key...), value)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) getSchema() []Entity {
	var result []Entity
	db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(privateSchema))
		if err != nil {
			return nil
		}
		item.Value(func(v []byte) error {
			json.Unmarshal(v, &result)
			return nil
		})
		return nil
	})
	return result
}

func (db *DB) storeSchema(schema []Entity) error {
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return err
	}
	err = db.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(privateSchema), schemaBytes)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}
