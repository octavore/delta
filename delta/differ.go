package delta

import (
	"math"
	"strconv"
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
	weights  Weights
	ab       [][]int32
	solution [][]lineSource
}

type lineSource string

const (
	Unknown          lineSource = ""
	LineFromA        lineSource = "<"
	LineFromB        lineSource = ">"
	LineFromBoth     lineSource = "="
	LineFromBothEdit lineSource = "~"
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
		weights:  defaultWeights,
		solution: solution,
	}
}

// scores for the algorithm
type Weights struct {
	Deletion int32
	Match    int32
	Mismatch int32
	NewMode  int32
}

var defaultWeights = Weights{
	Deletion: -1,
	Match:    1,
	Mismatch: 0,
	NewMode:  0,
}

func (d *Differ) Solve() *DiffSolution {
	s := &DiffSolution{}
	m := modeBeginning
	// copy over shared prefix
	var i int
	for ; i < len(d.a) && i < len(d.b); i++ {
		if strings.TrimSpace(d.a[i]) != strings.TrimSpace(d.b[i]) {
			break
		}
		s.addLine(d.a[i], d.b[i], LineFromBoth)
	}
	if i > 0 {
		m = modeMatch
	}

	// compute optimal
	_ = d.computeOptimal(i, i, m)
	d.getSolution(s, i, i)
	return s
}

type blockMode int

const (
	modeBeginning blockMode = iota
	modeDeleteA
	modeDeleteB
	modeMatch
	modeMismatch
)

// computeOptimal computes the optimal (maximum) score, which
// corresponds to the best diff
func (d *Differ) computeOptimal(ai, bi int, m blockMode) int32 {
	// return memoized result
	if i := d.solution[ai][bi]; i != Unknown {
		return d.ab[ai][bi]
	}

	// base case: no more lines to align
	if ai == len(d.a)-1 && bi == len(d.b)-1 {
		return 0
	}

	// initialize score to infinity
	d.ab[ai][bi] = -math.MaxInt32

	// skip a, addition in b/deletion in a
	if ai < len(d.a)-1 {
		s := d.computeOptimal(ai+1, bi, modeDeleteA) + d.weights.Deletion
		if m != modeDeleteA {
			s += d.weights.NewMode
		}
		if s > d.ab[ai][bi] {
			d.ab[ai][bi] = s
			d.solution[ai][bi] = LineFromA
		}
	}

	// skip b, addition in a/deletion in b
	if bi < len(d.b)-1 {
		s := d.computeOptimal(ai, bi+1, modeDeleteB) + d.weights.Deletion
		if m != modeDeleteB {
			s += d.weights.NewMode
		}
		if s > d.ab[ai][bi] {
			d.ab[ai][bi] = s
			d.solution[ai][bi] = LineFromB
		}
	}

	// align lines
	if ai < len(d.a)-1 && bi < len(d.b)-1 {
		var n blockMode
		var s int32
		if strings.TrimSpace(d.a[ai]) == strings.TrimSpace(d.b[bi]) {
			n = modeMatch
			s = d.weights.Match
		} else {
			n = modeMismatch
			s = d.weights.Mismatch
		}
		s += d.computeOptimal(ai+1, bi+1, n)
		if m != n {
			s += d.weights.NewMode
		}
		if s > d.ab[ai][bi] {
			d.ab[ai][bi] = s
			// todo: consolidate with n
			if n == modeMatch {
				d.solution[ai][bi] = LineFromBoth
			} else {
				d.solution[ai][bi] = LineFromBothEdit
			}
		}
	}

	return d.ab[ai][bi]
}

func (d *Differ) debug() string {
	o := ""
	for _, l := range d.ab {
		for _, score := range l {
			o += " " + strconv.Itoa(int(score))
		}
		o += "\n"
	}
	return o
}

// this needs a second pass which minimizes number of change blocks
func (d *Differ) getSolution(s *DiffSolution, a, b int) {
	// iterate until no more string
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
		case LineFromBoth, LineFromBothEdit:
			s.addLine(d.a[a], d.b[b], d.solution[a][b])
			a++
			b++
		}
	}
}
