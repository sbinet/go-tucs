package laser

import (
	"github.com/sbinet/go-tucs/pkg/tucs"
)

// readlaser is a tucs.Worker to read-laser data
type readlaser struct {
	tucs.CalibBase
	nevtcut int
	diode   int
	boxpar  bool
	runmap  map[int64]interface{}
	runs    []tucs.Run
	verbose bool
}

type ReadLaserCfg struct {
	WorkDir  string // name of the directory holding data
	DiodeNbr int
	BoxPar   bool
	Verbose  bool
}

// ReadLaser returns a read-laser worker
func ReadLaser(rtype tucs.RegionType, cfg ReadLaserCfg) tucs.Worker {
	w := &readlaser{
		CalibBase: tucs.NewCalibBase(rtype, cfg.WorkDir),
		nevtcut:   10,
		diode:     cfg.DiodeNbr,
		boxpar:    cfg.BoxPar,
		runmap:    make(map[int64]interface{}),
		runs:      make([]tucs.Run, 0),
		verbose:   cfg.Verbose,
	}

	return w
}

func (w *readlaser) ProcessStart() error {
	return nil
}

// check readlaser implements the tucs.Worker interface
var _ tucs.Worker = (*readlaser)(nil)

// EOF
