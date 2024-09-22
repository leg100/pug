package tui

import (
	"bytes"
)

func SanitizeColors(b []byte) []byte {
	var (
		ansi         bool
		buf          = new(bytes.Buffer)
		lastcolorseq = new(bytes.Buffer)
	)
	for _, c := range b {
		if c == '\x1B' {
			ansi = true
			lastcolorseq.Reset()
			_ = lastcolorseq.WriteByte(c)
		} else if ansi {
			_ = lastcolorseq.WriteByte(c)
			if isTerminator(c) {
				ansi = false
				if bytes.HasSuffix(lastcolorseq.Bytes(), []byte("[0m")) {
					// reset sequence
					lastcolorseq.Reset()
				} else if c != 'm' {
					// not a color code sequence
					lastcolorseq.Reset()
				}
			}
		} else if c == '\n' && lastcolorseq.Len() > 0 {
			// reset color sequence before adding new line
			buf.Write([]byte{'\x1B', '[', '0', 'm'})
			buf.WriteByte(c)
			// re-start color sequence on new line
			buf.Write(lastcolorseq.Bytes())
			continue
		}

		_ = buf.WriteByte(c)
	}
	return buf.Bytes()
}

func isTerminator(c byte) bool {
	return (c >= 0x40 && c <= 0x5a) || (c >= 0x61 && c <= 0x7a)
}
