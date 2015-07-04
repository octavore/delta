package delta

import (
	"bytes"
	"html/template"
	"strconv"
)

// DiffSolution contains a set of lines, where each element of
// lines comprises the left and right line, and whether the change
// was from A or B.
type DiffSolution struct {
	lines [][3]string
}

func (d *DiffSolution) addLineA(a string) {
	d.addLine(a, "", LineFromA)
}

func (d *DiffSolution) addLineB(b string) {
	d.addLine("", b, LineFromB)
}

func (d *DiffSolution) addLine(a, b string, l lineSource) {
	d.lines = append(d.lines, [3]string{a, b, string(l)})
}

func (d *DiffSolution) Raw() [][3]string {
	return d.lines
}

var tpl = template.Must(template.New("div").Parse("<div class='{{.Classes}}'>{{.Contents}}</div>\n"))

type div struct {
	Classes  string
	Contents interface{}
}

// HTML builds up a html-friendly diff.
func (d *DiffSolution) HTML() string {
	// closest contains the number of lines to the next changed lines
	maxContext := 10
	maxContext++ // + 1 for lines to hide
	nextChange := make([]int, len(d.lines))
	lastChangedLine := len(d.lines) + 10
	for i := len(d.lines) - 1; i > -1; i-- {
		switch lineSource(d.lines[i][2]) {
		case LineFromA, LineFromB, LineFromBothEdit:
			lastChangedLine = i
		case LineFromBoth:
			if d.lines[i][0] != d.lines[i][1] {
				lastChangedLine = i
			}
		}
		nextChange[i] = lastChangedLine - i
		if nextChange[i] > maxContext {
			nextChange[i] = maxContext
		}
	}

	li, ri := 0, 0
	lg := bytes.NewBufferString("<div class='gutter'>\n")
	rg := bytes.NewBufferString("<div class='gutter'>\n")
	lb := bytes.NewBufferString("<div id='diff-left' class='diff-pane'><div class='diff-pane-contents'>\n")
	rb := bytes.NewBufferString("<div id='diff-right' class='diff-pane'><div class='diff-pane-contents'>\n")
	lastChangedLine = -maxContext
	for i, l := range d.lines {
		ls := lineSource(l[2])
		closestChange := 0
		if ls == LineFromBoth && l[0] == l[1] {
			closestChange = i - lastChangedLine
			if closestChange > nextChange[i] {
				closestChange = nextChange[i]
			}
		} else {
			lastChangedLine = i
		}
		lc := "line-context-" + strconv.Itoa(closestChange) + " line "
		if ls == LineFromA {
			li++
			must(tpl.Execute(lg, div{lc + "line-a line-addition", li}))
			must(tpl.Execute(rg, div{lc + "line-b", ""}))
			must(tpl.Execute(lb, div{lc + "line-a line-addition", l[0]}))
			must(tpl.Execute(rb, div{lc + "line-b", ""}))
		} else if ls == LineFromB {
			ri++
			must(tpl.Execute(lg, div{lc + "line-a", ""}))
			must(tpl.Execute(rg, div{lc + "line-b line-addition", ri}))
			must(tpl.Execute(lb, div{lc + "line-a", ""}))
			must(tpl.Execute(rb, div{lc + "line-b line-addition", l[1]}))
		} else if ls == LineFromBothEdit {
			li++
			ri++
			must(tpl.Execute(lg, div{lc + "line-a line-mismatch", li}))
			must(tpl.Execute(rg, div{lc + "line-b line-mismatch", ri}))
			must(tpl.Execute(lb, div{lc + "line-a line-mismatch", l[0]}))
			must(tpl.Execute(rb, div{lc + "line-b line-mismatch", l[1]}))
		} else if l[0] != l[1] {
			li++
			ri++
			must(tpl.Execute(lg, div{lc + "line-a line-ws", li}))
			must(tpl.Execute(rg, div{lc + "line-b line-ws", ri}))
			must(tpl.Execute(lb, div{lc + "line-a line-ws", l[0]}))
			must(tpl.Execute(rb, div{lc + "line-b line-ws", l[1]}))
		} else if ls == LineFromBoth {
			li++
			ri++
			must(tpl.Execute(lg, div{lc + "line-a line-match", li}))
			must(tpl.Execute(rg, div{lc + "line-b line-match", ri}))
			must(tpl.Execute(lb, div{lc + "line-a line-match", l[0]}))
			must(tpl.Execute(rb, div{lc + "line-b line-match", l[1]}))
		}
	}

	lg.WriteString("</div>")
	rg.WriteString("</div>")
	lb.WriteString("</div></div>")
	rb.WriteString("</div></div>")
	if li == 0 {
		return rg.String() + rb.String()
	}
	if ri == 0 {
		return lg.String() + lb.String()
	}
	return lg.String() + lb.String() + rg.String() + rb.String()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
