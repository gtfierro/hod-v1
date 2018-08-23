//go:generate gocc -p github.com/gtfierro/hod/lang -a sparql.bnf
package sparql

import (
	"github.com/gtfierro/hod/lang/ast"
	"github.com/gtfierro/hod/lang/lexer"
	"github.com/gtfierro/hod/lang/parser"
)

var p *parser.Parser

func init() {
	p = parser.NewParser()
}

func Parse(s string) (*ast.Query, error) {
	lexed := lexer.NewLexer([]byte(s))
	_q, err := p.Parse(lexed)
	if err != nil {
		return nil, err
	}
	q := _q.(ast.Query)
	return &q, nil
}
