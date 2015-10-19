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
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/octavore/delta"
	"github.com/octavore/delta/delta/static"
	"github.com/octavore/delta/delta/vendor/browser"
	"github.com/octavore/delta/formatter"
)

const VERSION = "0.4.0"

func main() {
	// only one of the following should be provided
	cli := flag.Bool("cli", false, "open the file in the terminal")
	html := flag.Bool("html", false, "print out html")
	install := flag.Bool("install", false, "install to gitconfig")
	uninstall := flag.Bool("uninstall", false, "remove from gitconfig")
	version := flag.Bool("version", false, "display delta version")

	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
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
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if !*cli {
		openDiff(pathFrom, pathTo, pathBase)
	} else {
		printDiff(pathFrom, pathTo, *html)
	}
}

func getAsset(path string) string {
	a, err := static.Asset(path)
	if err != nil {
		panic(err)
	}
	return string(a)
}

// openDiffs diffs the given files and writes the result to a tempfile,
// then opens it in the gui.
func openDiff(pathFrom, pathTo, pathBase string) {
	d, err := diff(pathFrom, pathTo)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		return
	}

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
	tmpl := template.Must(template.New("compare").Parse(getAsset("compare.html")))
	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, map[string]interface{}{
		"metadata": template.JS(string(meta)),
		"content":  template.HTML(html),
		"CSS":      template.CSS(getAsset("app.css")),
		"JS": map[string]interface{}{
			"mithril":   template.JS(getAsset("vendor/mithril.min.js")),
			"mousetrap": template.JS(getAsset("vendor/mousetrap.min.js")),
			"highlight": template.JS(getAsset("vendor/highlight.js")),
			"pouchdb":   template.JS(getAsset("vendor/pouchdb.min.js")),
			"app":       template.JS(getAsset("app.js")),
		},
	})
	if err != nil {
		panic(err)
	}

	browser.OpenReader(buf)
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

func printDiff(pathFrom, pathTo string, html bool) {
	d, err := diff(pathFrom, pathTo)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		return
	}
	if html {
		fmt.Println(formatter.HTML(d))
		return
	}
	fmt.Println(formatter.ColoredText(d))
}

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

func md5sum(s string) string {
	hash := md5.New()
	_, _ = hash.Write([]byte(s))
	return hex.EncodeToString(hash.Sum(nil))
}
