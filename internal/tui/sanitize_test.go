package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeColors(t *testing.T) {
	tt := []struct {
		name  string
		input []byte
		want  []byte
	}{
		// ANSI control codes are reset at the end of the line and re-activated
		// on the next line
		{
			input: []byte("\x1B[31mmary \nhad a little \nlamb\x1B[0m"),
			want:  []byte("\x1B[31mmary \x1B[0m\n\x1B[31mhad a little \x1B[0m\n\x1B[31mlamb\x1B[0m"),
		},
		// Ensure only the most recent ANSI control code is reset and
		// re-activated. Here the color is set to green before being set to red,
		// and the remainder of the text should be red.
		{
			input: []byte("\x1B[32m\x1B[31mmary \nhad a little \nlamb\x1B[0m"),
			want:  []byte("\x1B[32m\x1B[31mmary \x1B[0m\n\x1B[31mhad a little \x1B[0m\n\x1B[31mlamb\x1B[0m"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := SanitizeColors([]byte(tc.input))
			assert.Equal(t, string(tc.want), string(got))
		})
	}
}
