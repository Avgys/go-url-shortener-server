package pkg1

import (
	"log"
	"os"
)

func mulfunc(i int) (int, error) {
	return i * 2, nil
}

func errCheckFunc() {
	panic(5)     // want "panic called here"
	panic("")    // want "panic called here"
	os.Exit(2)   // want "os.Exit found outside main function"
	log.Fatal(2) // want "log.Fatal found outside main function"
}
