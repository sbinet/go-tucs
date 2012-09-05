package laser

import (
	"github.com/sbinet/go-tucs/pkg/tucs"
)

// readlaser is a tucs.Worker to read-laser data
type readlaser struct {
	tucs.CalibBase
}

type ReadLaserCfg struct {
	WorkDir string // name of the directory holding data
}

// ReadLaser returns a read-laser worker
func ReadLaser(rtype tucs.RegionType, cfg ReadLaserCfg) tucs.Worker {
	w := &readlaser{
		CalibBase: tucs.NewCalibBase(rtype, cfg.WorkDir),
	}

	return w
}
// check readlaser implements the tucs.Worker interface
var _ tucs.Worker = (*readlaser)(nil)

// EOF
