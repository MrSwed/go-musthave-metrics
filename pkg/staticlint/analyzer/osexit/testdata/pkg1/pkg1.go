package pkg1

import (
	"os"
)

func errCheckFunc() {
	os.Exit(0)
	go os.Exit(0)
	defer os.Exit(0)
}

func main() {
	os.Exit(0)
	go os.Exit(0)
	defer os.Exit(0)
}
