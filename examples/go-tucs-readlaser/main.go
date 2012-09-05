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
			Runs:    []string{"-1 week"},
			Region:  "EBC_m62_c37_highgain",
			RunType: "Las",
			UseDateProg: true,
			//Verbose: false,
			KeepOnlyActive: true,
			//Filter: "",
			Amp: 23000.,
			// GetLast: false,
			UpdateSpecial: true,
			// AllowC10Errors: false,
			// CsComment: "",
			// TwoInput: false,
		}
		app.AddWorker(tucs.NewFilter(tucs.Readout, cfg))
	}
	{
		cfg := wtucs.PrintCfg{}
		app.AddWorker(wtucs.Print(tucs.Readout, cfg))
	}

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// EOF
