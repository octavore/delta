package delta

import (
	"fmt"
	"math"
	"strings"
)

// Diff two strings and return a Differ with the solution
func Diff(a, b string) *DiffSolution {
	aw := strings.Split(a, "\n")
	bw := strings.Split(b, "\n")
	d := NewDiffer(aw, bw)
	return d.Solve()
}

type Differ struct {
	a        []string
	b        []string
	ab       [][]int32
	solution [][]lineSource
}
type lineSource int

const (
	Unknown lineSource = iota
	LineFromA
	LineFromB
	LineFromBoth
)

func NewDiffer(a, b []string) *Differ {
	ab := make([][]int32, len(a))
	solution := make([][]lineSource, len(a))
	for i := range ab {
		ab[i] = make([]int32, len(b))
		solution[i] = make([]lineSource, len(b))
	}
	return &Differ{
		a:        a,
		b:        b,
		ab:       ab,
		solution: solution,
	}
}

// scores for the algorithm
const (
	deletionScore = -1
	matchScore    = 1
	mismatchScore = -1
)

func (d *Differ) Solve() *DiffSolution {
	_ = d.computeOptimal(0, 0)
	return d.getSolution(0, 0)
}

// computeOptimal computes the optimal (maximum) score, which
// corresponds to the best diff
func (d *Differ) computeOptimal(ai, bi int) int32 {
	// return memoized result
	if i := d.solution[ai][bi]; i != Unknown {
		return d.ab[ai][bi]
	}

	// base case
	if ai == len(d.a)-1 && bi == len(d.b)-1 {
		return 0
	}

	// initialize score to infinity
	d.ab[ai][bi] = -math.MaxInt32

	// skip a
	if ai < len(d.a)-1 {
		s := d.computeOptimal(ai+1, bi) + deletionScore
		if s > d.ab[ai][bi] {
			d.ab[ai][bi] = s
			d.solution[ai][bi] = LineFromA
		}
	}

	// skip b
	if bi < len(d.b)-1 {
		s := d.computeOptimal(ai, bi+1) + deletionScore
		if s > d.ab[ai][bi] {
			d.ab[ai][bi] = s
			d.solution[ai][bi] = LineFromB
		}
	}

	// skip both
	if ai < len(d.a)-1 && bi < len(d.b)-1 {
		s := d.computeOptimal(ai+1, bi+1)
		if d.a[ai] == d.b[bi] {
			s += matchScore
		} else {
			s += mismatchScore
		}
		if s > d.ab[ai][bi] {
			d.ab[ai][bi] = s
			d.solution[ai][bi] = LineFromBoth
		}
	}

	return d.ab[ai][bi]
}

func (d *Differ) debug() {
	for _, l := range d.ab {
		for _, score := range l {
			fmt.Printf("%3d", score)
		}
		fmt.Println()
	}
}

func (d *Differ) printSolution(a, b int) {
	// no more string
	if a == len(d.a)-1 && b == len(d.b)-1 {
		return
	}

	// no more a
	if a == len(d.a)-1 {
		fmt.Println("            ", d.b[b])
		b++
		d.printSolution(a, b)
		return
	}

	// no more b
	if b == len(d.b)-1 {
		fmt.Println(d.a[a])
		a++
		d.printSolution(a, b)
		return
	}

	switch d.solution[a][b] {
	case LineFromA:
		fmt.Println(d.a[a])
		a++
	case LineFromB:
		fmt.Println("            ", d.b[b])
		b++
	case LineFromBoth:
		fmt.Println(d.a[a], "            ", d.b[b])
		a++
		b++
	}
	d.printSolution(a, b)
}

func (d *Differ) getSolution(a, b int) *DiffSolution {
	s := &DiffSolution{}
	// no more string
	for a != len(d.a)-1 || b != len(d.b)-1 {
		// no more a
		if a == len(d.a)-1 {
			s.addLineB(d.b[b])
			b++
			continue
		}

		// no more b
		if b == len(d.b)-1 {
			s.addLineA(d.a[a])
			a++
			continue
		}

		switch d.solution[a][b] {
		case LineFromA:
			s.addLineA(d.a[a])
			a++
		case LineFromB:
			s.addLineB(d.b[b])
			b++
		case LineFromBoth:
			s.addLine(d.a[a], d.b[b])
			a++
			b++
		}
	}
	return s
}

type DiffSolution struct {
	lines [][2]*string
}

func (d *DiffSolution) addLineA(a string) {
	d.lines = append(d.lines, [2]*string{&a, nil})
}

func (d *DiffSolution) addLineB(b string) {
	d.lines = append(d.lines, [2]*string{nil, &b})
}

func (d *DiffSolution) addLine(a, b string) {
	d.lines = append(d.lines, [2]*string{&a, &b})
}

func (d *DiffSolution) HTML() string {
	li, ri := 0, 0
	lg := "<div class='gutter'>"
	rg := "<div class='gutter'>"
	left := "<div id='diff-left' class='diff-pane'>"
	right := "<div id='diff-right' class='diff-pane'>"
	for _, l := range d.lines {
		if l[0] != nil && l[1] == nil {
			li++
			lg += fmt.Sprintf("<div class='line line-number'>%d</div>", li)
			rg += "<div class='line'></div>"
			left += fmt.Sprintf("<div class='line line-a line-addition'>%s</div>", *l[0])
			right += "<div class='line line-b'></div>"
		} else if l[0] == nil && l[1] != nil {
			ri++
			lg += "<div class='line'></div>"
			rg += fmt.Sprintf("<div class='line line-number'>%d</div>", ri)
			left += "<div class='line line-a'></div>"
			right += fmt.Sprintf("<div class='line line-b line-addition'>%s</div>", *l[1])
		} else if *l[0] == *l[1] {
			li++
			ri++
			lg += fmt.Sprintf("<div class='line line-number'>%d</div>", li)
			rg += fmt.Sprintf("<div class='line line-number'>%d</div>", ri)
			left += fmt.Sprintf("<div class='line line-a line-match'>%s</div>", *l[0])
			right += fmt.Sprintf("<div class='line line-b line-match'>%s</div>", *l[1])
		} else if *l[0] != *l[1] {
			li++
			ri++
			lg += fmt.Sprintf("<div class='line line-number'>%d</div>", li)
			rg += fmt.Sprintf("<div class='line line-number'>%d</div>", ri)
			left += fmt.Sprintf("<div class='line line-a line-mismatch'>%s</div>", *l[0])
			right += fmt.Sprintf("<div class='line line-b line-mismatch'>%s</div>", *l[1])
		}
	}
	lg += "</div>"
	rg += "</div>"
	left += "</div>"
	right += "</div>"

	return lg + left + rg + right
}
