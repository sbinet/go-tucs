go-tucs
=======

An experimental re-write of TUCS (TileCal Unified Calibration Software) in Go.

https://twiki.cern.ch/twiki/bin/viewauth/Atlas/TileCalibrationTucs
https://svnweb.cern.ch/trac/atlasoff/browser/TileCalorimeter/TileCalib/TileCalibAlgs/trunk/share/Tucs/src

Installation
------------

    $ go get github.com/sbinet/go-tucs/cmd/go-tucs


Example
-------

    $ go-tucs
    Welcome to Go-TUCS (pid=85032). Building detector tree...
    Constructing TileCal detector tree:
        MBTS mapping enabled
        Special mapping in EBA15 and EBC18 enabled
    done.
    running [*main.printWorker]...
    ::worker-start...
    ::worker-start...[done]
    ::worker-stop...
      processed [29817] region(s) of type [readout]
    ::worker-stop... [done]



