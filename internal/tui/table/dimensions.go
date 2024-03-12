package table

// Updates column width in-place.  This could be optimized but should be called
// very rarely so we prioritize simplicity over performance here.
func (m *Model[T]) recalculateWidth() {
	var (
		// Where do these numbers come from? Padding?
		totalFlexWidth  = m.width - len(m.cols) - 1
		totalFlexFactor int
		flexGCD         int
	)

	for _, col := range m.cols {
		if col.FlexFactor == 0 {
			// Column not using flex
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
