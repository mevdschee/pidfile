# Pidfile package for Go

The `pidfile` package provides a simple way to ensure that only one instance of a Go application runs at any given time by using a PID (Process ID) file.

## Features

- Detects that the an instance of the application is already running.
- Allows you to prevents multiple instances of the application from running concurrently.
- Signals the first instance of the application when a next instance is started.

## Installation

To install the `pidfile` package, use the following `go get` command:

    go get github.com/mevdschee/pidfile

## Example

Here is a full example of an application that uses the pidfile package to ensure single instance execution

```go
package main

import (
	"log"
	"os"

	"github.com/mevdschee/pidfile"
)

func main() {

    // create PID file struct based on identifier
    pf := pidfile.New("app_identifier")
    // when a second instance is started
    pf.OnSecond = func() {
		log.Println("another instance was started")
	}
    // create PID file on application start
    err := pf.Create();
    if err != nil {
        log.Fatalf("can't create PID file: %v", err)
    }
    // remove PID file on application close
    defer pf.Remove();
    // if this is not the first instance, then close it
    if pf.FirstPid != os.Getpid() {
        return
    } 
    
    // application code
}
```

