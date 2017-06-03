package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

type EventHandler interface {
	OnPushEvent(event *github.PushEvent) error
	OnPingEvent(event *github.PingEvent) error
}

type eventHandler struct {
	client        *GithubClient
	reponame      string
	outputhandler OutputHandler

	MasterRef string
}

const (
	REPONAME_ANY = "ANY"
)

var (
	RepoFullName string
	MasterRef    string
)

var _ EventHandler = &eventHandler{}

func init() {
	flag.StringVar(&RepoFullName, "repo", REPONAME_ANY, "The full name of the repository we are watching from GitHub")
	flag.StringVar(&MasterRef, "master-ref", "refs/heads/master", "The ref with post-commit targets. Defaults to refs/heads/master")
}

func NewEventHandler(reponame string, client *GithubClient, outputhandler OutputHandler) *eventHandler {
	h := &eventHandler{}
	h.client = client
	h.reponame = reponame
	h.outputhandler = outputhandler
	h.MasterRef = MasterRef
	return h
}

func (h *eventHandler) OnPushEvent(event *github.PushEvent) error {
	glog.Infof("Received push event")
	if err := checkPushEvent(event); err != nil {
		return err
	}

	head, fullName := *event.HeadCommit.ID, *event.Repo.FullName
	if h.reponame != REPONAME_ANY && fullName != h.reponame {
		return fmt.Errorf("Will not handle requests for this repository: %s", fullName)
	}

	// TODO: REMOVE THIS
	content, _ := json.Marshal(event)
	fmt.Println(string(content))

	// Get a new job runner
	runner := NewRunner()
	defer runner.Cleanup()

	err := h.client.PostStatus(fullName, head, "pending", head)
	if err != nil {
		glog.Warningf("Failed to create pending status: %v", err)
	}

	// Construct the clone URL
	prefix := h.client.BaseURL()
	cloneURL := fmt.Sprintf("%s/%s.git", prefix, fullName)

	// Clone repository
	err = runner.CloneRepo(cloneURL, event.GetRef())
	if err != nil {
		h.outputhandler.AddOutput(head, "Failed to clone repo: %v", err)
		h.client.PostStatus(fullName, head, "failure", head)
		return err
	}

	// Checkout head commit
	err = runner.Checkout(head)
	if err != nil {
		h.client.PostStatus(fullName, head, "failure", head)
		h.outputhandler.AddOutput(head, "Failed to checkout head: %v", err)
		return err
	}

	// Write the head of the output
	h.outputhandler.AddOutput(head, "TARGET: jarvis-ci-test\n-------\n")

	// Append to the output continuously
	fn := func(line string) error {
		h.outputhandler.AddOutput(head, line)
		return nil
	}
	err = runner.WatchFn(fn, "make", "jarvis-ci-test")
	if err != nil {
		h.outputhandler.AddOutput(head, "-------\n%v", err)
	}

	// Append footer for main target
	h.outputhandler.AddOutput(head, "=======\n")

	// Handle the error now
	if err != nil {
		h.client.PostStatus(fullName, head, "failure", head)
		glog.Infof("Failed jarvis-ci-test: %v", err)
		return nil
	}

	// Check if the ref is the master ref
	if h.MasterRef != event.GetRef() {
		h.client.PostStatus(fullName, head, "success", head)
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

	// TODO: Watch those outputs too
	for _, target := range targets {
		h.outputhandler.AddOutput(head, "TARGET: %s\n-------", target)
		out, err := runner.Run("make", target)
		if err != nil {
			glog.Infof("Failed %s: %s | %v", target, string(out), err)
		} else {
			glog.Infof("Success %s: %s", target, string(out))
		}
		h.outputhandler.AddOutput(head, "%v\n-------\n%v\n-------\n", string(out), err)
	}
	return nil
}

func (h *eventHandler) OnPingEvent(event *github.PingEvent) error {
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
