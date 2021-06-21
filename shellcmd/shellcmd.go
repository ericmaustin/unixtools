package shellcmd

import (
	"os/exec"
	"strings"
)

type Builder struct {
	shell string
	shellArgs []string
	cmdSep string
	Commands []*exec.Cmd
	//cmd      *exec.Cmd
}

func (s *Builder) AddCmd(cmd ...*exec.Cmd) *Builder {
	for _, c := range cmd {
		s.Commands = append(s.Commands, c)
	}
	return s
}

func (s *Builder) SetCmdSep(sep string) *Builder {
	s.cmdSep = sep
	return s
}

func (s *Builder) Cmd() *exec.Cmd {
	cmdStrs := make([]string, len(s.Commands))
	for i, cmd := range s.Commands {
		cmdStrs[i] = cmd.String()
	}
	return exec.Command(s.shell, append(s.shellArgs, strings.Join(cmdStrs, s.cmdSep))...)
}

func NewBuilder(shell string, arg ...string) *Builder {
	return &Builder{
		shell:    shell,
		shellArgs: arg,
		cmdSep: "; ",
	}
}

func NewShellBuilder() *Builder {
	return &Builder{
		shell:    "/bin/sh",
		shellArgs: []string{"-c"},
		cmdSep: "; ",
	}
}

func NewBashBuilder() *Builder {
	return &Builder{
		shell:    "/usr/bin/bash",
		shellArgs: []string{"-c"},
		cmdSep: "; ",
	}
}