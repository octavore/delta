package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

const configFile = ".deltarc"

// Config represents user-provided options (via the file specified by configFile)
type Config struct {
	Context           *int     `json:"context"`
	ShowEmpty         *bool    `json:"showEmpty"`
	ShouldCollapse    *bool    `json:"shouldCollapse"`
	Highlight         *bool    `json:"highlight"`
	UnmodifiedOpacity *float32 `json:"unmodifiedOpacity"`
	DiffFontSize      *int32   `json:"diffFontSize"`
}

func loadConfig() (config Config, err error) {
	usr, err := user.Current()
	if err != nil {
		return
	}
	deltarc := filepath.Join(usr.HomeDir, configFile)
	f, err := os.Open(deltarc)
	if (err != nil && os.IsExist(err)) || f == nil {
		return
	}
	d, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}
	err = json.Unmarshal(d, &config)
	return
}
