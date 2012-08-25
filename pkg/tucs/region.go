package tucs

import (
	"fmt"
	"math"
	"strings"
)

type RegionType uint

const (
	Readout RegionType = iota
	Physical
	B175
	TestBeam
)

func (rt RegionType) String() string {
	switch rt {
	case Readout:
		return "readout"
	case Physical:
		return "physical"
	case B175:
		return "b175"
	case TestBeam:
		return "testbeam"
	}
	return "<unknown>"
}

type Region struct {
	names    []string
	parents  []*Region
	children []*Region
	hashes   map[string]string
	events   []Event
	Type     RegionType
}

func NewRegion(typ RegionType, name string, names ...string) *Region {
	r := &Region{
		names: []string{name},
		Type:  typ,
	}
	r.names = append(r.names, names...)

	return r
}

func (r *Region) Contains(rhs *Region) bool {
	//TODO
	return false
}

func (r *Region) Hash(nidx, pidx uint) string {
	//TODO
	key := fmt.Sprintf("%d_%s_%d", nidx, r.Type, pidx)
	return r.hashes[key]
}

func (r *Region) Name(idx uint) string {
	return r.names[idx]
}

func (r *Region) Children(regtype RegionType) []*Region {
	children := make([]*Region, 0)

	set := make(map[*Region]struct{}, 0)

	found := false
	for _, region := range r.children {
		if region.Type == regtype {
			if _, ok := set[region]; !ok {
				set[region] = struct{}{}
				children = append(children, region)
				found = true
			}
		}
	}

	if !found {
		children = r.children
	}
	return children
}

func (r *Region) SetChildren(children []*Region) {
	set := make(map[*Region]struct{})
	for _, p := range r.children {
		set[p] = struct{}{}
	}

	for _, p := range children {
		if _, ok := set[p]; !ok {
			set[p] = struct{}{}
			r.children = append(r.children, p)
		}
	}
}

func (r *Region) Parent(regtype RegionType, idx uint) *Region {
	if len(r.parents) == 0 {
		return nil
	}

	for _, p := range r.parents {
		if p.Type == regtype {
			return p
		}
	}

	return r.parents[idx]
}

func (r *Region) SetParent(parents ...*Region) {
	r.parents = append(r.parents, parents...)
}

type RegionFct func(t RegionType, region *Region) error

func (r *Region) IterRegions(t RegionType, fct RegionFct) error {
	children := r.Children(t)
	for _, child := range children {
		err := child.IterRegions(t, fct)
		if err != nil {
			return err
		}
	}
	return fct(t, r)
}

/*
func (r *Region) Regions(t RegionType) chan *Region {
	ch := make(chan *Region)

	go func() {
		children := r.Children(t)
		for _, child := range children {
			regions := child.Regions(t)
			for rchild := range regions {
				ch <- rchild
			}
		}
		ch <- r
	}()
	return ch
}
*/

func (r *Region) Number(nidx, pidx uint) []int {
	//TODO. or return 4 ints ? part, module, sample, tower
	nbr := []int{}
	return nbr
}

func (r *Region) Channels(useSpecialEBmods bool) []int {
	//TODO.
	chans := []int{}
	return chans
}

func (r *Region) EtaPhi() (eta, phi float64, err error) {
	if !strings.Contains(r.Hash(0, 0), "_t") {
		return eta, phi, fmt.Errorf("no eta/phi cell position")
	}

	nbr := r.Number(0, 0)
	part := nbr[0]
	module := float64(nbr[1])
	sample := nbr[2]
	tower := float64(nbr[3])

	if module < 33 {
		phi = (module - 0.5) / 32.0 * math.Pi
	} else {
		phi = (module - 64.5) / 32.0 * math.Pi
	}

	if sample < 2 {
		eta = tower*0.1 + 0.05
	} else if sample < 3 {
		eta = tower * 0.1
	} else {
		eta = 1.0
	}

	if part == 2 || part == 4 {
		eta *= -1.0
	}
	return
}

func (r *Region) MBTSType() int {
	//TODO
	return 0
}

func (r *Region) MBTSName() string {
	//TODO
	return "<MBTSName>"
}

func (r *Region) CrackPartner() string {
	//TODO
	return "<CrackPartner>"
}

// EOF
