package config

import (
	"os"
	"os/signal"
	"syscall"
)

// WaitSIGHUP blocks until a SIGHUP signal is received.
func WaitSIGHUP() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Signal(syscall.SIGHUP))
	<-ch
}
