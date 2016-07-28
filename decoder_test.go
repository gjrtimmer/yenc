package yenc_test

import (
    "io/ioutil"
    "reflect"
    "runtime"
    "sync"
    "testing"
    "time"

    "github.com/GJRTimmer/yenc"
)

func TestDecoderMaxCores(t *testing.T) {

    cpu := runtime.NumCPU()
    _, err := yenc.NewDecoder(cpu * 2, nil)
    if err == nil {
        t.Fatal("max cores restriction failed")
    }
}

func TestDecoder(t *testing.T) {

    workQueue := make(chan *yenc.Data)
    d, err := yenc.NewDecoder(4, workQueue)
    if err != nil {
        t.Fatal(err)
    }
    d.Start()

    completionTicker := time.NewTicker(1 * time.Second)
    wg := new(sync.WaitGroup)
    responses := d.Collect()

    results := make([]*yenc.Data, 2)

    wg.Add(1)
    go func(wg *sync.WaitGroup) {
        defer func() {
            wg.Done()
        }()
        for {
            select {
            case r := <- responses:
                results[r.Meta.Header.Part - 1] = r
            case <- completionTicker.C:
                if len(results) == 2 {
                    return
                }
            }
        }
    }(wg)

    // Send Work
    for _, e := range m.Encoded {

        source, err := ioutil.ReadFile(e)
        if err != nil {
            t.Fatal(err)
        }

        workQueue <- &yenc.Data {
            Content: source,
        }
    }

    wg.Wait()

    resultBytes := make([]byte, 0)
    for _, d := range results {
        if d.Error != nil {
            t.Fatal(err)
        }

        resultBytes = append(resultBytes, d.Content...)
    }

    verify, err := ioutil.ReadFile(m.Decoded)
    if err != nil {
        t.Fatal(err)
    }

    if !reflect.DeepEqual(resultBytes, verify) {
        t.Fatal("decoded data does not match")
    }
}

// EOF
