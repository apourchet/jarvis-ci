package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"context"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type GithubClient struct {
	*github.Client
}

func NewGithubClient() *GithubClient {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: string(Token)})
	client := github.NewClient(oauth2.NewClient(context.Background(), ts))
	return &GithubClient{client}
}

func (c *GithubClient) PostStatus(fullName, head, status string) error {
	// Create request
	url := fmt.Sprintf("https://api.github.com/repos/%s/statuses/%s", fullName, head)
	data := map[string]string{"state": status, "context": "jarvis-ci", "description": "Jarvis-CI testing"}
	dataString, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(dataString))
	if err != nil {
		return fmt.Errorf("Failed to create status creation request: %v", err)
	}

	// Send request
	resp, err := c.Do(context.Background(), req, nil)
	if err != nil {
		return err
	} else if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Failed to post pending status, bad status code: %s", resp.StatusCode)
	}

	glog.Infof("Successfully set status of %s/%s to %s", fullName, head, status)
	return nil
}
