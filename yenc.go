package yenc

import (
    "bufio"
    "bytes"
    //"fmt"
    "io"
    "hash/crc32"
)

// Decode yEnc data
func Decode(b []byte) ([]byte, *Meta, error) {

    var err error
    //data := make([]byte, 0)
    //crc := crc32.NewIEEE()
    br := bufio.NewReader(bytes.NewReader(b))

    // Read Header
    // h == YBegin Header
    m := &Meta {
        Header: &yBegin{},
        Footer: &yEnd{},
        hash: crc32.NewIEEE(),
    }
    m, err = parseMeta(br, m, yBeginID, parseHeader)
    if err != nil {
        return nil, nil, err
    }

    // If multipart, then h.Part > 0
    if m.Header.Part > 0 {
        // Parse Part Header
        m, err = parseMeta(br, m, yPartID, parsePartHeader)
        if err != nil {
            return nil, nil, err
        }
    }

    return decode(br, m)
}

// DecodeData *yenc.Data
func DecodeData(d *Data) *Data {
    d.Content, d.Meta, d.Error = Decode(d.Content)
    return d
}

func decode(r *bufio.Reader, m *Meta) ([]byte, *Meta, error) {

    var data []byte
    var line []byte
    var err error

    for {
        line, err = r.ReadBytes('\n')
        if err != nil {
            if err != io.EOF {
                return nil, m, err
            }
        }

        // Strip <CR><LF>
        line = bytes.TrimRight(line, "\r\n")

        // Check for =yend marker line
        if len(line) >= len(yEndID) && string(line[:len(yEndID)]) == yEndID {
            m, err = parseFooter(m, string(line), len(yEndID))
            return data, m, err
        }

        // Decode current line
        b, _, err := decodeLine(line)
        if err != nil {
            if err != io.EOF {
                return nil, m, err
            }
        }

        // Update hash
        m.hash.Write(b)
        if m.Part != nil {
            m.Part.hash.Write(b)
        }

        data = append(data, b...)
    }
}

func decodeLine(line []byte) (b []byte, n int, err error) {

    br := bufio.NewReader(bytes.NewReader(line))
    for n < len(line) {
        var token byte
        token, err = br.ReadByte()
        if err != nil {
            return
        }

        // Skip newline
        if token == '\n' {
            continue
        }

        // yEnc Escape char
        if token == '=' {
            // Get next byte
            token, err = br.ReadByte()
            if err != nil {
                return
            }

            // Parse
            token -= 64
        }

        var c byte
        c = token - 42
        b = append(b, c)
        n++
    }

    return
}

// EOF
