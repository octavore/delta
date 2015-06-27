package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"

	"bitbucket.org/pancakeio/delta/delta"
)

func main() {
	open := flag.Bool("open", false, "open the file in the gui")
	html := flag.Bool("html", false, "print out html")
	flag.Parse()

	pathFrom := flag.Arg(0)
	pathTo := flag.Arg(1)

	if *open {
		openDiff(pathFrom, pathTo)
	} else {
		printDiff(pathFrom, pathTo, *html)
	}
}

func openDiff(pathFrom, pathTo string) {
	dir, _ := os.Getwd()
	u, _ := url.Parse("delta://open")
	v := url.Values{}
	v.Add("base", dir)
	v.Add("left", pathFrom)
	v.Add("right", pathTo)
	u.RawQuery = v.Encode()
	exec.Command("open", u.String()).Run()
}

func printDiff(pathFrom, pathTo string, html bool) {
	from, err := ioutil.ReadFile(pathFrom)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("error reading %q: %v", pathFrom, err))
		return
	}
	to, err := ioutil.ReadFile(pathTo)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("error reading %q: %v", pathTo, err))
		return
	}

	d := delta.Diff(string(from), string(to))
	if html {
		fmt.Println(d.HTML())
	} else {
		fmt.Println(d.Raw())
	}
}
