package tucs

// filterWorker selects runs for TUCS to use.
type filterWorker struct {
	Base
	cfg FilterCfg
}

// FilterCfg is a helper struct to ease the configuration of NewFilter
type FilterCfg struct {
	Runs           []string
	Type           RegionType
	Region         string // fixme: use a const-type
	UseDateProg    bool
	Verbose        bool
	RunType        string // fixme: use a const-type
	KeepOnlyActive bool
	Filter         string // fixme: type-safety
	Amp            int64
	GetLast        bool
	UpdateSpecial  bool
	AllowC10Errors bool
	CsComment      string // fixme: type-safety
	TwoInput       bool
}

// NewFilter creates a new filterWorker
func NewFilter(cfg FilterCfg) Worker {
	w := &filterWorker{
		Base: NewBase(cfg.Type),
		cfg:  cfg,
	}
	return w
}

// check filterWorker implements tucs.Worker
var _ Worker = (*filterWorker)(nil)

// EOF
