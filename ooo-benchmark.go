package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/filipecosta90/hdrhistogram"
	"github.com/mediocregopher/radix/v3"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var totalCommands uint64
var totalErrors uint64
var latencies *hdrhistogram.Histogram

type testResult struct {
	StartTime       int64   `json:"StartTime"`
	Duration        float64 `json:"Duration"`
	Pipeline        int     `json:"Pipeline"`
	Compressed      bool    `json:"Compressed"`
	ChunkSize       int     `json:"ChunkSize"`
	OOOPercentage   float64 `json:"OOOPercentage"`
	CommandRate     float64 `json:"CommandRate"`
	TotalCommands   uint64  `json:"TotalCommands"`
	totalErrors     uint64  `json:"totalErrors"`
	TotalTimeseries uint64  `json:"TotalTimeseries"`
	TsMin           uint64  `json:"TsMin"`
	TsMax           uint64  `json:"TsMax"`
	CommandsPerTs   uint64  `json:"CommandsPerTs"`
	P50IngestionMs  float64 `json:"p50IngestionMs"`
	P95IngestionMs  float64 `json:"p95IngestionMs"`
	P99IngestionMs  float64 `json:"p99IngestionMs"`
}

func ingestionRoutine(addr string, continueOnError bool, pipelineSize int, clientName string, tsName string, compressed bool, chunk_size int, ooo_percent float64, number_samples uint64, debug_level int, wg *sync.WaitGroup) {
	// tell the caller we've stopped
	defer wg.Done()

	conn, err := radix.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	err = conn.Do(radix.FlatCmd(nil, "CLIENT", "SETNAME", clientName))
	if err != nil {
		log.Fatal(err)
	}
	createCmd := radix.FlatCmd(nil, "TS.CREATE", tsName, "CHUNK_SIZE", chunk_size)
	if !compressed {
		createCmd = radix.FlatCmd(nil, "TS.CREATE", tsName, "UNCOMPRESSED", "CHUNK_SIZE", chunk_size)
	}
	err = conn.Do(createCmd)
	if err != nil {
		log.Fatal(err)
	}
	ts := 100000
	currentPipelineSize := 0
	maxTs := ts
	cmds := make([]radix.CmdAction, 0, 0)
	times := make([]time.Time, 0, 0)
	clientIssuedCommands := 0
	for i := 0; uint64(i) < number_samples; i++ {
		if rand.Float64()*100.0 < ooo_percent {
			ts = rand.Intn(maxTs)
		} else {
			maxTs++
			ts = maxTs
		}
		value := rand.Float32() * 32.0

		cmd := radix.FlatCmd(nil, "TS.ADD", tsName, ts, value)
		cmds = append(cmds, cmd)
		currentPipelineSize++
		startT := time.Now()
		times = append(times, startT)
		if currentPipelineSize >= pipelineSize {
			err = conn.Do(radix.Pipeline(cmds...))
			endT := time.Now()
			if err != nil {
				if continueOnError {
					atomic.AddUint64(&totalErrors, uint64(pipelineSize))
					if debug_level > 0 {
						log.Println(fmt.Sprintf("Received an error with the following command(s): %v, error: %v", cmds, err))
					}
				} else {
					log.Fatal(err)
				}
			}

			for _, t := range times {
				duration := endT.Sub(t)
				latencies.RecordValue(duration.Microseconds())
			}
			clientIssuedCommands += currentPipelineSize
			atomic.AddUint64(&totalCommands, uint64(pipelineSize))
			cmds = nil
			cmds = make([]radix.CmdAction, 0, 0)
			times = nil
			times = make([]time.Time, 0, 0)
			currentPipelineSize = 0
		}
	}
}

