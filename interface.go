//command is a goroutine that allows an operator to interact with the tool through CLI.
//
//Implemented Commands :
//- stop
package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

// command handles CLI interactions
func command(syn *Sync) {

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for sig := range sigs {
		log.Info("command received signal :", sig.String())
		for n := 1; n < syn.nbReceivers; n++ {
			syn.syncChan <- struct{}{}
		}
		break
	}

	log.Info("command terminating.")
	syn.wg.Done()
}
