package tucs

type Event struct {
	Run  Run
	Data interface{}
}

type Run struct {
	Type   string
	Number int64
	Data   interface{}
}

// EOF
