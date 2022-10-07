package tucs

import (
	"fmt"

	"go-hep.org/x/hep/groot"
)

// Base implements a basic tucs.Worker
type Base struct {
	HistFile *groot.File
	rtype    RegionType
}

// NewBase creates a new Base worker ready for embedding
func NewBase(rtype RegionType) Base {
	return Base{
		HistFile: nil,
		rtype:    rtype,
	}
}

// InitHistFile grabs the ROOT file 'fname' and makes it the current gDirectory
func (b *Base) InitHistFile(fname string) error {
	hfile, err := groot.Open(fname)
	if err != nil {
		hfile, err = groot.Create(fname)
		if err != nil {
			return fmt.Errorf("tucs: could not create TUCS ROOT file %q: %w", fname, err)
		}
	}
	if hfile == nil {
		return fmt.Errorf("tucs: could not open file %q", fname)
	}
	b.HistFile = hfile

	return nil
}

func (b *Base) ProcessStart() error {
	//fmt.Printf("--process-start--\n")
	return nil
}

func (b *Base) ProcessStop() error {
	//fmt.Printf("--process-stop--\n")
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
