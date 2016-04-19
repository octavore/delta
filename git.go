package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/browser"
)

func installGit() {
	commands := [][]string{
		{"git", "config", "--global", "diff.tool", "delta"},
		{"git", "config", "--global", "difftool.prompt", "false"},
		{"git", "config", "--global", "difftool.delta.cmd", `delta "$LOCAL" "$REMOTE" "$MERGED"`},
	}

	for _, c := range commands {
		fmt.Println(strings.Join(c, " "))
		o, _ := exec.Command(c[0], c[1:]...).CombinedOutput()
		fmt.Print(string(o))
	}
}

// known issue: this does not remove the gitconfig section if the unset
// operation causes the section to become empty.
func uninstallGit() {
	commands := [][]string{
		{"git", "config", "--global", "--unset", "diff.tool"},
		{"git", "config", "--global", "--unset", "difftool.prompt"},
		{"git", "config", "--global", "--remove-section", "difftool.delta"},
	}

	for _, c := range commands {
		fmt.Println(strings.Join(c, " "))
		o, _ := exec.Command(c[0], c[1:]...).CombinedOutput()
		fmt.Print(string(o))
	}
}

type GistRequest struct {
	Description string                       `json:"description"`
	Public      bool                         `json:"public"`
	Files       map[string]map[string]string `json:"files"`
}

type GistResponse struct {
	Files map[string]struct {
		RawURL string `json:"raw_url"`
	} `json:"files"`
}

const gistURL = "https://api.github.com/gists"

func uploadGist(data []byte) {
	b, _ := json.Marshal(&GistRequest{
		Description: "diff created by delta https://github.com/octavore/delta",
		Public:      false,
		Files: map[string]map[string]string{
			"diff.html": map[string]string{"content": string(data)},
		},
	})

	resp, err := http.Post(gistURL, "application/json", bytes.NewBuffer(b))
	if err != nil {
		fmt.Fprintln(os.Stderr, resp)
		return
	}
	r := &GistResponse{}
	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	url := strings.Replace(r.Files["diff.html"].RawURL,
		"gist.githubusercontent", "cdn.rawgit", 1)
	fmt.Println(url)
	_ = browser.OpenURL(url)
}
