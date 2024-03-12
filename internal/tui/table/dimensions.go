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
			// Does this actually do anything?
			//m.cols[index].style = col.style.Width(col.width)
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

		// Take borders into account for the actual style
		//m.cols[index].style = m.cols[index].style.Width(width)
	}
}

//func (m *Model[T]) recalculateLastHorizontalColumn() {
//	//if m.horizontalScrollFreezeColumnsCount >= len(m.columns) {
//	//	m.maxHorizontalColumnIndex = 0
//
//	//	return
//	//}
//
//	//if m.totalWidth <= m.maxTotalWidth {
//	//	m.maxHorizontalColumnIndex = 0
//
//	//	return
//	//}
//
//	const (
//		leftOverflowWidth = 2
//		borderAdjustment  = 1
//	)
//
//	// Always have left border
//	visibleWidth := borderAdjustment + leftOverflowWidth
//
//	for i := 0; i < m.horizontalScrollFreezeColumnsCount; i++ {
//		visibleWidth += m.columns[i].width + borderAdjustment
//	}
//
//	m.maxHorizontalColumnIndex = len(m.columns) - 1
//
//	// Work backwards from the right
//	for i := len(m.columns) - 1; i >= m.horizontalScrollFreezeColumnsCount && visibleWidth <= m.maxTotalWidth; i-- {
//		visibleWidth += m.columns[i].width + borderAdjustment
//
//		if visibleWidth <= m.maxTotalWidth {
//			m.maxHorizontalColumnIndex = i - m.horizontalScrollFreezeColumnsCount
//		}
//	}
//}
