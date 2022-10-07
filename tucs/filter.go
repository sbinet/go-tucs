package tucs

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// filterWorker selects runs for TUCS to use.
type filterWorker struct {
	Base
	region           []string         // region selection
	runs             []Run            // run list
	runlst           []Run            // run list
	cs_atlas_runlst  []int64          // run-nbr set
	run_type         string           // requested run-type
	flags            map[int64]string // digiflags for each run-number
	keep_only_active bool             // only keep the active detector parts
	verbose          bool             // enable verbose output
	update_special   bool             // specified update
	allow_c10_err    bool             // allow errors for C10
	cs_comment       string           // cesium run/magnet description
	two_inputs       bool             // whether CIS runs should be between 2 dates
	filter           string           // store which laser filter is requested
	amp              float64          // requested amperage

	db     *sql.DB
	a_bad  []int         // list of special PMTs with cut-outs for A16
	ma_bad []int         // list of special modules with cut-outs for A16
	c_bad  []int         // list of special PMTs for C10
	d_bad  []int         // list of special PMTs for D5
	md_bad []int         // list of special modules EBA15 and EBC18
	period time.Duration // period of time used for cesium runs
}

// FilterCfg is a helper struct to ease the configuration of NewFilter
type FilterCfg struct {
	Runs           interface{} //[]string
	RunSet         string
	Region         string // fixme: use a const-type
	UseDateProg    bool
	Verbose        bool
	RunType        string  // fixme: use a const-type
	KeepOnlyActive bool    // only keep the active detector parts
	Filter         string  // requested laser filter. fixme: type-safety
	Amp            float64 // requested amperage
	GetLast        bool
	UpdateSpecial  bool
	AllowC10Errors bool   // allow errors for C10
	CsComment      string // cesium run/magnet description. fixme: type-safety
	TwoInput       bool   // whether CIS runs should be between 2 dates
}

// NewFilter creates a new filterWorker
func NewFilter(rtype RegionType, cfg FilterCfg) Worker {
	run, run2, two_inputs := translate_runs(&cfg)

	w := &filterWorker{
		Base:             NewBase(rtype),
		region:           region_from_cfg(&cfg),
		runs:             make([]Run, 0),
		runlst:           make([]Run, 0),
		cs_atlas_runlst:  make([]int64, 0),
		run_type:         cfg.RunType,
		flags:            make(map[int64]string),
		keep_only_active: cfg.KeepOnlyActive,
		verbose:          cfg.Verbose,
		update_special:   cfg.UpdateSpecial,
		allow_c10_err:    cfg.AllowC10Errors,
		cs_comment:       cfg.CsComment,
		two_inputs:       two_inputs,
		filter:           cfg.Filter,
		amp:              cfg.Amp,
	}

	iruns := []int64{}
	//fmt.Printf("run=%v (%T) run2=%v 2-inputs=%v\n", run, run, run2, two_inputs)
	//fmt.Printf("use-date-prog: %v\n", cfg.UseDateProg)

	// run-nbr selection
	if two_inputs {
		// charge injection for 2 dates
		if run, ok := run.(string); ok {
			if cfg.UseDateProg {
				iruns = w.date_prog(run, run2.(string))
			} else if _, ferr := os.Stat(run); ferr == nil {
				iruns = runs_from_file(run)
			}
		}
	} else {
		// laser or cesium or charge injection for a single date
		switch r := run.(type) {
		case int64:
			iruns = append(iruns, r)
		case []int64:
			iruns = append(iruns, r...)
		case string:
			if cfg.UseDateProg {
				iruns = w.date_prog(r, "")
			} else if _, ferr := os.Stat(r); ferr == nil {
				iruns = runs_from_file(r)
			}
		}
	}

	// select only the last run
	if cfg.GetLast {
		//FIXME: there was this additional check in tucs.worker.Use.py...
		//if _, ok := run.(string); ok {
		runmax := int64(0)
		for _, irun := range iruns {
			if irun > runmax {
				runmax = irun
			}
		}
		iruns = []int64{runmax}
		//}
	}

	db, err := sql.Open("mysql", "tcp:pcata007.cern.ch:3306*tile/reader/")
	if err != nil {
		panic("tucs.Filter.OpenSql: " + err.Error())
	}
	defer db.Close()

	if w.run_type == "cesium" {
		// cesium: each channel may have its own list of runs
		// as the runs are not partition-wide...
		for _, irun := range iruns {
			//fmt.Printf("--> irun=%v\n", irun)
			if irun > 40000 {
				rows, err := db.Query(
					`SELECT date FROM tile.comminfo WHERE run>=? ORDER BY run DESC LIMIT 1`,
					irun)
				if err != nil {
					panic("tucs.Filter.Sql.Query: " + err.Error())
				}
				for rows.Next() {
					var date time.Time
					err = rows.Scan(&date)
					if err != nil {
						panic("tucs.Filter.Sql.Scan: " + err.Error())
					}
					fmt.Printf("date: %v\n", date)
					w.runs = append(w.runs,
						Run{
							Type:   w.run_type,
							Number: irun,
							Time:   date,
							Data:   make(DataMap),
						})
				}
			} else {
				w.runs = append(w.runs,
					Run{
						Type:   w.run_type,
						Number: irun,
						Time:   time.Unix(0, 0), //FIXME: better default ?
						Data:   make(DataMap),
					})
			}
		}
	} else {
		for _, irun := range iruns {
			//fmt.Printf("--> irun=%v\n", irun)
			rows, err := db.Query(`select run, type, date, digifrags from tile.comminfo where run=?`,
				irun)
			if err != nil {
				panic("tucs.Filter.Sql.Query: " + err.Error())
			}
			irun2 := int64(-1)
			rtype := ""
			date := time.Unix(0, 0) // FIXME: better default ?
			digifrags := ""
			if !rows.Next() {
				irun2 = irun
			} else {
				err = rows.Scan(&irun2, &rtype, &date, &digifrags)
			}
			//fmt.Printf("==> %v, %v, %v, #%v\n", irun2, rtype, date, len(digifrags))
			if cfg.RunType == "all" || cfg.RunType == rtype || rtype == "" {
				w.runs = append(w.runs,
					Run{
						Type:   rtype,
						Number: irun2,
						Time:   date,
						Data:   make(DataMap),
					})
			}
			if w.keep_only_active && (rtype == cfg.RunType || rtype == "") {
				if w.verbose || (len(digifrags)/6 != 256) {
					fmt.Printf("in run %v, modules in readout: %v\n",
						irun, len(digifrags)/6)
				}
				if len(digifrags) == 0 {
					// turn off filter for active detector elements
					w.keep_only_active = false
				}
				w.flags[irun] = digifrags
			}
		}
	}
	return w
}

