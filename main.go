package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
)

var (
	TokenPath     string
	Token         []byte
	HubSecretPath string
	HubSecret     []byte

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

	// Read in the token
	var err error
	Token, err = ioutil.ReadFile(TokenPath)
	if err != nil {
		// glog.Fatalf("Failed to read token: %v", err)
	}

	// Read in the hub secret
	HubSecret, err = ioutil.ReadFile(HubSecretPath)
	if err != nil {
		// glog.Fatalf("Failed to read hub secret: %v", err)
	}

	// Start the server
	http.HandleFunc(fmt.Sprintf("%s/debug/status", BasePath), debug)
	http.HandleFunc(fmt.Sprintf("%s/push", BasePath), push)
	err = http.ListenAndServe(fmt.Sprintf(":%d", ServerPort), nil)
	glog.Fatalf("Error while serving: %v", err)
}

func logConfig() {
	glog.Infof("TokenPath: %s", TokenPath)
	glog.Infof("HubSecretPath: %s", HubSecretPath)
	glog.Infof("ServerPort: %d", ServerPort)
	glog.Infof("BasePath: %s", BasePath)
}

func debug(w http.ResponseWriter, req *http.Request) {
	glog.Infof("Handling debug request")
	fmt.Fprintf(w, "OK")
}

func push(w http.ResponseWriter, req *http.Request) {
	glog.Infof("Handling onpush request")
	glog.Infof("Request: %v", req)
	fmt.Fprintf(w, "OK")
}
