package main

import (
	"flag"
	"log"
	"time"

	"github.com/pico-db/pico/db"
)

// Test out the APIs for documents
func main() {
	which := flag.Int("which", 1, "The example to run")
	flag.Parse()
	switch *which {
	case 1:
		setSimpleDocument()
	case 2:
		marshalDocument()
	case 3:
		unmarshalDocument()
	case 4:
		toBytes()
	default:
		log.Println("test option not found")
	}
}

func setSimpleDocument() {
	doc := db.NewDocument()
	doc.Set("example", "1")
	js, err := doc.Json()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(js)
}

type Person struct {
	Name        string    `pson:"name"`
	DateOfBirth time.Time `pson:"dob"`
	Age         int       `pson:"age"`
}

func marshalDocument() {
	p := Person{
		Name:        "Nguyen Tran",
		DateOfBirth: time.Now(),
		Age:         21,
	}
	doc := db.Document{}
	err := doc.Marshal(&p)
	if err != nil {
		log.Fatalln(err)
	}
	js, err := doc.Json()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(js)
}

func unmarshalDocument() {
	p := Person{
		Name:        "Nguyen Tran",
		DateOfBirth: time.Now(),
		Age:         21,
	}
	doc := db.Document{}
	err := doc.Marshal(&p)
	if err != nil {
		log.Fatalln(err)
	}
	emptyp := Person{}
	err = doc.Unmarshal(&emptyp)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(emptyp)
}

func toBytes() {
	doc := db.NewDocument()
	doc.Set("example", "1")
	bytes, err := doc.Encode()
	if err != nil {
		log.Fatalln(err)
	}
	emptydoc := db.Document{}
	err = emptydoc.Decode(bytes)
	if err != nil {
		log.Fatalln(err)
	}
	js, err := emptydoc.Json()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(js)
}
