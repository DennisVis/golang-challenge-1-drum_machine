package drum

import (
	"fmt"
	"bytes"
	"strconv"
	"strings"
	"encoding/binary"
	"os"
	"io"
)

const (
	header = "SPLICE"
	nrSteps = 16
)

// track is the representation of a single instrument
// within a Pattern.
type track struct {
	id    uint8
	name  string
	steps [nrSteps]bool
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   float32
	tracks  []track
}

func (p *Pattern) String() string {
	var tBuff bytes.Buffer
	for i := range p.tracks {
		t := p.tracks[i]

		tBuff.WriteString("\n(")
		tBuff.WriteString(strconv.Itoa(int(t.id)))
		tBuff.WriteString(")")
		tBuff.WriteString(" ")
		tBuff.WriteString(t.name)
		tBuff.WriteString("\t")

		var sBuff bytes.Buffer
		for j := range t.steps {
			if j % 4 == 0 {
				sBuff.WriteString("|")
			}

			if t.steps[j] {
				sBuff.WriteString("x")
			} else {
				sBuff.WriteString("-")
			}
		}
		sBuff.WriteString("|")

		tBuff.WriteString(sBuff.String())
	}

	tempo := float64(p.tempo)
	roundedTempo := float64(int64(tempo / 0.5 + 0.5)) * 0.5
	tempoString := strconv.Itoa(int(p.tempo))
	if roundedTempo != tempo {
		tempoString = strconv.FormatFloat(roundedTempo, 'f', 1, 32)
	}

	return fmt.Sprintf(
		"Saved with HW Version: %s\nTempo: %s%s\n",
		p.version,
		tempoString,
		tBuff.String(),
	)
}


// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read file for payt %s: %v", path, err)
	}

	p := Pattern{}

	var hdr [len(header)]byte
	if err := binary.Read(r, binary.BigEndian, &hdr); err != nil {
		return nil, fmt.Errorf("unable to decode splice header: %v", err)
	}

	if header != string(hdr[:]) {
		return nil, fmt.Errorf("decoded file header does not match %v", header)
	}

	var patSize int64
	if err := binary.Read(r, binary.BigEndian, &patSize); err != nil {
		return nil, fmt.Errorf("unable to decode patSize: %v", err)
	}

	var version [32]byte
	if err := binary.Read(r, binary.BigEndian, &version); err != nil {
		return nil, fmt.Errorf("unable to decode version: %v", err)
	}
	p.version = strings.TrimRight(string(version[:]), string(0))

	if err := binary.Read(r, binary.LittleEndian, &p.tempo); err != nil {
		return nil, fmt.Errorf("unable to decode tempo: %v", err)
	}

	for {
		track, err := decodeTrack(r)
		if (err != nil) {
			break
		}

		p.tracks = append(p.tracks, *track)
	}

	return &p, nil
}

// decodeTrack uses the given Reader to read and use the bytes that make up a track.
// It returns a pointer to the newly created track.
func decodeTrack(r io.Reader) (*track, error) {
	t := track{}

	if err := binary.Read(r, binary.LittleEndian, &t.id); err != nil {
		return nil, fmt.Errorf("unable to decode track id: %v", err)
	}

	var nameLength int32
	if err := binary.Read(r, binary.BigEndian, &nameLength); err != nil {
		return nil, fmt.Errorf("unable to decode track name length: %v", err)
	}

	name := make([]byte, nameLength)
	if err := binary.Read(r, binary.BigEndian, &name); err != nil {
		return nil, fmt.Errorf("unable to decode track name: %v", err)
	}
	t.name = string(name[:])

	if err := binary.Read(r, binary.LittleEndian, &t.steps); err != nil {
		return nil, fmt.Errorf("unable to decode track pattern: %v", err)
	}

	return &t, nil
}
