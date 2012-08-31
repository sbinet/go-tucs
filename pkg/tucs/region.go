package tucs

import (
	"fmt"
	"io"
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

// Region is what a detector tree is made of.
//
// Each region has various attributes: its parent(s), its child(ren), and any
// event object associated with it.
// One can also call the Hash method to get a unique location for this region in
// detector geometry tree.
type Region struct {
	names    []string
	// the set of parent region(s) this region is attached to.
	parents  []*Region
	// the set of children regions this region is made of.
	children []*Region
	// hashes stores a unique identifier for each region.
	hashes   map[string]string
	// events is a list of Events associated with a particular region.
	events   []Event
	// the Type for any given region says if region is part of the
	// read-out electronics (partitions, modules, channels) or the physical
	// geometry (cells, towers).
	// In case of ambiguity, assume Readout.
	Type     RegionType
}

// NewRegion creates a new Region of type typ with primary name name
func NewRegion(typ RegionType, name string, names ...string) *Region {
	r := &Region{
		names:    []string{name},
		parents:  make([]*Region, 0),
		children: make([]*Region, 0),
		hashes:   make(map[string]string),
		events:   make([]Event, 0),
		Type:     typ,
	}
	r.names = append(r.names, names...)

	return r
}

func (r *Region) Contains(rhs *Region) bool {
	//TODO
	return false
}

func (r *Region) Hash(nidx, pidx uint) string {
	k := fmt.Sprintf("%d_%s_%d", nidx, r.Type, pidx)
	if hash, ok := r.hashes[k]; ok {
		return hash
	}
	parent := r.Parent(r.Type, pidx)
	if parent != nil {
		r.hashes[k] = parent.Hash(nidx, pidx) + "_" + r.Name(nidx)
	} else {
		r.hashes[k] = r.Name(nidx)
	}
	return r.hashes[k]
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

// SanityCheck checks whether the internal state of this Region is consistent
// with all the other Regions it is in relation (parents and children)
// It returns a non-nil error in case of inconsistency
func (r *Region) SanityCheck() error {
	var err error
	for _, p := range r.parents {
		found := false
		for _, child := range p.children {
			if child == r {
				found = true
				break
			}
		}
		if !found {
			err = fmt.Errorf("tucs.Region.SanityCheck: my parents disowned me")
			fmt.Printf("** %s\n", err.Error())
		}
	}
	
	for _, c := range r.children {
		found := false
		for _, parent := range c.parents {
			if parent == r {
				found = true
				break
			}
		}
		if !found {
			err = fmt.Errorf("tucs.Region.SanityCheck: my children disowned me")
			fmt.Printf("** %s\n", err.Error())
		}
	}
	return err
}

// Print dumps the tree structure of the Region into the out io.Writer
// If depth == -1, the whole tree is displayed
func (r *Region) Print(out io.Writer, depth int, nidx, pidx uint, rtype RegionType) {
	fmt.Fprint(out, r.Name(nidx), r.Hash(nidx, pidx))
	depth--

	if depth != 0 {
		for _, c := range r.Children(rtype) {
			c.Print(out, depth, nidx, pidx, rtype)
		}
	}
}

// RegionFct allows to apply an item of work on each sub-region of a given Region.
// See tucs.Region.IterRegion
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
		part := map[string]int{"LBA": 1, "LBC": 2, "EBA": 3, "EBC": 4}
		nbr = append(nbr, part[hash[0]])
	}

	if len(hash) >= 2 {
		// get module
		mid, err := strconv.ParseInt(hash[1][1:], 10, 64)
		if err != nil {
			panic(fmt.Sprintf("tucs.Region.Number: %v\n (hash: %v %v)",
				err.Error(), hash, hashstr))
		}
		nbr = append(nbr, int(mid))
	}

	if len(hash) >= 3 {
		// get channel or sample
		switch r.Type {
		case Physical:
			samp := map[string]int{"A": 0, "BC": 1, "D": 2, "E": 3}
			nbr = append(nbr, samp[hash[2][1:]])

		default:
			chid, err := strconv.ParseInt(hash[2][1:], 10, 64)
			if err != nil {
				panic("tucs.Region.Number: " + err.Error())
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
					panic("tucs.Region.Number: " + err.Error())
				}
				nbr = append(nbr, int(adcid))
			} else if hash[3][:4] == "MBTS" {
				nbr = append(nbr, 15)
			}
		default:
			gain := map[string]int{"lowgain": 0, "highgain": 1}
			nbr = append(nbr, gain[hash[3]])
		}
	}
	return nbr
}

