package sparql

import (
	"log"

	"github.com/gtfierro/hod/lang/lexer"
	"github.com/gtfierro/hod/lang/parser"
)

var p *parser.Parser

func init() {
	p = parser.NewParser()
}

func Parse(s string) {
	lexed := lexer.NewLexer([]byte(s))
	x, err := p.Parse(lexed)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("%+v", x)
	}
}
