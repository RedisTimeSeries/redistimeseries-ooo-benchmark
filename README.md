[![CircleCI](https://circleci.com/gh/RedisTimeSeries/redistimeseries-ooo-benchmark.svg?style=svg)](https://circleci.com/gh/RedisTimeSeries/redistimeseries-ooo-benchmark)
[![GitHub issues](https://img.shields.io/github/release/RedisTimeSeries/redistimeseries-ooo-benchmark.svg)](https://github.com/RedisTimeSeries/redistimeseries-ooo-benchmark/releases/latest)
[![Codecov](https://codecov.io/gh/RedisTimeSeries/redistimeseries-ooo-benchmark/branch/master/graph/badge.svg)](https://codecov.io/gh/RedisTimeSeries/redistimeseries-ooo-benchmark)
[![GoDoc](https://godoc.org/github.com/RedisTimeSeries/redistimeseries-ooo-benchmark?status.svg)](https://godoc.org/github.com/RedisTimeSeries/redistimeseries-ooo-benchmark)


# redistimeseries-ooo-benchmark
[![Forum](https://img.shields.io/badge/Forum-RedisTimeSeries-blue)](https://forum.redislabs.com/c/modules/redistimeseries)
[![Gitter](https://badges.gitter.im/RedisLabs/RedisTimeSeries.svg)](https://gitter.im/RedisLabs/RedisTimeSeries?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

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
go get github.com/RedisTimeSeries/redistimeseries-ooo-benchmark
cd $GOPATH/src/github.com/RedisTimeSeries/redistimeseries-ooo-benchmark
make
```

## Usage of redistimeseries-ooo-benchmark

```
Usage of redistimeseries-ooo-benchmark:
  -chunk-size int
        chunk size. (default 4096)
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