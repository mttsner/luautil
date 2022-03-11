package luautil

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

// Quote returns a double-quoted lua string literal representing s. The
// returned string uses lua escape sequences (\t, \n, \123, \u0100) for
// control characters and non-printable characters as defined by
// IsPrint.
func Quote(s string) string {
	b := &strings.Builder{}
	b.Grow(3*len(s)/2)
	quoteWith(b, s, '"')
	return b.String()
}

func convert(char int) string {
	if char < 10 {
		return "00" + strconv.Itoa(char)
	} else if char < 100 {
		return "0" + strconv.Itoa(char)
	}
	return strconv.Itoa(char)
}

func quoteWith(b *strings.Builder, s string, quote byte) {
	b.WriteByte(quote)
	for width := 0; len(s) > 0; s = s[width:] {
		r := rune(s[0])
		width = 1
		if r >= utf8.RuneSelf {
			r, width = utf8.DecodeRuneInString(s)
		}
		if width == 1 && r == utf8.RuneError {
			b.WriteRune('\\')
			b.WriteString(convert(int(s[0])))
			continue
		}
		appendEscapedRune(b, r, quote)
	}
	b.WriteByte(quote)
}

func appendEscapedRune(b *strings.Builder, r rune, quote byte) {
	var runeTmp [utf8.UTFMax]byte
	if r == rune(quote) || r == '\\' { // always backslashed
		b.WriteRune('\\')
		b.WriteRune(r)
		return
	}
	if strconv.IsPrint(r) {
		n := utf8.EncodeRune(runeTmp[:], r)
		b.Write(runeTmp[:n])
		return
	}
	switch r {
	case '\a':
		b.WriteString(`\a`)
	case '\b':
		b.WriteString(`\b`)
	case '\f':
		b.WriteString(`\f`)
	case '\n':
		b.WriteString(`\n`)
	case '\r':
		b.WriteString(`\r`)
	case '\t':
		b.WriteString(`\t`)
	case '\v':
		b.WriteString(`\v`)
	default:
		b.WriteRune('\\')
		b.WriteString(convert(int(r)))
	}
}