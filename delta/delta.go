package main

/*
	delta is a command-line diff utility.

	Usage:
		`delta <file1> <file2> <merged>`

		file1 is set to the name of the temporary file containing the contents of the diff pre-image.
		file2 is set to the name of the temporary file containing the contents of the diff post-image.
		merged is the name of the file which is being compared.
*/

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/octavore/delta"
	"github.com/octavore/delta/formatter"

	"github.com/pkg/browser"
)

const VERSION = "0.5.0"

func main() {
	cli := flag.Bool("cli", false, "print the diff to stdout")
	gist := flag.Bool("gist", false, "upload gist to github")
	version := flag.Bool("version", false, "display delta version")
	install := flag.Bool("install", false, "install to gitconfig")
	uninstall := flag.Bool("uninstall", false, "remove from gitconfig")

	html := flag.Bool("html", false, "use with --cli to output html instead of text diff")

	flag.Parse()
	switch {
	case *version:
		fmt.Println("delta", VERSION)
		return
	case *install:
		installGit()
		return
	case *uninstall:
		uninstallGit()
		return
	}
	if flag.NArg() < 2 {
		flag.PrintDefaults()
		return
	}
	pathFrom := flag.Arg(0)
	pathTo := flag.Arg(1)
	pathBase := flag.Arg(1)
	if flag.NArg() > 2 {
		pathBase = flag.Arg(2)
	}

	config, err := loadConfig()
	if err != nil {
		fmt.Println("warning: error parsing .deltarc file")
	}

	d, err := diff(pathFrom, pathTo)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		return
	}

	if err != nil {
		os.Stderr.WriteString(err.Error())
		return
	}
	if *cli && !*html {
		fmt.Println(formatter.ColoredText(d))
		return
	}
	page, err := render(d, pathFrom, pathTo, pathBase, config)
	switch {
	case *cli:
		page.WriteTo(os.Stdout)
	case *gist:
		uploadGist(page.Bytes())
	default:
		browser.OpenReader(page)
	}
}

// openDiffs diffs the given files and writes the result to a tempfile,
// then opens it in the gui.
func render(d *delta.DiffSolution, pathFrom, pathTo, pathBase string, config Config) (*bytes.Buffer, error) {
	change := changeModified
	if pathTo == "/dev/null" {
		change = changeDeleted
	} else if pathFrom == "/dev/null" {
		change = changeAdded
	}

	// normalize paths so we don't have tmp on the path
	tmpFrom := strings.HasPrefix(pathFrom, os.TempDir())
	tmpTo := strings.HasPrefix(pathTo, os.TempDir())
	if tmpFrom && !tmpTo {
		pathFrom = pathTo
	} else if !tmpFrom && tmpTo {
		pathTo = pathFrom
	}

	wd, _ := os.Getwd()
	html := formatter.HTML(d)
	m := &Metadata{
		From:      pathFrom,
		To:        pathTo,
		Merged:    pathBase,
		Dir:       wd,
		Change:    change,
		Hash:      md5sum(html),
		DirHash:   md5sum(wd),
		Timestamp: time.Now().UnixNano() / 1000000, // convert to millis
	}
	meta, _ := json.Marshal(m)
	cfg, _ := json.Marshal(config)
	tmpl := template.Must(template.New("compare").Parse(getAsset("compare.html")))
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, map[string]interface{}{
		"metadata": template.JS(string(meta)),
		"config":   template.JS(cfg),
		"content":  template.HTML(html),
		"CSS":      template.CSS(getAsset("app.css")),
		"JS": map[string]interface{}{
			"mithril":   template.JS(getAsset("vendor/mithril.min.js")),
			"mousetrap": template.JS(getAsset("vendor/mousetrap.min.js")),
			"highlight": template.JS(getAsset("vendor/highlight.min.js")),
			"pouchdb":   template.JS(getAsset("vendor/pouchdb.min.js")),
			"app":       template.JS(getAsset("app.js")),
		},
	})
	return buf, err
}

// diff reads in files in pathFrom and pathTo, and returns a diff
func diff(pathFrom, pathTo string) (*delta.DiffSolution, error) {
	from, err := ioutil.ReadFile(pathFrom)
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %v", pathFrom, err)
	}
	to, err := ioutil.ReadFile(pathTo)
	if err != nil {
		return nil, fmt.Errorf("error reading %q: %v", pathTo, err)
	}
	return delta.HistogramDiff(string(from), string(to)), nil
}
