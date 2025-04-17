package main

import (
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	LoggerInit()
	EnvInit()
	WaInit()

	// Listen to Ctrl+C interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	meow.Disconnect()
}
