package delta

import (
	"strconv"
)

type DiffSolution struct {
	lines [][3]string
}

func (d *DiffSolution) addLineA(a string) {
	d.lines = append(d.lines, [3]string{a, "", string(LineFromA)})
}

func (d *DiffSolution) addLineB(b string) {
	d.lines = append(d.lines, [3]string{"", b, string(LineFromB)})
}

func (d *DiffSolution) addLine(a, b string, l lineSource) {
	d.lines = append(d.lines, [3]string{a, b, string(l)})
}

func (d *DiffSolution) Raw() [][3]string {
	return d.lines
}

func (d *DiffSolution) HTML() string {
	li, ri := 0, 0
	lg := "<div class='gutter'>"
	rg := "<div class='gutter'>"
	left := "<div id='diff-left' class='diff-pane'><div class='diff-pane-contents'>"
	right := "<div id='diff-right' class='diff-pane'><div class='diff-pane-contents'>"
	for _, l := range d.lines {
		if l[2] == string(LineFromA) {
			li++
			lg += "<div class='line line-a line-number line-addition'>" + strconv.Itoa(li) + "</div>\n"
			rg += "<div class='line line-b'></div>\n"
			left += "<div class='line line-a line-addition'>" + l[0] + "</div>\n"
			right += "<div class='line line-b'></div>\n"
		} else if l[2] == string(LineFromB) {
			ri++
			lg += "<div class='line line-a'></div>\n"
			rg += "<div class='line line-b line-number line-addition'>" + strconv.Itoa(ri) + "</div>\n"
			left += "<div class='line line-a'></div>\n"
			right += "<div class='line line-b line-addition'>" + l[1] + "</div>\n"
		} else if l[2] == string(LineFromBothEdit) || l[0] != l[1] {
			li++
			ri++
			lg += "<div class='line line-a line-number line-mismatch'>" + strconv.Itoa(li) + "</div>\n"
			rg += "<div class='line line-b line-number line-mismatch'>" + strconv.Itoa(ri) + "</div>\n"
			left += "<div class='line line-a line-mismatch'>" + l[0] + "</div>\n"
			right += "<div class='line line-b line-mismatch'>" + l[1] + "</div>\n"
		} else if l[2] == string(LineFromBoth) {
			li++
			ri++
			lg += "<div class='line line-a line-number line-match'>" + strconv.Itoa(li) + "</div>\n"
			rg += "<div class='line line-b line-number line-match'>" + strconv.Itoa(ri) + "</div>\n"
			left += "<div class='line line-a line-match'>" + l[0] + "</div>\n"
			right += "<div class='line line-b line-match'>" + l[1] + "</div>\n"
		}
	}
	lg += "</div>\n"
	rg += "</div>\n"
	left += "</div></div>\n"
	right += "</div></div>\n"

	return lg + left + rg + right
}
