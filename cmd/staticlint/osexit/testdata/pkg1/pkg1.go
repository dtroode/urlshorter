package main

import (
	"os"
)

func main() {
	os.Exit(1) // want "os call detected"
}

func run() {
	os.Exit(1)
}
