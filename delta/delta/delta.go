package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"bitbucket.org/pancakeio/delta/delta"
)

func main() {
	open := flag.Bool("open", false, "open the file in the gui")
	html := flag.Bool("html", false, "print out html")
	flag.Parse()

	pathFrom := flag.Arg(0)
	pathTo := flag.Arg(1)
	base := os.Getenv("BASE")
	fmt.Println(pathTo)
	fmt.Println(os.Getenv("BASE"))
	fmt.Println()

	if *open {
		openDiff(pathFrom, pathTo)
	} else {
		printDiff(pathFrom, pathTo, *html)
	}
}

func openDiffPaths(pathFrom, pathTo string) {
	dir, _ := os.Getwd()
	u, _ := url.Parse("delta://open")
	v := url.Values{}
	v.Add("base", dir)
	v.Add("left", pathFrom)
	v.Add("right", pathTo)
	u.RawQuery = v.Encode()
	exec.Command("open", u.String()).Run()
}

// openDiffs diffs the given files and writes the result to a tempfile,
// then opens it in the gui.
func openDiff(pathFrom, pathTo string) {
	d, err := diff(pathFrom, pathTo)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		return
	}
	f, err := ioutil.TempFile("", "delta-diff")
	if err != nil {
		os.Stderr.WriteString(err.Error())
		return
	}
	io.WriteString(f, d.HTML())

	dir, _ := os.Getwd()
	u, _ := url.Parse("delta://openset")
	v := url.Values{}

	tmpFrom := strings.HasPrefix(pathFrom, os.TempDir())
	tmpTo := strings.HasPrefix(pathTo, os.TempDir())
	if tmpFrom && !tmpTo {
		pathFrom = pathTo
	} else if !tmpFrom && tmpTo {
		pathTo = pathFrom
	}

	v.Add("base", dir)
	v.Add("left", pathFrom)
	v.Add("right", pathTo)
	v.Add("diff", f.Name())
	u.RawQuery = v.Encode()
	exec.Command("open", u.String()).Run()
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
	return delta.Diff(string(from), string(to)), nil
}

func printDiff(pathFrom, pathTo string, html bool) {
	d, err := diff(pathFrom, pathTo)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		return
	}
	if html {
		fmt.Println(d.HTML())
	} else {
		for i, l := range d.Raw() {
			if l[2] == "=" && l[0] == l[1] {
				fmt.Printf("%d %s = %s \n", i, l[2], l[0])
				continue
			}
			if l[0] != "" {
				fmt.Printf("\x1b[31m%d %s < %s\x1b[0m\n", i, l[2], l[0])
			}
			if l[1] != "" {
				fmt.Printf("\x1b[32m%d %s > %s\x1b[0m\n", i, l[2], l[1])
			}
		}
	}
}
