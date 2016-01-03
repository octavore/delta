# Delta

Delta is both a Go library and a command-line utility for text diffs. Diffs
can be displayed in the browser, or in command-line using the `--cli` flag.

## Installation

    brew install octavore/tools/delta

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

## Other Usage

```
delta --cli <fileA> <fileB>         # print text diff to stdout
delta --cli --html <fileA> <fileB>  # print html diff to stdout
delta --gist <fileA> <fileB>        # upload html diff to a gist
```

## Configure Git

The `delta` binary must be on your `$PATH` in order for this work. The
following are helpers for adding `delta` to your `~/.gitconfig` file.

    delta --install   # makes delta the default for `git difftool`
    delta --uninstall # remove delta from your gitconfig

## User Config

You can configure `delta` using a `~/.deltarc` file, for example:

```
{
  "context": 9,
  "showEmpty": true,
  "shouldCollapse": false,
  "highlight": true,
  "unmodifiedOpacity": 0.8,
  "diffFontSize": 12
}
```

### Options

config key          | key type  | description
------------------- | --------- | ------------------------------------------
`context`           | `integer` | between 0 and 4 number of lines of context to show
`showEmpty`         | `bool`    | whether to hide empty lines
`shouldCollapse`    | `bool`    | whether to merge browser tabs
`highlight`         | `bool`    | toggles syntax highlighting
`unmodifiedOpacity` | `float`   | opacity of unmodified lines, between 0.1 and 1
`diffFontSize`      | `integer` | font size of the diff

## Browser Support

![Screenshot](https://raw.github.com/octavore/delta/master/screenshot.jpg)

Delta works best in Chrome and Safari. You will see a separate tab open for
each diff file, and then the tabs will consolidate into a single tab. In
Firefox, each diff will remain in separate tabs. This is because in Firefox,
each file receives its own IndexedDB instance, instead of a shared instance.

Browser support relies on the following open source libraries:

- [Mithril](http://mithril.js.org/)
- [highlight.js](https://highlightjs.org/)
- [Mousetrap](https://craig.is/killing/mice)
- [PouchDB](http://pouchdb.com/)

## Development

### Compiling From Source

    go get github.com/octavore/delta/delta

If `$GOPATH/bin` is on your `$PATH`, run `delta -h` for usage.

### Regenerating Assets

You will need `npm` to install the necessary node packages for compiling the
front-end assets.

[go-bindata](https://github.com/jteeuwen/go-bindata) is used to generate the
`bindata.go`, which allows us to embed static resources into the compiled
binary.

## TODO

- make differ/histogram diff functions more consistent and add an interface.
- more comments for godoc/golint.
- fix race condition in update
- favicon
- make sidebar resizable

