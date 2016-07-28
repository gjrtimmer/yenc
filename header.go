package yenc

import (
    "bufio"
    "fmt"
    "hash"
    "hash/crc32"
    "strconv"
    "strings"
)

const (
    yBeginID string   = "=ybegin"
    yPartID string    = "=ypart"
    yEndID string     = "=yend"
)

// Meta of yEnc
type Meta struct {
    Header *yBegin
    Part *yPart
    Footer *yEnd

    hash hash.Hash32
}

// YBegin Header
type yBegin struct {
    Name string
    Size Size
    Line uint
    Part uint
}

// YPart Header
type yPart struct {
    Begin uint64
    End uint64
    hash hash.Hash32
}

// YEnd Header
type yEnd struct {
    Size Size
    Part uint
}

func parseMeta(r *bufio.Reader, m *Meta, headerID string, parser func(m *Meta, line string, start int) (*Meta, error)) (*Meta, error) {
    var line string
    var err error

    for {
        line, err = r.ReadString('\n')
        if err != nil {
            return nil, err
        }

        // Locate the header provided by argument 'header'
        if len(line) > len(headerID) && line[:len(headerID)] == headerID {
            break // Current line is header line, start parsing
        }
    }

    return parser(m, line, len(headerID))
}

func parseHeader(m *Meta, line string, start int) (*Meta, error) {

    // Parse header line into key=val blocks
    headerBlocks := strings.Split(line[start:], " ")
    for i := range headerBlocks {
        // Get Key=Val from header blocks
        kv := strings.Split(strings.TrimSpace(headerBlocks[i]), "=")
        if len(kv) < 2 {
            // unknown additional yEnc key
            // Only allow key=val pairs
            continue
        }

        // Process keys
        switch strings.ToLower(kv[0]) {
            // =ybegin
        case "name":
            m.Header.Name = strings.TrimSpace(kv[1])
        case "size":
            s, _ := strconv.ParseUint(strings.TrimSpace(kv[1]), 10, 64)
            switch start {
            case len(yBeginID):
                m.Header.Size = Size(s)
            case len(yEndID):
                m.Footer.Size = Size(s)
            }
        case "part":
            p, _ := strconv.ParseUint(strings.TrimSpace(kv[1]), 10, 32)
            switch start {
            case len(yBeginID):
                m.Header.Part = uint(p)
            case len(yEndID):
                m.Footer.Part = uint(p)
            }
        case "line":
            l, _ := strconv.ParseUint(strings.TrimSpace(kv[1]), 10, 32)
            m.Header.Line = uint(l)
        case "total":
            // TODO: Not implemented
        default:
            return nil, fmt.Errorf("unknown yEnc attribute: %s", kv[0])
        }
    }

    return m, nil
}

func parsePartHeader(m *Meta, line string, start int) (*Meta, error) {

    m.Part = &yPart{
        hash: crc32.NewIEEE(),
    }

    // Parse header line into key=val blocks
    headerBlocks := strings.Split(line[start:], " ")
    for i := range headerBlocks {
        // Get Key=Val from header blocks
        kv := strings.Split(strings.TrimSpace(headerBlocks[i]), "=")
        if len(kv) < 2 {
            // unknown additional yEnc key
            // Only allow key=val pairs
            continue
        }

        // Process keys
        switch strings.ToLower(kv[0]) {
            // =ypart
        case "begin":
            m.Part.Begin, _ = strconv.ParseUint(strings.TrimSpace(kv[1]), 10, 64)
        case "end":
            m.Part.End, _ = strconv.ParseUint(strings.TrimSpace(kv[1]), 10, 64)
        default:
            return nil, fmt.Errorf("unknown yEnc attribute: %s", kv[0])
        }
    }

    return m, nil
}

func parseFooter(m *Meta, line string, start int) (*Meta, error) {

    // Parse header line into key=val blocks
    headerBlocks := strings.Split(line[start:], " ")
    for i := range headerBlocks {
        // Get Key=Val from header blocks
        kv := strings.Split(strings.TrimSpace(headerBlocks[i]), "=")
        if len(kv) < 2 {
            // unknown additional yEnc key
            // Only allow key=val pairs
            continue
        }

        // Process keys
        switch strings.ToLower(kv[0]) {
        case "size":
            s, _ := strconv.ParseUint(strings.TrimSpace(kv[1]), 10, 64)
            m.Footer.Size = Size(s)
        case "crc32":
            // Validate CRC for Single yEnc CRC
            if crc64, err := strconv.ParseUint(strings.TrimSpace(kv[1]), 16, 64); err == nil {
                if uint32(crc64) != m.hash.Sum32() {
                    return m, fmt.Errorf("CRC check failed (header crc: %x) != (decoded hash: %x)", uint32(crc64), m.hash.Sum32())
                }
            }
        case "pcrc32":
            // Validate CRC for yEnc Part
            if crc64, err := strconv.ParseUint(strings.TrimSpace(kv[1]), 16, 64); err == nil {
                if uint32(crc64) != m.Part.hash.Sum32() {
                    return m, fmt.Errorf("CRC check failed (header crc: %x) != (decoded hash: %x)", uint32(crc64), m.Part.hash.Sum32())
                }
            }
        }
    }

    return m, nil
}

// EOF
