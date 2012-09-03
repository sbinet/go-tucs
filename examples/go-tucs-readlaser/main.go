package main

import (
	"log"

	"github.com/sbinet/go-tucs/pkg/tucs"
	"github.com/sbinet/go-tucs/pkg/tucs/workers"
)

func main() {
	const useMBTS = true
	const useSpecialEBmods = true

	app := tucs.NewApp(useMBTS, useSpecialEBmods)

	{
		cfg := tucs.FilterCfg{
			Runs: []string{"-1 week",},
			Type: tucs.Readout,
			Region: "EBC_m62_c37_highgain",
			RunType: "Las",
		}
		app.AddWorker(tucs.NewFilter(cfg))
	}
	app.AddWorker(wtucs.Print(tucs.Readout, "", "", wtucs.PrintCfg{}))

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// EOF
