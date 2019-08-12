//command is a goroutine that allows an operator to interact with the tool through CLI.
//
//Implemented Commands :
//- stop
package main

import (
	"io"
	"os"
	"os/signal"
	"syscall"
)

// command handles CLI interactions
func command(syn *Sync) {
	defer syn.wg.Done()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for sig := range sigs {
		log.Info("Command received signal :", sig.String())
		// This Goroutine is not waiting for a stop signal/message, so we take one off

		log.SetOutput(io.MultiWriter(os.Stdout, log.Out))
		log.Info("Logging to both file and console.")

		for n := 1; n < int(syn.nbReceivers); n++ {
			syn.syncChan <- struct{}{}
		}
		break
	}

	log.Info("Command terminating.")
}
