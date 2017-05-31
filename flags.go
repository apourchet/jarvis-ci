package main

import "github.com/golang/glog"

func printFlags() {
	glog.Infof("Server port: %d", ServerPort)
	glog.Infof("Base path: %s", BasePath)
	glog.Infof("Token path: %s", TokenPath)
	glog.Infof("Hub secret path: %s", HubSecretPath)
	glog.Infof("Repository full name: %s", RepoFullName)
}
