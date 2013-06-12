package tucs

import (
	"fmt"

	"github.com/go-hep/croot"
)

// Base implements a basic tucs.Worker
type Base struct {
	HistFile croot.File
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
	var err error
	const compress = 1
	const netopt = 0

	hfile := croot.GRoot.GetFile(fname)
	if hfile == nil {
		hfile, err = croot.OpenFile(fname, "recreate", "TUCS histogram", compress, netopt)
	}
	if err != nil {
		return err
	}
	if hfile == nil {
		return fmt.Errorf("tucs.Base: could not open file [%s]", fname)
	}
	b.HistFile = hfile
	if !b.HistFile.Cd("") {
		return fmt.Errorf("tucs.Base: could not make [%s] the current directory", fname)
	}
	return err
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

// EOF
