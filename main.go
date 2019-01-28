package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

func main() {

	terms = make(map[string][]Term, 0)
	holidays = make([]KeyDate, 0)
	keyDates = make([]KeyDate, 0)

	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("ERROR: %s\n", e)
		}
	}()

	load := func(filename string, data interface{}) {
		raw, _ := ioutil.ReadFile(filename)
		if err := json.Unmarshal(raw, &data); err != nil {
			panic(err)
		}
	}

	load(_terms, &terms)
	load(_holidays, &holidays)
	load(_keyDates, &keyDates)

	port := os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(_port)
	}

	fmt.Printf("Starting server on port %s...\n", port)
	p, _ := strconv.Atoi(port)
	router(p)
}
