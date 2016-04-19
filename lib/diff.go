package delta

// Solver is an interface implemented by the diff algorithms.
type Solver interface {
	Solve() *DiffSolution
}

var (
	_ Solver = &HistogramDiffer{}
	_ Solver = &SequenceDiffer{}
)
