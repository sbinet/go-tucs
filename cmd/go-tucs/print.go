package main

import (
	"fmt"

	"github.com/sbinet/go-tucs/pkg/tucs"
)

type printWorker struct {
	rtype    tucs.RegionType
	nregions int
}

func PrintWorker(rtype tucs.RegionType) tucs.Worker {
	w := &printWorker{
		rtype:    rtype,
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
	fmt.Printf("  processed [%d] region(s)\n", w.nregions)
	fmt.Printf("::worker-stop... [done]\n")
	return nil
}

func (w *printWorker) ProcessRegion(region *tucs.Region) error {
	//fmt.Printf("::process-region [%s]...\n", region.Name(0))
	w.nregions += 1
	return nil
}

func (w *printWorker) RegionType() tucs.RegionType {
	return w.rtype
}

// EOF
