package main

import (
	"bufio"
	"io"
)

type Probe []string

/*
func (p *Probe) decompose() (*HitSet, error) {
	// 1. Parse
	// 2. decompose into logEntry
	// 3. append to hitset
}

 */

func NewProbe(s *Session) (*Probe, error) {

	// Check file size to see if it is necessary to read
	stat, err := s.dataFile.Stat()
	if err != nil {
		// TODO better error handling
		return nil, err
	}

	size := stat.Size()
	delta := size - s.offset

	var lines []string

	// Only if there's a delta, continue reading
	if delta > 0 {
		lines = getNewLines(s.reader)
	}

	s.offset = size
	s.alertWatcher.push(len(lines))

	//return &Probe(lines), nil
	return nil, nil
}


func getNewLines(reader *bufio.Reader) []string {

	var line string
	var err error
	var lines []string

	for {
		line, err = reader.ReadString('\n')

		if err != nil {
			// ReadString returns err != nil iff line not ending with delim :
			// - thus not totally written by writer, consider EOF
			// or hits EOF
			if err == io.EOF {
				// TODO : what about blank lines containing only '\n' ?
				lines = append(lines, line)
			}

			// In both cases, we return the buffer of lines
			return lines
		}

		lines = append(lines, line)
	}

	// Code won't reach here
}