func region_from_cfg(cfg *FilterCfg) []string {
	regions := []string{}
	if len(cfg.Region) == 0 {
		return regions
	}
	if strings.Contains(cfg.Region, ",") {
		regions = strings.Split(cfg.Region, ",")
	} else {
		regions = append(regions, cfg.Region)
	}

	if regions[0][0] == 'H' {
		// FIXME
		panic("tucs.Filter: converting from cell-hash is not implemented")
	}
	return regions
}

// translate_runs translates the string form of the runs list into a form that
// filterWorker can work with
func translate_runs(cfg *FilterCfg) (run, run2 interface{}, two_inputs bool) {
	switch r := cfg.Runs.(type) {
	case int64:
		two_inputs = false
		run = r
		run2 = ""

	case []string:
		two_inputs = false
		run = r[0]
		run2 = ""

	case []int:
		two_inputs = false
		run = r
		run2 = ""

	case string:
		if !strings.Contains(r, "-") {
			// it's a date, use date -28days
			run2 = r
			run = r + "-28 days"
			two_inputs = true
		} else if r[0] == '-' {
			// check string to see if it is -x days
			two_inputs = false
			run = r
			run2 = ""
		} else {
			//fmt.
		}
	default:
		panic(fmt.Sprintf("tucs.Filter: unhandled cfg.Runs type: %T", cfg.Runs))
	}
	return
}

