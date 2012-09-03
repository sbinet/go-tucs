package wtucs

import (
	"github.com/sbinet/go-tucs/pkg/tucs"
)

type PrintCfg struct {
	RunType bool
	RunNbr bool
	Time bool
	Data bool
	Region bool
}

type printWorker struct {
	tucs.Base
}

func Print(rtype tucs.RegionType, region string, data string, cfg PrintCfg) tucs.Worker {
	w := &printWorker{
		Base: tucs.NewBase(rtype),
		
	}
	return w
}

// check printWorker satisfies the tucs.Worker interface
var _ tucs.Worker = (*printWorker)(nil)

// EOF
