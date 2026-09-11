package services

import "errors"

var (
	errNoConsole           = errors.New("connect a Console profile before browsing Console resources")
	errMissingID           = errors.New("service id is required")
	errMissingCluster      = errors.New("cluster id is required")
	errMissingService      = errors.New("service was not found")
	errMissingCreateFields = errors.New("name, repository id, git ref, and git folder are required")
	errMissingTarball      = errors.New("service does not have a tarball")
)
