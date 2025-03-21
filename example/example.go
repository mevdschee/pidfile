package main

import (
	"log"
	"os"

	"github.com/mevdschee/pidfile"
)

func main() {

	// create pidfile struct based on identifier
	pf := pidfile.New("app_identifier")
	// when a second instance is started
	pf.OnSecond = func(args []string) {
		log.Printf("another instance was started")
	}
	// create pidfile on application start
	err := pf.Create()
	if err != nil {
		log.Fatalf("can't create pidfile: %v", err)
	}
	// remove pidfile on application close
	defer pf.Remove()
	// if this is not the first instance, then close it
	if pf.FirstPid != os.Getpid() {
		return
	}

	// application code
}
