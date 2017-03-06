package main

import (
	"flag"

	"github.com/golang/glog"
)

var (
	Debug bool

	TokenPath     string
	Token         []byte
	HubSecretPath string
	HubSecret     []byte

	RepoFullName string

	MaxClones  int
	MaxOutputs int

	BasePath   string
	ServerPort int
)

func init() {
	flag.BoolVar(&Debug, "d", true, "Debug mode")
	flag.StringVar(&TokenPath, "t", "/jarvis-ci/token", "The API token for authenticating with GitHub")
	flag.StringVar(&HubSecretPath, "s", "/jarvis-ci/hubsecret", "The secret key for validating a request from GitHub")
	flag.StringVar(&RepoFullName, "repo", "ANY", "The full name of the repository we are watching from GitHub")

	flag.IntVar(&MaxClones, "max-clones", 10, "Maximum number of clones to keep")
	flag.IntVar(&MaxOutputs, "max-outputs", 10, "Maximum number of outputs to keep")

	flag.StringVar(&BasePath, "b", "/jarvis-ci", "The root path of the webhooks Jarvis will make available to GitHub")
	flag.IntVar(&ServerPort, "p", 8080, "The port that Jarvis should listen for webhooks on")

	flag.Set("logtostderr", "true")
}

func logConfig() {
	glog.Infof("Server port: %d", ServerPort)
	glog.Infof("Base path: %s", BasePath)

	glog.Infof("Token path: %s", TokenPath)
	glog.Infof("Hub secret path: %s", HubSecretPath)
	glog.Infof("Repository full name: %s", RepoFullName)

	if Debug {
		glog.Infof("Token: '%s'", Token)
		glog.Infof("Hub secret: '%s'", HubSecret)
	}
}
