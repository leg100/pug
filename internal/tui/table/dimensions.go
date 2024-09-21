// Lifted much code from:
//
// https://github.com/Evertras/bubble-table/blob/main/table/dimensions.go
//
// Copyright (c) 2022 Brandon Fulljames

package table

import "github.com/leg100/pug/internal/tui"

// Update column widths in-place.
//
// TODO: unit test
func (m *Model[V]) setColumnWidths() {
	var (
		// total available flex width initialized to total table width minus the
		// padding on each col (2) and the scrollbar to the right
		totalFlexWidth  = m.width - tui.ScrollbarWidth - 2*len(m.cols)
		totalFlexFactor int
		flexGCD         int
	)

	for _, col := range m.cols {
		if col.FlexFactor == 0 {
			// Column not using flex so subtract its width from avail width
			totalFlexWidth -= col.Width
		} else {
			totalFlexFactor += col.FlexFactor
			flexGCD = gcd(flexGCD, col.FlexFactor)
		}
	}

	if totalFlexFactor == 0 {
		return
	}

	// We use the GCD here because otherwise very large values won't divide
	// nicely as ints
	totalFlexFactor /= flexGCD

	flexUnit := totalFlexWidth / totalFlexFactor
	leftoverWidth := totalFlexWidth % totalFlexFactor

	for index := range m.cols {
		if m.cols[index].FlexFactor == 0 {
			continue
		}

		width := flexUnit * (m.cols[index].FlexFactor / flexGCD)

		if leftoverWidth > 0 {
			width++
			leftoverWidth--
		}

		if index == len(m.cols)-1 {
			width += leftoverWidth
			leftoverWidth = 0
		}

		width = max(width, 1)

		m.cols[index].Width = width
	}
}
