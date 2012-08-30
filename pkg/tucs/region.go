package tucs

import (
	"fmt"
	"math"
	"strconv"
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
	key := fmt.Sprintf("%d_%s_%d", nidx, r.Type, pidx)
	if hash, ok := r.hashes[key]; ok {
		return hash
	}
	parent := r.Parent(r.Type, pidx)
	if parent != nil {
		r.hashes[key] = parent.Hash(nidx, pidx) + "_" + r.Name(nidx)
	} else {
		r.hashes[key] = r.Name(nidx)
	}
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
	hashstr := r.Hash(nidx, 0)
	hash := []string{}
	if strings.HasPrefix(hashstr, "TILECAL") {
		hash = strings.Split(hashstr, "_")[1:]
	} else {
		hash = strings.Split(hashstr, "_")
	}

	nbr := []int{}
	if len(hash) >= 1 {
		// get partition or side
		part := map[string]int{"LBA": 1, "LBC":2, "EBA":3, "EBC":4}
		nbr = append(nbr, part[hash[0]])
	}

	if len(hash) >= 2 {
		// get module
		mid, err := strconv.ParseInt(hash[1][1:], 10, 64)
		if err != nil {
			panic("tucs.Region.Number: "+err.Error())
		}
		nbr = append(nbr, int(mid))
	}

	if len(hash) >= 3 {
		// get channel or sample
		switch r.Type {
		case Physical:
			samp := map[string]int{"A": 0, "BC":1, "D":2, "E":3}
			nbr = append(nbr, samp[hash[2][1:]])

		default:
			chid, err := strconv.ParseInt(hash[2][1:], 10, 64)
			if err != nil {
				panic("tucs.Region.Number: "+err.Error())
			}
			nbr = append(nbr, int(chid))
		}
	}

	if len(hash) >= 4 {
		// get ADC
		switch r.Type {
		case Physical:
			if hash[3][0] == 't' {
				adcid, err := strconv.ParseInt(hash[3][1:], 10, 64)
				if err != nil {
					panic("tucs.Region.Number: "+err.Error())
				}
				nbr = append(nbr, int(adcid))
			} else if hash[3][:4] == "MBTS" {
				nbr = append(nbr, 15)
			}
		default:
			gain := map[string]int{"lowgain":0, "highgain":1}
			nbr = append(nbr, gain[hash[3]])
		}
	}
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
