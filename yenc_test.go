package yenc_test

import (
    "io/ioutil"
    "os"
    "path/filepath"
    "reflect"
    "testing"

    "github.com/GJRTimmer/yenc"
)

var (
    s single
    m multi
)

type single struct {
    Encoded string
    Decoded string
}

type multi struct {
    Encoded []string
    Decoded string
}

func init() {
    s = single {
        Encoded: filepath.Join(".", "test", "single", "00000005.ntx"),
        Decoded: filepath.Join(".", "test", "single", "testfile.txt"),
    }

    m = multi {
        Encoded: []string{filepath.Join(".", "test", "multi", "00000020.ntx"), filepath.Join(".", "test", "multi", "00000021.ntx")},
        Decoded: filepath.Join(".", "test", "multi", "joystick.jpg"),
    }
}

func TestMain(m *testing.M) {
    code := m.Run()
    os.Exit(code)
}

// TestHeaders .
func TestSingleHeader(t *testing.T) {

    data, err := ioutil.ReadFile(s.Encoded)
    if err != nil {
        t.Fatal(err)
    }
    
    _, meta, err := yenc.Decode(data)
    if err != nil {
        t.Fatal(err)
    }

    if meta.Header.Line != 128 {
        t.Fatal("expecting value '128' for yEnc attribute '=ybegin#line'")
    }

    if meta.Header.Size != 584 {
        t.Fatal("expecting value '584' for yEnc attribute =yBegin#size")
    }

    if meta.Header.Name != "testfile.txt" {
        t.Fatal("expectiing value 'testfile.txt' for yEnc attribute =begin#name")
    }

    // TODO: Vallidate end header
}

func TestMultiHeader(t *testing.T) {

    for i, e := range m.Encoded {

        data, err := ioutil.ReadFile(e)
        if err != nil {
            t.Fatal(err)
        }

        _, meta, err := yenc.Decode(data)
        if err != nil {
            t.Fatal(err)
        }

        if meta.Header.Line != 128 {
            t.Fatal("expecting value '128' for yEnc attribute '=ybegin#line'")
        }

        if meta.Header.Size != 19338 {
            t.Fatal("expecting value '19338' for yEnc attribute '=ybegin#size'")
        }

        if meta.Header.Name != "joystick.jpg" {
            t.Fatal("expectiing value 'joystick.jpg' for yEnc attribute =begin#name")
        }

        if meta.Header.Part != uint(i + 1) {
            t.Fatalf("expecting value '%d' for yEnc attribute '=ybegin#part'", i + 1)
        }

        switch i {
        case 0:
            if meta.Part.Begin != 1 {
                t.Fatalf("expecting value '1' for yEnc attribute '=ypart-%d#begin'", i + 1)
            }
            if meta.Part.End != 11250 {
                t.Fatalf("expecting value '11250' for yEnc attribute '=ypart-%d#begin'", i + 1)
            }
        case 1:
            if meta.Part.Begin != 11251 {
                t.Fatalf("expecting value '11251' for yEnc attribute '=ypart-%d#end'", i + 1)
            }
            if meta.Part.End != 19338 {
                t.Fatalf("expecting value '19338' for yEnc attribute '=ypart-%d#end'", i + 1)
            }
        }
    }


    // Validate END Header
}

func TestSingleDecode(t *testing.T) {

    source, err := ioutil.ReadFile(s.Encoded)
    if err != nil {
        t.Fatal(err)
    }

    verify, err := ioutil.ReadFile(s.Decoded)
    if err != nil {
        t.Fatal(err)
    }

    data, _, err := yenc.Decode(source)
    if err != nil {
        t.Fatal(err)
    }

    if !reflect.DeepEqual(data, verify) {
        t.Fatal("decoded data does not match")
    }

    // Validate CRC
}

func TestMultiDecode(t *testing.T) {

    var result []byte
    for _, e := range m.Encoded {

        source, err := ioutil.ReadFile(e)
        if err != nil {
            t.Fatal(err)
        }

        data, _, err := yenc.Decode(source)
        if err != nil {
            t.Fatal(err)
        }

        result = append(result, data...)
    }

    verify, err := ioutil.ReadFile(m.Decoded)
    if err != nil {
        t.Fatal(err)
    }

    if !reflect.DeepEqual(result, verify) {
        t.Fatal("decoded data does not match")
    }
}

// EOF
