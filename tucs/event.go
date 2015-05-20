package tucs

import (
	"fmt"
	"sort"
	"time"
)

type DataMap map[string]interface{}

type Event struct {
	Run  Run
	Data DataMap
}

type Run struct {
	Type   string
	Number int64
	Time   time.Time
	Data   DataMap
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

// RunList is a slice of Runs with some more refinements
type RunList []Run

// Runs is the global variable holding a RunList
var Runs = make(RunList, 0)

// RunsOfType returns a slice of Runs from this RunList with the correct run_type
func (r RunList) RunsOfType(rtype string) []Run {
	lst := make([]Run, 0, len(r))
	for _, v := range r {
		if v.Type == rtype {
			lst = append(lst, v)
		}
	}
	return lst
}

// Remove removes a given Run from this RunList.
// Note it only checks the run number for the equality comparison.
func (r *RunList) Remove(run Run) {
	lst := make([]Run, 0, len(*r))
	//TODO: use slice tricks ?
	for _, v := range *r {
		if v.Number != run.Number {
			lst = append(lst, v)
		}
	}
	*r = make([]Run, len(lst))
	copy(*r, lst)
}

// EOF
