package main

import (
	"fmt"
	pwd "github.com/image357/password"
	"github.com/image357/password/log"
	"github.com/image357/password/rest"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// setup logging
	err := log.SetMultiJSON("log.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Level(slog.LevelDebug)

	// enable recovery
	pwd.EnableRecovery("recovery_key")

	// start rest service
	err = rest.StartMultiService(":8080", "/prefix", "storage_key", rest.DebugAccessCallback)
	if err != nil {
		fmt.Println(err)
		return
	}

	// wait for signal term or interrupt
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	sig := <-cancelChan
	fmt.Printf("Caught signal %v\n", sig)

	// stop rest service
	err = rest.StopService(5000, ":8080", "/prefix")
	if err != nil {
		fmt.Println(err)
	}
}
