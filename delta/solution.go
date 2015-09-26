package delta

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"
	"strings"
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

func (d *DiffSolution) addSolution(e *DiffSolution) {
	d.lines = append(d.lines, e.lines...)
}

func (d *DiffSolution) Raw() [][3]string {
	return d.lines
}

var div = template.Must(template.New("div").Parse("<div class='{{.Classes}}'>{{.Contents}}</div>\n"))
var span = template.Must(template.New("span").Parse("<span class='{{.Classes}}'>{{.Contents}}</span>"))

type elem struct {
	Classes  string
	Contents interface{}
}

var svgClasses = map[lineSource]string{
	LineFromA:        "add",
	LineFromB:        "del",
	LineFromBothEdit: "edit",
}

func (d *DiffSolution) HTMLLine() (string, string) {
	a := &bytes.Buffer{}
	b := &bytes.Buffer{}
	for _, word := range d.lines {
		switch lineSource(word[2]) {
		case LineFromA:
			span.Execute(a, elem{"w-add", word[0]})
			span.Execute(b, elem{"w-del", ""})
		case LineFromB:
			span.Execute(a, elem{"w-del", ""})
			span.Execute(b, elem{"w-add", word[1]})
		case LineFromBothEdit:
			span.Execute(a, elem{"w-edit", word[0]})
			span.Execute(b, elem{"w-edit", word[1]})
		case LineFromBoth:
			a.WriteString(template.HTMLEscapeString(word[0]))
			b.WriteString(template.HTMLEscapeString(word[1]))
		}
	}
	return a.String(), b.String()
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

	lastSource := LineFromBoth
	lineHeight := 16
	ll := bytes.NewBufferString(fmt.Sprintf(`<div><svg width="16" height="%d">`, lineHeight*len(d.lines)))

	for i, l := range d.lines {
		ls := lineSource(l[2])
		if ls != lastSource {
			lastSource = ls
			if l[0] != l[1] {
				ll.WriteString(
					fmt.Sprintf(`<line x1="%d" x2="%d" y1="%d" y2="%d" stroke-width="1" class="connector-%s" />`,
						0, 16, lineHeight*li, lineHeight*ri, svgClasses[lastSource],
					),
				)
			}
		}

		closestChange := 0
		if ls == LineFromBoth && l[0] == l[1] {
			closestChange = i - lastChangedLine
			if closestChange > nextChange[i] {
				closestChange = nextChange[i]
			}
			if closestChange == maxContext {
				closestChange = -1
			}
		} else {
			lastChangedLine = i
		}
		lc := "line-context-" + strconv.Itoa(closestChange) + " line "
		if ls == LineFromA {
			li++
			must(div.Execute(lg, elem{lc + "line-a line-addition", li}))
			must(div.Execute(rg, elem{lc + "line-b", ""}))
			must(div.Execute(lb, elem{lc + "line-a line-addition", l[0]}))
			must(div.Execute(rb, elem{lc + "line-b", ""}))
		} else if ls == LineFromB {
			ri++
			must(div.Execute(lg, elem{lc + "line-a", ""}))
			must(div.Execute(rg, elem{lc + "line-b line-addition", ri}))
			must(div.Execute(lb, elem{lc + "line-a", ""}))
			must(div.Execute(rb, elem{lc + "line-b line-addition", l[1]}))
		} else if ls == LineFromBothEdit {
			li++
			ri++
			dl, dr := "", ""
			sol := DiffLine(l[0], l[1])
			if sol != nil {
				dl, dr = sol.HTMLLine()
			} else {
				dl = template.HTMLEscapeString(l[0])
				dr = template.HTMLEscapeString(l[1])
			}
			must(div.Execute(lg, elem{lc + "line-a line-mismatch", li}))
			must(div.Execute(rg, elem{lc + "line-b line-mismatch", ri}))
			must(div.Execute(lb, elem{lc + "line-a line-mismatch", template.HTML(dl)}))
			must(div.Execute(rb, elem{lc + "line-b line-mismatch", template.HTML(dr)}))
		} else if l[0] != l[1] {
			li++
			ri++
			must(div.Execute(lg, elem{lc + "line-a line-ws", li}))
			must(div.Execute(rg, elem{lc + "line-b line-ws", ri}))
			must(div.Execute(lb, elem{lc + "line-a line-ws", l[0]}))
			must(div.Execute(rb, elem{lc + "line-b line-ws", l[1]}))
		} else if ls == LineFromBoth {
			li++
			ri++
			must(div.Execute(lg, elem{lc + "line-a line-match", li}))
			must(div.Execute(rg, elem{lc + "line-b line-match", ri}))
			must(div.Execute(lb, elem{lc + "line-a line-match", l[0]}))
			must(div.Execute(rb, elem{lc + "line-b line-match", l[1]}))
		}
	}

	lg.WriteString("</div>")
	rg.WriteString("</div>")
	lb.WriteString("</div></div>")
	rb.WriteString("</div></div>")
	ll.WriteString("</div>")
	lbs := strings.Replace(lb.String(), "\t", "<span class='delta-tab'>\t</span>", -1)
	rbs := strings.Replace(rb.String(), "\t", "<span class='delta-tab'>\t</span>", -1)
	if li == 0 {
		return rg.String() + rbs
	}
	if ri == 0 {
		return lg.String() + lbs
	}
	return lg.String() + lbs + rg.String() + rbs
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
