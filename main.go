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

	var err error

	// Read in the token
	Token, err = ioutil.ReadFile(TokenPath)
	if err != nil {
		glog.Warningf("Failed to read token: %v", err)
	}

	// Read in the hub secret
	HubSecret, err = ioutil.ReadFile(HubSecretPath)
	if err != nil {
		glog.Warningf("Failed to read hub secret: %v", err)
		HubSecret = []byte("mysecret")
	}

	// Log the configuration
	logConfig()

	// Start the server
	http.HandleFunc(fmt.Sprintf("%s/debug/status", BasePath), debug)
	http.HandleFunc(fmt.Sprintf("%s/hook", BasePath), hook)
	err = http.ListenAndServe(fmt.Sprintf(":%d", ServerPort), nil)
	glog.Fatalf("Error while serving: %v", err)
}

func debug(w http.ResponseWriter, req *http.Request) {
	glog.Infof("Handling debug request")
	fmt.Fprintf(w, "OK")
}

func hook(w http.ResponseWriter, req *http.Request) {
	glog.Infof("Handling hook request")
	glog.Infof("Request: %v", req)

	// Verify that it's coming from github
	payload, err := github.ValidatePayload(req, HubSecret)
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

	// Switch on its type
	switch event := event.(type) {
	case *github.PingEvent:
		glog.Infof("Received ping event")
	case *github.PushEvent:
		glog.Infof("Received push event")
		err = handlePushEvent(event)
	}

	if err != nil {
		glog.Errorf("Failed to handle github hook: %v", err)
		return
	}
	// Done
	fmt.Fprintf(w, "OK")
}

// Returns an error if it is a bad request
func handlePushEvent(event *github.PushEvent) error {
	// Make sure every field exists
	if event.Repo == nil {
		return fmt.Errorf("Missing PushEvent.Repo")
	}
	if event.HeadCommit == nil {
		return fmt.Errorf("Missing PushEvent head commit")
	}
	if event.Repo.FullName == nil {
		return fmt.Errorf("Missing PushEvent repo full name")
	}
	if event.HeadCommit.ID == nil {
		return fmt.Errorf("Missing PushEvent head commit ID")
	}

	head := *event.HeadCommit.ID
	fullName := *event.Repo.FullName

	// Check the full name for security
	if RepoFullName != "ANY" && fullName != RepoFullName {
		return fmt.Errorf("Will not handle requests for this repository: %s", fullName)
	}

	err := NewGithubClient().PostStatus(fullName, head, "pending")
	if err != nil {
		glog.Warningf("Failed to create pending status: %v", err)
	}

	// Construct the clone URL
	prefix := fmt.Sprintf("https://%s@github.com", Token)
	cloneURL := fmt.Sprintf("%s/%s.git", prefix, fullName)
	cloneDir := getCloneDir()

	// Clone repository
	err = CloneRepo(cloneDir, cloneURL)
	if err != nil {
		NewGithubClient().PostStatus(fullName, head, "failure")
		return err
	}

	// Checkout head commit
	err = CheckoutHead(cloneDir, head)
	if err != nil {
		NewGithubClient().PostStatus(fullName, head, "failure")
		return err
	}

	// Execute test command
	outputFile := getOutputFile()
	out, err := DoTest(cloneDir)
	WriteOutput(out, outputFile)
	if err != nil {
		NewGithubClient().PostStatus(fullName, head, "failure")
		return fmt.Errorf("Failed test command: %v", err)
	}

	NewGithubClient().PostStatus(fullName, head, "success")
	return nil
}