func (w *filterWorker) date_prog(run, run2 string) []int64 {
	iruns := []int64{}
	var date time.Time
	var date2 time.Time
	var err error
	if w.two_inputs == false {
		cmd := exec.Command("date", "-d", run, "--rfc-2822")
		if cmd == nil {
			panic("tucs.Filter.date_prog: could not create command 'date'")
		}
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("err: %v\n", string(out))
			panic("tucs.Filter.date_prog: " + err.Error())
		}
		datestr := strings.Trim(string(out), " \n\r")
		date, err = time.Parse(time.RFC1123Z, datestr)
		if err != nil {
			panic("tucs.Filter.date_prog: " + err.Error())
		}
	} else {
		{
			cmd := exec.Command("date", "-d", run, "--rfc-2822")
			if cmd == nil {
				panic("tucs.Filter.date_prog: could not create command 'date'")
			}
			out, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("err: %v\n", string(out))
				panic("tucs.Filter.date_prog: " + err.Error())
			}
			datestr := strings.Trim(string(out), " \n\r")
			date, err = time.Parse(time.RFC1123Z, datestr)
			if err != nil {
				panic("tucs.Filter.date_prog: " + err.Error())
			}
		}
		{
			cmd := exec.Command("date", "-d", run2, "--rfc-2822")
			if cmd == nil {
				panic("tucs.Filter.date_prog: could not create command 'date'")
			}
			out, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("err: %v\n", string(out))
				panic("tucs.Filter.date_prog: " + err.Error())
			}
			datestr := strings.Trim(string(out), " \n\r")
			date2, err = time.Parse(time.RFC1123Z, datestr)
			if err != nil {
				panic("tucs.Filter.date_prog: " + err.Error())
			}
		}
	}

	//fmt.Printf("date:  %v\ndate2: %v\n", date, date2)
	db, err := sql.Open("mymysql", "tcp:pcata007.cern.ch:3306*tile/reader/")
	if err != nil {
		panic("tucs.Filter.OpenSql: " + err.Error())
	}
	defer db.Close()

	query := []string{}
	args := []interface{}{}

	if w.run_type == "Las" {
		query = append(query, "select run from tile.comminfo where")
		// special treatment for LASER
		if w.two_inputs {
			query = append(query, "date>? and date<?")
			args = append(args, date, date2)
		} else {
			query = append(query, "date>?")
			args = append(args, date)
		}
		query = append(query, "and")
		if w.filter == "" || w.filter == " " {
			query = append(query, "((lasfilter='6' and events>10000) or (lasfilter='8' and events>100000))")
		} else {
			switch w.filter {
			case "6":
				query = append(query, ` (lasfilter="6" and events>10000)`)
			case "8":
				query = append(query, ` (lasfilter="8" and events>100000)`)
			default:
				query = append(query, `lasfilter=?`)
				args = append(args, w.filter)
			}
		}

		query = append(query, `and lasreqamp=? and type ='Las' and not (recofrags like '%005%' or recofrags like '%50%' or lasshopen=1 ) and comments is NULL`)
		args = append(args, w.amp)

	} else if w.run_type == "cesium" {
		query = append(query, `select run from tile.runDescr where time>? and module<65`)
		args = append(args, date)
		if w.cs_comment != "" {
			query = append(query, "and comment=?")
			args = append(args, w.cs_comment)
		}
		if w.two_inputs {
			query = append(query, `and time<?`)
			args = append(args, date2)
		}
	} else if w.run_type == "CIS" && w.two_inputs == true {
		query = append(query, `select run from title.comminfo where date<? and date>?`)
		args = append(args, date2, date)
	} else {
		query = append(query, `select run from title.comminfo where run<9999999 and date>?`)
		args = append(args, date)
	}

	fmt.Printf("%v\n", strings.Join(query, " "))
	//fmt.Printf("%v %v\n", len(args), args)
	rows, err := db.Query(strings.Join(query, " "), args...)
	if err != nil {
		fmt.Printf("tucs.Filter.QuerySql: %v\n", err.Error())
		panic(err.Error())
	}
	//fmt.Printf("cols: %v\n", cols)
	for rows.Next() {
		irun := int64(-1)
		err = rows.Scan(&irun)
		if err != nil {
			panic(err.Error())
		}
		//fmt.Printf("--> %v\n", irun)
		iruns = append(iruns, irun)
	}
	return iruns
}

func runs_from_file(fname string) []int64 {
	iruns := make([]int64, 0)
	f, err := os.Open(fname)
	if err != nil {
		panic("tucs.Filter: " + err.Error())
	}
	for {
		irun := int64(-1)
		_, err = fmt.Fscanf(f, "%d\n", &irun)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic("tucs.Filter.ReadLine: " + err.Error())
		}
		iruns = append(iruns, irun)
	}
	return iruns
}

