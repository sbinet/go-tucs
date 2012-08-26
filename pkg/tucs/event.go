package tucs

import (
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

// RunList is a slice of Runs which implements sort.Interface
type RunList []Run

func (r RunList) Len() int           { return len(r) }
func (r RunList) Less(i, j int) bool { return r[i].Number < r[j].Number }
func (r RunList) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }

func (r RunList) Sort() { sort.Sort(r) }

func SortRunList(runs []Run) { sort.Sort(RunList(runs)) }

// EOF
