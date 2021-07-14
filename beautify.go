package luautil

import (
	"io"

	"github.com/notnoobmaster/luautil/parse"
)

func Beautify(input io.Reader) (string, error) {
	chunk, err := parse.Parse(input, "")
	if err != nil {
		return "", err
	}
	return chunk.String(), nil
}