func (w *filterWorker) ProcessStart() error {
	var err error = nil
	printf := fmt.Printf
	printf("Regions: %v\n", w.region)

	for _, run := range w.runs {
		fmt.Printf("%v\n", &run)
	}

	if len(w.region) == 0 {
		fmt.Printf("Filter: using the whole detector\n")
	} else {
		fmt.Printf("Filter: only using region(s) %v\n", w.region)
	}

	w.runlst = make([]Run, 0, len(w.runs))
	for _, run := range w.runs {
		w.runlst = append(w.runlst, run)
	}

	if w.run_type == "cesium" {
		w.db, err = sql.Open("mymysql", "tcp:pcata007.cern.ch:3306*tile/reader/")
		if err != nil {
			return err
		}
		w.a_bad = []int{41, 42}
		w.ma_bad = []int{36, 61}
		if w.allow_c10_err {
			w.c_bad = []int{5, 6}
		} else {
			w.c_bad = []int{0}
		}
		w.d_bad = []int{17, 18}
		w.md_bad = []int{15, 18}
		w.period, err = time.ParseDuration("168h") // = 7*24h = 7 days = 1week
		if err != nil {
			return err
		}
	}
	return err
}

func (w *filterWorker) ProcessStop() error {
	var err error
	if w.run_type == "cesium" {
		err = w.db.Close()
	}

	return err
}

func (w *filterWorker) ProcessRegion(region *Region) error {
	var err error
	use_region := false
	if len(w.region) > 0 {
		for _, reg := range w.region {
			if strings.Contains(region.Hash(0, 0), reg) ||
				strings.Contains(region.Hash(1, 0), reg) {
				use_region = true
				break
			}
		}
	} else {
		use_region = true
	}

	if use_region {
		//TODO
		if w.run_type == "cesium" {
			panic("FIXME: not implemented")
			// nbr := region.Number(1, 0)
			// if len(nbr) == 3 {
			// 	module := region.Parent(Readout, 0)
			// 	if len(module.Events()) == 0 {
			// 		for _, run := range w.runlst {
			// 			w.fill_module(module, run)
			// 		}
			// 	}
			// 	for ievt, _ := range module.Events() {

			// 	}
			// }
		} else {
			for _, run := range w.runlst {
				hash := region.Hash(0, 0)
				if w.keep_only_active && !w.is_active(hash, run.Number) {
					if w.verbose {
						fmt.Printf("Region not in readout, removing: %v\n",
							hash)
					}
				} else if w.run_type == "Las" {
					// region is an ADC ?
					if !strings.Contains(hash, "gain") {
						continue
					}
					data := make(DataMap)
					data["region"] = hash
					region.AddEvent(Event{Run: run, Data: data})
				} else {
					data := make(DataMap)
					data["region"] = hash
					region.AddEvent(Event{Run: run, Data: data})
				}
			}
		}
	}
	// update global run-list...
	Runs = make([]Run, len(w.runlst))
	copy(Runs, w.runlst)
	return err
}

func (w *filterWorker) is_active(hash string, run int64) bool {
	v, ok := w.flags[run]
	if !ok {
		// region *is* used
		return true
	} else {
		if len(v) == 0 {
			fmt.Printf("flags[%d]=%v\n", run, v)
			return true
		}
	}

	hex := uint16(0)
	switch {
	case strings.Contains(hash, "B"):
		switch {
		case strings.Contains(hash, "LBA"):
			hex = uint16(0x1 << 8)
		case strings.Contains(hash, "LBC"):
			hex = uint16(0x2 << 8)
		case strings.Contains(hash, "EBA"):
			hex = uint16(0x3 << 8)
		case strings.Contains(hash, "EBC"):
			hex = uint16(0x4 << 8)
		}
		if strings.Contains(hash, "_m") {
			idx := strings.Index(hash, "m")
			if idx == -1 {
				panic("tucs.Filter.is_active: index logic error")
			}
			modstr := hash[idx+1 : idx+3]
			mod, err := strconv.ParseInt(modstr, 10, 64)
			if err != nil {
				panic("tucs.Filter.is_active: parse pb: " + err.Error())
			}
			mod -= 1
			hex += uint16(mod)
			hexstr := fmt.Sprintf("0x%x", hex)
			return strings.Contains(w.flags[run], hexstr)
		}
	case hash == "TILECAL":
		return true

	default:
		fmt.Printf("**error** tucs.Filter: unknown hash: %v\n", hash)
		return false
	}
	panic("unreachable")
}

// check filterWorker implements tucs.Worker
var _ Worker = (*filterWorker)(nil)
