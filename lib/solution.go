package delta

// LineSource indicates the origin of the solution line.
type LineSource string

// These are valid values for LineSource.
const (
	Unknown          LineSource = ""
	LineFromA        LineSource = "<"
	LineFromB        LineSource = ">"
	LineFromBoth     LineSource = "="
	LineFromBothEdit LineSource = "~"
)

// DiffSolution contains a set of lines, where each element of
// lines comprises the left and right line, and whether the change
// was from A or B.
type DiffSolution struct {
	Lines [][3]string
}

func (d *DiffSolution) addLineA(a string) {
	d.addLine(a, "", LineFromA)
}

func (d *DiffSolution) addLineB(b string) {
	d.addLine("", b, LineFromB)
}

func (d *DiffSolution) addLine(a, b string, l LineSource) {
	d.Lines = append(d.Lines, [3]string{a, b, string(l)})
}

func (d *DiffSolution) addSolution(e *DiffSolution) {
	d.Lines = append(d.Lines, e.Lines...)
}

// PostProcess loops over the solution. For each changed region, see if we can
// move it forward. i.e. if we have the following changeset:
// 	 a [b c d] b c
// then we move the modified region forward so we have instead:
//   a b c [d b c]
// this heuristic only moves additions or deletions (but never both in a move).
func (d *DiffSolution) PostProcess() {
	lastChangeStartIndex := -1
	lastChangeType := Unknown
	lastLineType := LineFromBoth
	for i, word := range d.Lines {
		currentLineType := LineSource(word[2])
		// we've reached the end of a region. Now we try find a section to move forward.
		if currentLineType == LineFromBoth && currentLineType != lastLineType {
			if lastChangeType != LineFromB && lastChangeType != LineFromA {
				// don't try to move if it wasn't an addition or deletion
				goto ContinueProcessing
			}

			// walk the change region to find a match
			p1 := lastChangeStartIndex
			p2 := i
			for ((lastChangeType == LineFromA && d.Lines[p1][0] == d.Lines[p2][0]) ||
				(lastChangeType == LineFromB && d.Lines[p1][1] == d.Lines[p2][1])) &&
				LineSource(d.Lines[p2][2]) == LineFromBoth {
				d.Lines[p1], d.Lines[p2] = d.Lines[p2], d.Lines[p1]
				p1++
				p2++
				if p2 >= len(d.Lines) {
					break
				}
			}
		}

		// we've reached the beginning of a region. Update pointers.
		if lastLineType != currentLineType {
			lastChangeStartIndex = i
			lastChangeType = currentLineType
		}

	ContinueProcessing:
		lastLineType = currentLineType
	}
}