func (r *Region) Channels(useSpecialEBmods bool) []int {
	if strings.Contains(r.Name(0), "MBTS") {
		return []int{0}
	} else if !strings.Contains(r.Name(0), "t") {
		fmt.Printf("**tucs.Region.Channels only meaningful for tower regions\n")
		return []int{}
	}
	type chan_t []int
	type chans_t []chan_t
	cell2chan := [][]chans_t{
		// LB
		{
			// A
			chans_t{
				{1, 4}, {5, 8}, {9, 10}, {15, 18}, {19, 20}, {23, 26}, {29, 32}, {35, 38}, {37, 36}, {45, 46},
			},
			// BC
			chans_t{
				{3, 2}, {7, 6}, {11, 12}, {17, 16}, {21, 22}, {27, 28}, {33, 34}, {39, 40}, {47, 42},
			},
			// D
			chans_t{
				{-1, 0}, {}, {13, 14}, {}, {25, 24}, {}, {41, 44},
			},
			// E
			chans_t{},
		},
		// EB
		{
			// A
			chans_t{
				{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {7, 6}, {11, 10}, {21, 20}, {32, 31}, {40, 41},
			},
			// BC
			chans_t{
				{}, {}, {}, {}, {}, {}, {}, {}, {}, {5, 4}, {9, 8}, {15, 14}, {23, 22}, {35, 30}, {36, 39},
			},
			// D
			chans_t{
				{}, {}, {}, {}, {}, {}, {}, {}, {3, 2}, {}, {17, 16}, {}, {37, 38},
			},
			// E
			chans_t{
				{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {13}, {12}, {}, {1}, {}, {0},
			},
		},
	}
	nbr := r.Number(0, 0)
	part := nbr[0]
	module := nbr[1]
	sample := nbr[2]
	tower := nbr[3]
	barrel := 0
	if part <= 2 {
		barrel = 0
	} else {
		barrel = 1
	}

	// special modules: EBA15 and EBC18
	if useSpecialEBmods && ((part == 3 && module == 15) || (part == 4 && module == 18)) {
		// fixit
		cell2chan = [][]chans_t{
			// LB
			{},
			// EB
			{
				// A
				chans_t{
					{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {7, 6}, {11, 10}, {21, 20}, {32, 31}, {40, 41},
				},

				// BC
				chans_t{
					{}, {}, {}, {}, {}, {}, {}, {}, {}, {5, 4}, {9, 8}, {15, 14}, {23, 22}, {35, 30}, {36, 39},
				},

				// D - D5 (or D10) merged with D4 (or D08)
				chans_t{
					{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {17, 16}, {}, {37, 38},
				},

				// E - E3, E4 -> chan 18, 19
				chans_t{
					{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {13}, {12}, {}, {19}, {}, {18},
				},
			},
		}
	}
	ch := cell2chan[barrel][sample][tower]
	// copy to prevent from pinning these huge slices...
	chans := make([]int, len(ch))
	copy(chans, ch)
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

// MBTSType checks if a module has MBTS connected to channel 0 and whether the
// crack scintillator is missing.
// It returns:
//  0 if no MBTS
//  1 if the MBTS is present but the crack missing
//  2 if the MBTS is present and the crack present
func (r *Region) MBTSType() int {
	nbr := r.Number(0, 0)
	if len(nbr) < 2 {
		panic("tucs.Region.MBTSType: this should only be called at the module level or lower")
	}
	part := nbr[0]
	module := nbr[1]

	switch part {
	case 3:
		if in_intslice(module,
			[]int{3, 12, 23, 30, 35, 44, 53, 60}) {
			return 1
		} else if in_intslice(module,
			[]int{4, 13, 24, 31, 36, 45, 54, 61}) {
			return 2
		} else {
			return 0
		}
	case 4:
		if in_intslice(module,
			[]int{4, 13, 20, 28, 37, 45, 54, 61}) {
			return 1
		} else if in_intslice(module,
			[]int{5, 12, 19, 27, 36, 44, 55, 62}) {
			return 2
		} else {
			return 0
		}
	default:
		return 0
	}
	panic("unreachable")
}

// MBTSName returns a stub name consistent with L1 trigger name
func (r *Region) MBTSName() string {
	nbr := r.Number(0, 0)
	if len(nbr) < 2 {
		panic("tucs.Region.MBTSName: this should only be called at the module level or lower")
	}
	part := nbr[0]
	module := nbr[1]
	name := []string{}
	switch part {
	case 3:
		name = append(name, "A")
		idx := idx_intslice(module, []int{4, 13, 24, 31, 36, 44, 53, 61, 03, 12, 23, 30, 35, 45, 54, 60})
		if idx >= 0 {
			// FIXME: should the format be %02d instead ?
			name = append(name, fmt.Sprintf("%d", idx))
		} else {
			panic("tucs.Region.MBTSName: invalid index or module")
		}

	case 4:
		name = append(name, "C")
		idx := idx_intslice(module, []int{5, 13, 20, 28, 37, 45, 55, 62, 04, 12, 19, 27, 36, 44, 54, 61})
		if idx >= 0 {
			// FIXME: should the format be %02d instead ?
			name = append(name, fmt.Sprintf("%d", idx))
		} else {
			panic("tucs.Region.MBTSName: invalid index or module")
		}

	}
	return strings.Join(name, "")
}

// CrackPartner returns the module name of the module partner with which that
// region shares the crack scintillator.
func (r *Region) CrackPartner() string {
	nbr := r.Number(0, 0)
	if len(nbr) < 2 {
		panic("tucs.Region.CrackPartner: this should only be called at the module level or lower")
	}
	part := nbr[0]
	module := nbr[1]
	name := ""

	pairs := [][]int{}
	switch part {
	case 3:
		pairs = [][]int{
			{3, 4}, {12, 13}, {23, 24}, {30, 31}, {35, 36}, {44, 45}, {53, 54}, {60, 61},
		}

	case 4:
		pairs = [][]int{
			{4, 5}, {13, 12}, {20, 19}, {28, 27}, {37, 36}, {45, 44}, {54, 55}, {61, 62},
		}

	default:
	}

	for _, v := range pairs {
		if module == v[0] {
			name = fmt.Sprintf("m%02d", v[1])
			break
		} else if module == v[1] {
			name = fmt.Sprintf("m%02d", v[0])
			break
		}
	}
	return name
}

// EOF
