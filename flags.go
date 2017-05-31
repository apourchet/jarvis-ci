package main

import (
	"flag"

	"github.com/golang/glog"
)

var (
	TokenPath     string
	HubSecretPath string

	RepoFullName string

	MaxOutputs int

	BasePath   string
	ServerPort int
)

const (
	REPONAME_ANY = "ANY"
)

func init() {
	flag.StringVar(&TokenPath, "t", "/jarvis-ci/token", "The API token for authenticating with GitHub")
	flag.StringVar(&HubSecretPath, "s", "/jarvis-ci/hubsecret", "The secret key for validating a request from GitHub")
	flag.StringVar(&RepoFullName, "repo", REPONAME_ANY, "The full name of the repository we are watching from GitHub")

	flag.IntVar(&MaxOutputs, "max-outputs", 10, "Maximum number of outputs to keep")

	flag.StringVar(&BasePath, "b", "/jarvis-ci", "The root path of the webhooks Jarvis will make available to GitHub")
	flag.IntVar(&ServerPort, "p", 8080, "The port that Jarvis should listen for webhooks on")

	flag.Set("logtostderr", "true")
}

func printFlags() {
	glog.Infof("Server port: %d", ServerPort)
	glog.Infof("Base path: %s", BasePath)

	glog.Infof("Token path: %s", TokenPath)
	glog.Infof("Hub secret path: %s", HubSecretPath)
	glog.Infof("Repository full name: %s", RepoFullName)
}
