package tucs

// filterWorker selects runs for TUCS to use.
type filterWorker struct {
	Base
}

// NewFilter creates a new filterWorker
func NewFilter(rtype RegionType) Worker {
	return &filterWorker{
		Base: NewBase(rtype),
	}
}

// check filterWorker implements tucs.Worker
var _ Worker = (*filterWorker)(nil)

// EOF
