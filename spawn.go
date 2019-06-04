package liblpc

import (
	"os"
	"os/exec"
)

func Spawn(exePath string, extraFd int) (*exec.Cmd, error) {
	cmd := &exec.Cmd{
		Path:       exePath,
		Stdin:      os.Stdin,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		ExtraFiles: []*os.File{os.NewFile(uintptr(extraFd), "")},
	}
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	return cmd, nil
}
