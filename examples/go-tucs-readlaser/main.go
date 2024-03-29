package main

import (
	"log"

	"github.com/sbinet/go-tucs/tucs"
	wtucs "github.com/sbinet/go-tucs/tucs/workers"
	"github.com/sbinet/go-tucs/tucs/workers/laser"
)

func main() {
	const useMBTS = true
	const useSpecialEBmods = true

	app := tucs.NewApp(useMBTS, useSpecialEBmods)

	{
		cfg := tucs.FilterCfg{
			Runs:        []string{"-1 week"},
			Region:      "EBC_m62_c37_highgain",
			RunType:     "Las",
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
		cfg := wtucs.PrintCfg{
			PrintRunType: true,
			PrintRunNbr:  true,
			PrintTime:    true,
			PrintData:    true,
			PrintRegion:  true,
			//Verbose     : true,
			//Region : "some region",
			//Data      : "some data",
		}
		app.AddWorker(wtucs.Print(tucs.Readout, cfg))
	}

	{
		cfg := laser.ReadLaserCfg{
			WorkDir:  "/afs/cern.ch/user/t/tilecali/w0/ntuples/las",
			DiodeNbr: -1,
			BoxPar:   false,
			Verbose:  false,
		}
		app.AddWorker(laser.ReadLaser(tucs.Readout, cfg))
	}

	{
		cfg := wtucs.PrintCfg{
			PrintRunType: true,
			PrintRunNbr:  true,
			PrintTime:    true,
			PrintData:    true,
			PrintRegion:  true,
			//Verbose     : true,
			//Region : "some region",
			//Data      : "some data",
		}
		app.AddWorker(wtucs.Print(tucs.Readout, cfg))
	}

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}
