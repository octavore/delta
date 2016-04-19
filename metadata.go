package main

type change string

const (
	changeModified change = "modified"
	changeAdded    change = "added"
	changeDeleted  change = "deleted"
)

type Metadata struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Merged    string `json:"merged"`
	Dir       string `json:"dir"`
	Change    change `json:"change"`
	Hash      string `json:"hash"`
	DirHash   string `json:"dirhash"`
	Timestamp int64  `json:"timestamp"`
}
