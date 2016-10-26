package turtle

import (
	"io"
)

func Parse(r io.Reader) {
	l := newlexer(r)
	ttlParse(l)
}
