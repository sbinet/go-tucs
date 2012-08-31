package tucs

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Worker interface {
	ProcessStart() error
	ProcessStop() error
	ProcessRegion(region *Region) error

	RegionType() RegionType
}

type App struct {
	workers  []Worker
	detector *Region
}

func NewApp(useMBTS, useSpecialEBmods bool) *App {
	app := &App{
		workers:  []Worker{},
		detector: nil,
	}
	app.msg("Welcome to Go-TUCS (pid=%d). Building detector tree...\n", os.Getpid())
	app.detector = TileCal(useMBTS, useSpecialEBmods)
	app.msg("done.\n")
	return app
}

func (app *App) msg(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (app *App) Run() error {
	var err error

	msg := fmt.Printf
	for _, w := range app.workers {

		msg("running [%T]...\n", w)
		err = w.ProcessStart()
		if err != nil {
			return err
		}

		err = app.detector.IterRegions(
			w.RegionType(),
			func(t RegionType, region *Region) error {
				return w.ProcessRegion(region)
			})

		err = w.ProcessStop()
		if err != nil {
			return err
		}

	}
	return err
}

func (app *App) AddWorker(w ...Worker) {
	app.workers = append(app.workers, w...)
}

func TileCal(useMBTS, useSpecialEBmods bool) *Region {
	printf := fmt.Printf

	printf("Constructing TileCal detector tree:\n")
	if useMBTS {
		printf("\tMBTS mapping enabled\n")
	}
	if useSpecialEBmods {
		printf("\tSpecial mapping in EBA15 and EBC18 enabled\n")
	}

	// Level 1: tilecal and its partitions

	rtype := Readout
	tilecal := NewRegion(rtype, "TILECAL")

	// there are 4 partitions
	partitions := []*Region{
		NewRegion(rtype, "EBA"),
		NewRegion(rtype, "LBA"),
		NewRegion(rtype, "LBC"),
		NewRegion(rtype, "EBC"),
	}

	tilecal.SetChildren(partitions)
	for _, partition := range partitions {
		partition.SetParent(tilecal)
	}

	// Level 2: tell the partitions what they have as modules

	type RegionMap map[string]*Region
	type CrackDb map[string]RegionMap

	// holder for LBC's D0 channel
	chD0 := make(RegionMap)

	// holder for EB's cross module crack scintillators
	chCrack := make(CrackDb)
	chCrack["EBA"] = make(RegionMap)
	chCrack["EBC"] = make(RegionMap)

	for _, partition := range tilecal.Children(Readout) {
		// construct each of the 64 modules
		modules := make([]*Region, 0, 64)
		for i := 0; i < 64; i++ {
			n := fmt.Sprintf("m%02d", i+1)
			m := NewRegion(rtype, n)
			modules = append(modules, m)
		}

		// modules = append(modules, NewRegion(TestBeam, "m00"))
		// modules = append(modules, NewRegion(B175, "m65"))

		// make them children of partition
		partition.SetChildren(modules)
		for _, m := range modules { //partition.Children(Readout) {
			m.SetParent(partition)
		}

		//
		// The chan2pmt variable provides the mapping between channel number
		// and PMT number.  Negative values means the PMT doesn't exist. Use
		// the variable as follows: pmt_number = chan2pmt[channel_number]
		chan2pmt := []int{}
		EB := false

		pname := partition.Name(0)
		switch pname {
		case "EBA", "EBC":
			EB = true
			chan2pmt = []int{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
				13, 14, 15, 16, 17, 18, -19, -20, 21, 22, 23, 24,
				-27, -26, -25, -31, -32, -28, 33, 29, 30, -36, -35, 34,
				44, 38, 37, 43, 42, 41, -45, -39, -40, -48, -47, -46,
			}
		case "LBA", "LBC":
			EB = false
			chan2pmt = []int{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
				13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24,
				27, 26, 25, 30, 29, 28, -33, -32, 31, 36, 35, 34,
				39, 38, 37, 42, 41, 40, 45, -44, 43, 48, 47, 46,
			}
		}

		// EBA15 and EBC18 are special
		// https://twiki.cern.ch/twiki/bin/view/Atlas/SpecialModules#Module_Type_11
		// The PMT mapping is a little different since it's a physically
		// smaller drawer
		chan2pmtSpecial := []int{
			-1, -2, -3, -4, 5, 6, 7, 8, 9, 10, 11, 12,
			13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24,
			-27, -26, -25, -31, -32, -28, 33, 29, 30, -36, -35, 34,
			44, 38, 37, 43, 42, 41, -45, -39, -40, -48, -47, -46,
		}

		for _, module := range partition.Children(Readout) {
			mname := module.Name(0)
			channels := []*Region{}
			table := chan2pmtSpecial
			// for each module, create the PMTs/channels within it
			if useSpecialEBmods && ((pname == "EBA" && mname == "m15") ||
				(pname == "EBC" && mname == "m18")) {
				table = chan2pmtSpecial
			} else {
				table = chan2pmt
			}
			for x := 0; x < 48; x++ {
				if table[x] > 0 {
					n1 := fmt.Sprintf("c%02d", x)
					n2 := fmt.Sprintf("p%02d", table[x])
					channels = append(channels,
						NewRegion(Readout, n1, n2))
				}
			}
			module.SetChildren(channels)
			for _, channel := range module.Children(Readout) {
				chname := channel.Name(0)
				channel.SetParent(module)
				// save D0 channel for later matching with cell
				if pname == "LBC" && chname == "c00" {
					chD0[mname] = channel
				}
				// save E channels connected to neighboring module
				if EB && chname == "c01" && module.MBTSType() == 1 {
					chCrack[pname][mname] = channel
				}
			}

			// create both gains
			for _, channel := range module.Children(Readout) {
				gains := []*Region{
					NewRegion(rtype, "lowgain"),
					NewRegion(rtype, "highgain"),
				}
				channel.SetChildren(gains)
				for _, gain := range channel.Children(Readout) {
					gain.SetParent(channel)
				}
			} // channels
		} // modules
	} // partitions

	// For each module, create the cells ('physical' part of detector tree)
	for _, partition := range tilecal.Children(Readout) {
		for _, module := range partition.Children(Readout) {
			pname := partition.Name(0)
			mname := module.Name(0)
			samples := []*Region{}
			for _, x := range []string{"A", "BC", "D", "E"} {
				samples = append(samples, NewRegion(Physical, "s"+x))
			}
			module.SetChildren(samples)
			mbts := module.MBTSType()
			for _, sample := range module.Children(Physical) {
				sample.SetParent(module)
				towers := []*Region{}
				add_tower := func(ids ...int) {
					for _, id := range ids {
						towers = append(towers,
							NewRegion(Physical,
								fmt.Sprintf("t%02d", id)))
					}
				}
				sname := sample.Name(0)
				if strings.Contains(pname, "LB") {
					// towers in long barrel
					if strings.Contains(sname, "A") {
						for i := 0; i < 10; i++ {
							add_tower(i)
						}
					} else if strings.Contains(sname, "BC") {
						for i := 0; i < 9; i++ {
							add_tower(i)
						}
					} else if strings.Contains(sname, "D") {
						if "LBA" == partition.Name(0) {
							// Draw D0 cell on A-side
							add_tower(0, 2, 4, 6)
						} else {
							add_tower(2, 4, 6)
						}
					}
				} else {
					// towers in extended barrel
					if strings.Contains(sname, "A") {
						add_tower(11, 12, 13, 14, 15)
					} else if strings.Contains(sname, "BC") {
						add_tower(9, 10, 11, 12, 13, 14)
					} else if strings.Contains(sname, "D") {
						// special modules have D08 merged with D10
						if useSpecialEBmods && ((pname == "EBA" && mname == "m15") ||
							(pname == "EBC" && mname == "m18")) {
							add_tower(10, 12)
						} else {
							add_tower(8, 10, 12)
						}
					} else if strings.Contains(sname, "E") {
						if !useMBTS || mbts == 0 {
							add_tower(10, 11, 13, 15)
						} else if mbts == 1 {
							// MBTS pass through, crack scintillator missing
							add_tower(10, 11)
							towers = append(towers,
								NewRegion(Physical, "MBTS"+sample.MBTSName()),
							)
						} else if mbts == 2 {
							add_tower(10, 11, 13, 15)
							towers = append(towers,
								NewRegion(Physical, "MBTS"+sample.MBTSName()),
							)
						}
					}
				}
				sample.SetChildren(towers)
				for _, tower := range sample.Children(Physical) {
					tower.SetParent(sample)
					// connect readout channels to physical cells
					chans := []*Region{}
					chanNbrs := tower.Channels(useSpecialEBmods)
					if in_intslice(-1, chanNbrs) {
						// take care of D0 cell
						chans = append(chans, chD0[mname])
						for _, ch := range module.Children(Readout) {
							chnbr, err := strconv.ParseInt(ch.Name(0)[1:], 10, 64)
							if err == nil && chnbr == 0 {
								chans = append(chans, ch)
							}
						}
					} else if useMBTS && mbts == 2 &&
						strings.Contains(tower.Hash(0, 0), "sE_t15") {
						// take care of cross-module crack scintillators
						chans = append(chans, chCrack[pname][tower.CrackPartner()])
					} else {
						// everything else
						for _, ch := range module.Children(Readout) {
							chnbr, err := strconv.ParseInt(ch.Name(0)[1:], 10, 64)
							if err == nil && in_intslice(int(chnbr), chanNbrs) {
								chans = append(chans, ch)
							}
						}
					}
					tower.SetChildren(chans)
					for _, ch := range tower.Children(Readout) {
						ch.SetParent(tower)
					}
				} // towers
			} // samples
		} // modules
	} // partitions
	return tilecal
}

// EOF
