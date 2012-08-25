package main

import (
	"log"

	// 3rd party
	"github.com/sbinet/go-tucs/pkg/tucs"
)

func main() {
	app := tucs.NewApp(true, true)

	app.AddWorker(NewPrintWorker(tucs.Readout))

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// EOF