func main() {
	host := flag.String("host", "127.0.0.1:6379", "redis host.")
	pipeline := flag.Int("pipeline", 1, "pipeline.")
	seed := flag.Int64("random-seed", 12345, "random seed to be used.")
	compressed := flag.Bool("compressed", false, "test for compressed TS")
	ooo_percentage := flag.Float64("ooo-percentage", 0.0, "out of order percentage [0.0,100.0]")
	chunk_size := flag.Int("chunk-size", 256, "chunk size.")
	debug_level := flag.Int("debug-level", 0, "debug level.")
	ts_minimum := flag.Uint64("ts-minimum", 1, "time-serie ID minimum value ( each time-serie has a dedicated thread ).")
	ts_maximum := flag.Uint64("ts-maximum", 100, "time-serie ID maximum value ( each time-serie has a dedicated thread ).")
	samples_per_ts := flag.Uint64("samples-per-ts", 100000, "Number of total samples per timeseries.")
	json_out_file := flag.String("json-out-file", "", "Name of json output file, if not set, will not print to json.")
	client_update_tick := flag.Int("client-update-tick", 1, "client update tick.")
	flag.Parse()
	totalWorkers := *ts_maximum - *ts_minimum + 1
	stopChan := make(chan struct{})
	rand.Seed(*seed)
	latencies = hdrhistogram.New(1, 90000000, 3)

	// a WaitGroup for the goroutines to tell us they've stopped
	wg := sync.WaitGroup{}

	totalM := totalWorkers * *samples_per_ts
	fmt.Println(fmt.Sprintf("Total time-series: %d. Datapoints per TS: %d Total commands: %d", totalWorkers, *samples_per_ts, totalM))
	fmt.Println(fmt.Sprintf("Simulating OOO %%: %3.2f", *ooo_percentage))
	fmt.Println(fmt.Sprintf("Compressed chunks: %v. Chunk size %d Bytes.", *compressed, *chunk_size))
	fmt.Println(fmt.Sprintf("Pipeline: %d", *pipeline))
	fmt.Println(fmt.Sprintf("Using random seed: %d", *seed))

	for channel_id := *ts_minimum; channel_id <= *ts_maximum; channel_id++ {
		clientName := fmt.Sprintf("ooo-client#%d", channel_id)
		tsName := fmt.Sprintf("ooo-client#%d", channel_id)
		wg.Add(1)
		go ingestionRoutine(*host, true, *pipeline, clientName, tsName, *compressed, *chunk_size, *ooo_percentage, *samples_per_ts, *debug_level, &wg)
	}

	// listen for C-c
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	tick := time.NewTicker(time.Duration(*client_update_tick) * time.Second)
	closed, start_time, duration, totalMessages, _ := updateCLI(tick, c, totalM)
	messageRate := float64(totalMessages) / float64(duration.Seconds())
	p50IngestionMs := float64(latencies.ValueAtQuantile(50.0)) / 1000.0
	p95IngestionMs := float64(latencies.ValueAtQuantile(95.0)) / 1000.0
	p99IngestionMs := float64(latencies.ValueAtQuantile(99.0)) / 1000.0

	fmt.Fprint(os.Stdout, fmt.Sprintf("\n"))
	fmt.Fprint(os.Stdout, fmt.Sprintf("#################################################\n"))
	fmt.Fprint(os.Stdout, fmt.Sprintf("Total Duration %.3f Seconds\n", duration.Seconds()))
	fmt.Fprint(os.Stdout, fmt.Sprintf("Total Errors %d\n", totalErrors))
	fmt.Fprint(os.Stdout, fmt.Sprintf("Throughput summary: %.0f requests per second\n", messageRate))
	fmt.Fprint(os.Stdout, "Latency summary (msec):\n")
	fmt.Fprint(os.Stdout, fmt.Sprintf("    %9s %9s %9s\n", "p50", "p95", "p99"))
	fmt.Fprint(os.Stdout, fmt.Sprintf("    %9.3f %9.3f %9.3f\n", p50IngestionMs, p95IngestionMs, p99IngestionMs))

	if strings.Compare(*json_out_file, "") != 0 {

		res := testResult{
			StartTime:       start_time.Unix(),
			Pipeline:        *pipeline,
			Compressed:      *compressed,
			ChunkSize:       *chunk_size,
			OOOPercentage:   *ooo_percentage,
			Duration:        duration.Seconds(),
			CommandRate:     messageRate,
			TotalCommands:   totalMessages,
			totalErrors:     totalErrors,
			TotalTimeseries: totalWorkers,
			TsMin:           *ts_minimum,
			TsMax:           *ts_maximum,
			CommandsPerTs:   *samples_per_ts,
			P50IngestionMs:  p50IngestionMs,
			P95IngestionMs:  p95IngestionMs,
			P99IngestionMs:  p99IngestionMs,
		}
		file, err := json.MarshalIndent(res, "", " ")
		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile(*json_out_file, file, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	if closed {
		return
	}

	// tell the goroutine to stop
	close(stopChan)
	// and wait for them both to reply back
	wg.Wait()
}

func updateCLI(tick *time.Ticker, c chan os.Signal, message_limit uint64) (bool, time.Time, time.Duration, uint64, []float64) {

	start := time.Now()
	prevTime := time.Now()
	prevMessageCount := uint64(0)
	messageRateTs := []float64{}
	fmt.Fprint(os.Stdout, fmt.Sprintf("%26s %7s %25s %25s %7s %25s %25s\n", "Test time", " ", "Total Commands", "Total Errors", "", "Command Rate", "p50 lat. (msec)"))
	for {
		select {
		case <-tick.C:
			{
				now := time.Now()
				took := now.Sub(prevTime)
				messageRate := float64(totalCommands-prevMessageCount) / float64(took.Seconds())
				completionPercent := float64(totalCommands) / float64(message_limit) * 100.0
				errorPercent := float64(totalErrors) / float64(totalCommands) * 100.0

				p50 := float64(latencies.ValueAtQuantile(50.0)) / 1000.0

				if prevMessageCount == 0 && totalCommands != 0 {
					start = time.Now()
				}
				if totalCommands != 0 {
					messageRateTs = append(messageRateTs, messageRate)
				}
				prevMessageCount = totalCommands
				prevTime = now

				fmt.Fprint(os.Stdout, fmt.Sprintf("%25.0fs [%3.1f%%] %25d %25d [%3.1f%%] %25.2f %25.2f\t", time.Since(start).Seconds(), completionPercent, totalCommands, totalErrors, errorPercent, messageRate, p50))
				fmt.Fprint(os.Stdout, "\r")
				//w.Flush()
				if message_limit > 0 && totalCommands >= uint64(message_limit) {
					return true, start, time.Since(start), totalCommands, messageRateTs
				}

				break
			}

		case <-c:
			fmt.Println("received Ctrl-c - shutting down")
			return true, start, time.Since(start), totalCommands, messageRateTs
		}
	}
	return false, start, time.Since(start), totalCommands, messageRateTs
}
