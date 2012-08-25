package main

import (
	"fmt"

	"github.com/sbinet/go-tucs/pkg/tucs"
)

type PrintWorker struct {
	rtype    tucs.RegionType
	nregions int
}

func NewPrintWorker(rtype tucs.RegionType) *PrintWorker {
	w := &PrintWorker{
		rtype:    rtype,
		nregions: 0,
	}
	return w
}

func (w *PrintWorker) ProcessStart() error {
	fmt.Printf("::worker-start...\n")
	fmt.Printf("::worker-start...[done]\n")
	return nil
}

func (w *PrintWorker) ProcessStop() error {
	fmt.Printf("::worker-stop...\n")
	fmt.Printf("  processed [%d] region(s)\n", w.nregions)
	fmt.Printf("::worker-stop... [done]\n")
	return nil
}

func (w *PrintWorker) ProcessRegion(region *tucs.Region) error {
	//fmt.Printf("::process-region [%s]...\n", region.Name(0))
	w.nregions += 1
	return nil
}

func (w *PrintWorker) RegionType() tucs.RegionType {
	return w.rtype
}

// EOF
