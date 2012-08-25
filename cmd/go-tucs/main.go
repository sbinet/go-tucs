package main

import (
	"log"

	"github.com/sbinet/go-tucs/pkg/tucs"
)

func main() {
	app := tucs.NewApp(true, true)

	app.AddWorker(
		PrintWorker(tucs.Readout),
	)

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// EOF
