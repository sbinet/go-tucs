package wtucs

import (
	"fmt"
	"strings"

	"github.com/sbinet/go-tucs/pkg/tucs"
)

type PrintCfg struct {
	PrintRunType bool   // enable printing of the run-type infos
	PrintRunNbr  bool   // enable printing of the run-nbr infos
	PrintTime    bool   // enable printing of the time
	PrintData    bool   // enable printing of data infos
	PrintRegion  bool   // enable printing of region hash
	Verbose      bool   // enable verbose output
	Region       string // the region to print infos for
	Data         string // the data to print infos for
}

type printWorker struct {
	tucs.Base
	cfg PrintCfg
}

func Print(rtype tucs.RegionType, cfg PrintCfg) tucs.Worker {
	w := &printWorker{
		Base: tucs.NewBase(rtype),
		cfg:  cfg,
	}
	return w
}

func (w *printWorker) ProcessRegion(region *tucs.Region) error {
	var err error = nil

	if w.cfg.Region != "" && !strings.Contains(region.Hash(0, 0), w.cfg.Region) {
		return err
	}

	if len(region.Events()) == 0 {
		return err
	}

	printf := fmt.Printf
	for i, _ := range region.Events() {
		evt := &region.Events()[i]
		if w.cfg.PrintRunType {
			printf("%v, ", evt.Run.Type)
		}
		if w.cfg.PrintRunNbr {
			printf("%v, ", evt.Run.Number)
		}
		if w.cfg.PrintTime {
			printf("%v, ", evt.Run.Time)
		}
		if w.cfg.Verbose {
			printf("%v\n", evt.Data)
		} else {
			if w.cfg.PrintData {
				for k, v := range evt.Data {
					if w.cfg.Data != "" {
						if k == w.cfg.Data {
							printf("%v: %v, ", k, v)
						}
					} else {
						printf("%v: %v, ", k, v)
					}
				}
				printf("\n")
			}
		}
	}
	return err
}

// check printWorker satisfies the tucs.Worker interface
var _ tucs.Worker = (*printWorker)(nil)

// EOF
