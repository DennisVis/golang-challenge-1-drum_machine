package drum

import (
	"io/ioutil"
	"fmt"
	"bytes"
	"strconv"
	"strings"
)

func decodeSteps(bytes []byte) []bool {
	steps := make([]bool, len(bytes))
	for i := range bytes {
		b := bytes[i]
		if int(b) > 0 {
			steps[i] = true
		} else {
			steps[i] = false
		}
	}
	return steps
}

func decodeTrack(bytes []byte) (track, []byte) {
	nameLengthIndex := 4
	nameLength := int(bytes[nameLengthIndex])
	totalLength := 5 + nameLength + 16
	nameIndex := nameLengthIndex + 1

	id := strconv.Itoa(int(bytes[0]))
	name := string(bytes[nameIndex:nameIndex + nameLength])
	steps := decodeSteps(bytes[nameIndex + nameLength:totalLength])
	track := track{id, name, steps}

	if totalLength > len(bytes) {
		return track, nil
	}

	return track, bytes[totalLength:]
}

func decodeTracks(bytes []byte, tracks []track) []track {
	track, rest := decodeTrack(bytes)
	if rest != nil {
		tracks = append(tracks, track)

		if len(rest) > 0 {
			return decodeTracks(rest, tracks)
		}
	}
	return tracks
}

func decodeVersion(bytes []byte) string {
	versionBytes := make([]byte, 0)
	startIndex := 14
	i := 0
	for {
		b := bytes[startIndex + i]
		if int(b) == 0 {
			break
		}
		versionBytes = append(versionBytes, b)
		i++
	}
	return string(versionBytes)
}

func decodeTempo(bytes []byte, version string) string {
	tempo := float64(bytes[48])

	if strings.Contains(version, "909") {
		return "240"
	} else if strings.Contains(version, "708") {
		return "999"
	}

	tempo = tempo / 2.0
	if tempo < 100.0 {
		return strconv.FormatFloat(tempo + 0.5, 'f', 1, 64)
	}

	return strconv.FormatFloat(tempo, 'f', 0, 64)
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
func DecodeFile(path string) (*Pattern, error) {
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	version := decodeVersion(fileBytes)
	tempo := decodeTempo(fileBytes, version)
	tracks := decodeTracks(fileBytes[50:], make([]track, 0))

	p := &Pattern{version, tempo, tracks}

	return p, nil
}

type track struct {
	id    string
	name  string
	steps []bool
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
// TODO: implement
type Pattern struct {
	version string
	tempo   string
	tracks  []track
}

func (p *Pattern) String() string {
	var tracksBuffer bytes.Buffer
	for i := range p.tracks {
		track := p.tracks[i]

		tracksBuffer.WriteString("\n(")
		tracksBuffer.WriteString(track.id)
		tracksBuffer.WriteString(")")
		tracksBuffer.WriteString(" ")
		tracksBuffer.WriteString(track.name)
		tracksBuffer.WriteString("\t")

		var stepsBuffer bytes.Buffer
		for j := range track.steps {
			if j % 4 == 0 {
				stepsBuffer.WriteString("|")
			}
			if track.steps[j] {
				stepsBuffer.WriteString("x")
			} else {
				stepsBuffer.WriteString("-")
			}
		}
		stepsBuffer.WriteString("|")

		tracksBuffer.WriteString(stepsBuffer.String())
	}

	return fmt.Sprintf("Saved with HW Version: %s\nTempo: %s%s\n", p.version, p.tempo, tracksBuffer.String())
}
