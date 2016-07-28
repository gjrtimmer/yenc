package yenc

// Content .
type Content []byte

// Bytes of content
func (c Content) Bytes() []byte {
    return []byte(c)
}

// Data structure which holds yEnc data to be decoded by Decoder
type Data struct {
    Content Content
    Meta *Meta
    Error error
}

type worker struct {
    id          int
    workQueue   chan *Data // Queue to get work which needs to be processed
    respChan    chan *Data // Response Channel for completed work
    pool        chan chan *Data // Worker pool
    quit        chan bool //quit channel to shutdown the worker
}

func (w *worker) start() {
    go func() {
        for {
            // Add worker data queue to
            // the worker pool
            w.pool <- w.workQueue

            select {
            case work := <- w.workQueue:
                // Process incoming work and return over response channel
                w.respChan <- DecodeData(work)
            case <- w.quit:
                // Stop work
                return
            }
        }
    }()
}

func (w *worker) stop() {
    go func() {
        // Stop worker go routine
        w.quit <- true

        // Close response channel
        close(w.respChan)
    }()
}

func (w *worker) responseChannel() chan *Data {
    return w.respChan
}

// Defined within work.go because it's easy to mis
// changes made to worker struct
func (d *Decoder) newWorker(id int, pool chan chan *Data) *worker {
    return &worker {
        id: id,
        workQueue: make(chan *Data),
        respChan: make(chan *Data),
        pool: pool,
        quit: make(chan bool),
    }
}
// EOF
