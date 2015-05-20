package laser

import (
	"fmt"
	"path"

	"github.com/sbinet/go-tucs/tucs"
)

// readlaser is a tucs.Worker to read-laser data
type readlaser struct {
	tucs.CalibBase
	nevtcut int
	diode   int
	boxpar  bool
	runmap  map[int64][]tucs.Event
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
		runmap:    make(map[int64][]tucs.Event),
		runs:      make([]tucs.Run, 0),
		verbose:   cfg.Verbose,
	}

	return w
}

func (w *readlaser) ProcessStart() error {
	fmt.Printf("grun-list: %v\n", len(tucs.Runs))
	for _, run := range tucs.Runs {
		fname := path.Join(w.CalibBase.Dir(),
			fmt.Sprintf("tileCalibLAS_%v_Las.0.root", run.Number))
		if tucs.PathExists(fname) {
			run.Data["filename"] = path.Base(fname)
			w.runs = append(w.runs, run)
			w.runmap[run.Number] = nil
			fmt.Printf("file: %v\n", fname)
		} else {
			fmt.Printf("not yet processed, removing: %v\n", run)
			tucs.Runs.Remove(run)
		}
	}

	return nil
}

func (w *readlaser) ProcessRegion(region *tucs.Region) error {

	for _, evt := range region.Events() {
		if evt.Run.Type != "Las" {
			continue
		}
		if _, ok := evt.Run.Data["filename"]; ok {
			w.runmap[evt.Run.Number] = append(w.runmap[evt.Run.Number], evt)
		}
	}
	return nil
}

// check readlaser implements the tucs.Worker interface
var _ tucs.Worker = (*readlaser)(nil)

// EOF
