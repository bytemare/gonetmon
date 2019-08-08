//command is a goroutine that allows an operator to interact with the tool through CLI.
//
//Implemented Commands :
//- stop
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// command handles CLI interactions
func command(syncChan chan<- struct{}, nbReceivers int) {

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for sig := range sigs {
		fmt.Println("command received signal :", sig.String())
		for n := 0; n < nbReceivers; n++ {
			syncChan <- struct{}{}
		}
		break
	}

	fmt.Print("command terminating\n")
}
