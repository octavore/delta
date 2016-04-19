package main

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/octavore/delta/browser"
)

func getAsset(path string) string {
	a, err := browser.Asset(path)
	if err != nil {
		panic(err)
	}
	return string(a)
}

func md5sum(s string) string {
	hash := md5.New()
	_, _ = hash.Write([]byte(s))
	return hex.EncodeToString(hash.Sum(nil))
}
