package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"

	"github.com/golang/glog"
)

func CloneRepo(cloneDir, cloneURL string) error {
	glog.Infof("Cloning from %s into %s", cloneURL, cloneDir)
	err := exec.Command("git", "clone", cloneURL, cloneDir).Run()
	if err != nil {
		return fmt.Errorf("Failed to clone directory %s into %s: %v", cloneURL, cloneDir, err)
	}
	return nil
}

func CheckoutHead(cloneDir, head string) error {
	glog.Infof("Checking out head %s", head)
	cmd := exec.Command("git", "checkout", head)
	cmd.Dir = cloneDir
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to checkout head %s in %s: %v", head, cloneDir, err)
	}
	return nil
}

func DoTest(cloneDir string) ([]byte, error) {
	glog.Infof("Running `make test`")
	cmd := exec.Command("make", "test")
	cmd.Dir = cloneDir
	return cmd.CombinedOutput()
}

func WriteOutput(out []byte, outputFile string) {
	glog.Infof("Test command succeeded, writing output to %s", outputFile)
	err := ioutil.WriteFile(outputFile, out, 0644)
	if err != nil {
		glog.Warningf("Failed to write output to %s", outputFile)
	}
}
