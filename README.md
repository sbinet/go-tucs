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



Performances comparisons
------------------------

The `python` version:

    $ time python macros/examples/01_hello_world.py
    [...]
    Entering worker loop:
    Running PrintHelloWorld - A demo class that just prints hello world
    Hello world from the worker PrintHelloWorld!
    processed [29817] regions of type [readout]
    
    TUCS finished in: 0:00:04.679538
    python macros/examples/01_hello_world.py  36.69s user 0.68s system 97% cpu 38.212 total



The `go` one:

    $ time (go get go-tucs && go-tucs)
    [...]
    running [*main.printWorker]...
    ::worker-start...
    ::worker-start...[done]
    ::worker-stop...
      processed [29817] region(s) of type [readout]
    ::worker-stop... [done]
    ( go get . && go-tucs; )  1.32s user 0.05s system 99% cpu 1.387 total

so, even when recompiling the whole `tucs` package, it still beats the `python` version hands down.

