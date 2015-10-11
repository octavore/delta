package formatter

import (
	"bytes"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/octavore/delta"
)

const (
	divTmpl  = "<div class='{{.Classes}}'>{{.Contents}}</div>\n"
	spanTmpl = "<span class='{{.Classes}}'>{{.Contents}}</span>"
)

var (
	div  = template.Must(template.New("div").Parse(divTmpl))
	span = template.Must(template.New("span").Parse(spanTmpl))
)

type elem struct {
	Classes  string
	Contents interface{}
}

var svgClasses = map[delta.LineSource]string{
	delta.LineFromA:        "add",
	delta.LineFromB:        "del",
	delta.LineFromBothEdit: "edit",
}

// HTMLLine renders a diff solution into a before and after string.
// Words are added one at a time, and changes are marked with spans.
func HTMLLine(d *delta.DiffSolution) (string, string) {
	a := &bytes.Buffer{}
	b := &bytes.Buffer{}
	for _, word := range d.Lines {
		switch delta.LineSource(word[2]) {
		case delta.LineFromA:
			span.Execute(a, elem{"w-add", word[0]})
			span.Execute(b, elem{"w-del", ""})
		case delta.LineFromB:
			span.Execute(a, elem{"w-del", ""})
			span.Execute(b, elem{"w-add", word[1]})
		case delta.LineFromBothEdit:
			span.Execute(a, elem{"w-edit", word[0]})
			span.Execute(b, elem{"w-edit", word[1]})
		case delta.LineFromBoth:
			a.WriteString(template.HTMLEscapeString(word[0]))
			b.WriteString(template.HTMLEscapeString(word[1]))
		}
	}
	return a.String(), b.String()
}

// HTML builds up a html diff. Here be dragons! This is meant for the delta GUI.
func HTML(d *delta.DiffSolution) string {
	// closest contains the number of lines to the *next* changed lines
	maxContext := 10
	maxContext++ // + 1 for lines to hide
	nextChange := make([]int, len(d.Lines))
	lastChangedLine := len(d.Lines) + 10
	for i := len(d.Lines) - 1; i > -1; i-- {
		switch delta.LineSource(d.Lines[i][2]) {
		case delta.LineFromA, delta.LineFromB, delta.LineFromBothEdit:
			lastChangedLine = i
		case delta.LineFromBoth:
			if d.Lines[i][0] != d.Lines[i][1] {
				lastChangedLine = i
			}
		}
		nextChange[i] = lastChangedLine - i
		if nextChange[i] > maxContext {
			nextChange[i] = maxContext
		}
	}

	li, ri := 0, 0
	lg := bytes.NewBufferString("<div id='gutter-left' class='gutter'>\n")
	rg := bytes.NewBufferString("<div id='gutter-right' class='gutter'>\n")
	lb := bytes.NewBufferString("<div id='diff-left' class='diff-pane'><div class='diff-pane-contents'>\n")
	rb := bytes.NewBufferString("<div id='diff-right' class='diff-pane'><div class='diff-pane-contents'>\n")
	lastChangedLine = -maxContext

	lastSource := delta.LineFromBoth
	lineHeight := 16
	ll := bytes.NewBufferString(fmt.Sprintf(`<div><svg width="16" height="%d">`, lineHeight*len(d.Lines)))

	for i, l := range d.Lines {
		ls := delta.LineSource(l[2])
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

		// closestChange keeps track of how close we are to the *previous* change.
		closestChange := 0
		if ls == delta.LineFromBoth && l[0] == l[1] {
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
		lc := "lc-" + strconv.Itoa(closestChange) + " line "
		if ls == delta.LineFromA {
			li++
			must(div.Execute(lg, elem{lc + "la", li}))
			must(div.Execute(rg, elem{lc, ""}))
			must(div.Execute(lb, elem{lc + "la", l[0]}))
			must(div.Execute(rb, elem{lc, ""}))
		} else if ls == delta.LineFromB {
			ri++
			must(div.Execute(lg, elem{lc, ""}))
			must(div.Execute(rg, elem{lc + "la", ri}))
			must(div.Execute(lb, elem{lc, ""}))
			must(div.Execute(rb, elem{lc + "la", l[1]}))
		} else if ls == delta.LineFromBothEdit {
			li++
			ri++
			dl, dr := "", ""
			sol := delta.DiffLine(l[0], l[1])
			if sol != nil {
				dl, dr = HTMLLine(sol)
			} else {
				dl = template.HTMLEscapeString(l[0])
				dr = template.HTMLEscapeString(l[1])
			}
			must(div.Execute(lg, elem{lc + "ln", li}))
			must(div.Execute(rg, elem{lc + "ln", ri}))
			must(div.Execute(lb, elem{lc + "ln", template.HTML(dl)}))
			must(div.Execute(rb, elem{lc + "ln", template.HTML(dr)}))
		} else if l[0] != l[1] {
			li++
			ri++
			must(div.Execute(lg, elem{lc + "line-ws", li}))
			must(div.Execute(rg, elem{lc + "line-ws", ri}))
			must(div.Execute(lb, elem{lc + "line-ws", l[0]}))
			must(div.Execute(rb, elem{lc + "line-ws", l[1]}))
		} else if ls == delta.LineFromBoth {
			li++
			ri++
			must(div.Execute(lg, elem{lc + "lm", li}))
			must(div.Execute(rg, elem{lc + "lm", ri}))
			must(div.Execute(lb, elem{lc + "lm", l[0]}))
			must(div.Execute(rb, elem{lc + "lm", l[1]}))
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
