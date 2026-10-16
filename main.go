package main

import (
	"log"

	"github.com/kirick13/newborn/ui"
)

func main() {
	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}
}
