package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

func main() {
	flag.Parse()

	// Read in the token
	token, err := ioutil.ReadFile(TokenPath)
	if err != nil {
		glog.Warningf("Failed to read token: %v", err)
	}

	// Read in the hub secret
	hubSecret, err := ioutil.ReadFile(HubSecretPath)
	if err != nil {
		glog.Warningf("Failed to read hub secret: %v", err)
		hubSecret = []byte("antoinesecret")
	}

	// Log the configuration
	printFlags()

	// Create the github client
	client := NewGithubClient(string(token))

	// Create the event handler
	eventHandler := NewEventHandler(RepoFullName, client)

	// Start the server
	http.HandleFunc(fmt.Sprintf("%s/debug/status", BasePath), debug)
	http.HandleFunc(fmt.Sprintf("%s/hook", BasePath), hook(hubSecret, eventHandler))
	err = http.ListenAndServe(fmt.Sprintf(":%d", ServerPort), nil)
	glog.Fatalf("Error while serving: %v", err)
}

func debug(w http.ResponseWriter, req *http.Request) {
	glog.Infof("Handling debug request")
	fmt.Fprintf(w, "OK")
}

func hook(hubscrt []byte, eventHandler EventHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		glog.Infof("Handling hook request")
		glog.Infof("Request: %v", req)

		payload, err := github.ValidatePayload(req, hubscrt)
		// Verify that it's coming from github
		if err != nil {
			glog.Errorf("Failed to validate github hook payload: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Parse the event
		event, err := github.ParseWebHook(github.WebHookType(req), payload)
		if err != nil {
			glog.Errorf("Failed to parse github hook event: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		go func() {
			switch event := event.(type) {
			case *github.PingEvent:
				err = eventHandler.OnPingEvent(event)
			case *github.PushEvent:
				err = eventHandler.OnPushEvent(event)
			}

			// If there is an error, log it
			if err != nil {
				glog.Errorf("Failed to handle github hook: %v", err)
				return
			}
		}()

		// Hook handled successfully
		fmt.Fprintf(w, "OK")
	}
}
