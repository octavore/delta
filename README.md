# Delta

Delta is both a Go library and a command-line utility.

## Installation

    go get github.com/octavore/delta/delta

If `$GOPATH/bin` is on your `$PATH`, run `delta -h` for usage.

## Description

Delta implements two diff functions: Smith-Waterman, and histogram diff.

[Smith-Waterman](https://en.wikipedia.org/wiki/Smith%E2%80%93Waterman_algorithm) 
is a dynamic programming algorithm for aligning two sequences, in this case text
sequences. It originates from bioinformatics, where it is used for aligning DNA sequences.

[histogram diff](http://download.eclipse.org/jgit/docs/jgit-2.0.0.201206130900-r/apidocs/org/eclipse/jgit/diff/HistogramDiff.html)
is a diff algorithm first implemented in `JGit` and subsequently ported over
to `git`, where it can be used with the `git diff --histogram` command. This
implementation post processes the histogram diff in order to push down match
regions as far as possible. `git` also post processes diffs.

## TODO

- add support for hiding unchanged parts of the diff.
- make differ/histogram diff functions more consistent and add an interface. 
- more comments for godoc/golint.