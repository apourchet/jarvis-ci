package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/golang/glog"
)

var (
	TokenPath     string
	HubSecretPath string

	ServerPort int
	BasePath   string
)

func init() {
	flag.StringVar(&TokenPath, "t", "/jarvis-ci/token", "The API token for authenticating with GitHub")
	flag.StringVar(&HubSecretPath, "s", "/jarvis-ci/hubsecret", "The secret key for validating a request from GitHub")

	flag.IntVar(&ServerPort, "p", 8080, "The port that Jarvis should listen for webhooks on")
	flag.StringVar(&BasePath, "b", "/jarvis-ci/", "The root path of the webhooks Jarvis will make available to GitHub")

	flag.Set("logtostderr", "true")
}

func main() {
	flag.Parse()

	// Log the configuration
	logConfig()

	// Start the server
	http.HandleFunc(fmt.Sprintf("%s/push", BasePath), push)
	err := http.ListenAndServe(fmt.Sprintf(":%d", ServerPort), nil)
	glog.Fatalf("Error while serving: %v", err)
}

func logConfig() {
	glog.Infof("TokenPath: %s", TokenPath)
	glog.Infof("HubSecretPath: %s", HubSecretPath)
	glog.Infof("ServerPort: %d", ServerPort)
	glog.Infof("BasePath: %s", BasePath)
}

func push(w http.ResponseWriter, req *http.Request) {
	glog.Infof("Handling onpush request")
}
