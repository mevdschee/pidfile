package pidfile

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Pidfile represents the name
type Pidfile struct {
	Signal   syscall.Signal
	AppId    string
	FullPath string
	FirstPid int
	OnSecond func([]string)
}

// New creates a Pidfile instance based on the application ID
func New(appId string) *Pidfile {
	return &Pidfile{Signal: syscall.SIGTERM, AppId: appId, FullPath: fullPath(appId)}
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
	// if OnSecond is set, listen for SIGTERM and call it
	if pf.OnSecond != nil {
		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, pf.Signal)
			for {
				<-c
				_, err := os.Stat(pf.FullPath + ".args")
				if errors.Is(err, os.ErrNotExist) {
					os.Exit(1)
				}
				args, _ := os.ReadFile(pf.FullPath + ".args")
				os.Remove(pf.FullPath + ".args")
				pf.OnSecond(strings.Split(string(args), "\b"))
			}
		}()
	}
	// check if the file exists
	pf.FirstPid = os.Getpid()
	pid := os.Getpid()
	pids := fmt.Sprint(pid)
	f, err := os.OpenFile(pf.FullPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err == nil {
		f.WriteString(pids)
		f.Sync()
		f.Close()
		return nil
	}
	if !os.IsExist(err) {
		return err
	}
	time.Sleep(time.Millisecond * 100)
	pidb, err := os.ReadFile(pf.FullPath)
	if err != nil {
		return err
	}
	pids = string(pidb)
	pid, err = strconv.Atoi(pids)
	if err != nil {
		return err
	}
	process := findProcess(pid)
	if process != nil {
		pf.FirstPid = pid
		if pf.OnSecond != nil {
			var file *os.File
			for i := 0; i < 10; i++ {
				file, err = os.OpenFile(pf.FullPath+".args.lock", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
				if err == nil {
					defer func() {
						file.Close()
						os.Remove(pf.FullPath + ".args.lock")
					}()
					break
				}
				time.Sleep(time.Millisecond * 100)
			}
			os.WriteFile(pf.FullPath+".args", []byte(strings.Join(os.Args, "\b")), 0644)
			time.Sleep(time.Millisecond * 100)
			process.Signal(pf.Signal)
			for i := 0; i < 10; i++ {
				_, err = os.Stat(pf.FullPath + ".args")
				if os.IsNotExist(err) {
					break
				}
			}
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
	if pf.OnSecond != nil {
		os.Remove(pf.FullPath + ".args.lock")
		os.Remove(pf.FullPath + ".args")
	}
	return nil
}
