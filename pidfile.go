package pidfile

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
)

// Pidfile represents the name
type Pidfile struct {
	Signal   syscall.Signal
	AppId    string
	FullPath string
	FirstPid int
	OnSecond func()
}

// New creates a Pidfile instance based on the application ID
func New(appId string) *Pidfile {
	return &Pidfile{Signal: syscall.SIGUSR1, AppId: appId, FullPath: fullPath(appId)}
}

// fullPath returns an absolute filename, appropriate for the operating system
func fullPath(appId string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s.pid", appId))
}

func findProcess(pid int) *os.Process {
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}
	err = process.Signal(syscall.Signal(0))
	if err == syscall.ESRCH {
		// The process does not exist
		return nil
	}
	if err != nil {
		// Some other unexpected error
		return nil
	}
	// The process exists and is active
	return process
}

// Create creates a pidfile
func (pf *Pidfile) Create() error {
	// if OnSecond is set, listen for SIGUSR1 and call it
	if pf.OnSecond != nil {
		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, pf.Signal)
			for {
				<-c
				pf.OnSecond()
			}
		}()
	}
	// check if the file exists
	pf.FirstPid = os.Getpid()
	pids, err := os.ReadFile(pf.FullPath)
	if os.IsNotExist(err) {
		return os.WriteFile(pf.FullPath, []byte(fmt.Sprint(os.Getpid())), 0644)
	}
	if err != nil {
		return err
	}
	pid, err := strconv.Atoi(string(pids))
	if err != nil {
		return err
	}
	process := findProcess(pid)
	if process != nil {
		pf.FirstPid = pid
		if pf.OnSecond != nil {
			process.Signal(pf.Signal)
		}
		return nil
	}
	err = os.Remove(pf.FullPath)
	if err != nil {
		return err
	}
	err = os.WriteFile(pf.FullPath, []byte(fmt.Sprint(os.Getpid())), 0644)
	if err != nil {
		return err
	}
	return nil
}

// Remove removes a pidfile
func (pf *Pidfile) Remove() error {
	pid, err := os.ReadFile(pf.FullPath)
	if err != nil {
		return err
	}
	if string(pid) != fmt.Sprint(os.Getpid()) {
		return nil
	}
	err = os.Remove(pf.FullPath)
	if err != nil {
		return err
	}
	return nil
}
