package tucs

import (
	"fmt"
)

type Event struct {
}

type Worker interface {
	ProcessStart() error
	ProcessStop() error
	ProcessRegion(region *Region) error

	RegionType() RegionType
}

type App struct {
	workers  []Worker
	detector Region
}

func NewApp() *App {
	app := &App{
		workers:  []Worker{},
		detector: Region{},
	}
	return app
}

func (app *App) msg(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (app *App) Run() error {
	var err error

	for _, w := range app.workers {

		app.msg("running [%T]...\n", w)
		err = w.ProcessStart()
		if err != nil {
			return err
		}

		for region := range app.detector.Regions(w.RegionType()) {
			err = w.ProcessRegion(region)
			if err != nil {
				return err
			}
		}

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

// EOF
