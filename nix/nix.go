package nix

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func SetupSignalHandler() (stopCh <-chan struct{}) {
	stop := make(chan struct{})
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		close(stop)
		for i := 3; i > 0; i-- {
			log.Printf("There are %ds left to exit", i)
			time.Sleep(1 * time.Second)
		}
		os.Exit(1)
	}()
	return stop
}
