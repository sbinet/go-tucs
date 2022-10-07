package main

import (
	"log"

	"github.com/sbinet/go-tucs/tucs"
)

func main() {
	const useMBTS = true
	const useSpecialEBmods = true

	app := tucs.NewApp(useMBTS, useSpecialEBmods)

	app.AddWorker(
		PrintWorker(tucs.Readout),
	)

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
