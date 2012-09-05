package wlaser

import (
	"github.com/sbinet/go-tucs/pkg/tucs"
)

// readlaser is a tucs.Worker to read-laser data
type readlaser struct {
	tucs.Base
}

type ReadLaserCfg struct {
	
}

// ReadLaser returns a read-laser worker
func ReadLaser(rtype tucs.RegionType, cfg ReadLaserCfg) tucs.Worker {
	w := &readlaser{
		
	}

	return w
}
// check readlaser implements the tucs.Worker interface
var _ tucs.Worker = (*readlaser)(nil)

// EOF
