package tucs

import (
	
	"github.com/sbinet/go-croot/pkg/croot"
	)

/*
class ReadGenericCalibration(GenericWorker):
    "The Generic Calibration Template"

    tfile_cache = {}
    processingDir = 'tmp'
    
    def getFileTree(self, fileName, treeName):
        f, t = [None, None]

        if self.tfile_cache.has_key(self.processingDir+fileName):
            f, t = self.tfile_cache[self.processingDir+fileName]
        else:
            if os.path.exists(os.path.join(self.processingDir, fileName)) or 'rfio:/' == self.processingDir[0:6]:
                f = TFile.Open(os.path.join(self.processingDir, fileName), "read")

            if not f:
                return [None, None]
            
            t = f.Get(treeName)
            if not t:
                print "Tree failed to be grabbed: " + fileName
                return [None, None]

            self.tfile_cache[self.processingDir+fileName] = [f, t]
        
        return [f, t]
*/

type centry struct {
	file *croot.File
	tree *croot.Tree
}

// CalibBase is a generic calibration worker with a cache of ROOT files/trees
type CalibBase struct {
	Base
	cache map[string]centry
	workdir string
}

// NewCalibBase creates a new CalibBase worker ready for embedding
func NewCalibBase(rtype RegionType, workdir string) CalibBase {
	w := CalibBase{
		Base: NewBase(rtype),
		cache: make(map[string]centry),
		workdir: workdir,
	}
	return w
}

// checks CalibBase implements tucs.Worker
var _ Worker = (*CalibBase)(nil)

// EOF
