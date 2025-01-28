package util

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func SetupSignalHandler() (stopCh <-chan struct{}) {
	stop := make(chan struct{})
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		close(stop)
		// for i := 3; i > 0; i-- {
		// 	log.Printf("There are %ds left to exit", i)
		// 	time.Sleep(1 * time.Second)
		// }
		os.Exit(1)
	}()
	return stop
}

const maxLogStrLen = 80

func TruncateLogStr(text string) string {
	ellipsis := "..."
	maxLen := maxLogStrLen - len(ellipsis)
	if len(text) > maxLogStrLen {
		text = text[:maxLen] + "..."
	}
	return strconv.Quote(text)
}
