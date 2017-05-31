package main

import (
	"fmt"
	"os/exec"

	"github.com/golang/glog"
)

type Runner struct {
	clonedir string
}

func NewRunner() Runner {
	runner := Runner{}
	runner.clonedir = getCloneDir()
	return runner
}

func (r Runner) CloneRepo(cloneURL string) error {
	glog.Infof("Cloning from %s into %s", cloneURL, r.clonedir)
	err := exec.Command("git", "clone", cloneURL, r.clonedir).Run()
	if err != nil {
		return fmt.Errorf("Failed to clone directory %s into %s: %v", cloneURL, r.clonedir, err)
	}
	return nil
}

func (r Runner) Checkout(head string) error {
	glog.Infof("Checking out head %s", head)
	cmd := exec.Command("git", "checkout", head)
	cmd.Dir = r.clonedir
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to checkout head %s in %s: %v", head, r.clonedir, err)
	}
	return nil
}

func (r Runner) Run(program string, args ...string) ([]byte, error) {
	glog.Infof("Running `%s %v`", program, args)
	cmd := exec.Command(program, args...)
	cmd.Dir = r.clonedir
	return cmd.CombinedOutput()
}
