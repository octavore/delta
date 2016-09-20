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

	"github.com/octavore/delta/lib"
	"github.com/octavore/delta/lib/formatter"

	"github.com/pkg/browser"
)

// constants for command line options
const (
	OutputOptionCLI     = "cli"
	OutputOptionBrowser = "browser"
	OutputOptionGist    = "gist"

	FormatOptionHTML    = "html"
	FormatOptionText    = "text"
	FormatOptionDefault = "default"
)

var (
	// commands
	install   = flag.Bool("install", false, "Install to gitconfig.")
	uninstall = flag.Bool("uninstall", false, "Remove from gitconfig.")
	version   = flag.Bool("version", false, "Display delta version.")

	// diff settings
	output = flag.String("output", "cli", "Where to send the output. Valid values: browser (default), cli, gist.")
	format = flag.String("format", "default", `Format of the output. `)
)

func main() {
	flag.CommandLine.Usage = printHelp
	flag.Parse()
	if *install || *uninstall || *version {
		switch {
		case *version:
			printVersion()
		case *install:
			installGit()
		case *uninstall:
			uninstallGit()
		}
		return
	}
	if flag.NArg() < 2 {
		printVersion()
		printHelp()
		return
	}
	pathFrom, pathTo := flag.Arg(0), flag.Arg(1)
	pathBase := pathTo
	if flag.NArg() > 2 {
		pathBase = flag.Arg(2)
	}
	runDiff(pathFrom, pathTo, pathBase)
}

func printHelp() {
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println()
	fmt.Println("delta OPTION_COMMAND")

	fmt.Printf("%-20s %s\n", "  --install", "Install delta to gitconfig.")
	fmt.Printf("%-20s %s\n", "  --uninstall", "Remove delta from gitconfig.")
	fmt.Printf("%-20s %s\n", "  --version", "Display delta version.")

	// diff settings
	fmt.Println("\ndelta [OPTIONS] FILE1 FILE2")
	fmt.Printf("%-20s %s\n", "  --output", "Where to send the output. Valid values: browser, cli (default), gist.")
	fmt.Printf("%-20s %s\n", "  --format", `Valid values: default (text for cli, html otherwise), html, text.`)
	fmt.Println()
}

func printVersion() {
	fmt.Println("delta", Version)
}

func runDiff(pathFrom, pathTo, pathBase string) {
	config, err := loadConfig()
	if err != nil {
		os.Stderr.WriteString("warning: error parsing .deltarc file")
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
	if *format == FormatOptionDefault {
		switch *output {
		case OutputOptionBrowser, OutputOptionGist:
			*format = FormatOptionHTML
		case OutputOptionCLI:
			*format = FormatOptionText
		}
	}

	switch *format {
	case FormatOptionHTML:
		page, err := html(d, pathFrom, pathTo, pathBase, config)
		if err != nil {
			os.Stderr.WriteString(err.Error())
			return
		}
		switch *output {
		case OutputOptionCLI:
			page.WriteTo(os.Stdout)
		case OutputOptionGist:
			uploadGist(page.Bytes())
		case OutputOptionBrowser:
			browser.OpenReader(page)
		}

	case FormatOptionText:
		switch *output {
		case OutputOptionCLI:
			fmt.Println(formatter.ColoredText(d))
		case OutputOptionGist:
			uploadGist([]byte(formatter.Text(d)))
		case OutputOptionBrowser:
			browser.OpenReader(bytes.NewBufferString(formatter.Text(d)))
		}
	}
}

// openDiffs diffs the given files and writes the result to a tempfile,
// then opens it in the gui.
func html(d *delta.DiffSolution, pathFrom, pathTo, pathBase string, config Config) (*bytes.Buffer, error) {
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
