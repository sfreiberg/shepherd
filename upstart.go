package main

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
)

type Upstart struct {
	Name   string
	Action string // start, stop, restart
	Status int    // starting, started, stopped, stopping
	Result
}

func NewUpstart(name, action string) *Upstart {
	return &Upstart{Name: name, Action: action}
}

func (u *Upstart) Run() {
	u.UpdateStatus()

	switch u.Action {
	case "start":
		if err := u.Start(); err != nil {
			u.Result.Error = err.Error()
			return
		}
	case "restart":
		if err := u.Restart(); err != nil {
			u.Result.Error = err.Error()
			return
		}
	case "stop":
		if err := u.Stop(); err != nil {
			u.Result.Error = err.Error()
			return
		}
	case "reload":
		if err := u.Reload(); err != nil {
			u.Result.Error = err.Error()
			return
		}
	}

	u.Result.Success = true
	return
}

func (u *Upstart) Start() error {
	if u.Status == SvcUnset {
		u.UpdateStatus()
	}
	switch u.Status {
	case SvcStopped:
		cmd := exec.Command("/sbin/start", u.Name)
		_, err := cmd.Output()
		if err != nil {
			return err
		}
		u.Result.Changed = true
	case SvcStopping:
		return errors.New(fmt.Sprintf("Upstart: service %v is still stopping", u.Name))
	case SvcStarting:
		return errors.New(fmt.Sprintf("Upstart: service %v is still starting", u.Name))
	}
	return nil
}

func (u *Upstart) Restart() error {
	if u.Status == SvcUnset {
		u.UpdateStatus()
	}

	switch u.Status {
	case SvcStarted:
		cmd := exec.Command("/sbin/restart", u.Name)
		_, err := cmd.Output()
		if err != nil {
			return err
		}
		u.Result.Changed = true
	case SvcStopped:
		cmd := exec.Command("/sbin/start", u.Name)
		_, err := cmd.Output()
		if err != nil {
			return err
		}
		u.Result.Changed = true
	case SvcStopping:
		return errors.New(fmt.Sprintf("Upstart: service %v is still stopping", u.Name))
	case SvcStarting:
		return errors.New(fmt.Sprintf("Upstart: service %v is still starting", u.Name))
	}

	return nil
}

func (u *Upstart) Reload() error {
	if u.Status == SvcUnset {
		u.UpdateStatus()
	}

	switch u.Status {
	case SvcStarted:
		cmd := exec.Command("/sbin/reload", u.Name)
		_, err := cmd.Output()
		if err != nil {
			return err
		}
		u.Result.Changed = true
	case SvcStopping:
		return errors.New(fmt.Sprintf("Upstart: service %v is still stopping", u.Name))
	case SvcStarting:
		return errors.New(fmt.Sprintf("Upstart: service %v is still starting", u.Name))
	}

	return nil
}

func (u *Upstart) Stop() error {
	if u.Status == SvcUnset {
		u.UpdateStatus()
	}

	switch u.Status {
	case SvcStarted:
		cmd := exec.Command("/sbin/restart", u.Name)
		_, err := cmd.Output()
		if err != nil {
			return err
		}
		u.Result.Changed = true
	case SvcStopping:
		return errors.New(fmt.Sprintf("Upstart: service %v is still stopping", u.Name))
	case SvcStarting:
		return errors.New(fmt.Sprintf("Upstart: service %v is still starting", u.Name))
	}

	return nil
}

func (u *Upstart) UpdateStatus() {
	cmd := exec.Command("/sbin/status", u.Name)
	status, err := cmd.CombinedOutput()
	// error is most likely for non-existant svc so we're ignoring
	if err != nil {
		u.Status = SvcUnknown
		return
	}
	if matched, _ := regexp.Match("/running", status); matched {
		u.Status = SvcStarted
	} else if matched, _ := regexp.Match("/starting|/pre-start|/post-start", status); matched {
		u.Status = SvcStarting
	} else if matched, _ := regexp.Match("/stopping|/pre-stop|/killed|/post-stop", status); matched {
		u.Status = SvcStopping
	} else if matched, _ := regexp.Match("/waiting", status); matched {
		u.Status = SvcStopped
	} else {
		u.Status = SvcUnknown
	}
}
