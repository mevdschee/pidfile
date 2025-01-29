package Pidfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

// Pidfile represents the name
type Pidfile struct {
	FullPath string
	FirstPid int
}

// New creates a Pidfile instance based on the application ID
func New(appId string) *Pidfile {
	return &Pidfile{FullPath: fullPath(appId)}
}

// fullPath returns an absolute filename, appropriate for the operating system
func fullPath(appId string) string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s.pid", appId))
}

func processExists(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	if err == syscall.ESRCH {
		// The process does not exist
		return false
	}
	if err != nil {
		// Some other unexpected error
		return false
	}
	// The process exists and is active
	return true
}

// Create creates a pidfile
func (pf *Pidfile) Create() error {
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
	if processExists(pid) {
		pf.FirstPid = pid
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
