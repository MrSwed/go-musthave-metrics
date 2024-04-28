package pkg1

import (
	"os"
)

func errCheckFunc() {
	os.Exit(0)       // want "os.Exit is not allowed to use"
	go os.Exit(0)    // want "os.Exit is not allowed to use"
	defer os.Exit(0) // want "os.Exit is not allowed to use"
}
