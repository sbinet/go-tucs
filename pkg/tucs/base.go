package tucs

import (
	"fmt"

	"github.com/sbinet/go-croot/pkg/croot"
)

// Base implements a basic tucs.Worker
type Base struct {
	HistFile *croot.File
	rtype    RegionType
}

// NewBase creates a new *Base worker reading for embedding
func NewBase(rtype RegionType) Base {
	return Base{
		HistFile: nil,
		rtype: rtype,
	}
}

// InitHistFile grabs the ROOT file 'fname' and makes it the current gDirectory
// FIXME: wrap croot.GRoot
func (b *Base) InitHistFile(fname string) error {
	var err error
	const compress = 1
	const netopt = 0

	hfile := croot.OpenFile(fname, "recreate", "TUCS histogram", compress, netopt)
	if hfile == nil {
		return fmt.Errorf("tucs.Base: could not open file [%s]", fname)
	}
	b.HistFile = hfile
	return err
}

func (b *Base) ProcessStart() error {
	fmt.Printf("--process-start--\n")
	return nil
}

func (b *Base) ProcessStop() error {
	return nil
}

func (b *Base) ProcessRegion(region *Region) error {
	return nil
}

func (b *Base) RegionType() RegionType {
	return b.rtype
}

// checks Base implements tucs.Worker
var _ Worker = (*Base)(nil)

// EOF