package tucs

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	_ "github.com/ziutek/mymysql/godrv"
)

// filterWorker selects runs for TUCS to use.
type filterWorker struct {
	Base
	region           []string            // region selection
	runs             []Run               // run list
	runlst           []Run               // run list
	cs_atlas_runlst  []int64             // run-nbr set
	run_type         string              // requested run-type
	flags            map[int]interface{} // digiflags for each run-number
	keep_only_active bool                // only keep the active detector parts
	verbose          bool                // enable verbose output
	update_special   bool                // specified update
	allow_c10_err    bool                // allow errors for C10
	cs_comment       string              // cesium run/magnet description
	two_inputs       bool                // whether CIS runs should be between 2 dates
	filter           string              // store which laser filter is requested
	amp              float64             // requested amperage
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
		Base:            NewBase(rtype),
		region:          region_from_cfg(&cfg),
		runs:            make([]Run, 0),
		runlst:          make([]Run, 0),
		cs_atlas_runlst: make([]int64, 0),
		run_type:        cfg.RunType,
		flags:           make(map[int]interface{}),
		two_inputs:      two_inputs,
	}

	iruns := []int64{}
	fmt.Printf("run=%v (%T) run2=%v 2-inputs=%v\n", run, run, run2, two_inputs)
	fmt.Printf("use-date-prog: %v\n", cfg.UseDateProg)

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

	db, err := sql.Open("mymysql", "tcp:pcata007.cern.ch:3306*tile/reader/")
	if err != nil {
		panic("tucs.Filter.OpenSql: " + err.Error())
	}
	defer db.Close()

	if w.run_type == "cesium" {
		// cesium: each channel may have its own list of runs
		// as the runs are not partition-wide...
		for _, irun := range iruns {
			fmt.Printf("--> irun=%v\n", irun)
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
							Data:   nil,
						})
				}
			} else {
				w.runs = append(w.runs,
					Run{
						Type:   w.run_type,
						Number: irun,
						Time:   time.Now(), //FIXME: find a better default ?
						Data:   nil,
					})
			}
		}
	} else {

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

	fmt.Printf("date:  %v\ndate2: %v\n", date, date2)
	db, err := sql.Open("mymysql", "tcp:pcata007.cern.ch:3306*tile/reader/")
	if err != nil {
		panic("tucs.Filter.OpenSql: " + err.Error())
	}
	defer db.Close()

	var stmt *sql.Stmt = nil
	query := []string{}
	args := []interface{}{}

	if w.run_type == "Las" {
		// special treatment for LASER
		if w.two_inputs {
			query = append(query, "date>? and date<?")
			args = append(args, date, date2)
		} else {
			query = append(query, "date>?")
			args = append(args, date)
		}

		if w.filter == "" || w.filter == " " {
			
		}
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

	return err
}

// check filterWorker implements tucs.Worker
var _ Worker = (*filterWorker)(nil)

// EOF
