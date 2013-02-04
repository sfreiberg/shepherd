package actions

import (
	"os/exec"
)

type Command struct {
	Command string
	Result
}

func NewCommand(cmd string) *Command {
	return &Command{Command: cmd}
}

func (c *Command) Run() {
	cmd := exec.Command("/bin/sh", "-c", c.Command)

	out, err := cmd.CombinedOutput()
	c.Result.Output = string(out)
	if err != nil {
		c.Result.Error = err.Error()
		return
	}

	c.Result.Success = true
	return
}
