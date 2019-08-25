package gonetmon

import (
	"io"
	"os"
	"os/signal"
	"syscall"
)

// CLI acts as a command interface that allows an operator to interact with the tool through CLI.
//
// Implemented commands :
// - stop : through SIGINT or SIGTERM signals
func CLI(syn *synchronisation) {
	defer syn.wg.Done()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for sig := range sigs {
		log.Info("CLI received signal :", sig.String())
		log.SetOutput(io.MultiWriter(os.Stdout, log.Out))
		log.Info("Shutting down.")
		log.Info("Logging to both file and console.")

		// This Goroutine is not waiting for a stop signal/message, so we take one off
		for n := 1; n < int(syn.nbReceivers); n++ {
			syn.syncChan <- struct{}{}
		}
		break
	}

	log.Info("CLI terminating.")
}
