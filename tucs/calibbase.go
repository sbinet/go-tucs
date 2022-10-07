package tucs

import (

	//"os"
	"path"

	"go-hep.org/x/hep/groot"
	"go-hep.org/x/hep/groot/rtree"
)

type centry struct {
	file *groot.File
	tree rtree.Tree
}

// CalibBase is a generic calibration worker with a cache of ROOT files/trees
type CalibBase struct {
	Base
	cache   map[string]centry
	workdir string
}

// NewCalibBase creates a new CalibBase worker ready for embedding
func NewCalibBase(rtype RegionType, workdir string) CalibBase {
	w := CalibBase{
		Base:    NewBase(rtype),
		cache:   make(map[string]centry),
		workdir: workdir,
	}
	return w
}

func (w *CalibBase) Dir() string {
	return w.workdir
}

func (w *CalibBase) FileTree(file, tree string) (*groot.File, rtree.Tree) {
	var (
		err error
		f   *groot.File
		t   rtree.Tree
	)

	key := w.workdir + file
	c, ok := w.cache[key]
	if ok {
		return c.file, c.tree
	}

	fname := path.Join(w.workdir, file)
	f, err = groot.Open(fname)
	if err != nil {
		return nil, nil
	}

	o, err := f.Get(tree)
	if err != nil {
		return f, nil
	}
	t = o.(rtree.Tree)

	w.cache[key] = centry{
		file: f,
		tree: t,
	}
	return f, t
}

// checks CalibBase implements tucs.Worker
var _ Worker = (*CalibBase)(nil)

// EOF
