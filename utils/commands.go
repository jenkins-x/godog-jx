package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func RunCommand(dir, name string, args ...string) error {
	e := exec.Command(name, args...)
	if dir != "" {
		e.Dir = dir
	}
	e.Stdout = os.Stdout
	e.Stderr = os.Stderr
	err := e.Run()
	if err != nil {
		fmt.Printf("Error: Command failed  %s %s\n", name, strings.Join(args, " "))
	}
	return err
}

func RunCommandInteractive(interactive bool, dir string, name string, args ...string) error {
	e := exec.Command(name, args...)
	e.Stdout = os.Stdout
	e.Stderr = os.Stderr
	if dir != "" {
		e.Dir = dir
	}
	if interactive {
		e.Stdin = os.Stdin
	}
	err := e.Run()
	if err != nil {
		fmt.Printf("Error: Command failed  %s %s\n", name, strings.Join(args, " "))
	}
	return err
}

// GetCommandOutput evaluates the given command and returns the trimmed output
func GetCommandOutput(dir string, name string, args ...string) (string, error) {
	e := exec.Command(name, args...)
	if dir != "" {
		e.Dir = dir
	}
	data, err := e.CombinedOutput()
	text := string(data)
	text = strings.TrimSpace(text)
	return text, err
}
