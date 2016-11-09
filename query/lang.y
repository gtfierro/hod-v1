%{
package lang

import (
    "io"
    turtle "github.com/gtfierro/hod/goraptor"
    "fmt"
)

%}


%union{
    str string
    val turtle.URI
    triple turtle.Triple
    triples []turtle.Triple
    varlist []turtle.URI
    distinct bool
}

%token SELECT DISTINCT WHERE
%token COMMA LBRACE RBRACE DOT
%token VAR URI

%%

query        : selectClause WHERE LBRACE whereTriples RBRACE
             {
               yylex.(*lexer).varlist = $1.varlist
               yylex.(*lexer).distinct = $1.distinct
               yylex.(*lexer).triples = $4.triples
             }
             ;

selectClause : SELECT varList
             {
                $$.varlist = $2.varlist
                $$.distinct = false
             }
             | SELECT DISTINCT varList
             {
                $$.varlist = $3.varlist
                $$.distinct = true
             }
             ;

varList      : VAR
             {
                $$.varlist = []turtle.URI{turtle.ParseURI($1.str)}
             }
             | VAR varList
             {
                $$.varlist = append([]turtle.URI{turtle.ParseURI($1.str)}, $2.varlist...)
             }
             ;

whereTriples : triple
             {
                $$.triples = []turtle.Triple{$1.triple}
             }
             | triple whereTriples
             {
                $$.triples = append($2.triples, $1.triple)
             }
             ;

triple       : term term term DOT
             {
                $$.triple = turtle.Triple{$1.val, $2.val, $3.val}
             }
             | LBRACE term term term RBRACE DOT
             {
                $$.triple = turtle.Triple{$1.val, $2.val, $3.val}
             }
             ;

term         : VAR
             {
                $$.val = turtle.ParseURI($1.str)
             }
             | URI
             {
                $$.val = turtle.ParseURI($1.str)
             }
             ;
%%

const eof = 0

type lexer struct {
    scanner *Scanner
    error   error
    varlist []turtle.URI
    triples []turtle.Triple
    distinct bool
}

func newlexer(r io.Reader) *lexer {
	scanner := NewScanner(
		[]Definition{
            {Token: LBRACE,  Pattern: "\\{"},
            {Token: RBRACE,  Pattern: "\\}"},
            {Token: COMMA,  Pattern: "\\,"},
            {Token: DOT,  Pattern: "\\."},
            {Token: SELECT,  Pattern: "SELECT"},
            {Token: DISTINCT,  Pattern: "DISTINCT"},
            {Token: WHERE,  Pattern: "WHERE"},
            {Token: URI,  Pattern: "[a-zA-Z]+:[a-zA-Z0-9_\\-+%$#@]+"},
            {Token: VAR,  Pattern: "?[a-zA-Z0-9_]+"},
        })
	scanner.SetInput(r)
    return &lexer{
        scanner: scanner,
    }
}

func (l *lexer) Lex(lval *yySymType) int {
    r := l.scanner.Next()
    if r.Token == Error {
        if len(r.Value) > 0 {
            fmt.Println("ERROR",string(r.Value))
        }
        return eof
    }
    lval.str = string(r.Value)
    return int(r.Token)
}

func (l *lexer) Error(s string) {
    l.error = fmt.Errorf(s)
}

func TokenName(t Token) string {
    return yyTokname(int(t)-57342)
}
