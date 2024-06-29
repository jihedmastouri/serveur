package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"

	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
)

// Returns a fake value for a given field
func GetFake(f Field) (any, error) {
	switch f.Kind {
	case NumberType:
		val, _ := faker.RandomInt(0, 100)
		return val, nil
	case BooleanType:
		if rand.Intn(10) >= 5 {
			return true, nil
		} else {
			return false, nil
		}
	case StringType:
		return faker.Paragraph(), nil
	case NameType:
		return faker.Name(), nil
	case UsernameType:
		return faker.Username(), nil
	case FullnameType:
		return faker.FirstName() + " " + faker.LastName(), nil
	case EmailType:
		return faker.Email(), nil
	case DateType:
		return faker.Date(), nil
	case UrlType:
		return faker.URL(), nil
	case IpType:
		return faker.IPv4(), nil
	case UuidType:
		return faker.UUIDHyphenated(), nil
	case IdType:
		return faker.UUIDDigit(), nil
	case AddressType:
		return faker.GetAddress(), nil
	case PhoneType:
		return faker.Phonenumber(), nil
	default:
		s := fmt.Sprintf("Unknown Field Type : %s", f.Kind)
		return nil, fmt.Errorf(s)
	}
}

// Generates fake data for a given schema
func GenerateFakeData(schema []Field) (map[string]any, error) {
	data := make(map[string]any)
	for _, f := range schema {
		val, err := GetFake(f)
		if err != nil {
			return nil, err
		}
		data[f.Name] = val
	}
	return data, nil
}

// Fills the database with fake data
func FillDatabase(entities []Entity, s Store) {
	w := sync.WaitGroup{}
	for _, e := range entities {
		w.Add(1)
		log.Println("Generating fake data for entity:", e.Name)
		go func(e Entity, w *sync.WaitGroup) {
			for i := 0; i < e.Count; i++ {
				m, err := GenerateFakeData(e.Schema)
				if err != nil {
					log.Println(err)
					return
				}

				if m["id"] == nil {
					m["id"] = faker.UUIDDigit()
				}
				id := m["id"].(string)

				b, err := json.Marshal(m)
				if err != nil {
					fmt.Println(err)
					return
				}

				err = s.Set(e.Name, id, b)
				if err != nil {
					log.Println(err)
					return
				}
			}
			w.Done()
		}(e, &w)
	}
	w.Wait()
	log.Println("Done!")
}

func getStringOptions(f Field) *options.Options {
	opts := options.Options{}
	return &opts
}
