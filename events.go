package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

type EventHandler interface {
	OnPushEvent(event *github.PushEvent) error
	OnPingEvent(event *github.PingEvent) error
}

type handler struct {
	client   *GithubClient
	reponame string

	MasterRef string
}

var DefaultEventHandler EventHandler = &handler{}

func NewEventHandler(reponame string, client *GithubClient) *handler {
	h := &handler{}
	h.client = client
	h.reponame = reponame
	h.MasterRef = "refs/heads/master"
	return h
}

func (h *handler) OnPushEvent(event *github.PushEvent) error {
	glog.Infof("Received push event")
	if err := checkPushEvent(event); err != nil {
		return err
	}

	head := *event.HeadCommit.ID
	fullName := *event.Repo.FullName

	if h.reponame != REPONAME_ANY && fullName != h.reponame {
		return fmt.Errorf("Will not handle requests for this repository: %s", fullName)
	}

	// TODO: REMOVE THIS
	content, _ := json.Marshal(event)
	fmt.Println(string(content))

	err := h.client.PostStatus(fullName, head, "pending")
	if err != nil {
		glog.Warningf("Failed to create pending status: %v", err)
	}

	// Construct the clone URL
	prefix := h.client.BaseURL()
	cloneURL := fmt.Sprintf("%s/%s.git", prefix, fullName)

	// Get a new job runner
	runner := NewRunner()

	// Clone repository
	err = runner.CloneRepo(cloneURL)
	if err != nil {
		h.client.PostStatus(fullName, head, "failure")
		return err
	}

	// Checkout head commit
	err = runner.Checkout(head)
	if err != nil {
		h.client.PostStatus(fullName, head, "failure")
		return err
	}

	// Execute test command
	outputFile := getOutputFile()
	out, err := runner.Run("make", "jarvis-ci-test")

	// Write the output anyway
	writeOutput(out, outputFile)

	// Handle the error now
	if err != nil {
		h.client.PostStatus(fullName, head, "failure")
		return fmt.Errorf("Failed test command: %v", err)
	}

	h.client.PostStatus(fullName, head, "success")

	// Check if the ref is the master ref
	if h.MasterRef != event.GetRef() {
		return nil
	}

	// Parse the head commit message to find make targets
	msg := event.HeadCommit.GetMessage()
	targets := []string{}
	for _, line := range strings.Split(msg, "\n") {
		if strings.HasPrefix(line, "JARVIS: ") {
			targetstring := strings.TrimPrefix(line, "JARVIS: ")
			targets = strings.Split(targetstring, " ")
		}
	}

	for _, target := range targets {
		out, err := runner.Run("make", target)
		fmt.Println(string(out), err)
	}
	return nil
}

func (h *handler) OnPingEvent(event *github.PingEvent) error {
	glog.Infof("Received ping event")
	return nil
}

// Make sure every field exists
func checkPushEvent(event *github.PushEvent) error {
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
	return nil
}
