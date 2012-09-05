package tucs

import (
	"fmt"
	//"os"
	"path"

	"github.com/sbinet/go-croot/pkg/croot"
)

type centry struct {
	file *croot.File
	tree *croot.Tree
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

func (w *CalibBase) FileTree(file, tree string) (*croot.File, *croot.Tree) {
	var f *croot.File = nil
	var t *croot.Tree = nil

	key := w.workdir + file
	if c, ok := w.cache[key]; ok {
		f = c.file
		t = c.tree
	} else {
		fname := path.Join(w.workdir, file)
		f = croot.OpenFile(fname, "read", "TUCS ROOT file", 1, 0)
		if f != nil {
			t = f.GetTree(tree)
			if t == nil {
				fmt.Printf("**error** tucs.FileTree failed to grab file=%s tree=%s\n",
					file, tree)
			} else {
				w.cache[key] = centry{
					file: f,
					tree: t,
				}
			}
		}
	}
	return f, t
}

// checks CalibBase implements tucs.Worker
var _ Worker = (*CalibBase)(nil)

// EOF
