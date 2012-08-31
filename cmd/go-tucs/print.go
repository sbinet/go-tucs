package main

import (
	"fmt"

	"github.com/sbinet/go-tucs/pkg/tucs"
)

type printWorker struct {
	tucs.Base
	nregions int
}

func PrintWorker(rtype tucs.RegionType) tucs.Worker {
	w := &printWorker{
		Base:    tucs.NewBase(rtype),
		nregions: 0,
	}
	return w
}

func (w *printWorker) ProcessStart() error {
	fmt.Printf("::worker-start...\n")
	fmt.Printf("::worker-start...[done]\n")
	return nil
}

func (w *printWorker) ProcessStop() error {
	fmt.Printf("::worker-stop...\n")
	fmt.Printf("  processed [%d] region(s) of type [%s]\n", 
		w.nregions, w.RegionType())
	fmt.Printf("::worker-stop... [done]\n")
	return nil
}

func (w *printWorker) ProcessRegion(region *tucs.Region) error {
	//fmt.Printf("::process-region [%s]...\n", region.Name(0))
	w.nregions += 1
	return nil
}

// check printWorker satisfies the tucs.Worker interface
var _ tucs.Worker = (*printWorker)(nil)

// EOF
