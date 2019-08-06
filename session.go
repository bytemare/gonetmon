// Session object holds the current state of a monitoring session
package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)


type Session struct {
	alertWatcher 	*HitQueue
	hits			[]*HitSet
	report   		*Report
	dataFile 		*os.File
	reader	 		*bufio.Reader
	offset	 		int64
}

func (s *Session) isAlert() bool {
	return s.alertWatcher.alert
}

func (s *Session) addHits(h *HitSet) {
	s.hits = append(s.hits, h)
}

func (s *Session) SendReport(t *time.Time) *Report {
	// TODO : analyse hitset, generate report and return it

}

func NewSession(parameters *Parameters) *Session {

	var err error

	file, err := os.Open(parameters.CollectorFile)
	if err != nil {
		// Todo : handle error
		fmt.Println("PANIC : NewSession could not open collector file")
		return nil
	}

	reader := bufio.NewReader(file)

	// Track the file size during monitoring to avoid read conflicts
	stat, err := file.Stat()
	if err != nil {
		// TODO : handle this error
		panic(err)
	}
	offset := stat.Size()

	/* Necessary ?
	if _, err := file.Seek(offset, 0); err != nil {
		panic(err)
	}
	*/

	return &Session{
		alertWatcher: NewHitQueue(parameters.AlertSpan/parameters.ProbePeriod, parameters.AlertThreshold),
		dataFile:     file,
		reader:       reader,
		offset:       offset,
	}
}

func (s *Session) Close() {
	if s.dataFile != nil {
		err := s.dataFile.Close()
		if err != nil {
			//TODO : handle error
		}
	}
}

