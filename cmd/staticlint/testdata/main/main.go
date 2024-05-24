package main

import (
	"os"
)

func errCheckFunc2() {
	os.Exit(0)
	go os.Exit(0)
	defer os.Exit(0)
}

func main() {
	os.Exit(0)       // want "os.Exit is not allowed at main func of main package"
	go os.Exit(0)    // want "os.Exit is not allowed at main func of main package"
	defer os.Exit(0) // want "os.Exit is not allowed at main func of main package"
}
