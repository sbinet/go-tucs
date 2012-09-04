package tucs

import (
	"fmt"
	"sort"
	"time"
)

type Event struct {
	Run  Run
	Data map[string]interface{}
}

type Run struct {
	Type   string
	Number int64
	Time   time.Time
	Data   interface{}
}

func (r *Run) String() string {
	return fmt.Sprintf(
		"Run{Number: %v, Type: %v, Time: %v, Data: %v}",
		r.Number, r.Type, r.Time, r.Data,
	)
}

// runList is a slice of Runs which implements sort.Interface
// TODO: make it a slice of *pointers* to Run ? (if too slow)
type runList struct {
	runs []Run
	fct  func(i, j Run) bool
}

func (r runList) Len() int           { return len(r.runs) }
func (r runList) Less(i, j int) bool { return r.fct(r.runs[i], r.runs[j]) }
func (r runList) Swap(i, j int)      { r.runs[i], r.runs[j] = r.runs[j], r.runs[i] }

func (r runList) Sort() { sort.Sort(r) }

// SortRunList sorts runs by run number
func SortRunList(runs []Run) {
	sort.Sort(runList{
		runs: runs,
		fct:  func(i, j Run) bool { return i.Number < j.Number },
	})
}

// SortRunListBy sorts runs by the provided function fct
func SortRunListBy(runs []Run, fct func(i, j Run) bool) {
	sort.Sort(runList{runs, fct})
}

// EOF
