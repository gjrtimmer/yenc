package yenc

import (
    "fmt"
    "runtime"
    "sync"
)

// Decoder for multi core yEnc decoding
type Decoder struct {
    maxCores        int
    poolChannel     chan chan *Data
    workQueue       chan *Data
    pool            map[int]*worker
    quit            chan bool
}

// NewDecoder creates a new Decoder
func NewDecoder(maxCores int, workQueue chan *Data) (*Decoder, error) {

    if maxCores > runtime.NumCPU() {
        return nil, fmt.Errorf("max cores (%d) cannot exceed max number of available CPU (%d)", maxCores, runtime.NumCPU())
    }

    poolChan := make(chan chan *Data, maxCores)

    return &Decoder {
        maxCores: maxCores,

        poolChannel: poolChan,
        workQueue: workQueue,
        pool: make(map[int]*worker),
        quit: make(chan bool),
    }, nil
}

// Start Decoder
func (d *Decoder) Start() {

    // Start Workers
    for i := 0; i < d.maxCores; i++ {
        d.pool[i] = d.newWorker(i, d.poolChannel)
        d.pool[i].start()
    }

    go d.dispatch()
}

// Stop Decoder
func (d *Decoder) Stop() {
    // Stop Workers
    for _, w := range d.pool {
        w.stop()
    }

    // Signal Decoder to stop
    d.quit <- true

    // Clean pool
    d.pool = make(map[int]*worker)
}

func (d *Decoder) dispatch() {
    for {
        select {
        case work := <- d.workQueue:
            // Dispatch work to available worker
            go func() {
                // Fetch worker queue from pool
                workerQueue := <- d.poolChannel
                // Send work to worker
                workerQueue <- work
            }()
        case <- d.quit:
            return
        }
    }
}

// Collect completed work
func (d *Decoder) Collect() chan *Data {

    wg := new(sync.WaitGroup)
    out := make(chan *Data)

    wg.Add(d.maxCores)
    for _, worker := range d.pool {
        go collect(worker.responseChannel(), out, wg)
    }

    // Start goroutine to close out channel
    // when all collectors have been stopped
    // when alll connections are closed
    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}

func collect(c <-chan *Data, out chan<- *Data, wg *sync.WaitGroup) {

    // Collect completed work
    for r := range c {
        out <- r
    }

    // Channel has been closed
    // Signal Done
    wg.Done()
}
// EOF
