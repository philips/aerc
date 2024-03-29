package commands

import (
	"errors"
	"os"

	"git.sr.ht/~sircmpwn/aerc/widgets"
	"github.com/mitchellh/go-homedir"
)

var (
	previousDir string
)

func init() {
	register("cd", ChangeDirectory)
}

func ChangeDirectory(aerc *widgets.Aerc, args []string) error {
	if len(args) < 1 || len(args) > 2 {
		return errors.New("Usage: cd [directory]")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	var target string
	if len(args) == 1 {
		target = "~"
	} else if args[1] == "-" {
		if previousDir == "" {
			return errors.New("No previous folder to return to")
		} else {
			target = previousDir
		}
	} else {
		target = args[1]
	}
	target, err = homedir.Expand(target)
	if err != nil {
		return err
	}
	if err := os.Chdir(target); err == nil {
		previousDir = cwd
	}
	return err
}
