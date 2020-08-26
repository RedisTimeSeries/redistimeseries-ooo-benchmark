# redistimeseries-ooo-benchmark

## Overview

This repo contains code to mimic the out ot order / backfilled workloads on RedisTimeSeries >= v1.4.

Several aspects can dictate the overall system performance, like the:
- Pipeline size 
- Number of distinct clients ( each client has a dedicated time-serie )
- Compressed / Uncompressed series
- Out of order ratio 

## Installation

The easiest way to get and install the Subscriber Go program is to use
`go get` and then `go install`:
```bash
# Fetch this repo
go get github.com/filipecosta90/redistimeseries-ooo-benchmark
cd $GOPATH/src/github.com/filipecosta90/redistimeseries-ooo-benchmark
go get ./...
go build .
```

## Usage of redistimeseries-ooo-benchmark

```
Usage of redistimeseries-ooo-benchmark:
  -chunk-size int
        chunk size. (default 256)
  -client-update-tick int
        client update tick. (default 1)
  -compressed
        test for compressed TS
  -debug-level int
        debug level.
  -host string
        redis host. (default "127.0.0.1:6379")
  -json-out-file string
        Name of json output file, if not set, will not print to json.
  -ooo-percentage float
        out of order percentage [0.0,100.0]
  -pipeline int
        pipeline. (default 1)
  -random-seed int
        random seed to be used. (default 12345)
  -samples-per-ts uint
        Number of total samples per timeseries. (default 100000)
  -ts-maximum uint
        channel ID maximum value ( each channel has a dedicated thread ). (default 100)
  -ts-minimum uint
        channel ID minimum value ( each channel has a dedicated thread ). (default 1)

```