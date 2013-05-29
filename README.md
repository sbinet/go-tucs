go-tucs
=======

An experimental re-write of TUCS (TileCal Unified Calibration Software) in Go.

https://twiki.cern.ch/twiki/bin/viewauth/Atlas/TileCalibrationTucs
https://svnweb.cern.ch/trac/atlasoff/browser/TileCalorimeter/TileCalib/TileCalibAlgs/trunk/share/Tucs/src

## Installation

``` sh
$ go get github.com/sbinet/go-tucs/examples/go-tucs-helloworld
```

## Documentation

http://godoc.org/github.com/sbinet/go-tucs/pkg/tucs

## Notes

You'll need ``go-croot`` to be able to read ``ROOT`` files.
``go-croot`` is go-get-able like so:

``` sh
$ go get github.com/sbinet/go-croot
```

``go-croot`` itself needs ``croot``, a subset of the ``ROOT`` API
exposed thru ``C``.
Instructions on how to install ``croot`` are here:
 http://github.com/sbinet/croot
 

## Example

``` sh
$ go-tucs-helloworld
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
```


## Performances comparisons

The `python` version:

``` sh
$ time python macros/examples/01_hello_world.py
[...]
Entering worker loop:
Running PrintHelloWorld - A demo class that just prints hello world
Hello world from the worker PrintHelloWorld!
processed [29817] regions of type [readout]

TUCS finished in: 0:00:04.679538
python macros/examples/01_hello_world.py  36.69s user 0.68s system 97% cpu 38.212 total
```


The `go` one:

``` sh
$ time (go get github.com/sbinet/go-tucs/examples/go-tucs-helloworld && go-tucs-helloworld)
[...]
running [*main.printWorker]...
::worker-start...
::worker-start...[done]
::worker-stop...
  processed [29817] region(s) of type [readout]
::worker-stop... [done]
( go get . && go-tucs-helloworld; )  1.32s user 0.05s system 99% cpu 1.387 total
```

so, even when recompiling the whole `tucs` package, it still beats the `python` version hands down.

