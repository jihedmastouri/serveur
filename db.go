package main

import (
	"log"

	badger "github.com/dgraph-io/badger/v4"
)

type Store interface {
	GetAll(entityname string, validator Validtor) ([][]byte, error)
	Get(entityname string, key []byte) ([]byte, error)
	Set(entityname string, key []byte, value []byte) error
	Delete(entityname string, key []byte) error
	Patch(entityname string, key []byte, value []byte) error
}

type DB struct {
	db *badger.DB
}

type Validtor struct {
	validate  []func([]byte) bool
	terminate []func([][]byte) bool
}

func NewDB(isInMemory bool) *DB {
	opt := badger.DefaultOptions("./db")
	if isInMemory {
		opt = opt.WithInMemory(true)
	}
	db, err := badger.Open(opt)
	if err != nil {
		log.Fatal("Couldn't Open Database")
	}
	return &DB{db}
}

func (db *DB) Close() {
	db.db.Close()
}

func (db *DB) GetAll(entityname string, valid Validtor) ([][]byte, error) {
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
				for _, f := range valid.validate {
					if !f(v) {
						return nil
					}
				}
				result = append(result, v)
				return nil
			})
			if err != nil {
				return err
			}
			for _, f := range valid.terminate {
				if f(result) {
					break seeker
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

func (db *DB) Get(entityname string, key []byte) ([]byte, error) {
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

func (db *DB) Set(entityname string, key []byte, value []byte) error {
	err := db.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(append([]byte(entityname+"-"), key...), value)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) Delete(entityname string, key []byte) error {
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

func (db *DB) Patch(entityname string, key []byte, valye []byte) error {
	err := db.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(append([]byte(entityname+"-"), key...))
		if err != nil {
			return err
		}
		var value []byte
		err = item.Value(func(v []byte) error {
			value = v
			return nil
		})
		if err != nil {
			return err
		}
		value = append(value, valye...)
		err = txn.Set(append([]byte(entityname+"-"), key...), value)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}
