package delta

import (
	"strings"
)

// HistogramDiff uses the histogram diff algorithm to generate
// a line-based diff between two strings
func HistogramDiff(a, b string) *DiffSolution {
	aw := strings.Split(a, "\n")
	bw := strings.Split(b, "\n")
	return NewHistogramDiffer(aw, bw).Solve()
}

// matchRegion delineates a region of A and B which are equal.
// start is inclusive but end is exclusive
type matchRegion struct {
	aStart, aEnd int
	bStart, bEnd int
	matchScore   int
}

func (m *matchRegion) validStart(as, bs int) bool {
	return as < m.aStart && bs < m.bStart
}

func (m *matchRegion) validEnd(ae, be int) bool {
	return m.aEnd < ae && m.bEnd < be
}

func (m *matchRegion) length() int {
	if m == nil {
		return 0
	}
	return m.aEnd - m.aStart
}

// HistogramDiffer implements the histogram diff algorithm.
type HistogramDiffer struct {
	a []string
	b []string
}

// NewHistogramDiffer returns a HistogramDiffer which diffs the given sequence of words.
func NewHistogramDiffer(a, b []string) *HistogramDiffer {
	return &HistogramDiffer{a: a, b: b}
}

// createAHistogram creates a map of lines to an array of line numbers
// corresponding to occurrences of that particular line in the A input.
func (h *HistogramDiffer) createAHistogram(aStart, aEnd int) map[string][]int {
	histogram := map[string][]int{}
	for i := aStart; i < aEnd; i++ {
		line := strings.TrimSpace(h.a[i])
		histogram[line] = append(histogram[line], i)
	}
	return histogram
}

func (h *HistogramDiffer) eq(aIdx, bIdx int) bool {
	return strings.TrimSpace(h.a[aIdx]) == strings.TrimSpace(h.b[bIdx])
}

// longestSubstring finds the longest matching region in the given area of the
// inputs A and B.
func (h *HistogramDiffer) longestSubstring(aStart, aEnd, bStart, bEnd int) *matchRegion {
	var bestMatch *matchRegion
	bestMatchScore := aEnd - aStart
	histogram := h.createAHistogram(aStart, aEnd)
	for bIdx := bStart; bIdx < bEnd; {
		nextB := bIdx + 1
		lineB := strings.TrimSpace(h.b[bIdx])

		// only consider low-occurence elements
		if len(histogram[lineB]) > bestMatchScore {
			bIdx = nextB
			continue
		}

		// for all matching lines in A, find the longest substring
		// implict: if _, ok := histogram[lineB]; !ok { continue }
		prevA := aStart
		for _, as := range histogram[lineB] {
			// skip if the region was in the last match
			if as < prevA {
				continue
			}

			// we start off with the minimal matching region and then expand the
			// match region.
			r := matchRegion{
				aStart:     as,
				aEnd:       as + 1,
				bStart:     bIdx,
				bEnd:       bIdx + 1,
				matchScore: aEnd - aStart,
			}

			// expand beginning of match region
			for r.validStart(aStart, bStart) && h.eq(r.aStart-1, r.bStart-1) {
				r.aStart--
				r.bStart--
				if r.matchScore > 1 {
					trimmedAStart := strings.TrimSpace(h.a[r.aStart])
					r.matchScore = min(r.matchScore, len(histogram[trimmedAStart]))
				}
			}

			// expand end of match region
			for r.validEnd(aEnd, bEnd) && h.eq(r.aEnd, r.bEnd) {
				if r.matchScore > 1 {
					trimmedAEnd := strings.TrimSpace(h.a[r.aEnd])
					r.matchScore = min(r.matchScore, len(histogram[trimmedAEnd]))
				}
				r.aEnd++
				r.bEnd++
			}

			// see if we have a good match
			if bestMatch.length() < r.length() || r.matchScore < bestMatchScore {
				bestMatch = &r
				bestMatchScore = r.matchScore
			}

			// update cursors to skip regions we've already matched
			if nextB < r.bEnd {
				nextB = r.bEnd
			}
			prevA = r.aEnd
		}
		bIdx = nextB
	}
	return bestMatch
}

// solveRange finds the set of matching regions for the given sections
// of A and B. First the longest matching region is found, then we recurse
// on the area before the match, and then on the area after the match.
func (h *HistogramDiffer) solveRange(aStart, aEnd, bStart, bEnd int) []*matchRegion {
	if bEnd-bStart <= 1 {
		return nil
	}
	if aEnd-aStart <= 1 {
		return nil
	}
	region := h.longestSubstring(aStart, aEnd, bStart, bEnd)
	if region == nil {
		return nil
	}
	regions := append(h.solveRange(aStart, region.aStart, bStart, region.bStart), region)
	regions = append(regions, h.solveRange(region.aEnd, aEnd, region.bEnd, bEnd)...)
	return regions
}

// Solve returns a DiffSolution. Internally it uses solveRange to find
// all matching regions, then it uses the standard differ to create diffs
// on the intra-region area.
func (h *HistogramDiffer) Solve() *DiffSolution {
	s := &DiffSolution{}
	prevRegion := &matchRegion{aStart: 0, aEnd: 0, bStart: 0, bEnd: 0}
	regions := h.solveRange(0, len(h.a), 0, len(h.b))
	for _, region := range regions {
		// compute intra-region differences
		a := h.a[prevRegion.aEnd:region.aStart]
		b := h.b[prevRegion.bEnd:region.bStart]
		s.addSolution(NewSequenceDiffer(a, b).Solve())

		// copy match region
		for _, l := range h.a[region.aStart:region.aEnd] {
			s.addLine(l, l, LineFromBoth)
		}

		// update for loop
		prevRegion = region
	}

	// compute diff for final unmatched section
	a := h.a[prevRegion.aEnd:len(h.a)]
	b := h.b[prevRegion.bEnd:len(h.b)]
	s.addSolution(NewSequenceDiffer(a, b).Solve())
	s.PostProcess()
	return s
}

func min(a, b int) int {
	if b < a {
		return b
	}
	return a
}
