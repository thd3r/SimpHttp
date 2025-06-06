package utils

import (
	"encoding/json"
	"io"

	"github.com/thd3r/SimpHttp/pkg/net/client"
)

var CurrentVersion = "v0.1.2"

func Version() string {
	clients := client.NewClient(10)

	resp, err := clients.Do("GET", "https://api.github.com/repos/thd3r/SimpHttp/releases/latest")
	if err != nil {
		return CurrentVersion + " " + ColoredText("magenta", "unknown")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CurrentVersion + " " + ColoredText("magenta", "unknown")
	}

	var dataRelease = struct {
		ReleaseVersion string `json:"tag_name"`
	}{}

	if err := json.Unmarshal(body, &dataRelease); err != nil {
		return CurrentVersion + " " + ColoredText("magenta", "unknown")
	}

	if CurrentVersion < dataRelease.ReleaseVersion {
		return CurrentVersion + " " + ColoredText("red", "outdated")
	}
	if CurrentVersion == dataRelease.ReleaseVersion {
		return CurrentVersion + " " + ColoredText("green", "latest")
	}

	return CurrentVersion + " " + ColoredText("magenta", "unknown")
}
