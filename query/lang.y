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
    pred []PathPattern
    triple Filter
    triples []Filter
    varlist []turtle.URI
    distinct bool
    count bool
}

%token SELECT COUNT DISTINCT WHERE
%token COMMA LBRACE RBRACE DOT SEMICOLON SLASH PLUS QUESTION ASTERISK
%token VAR URI

%%

query        : selectClause WHERE LBRACE whereTriples RBRACE SEMICOLON
             {
               yylex.(*lexer).varlist = $1.varlist
               yylex.(*lexer).distinct = $1.distinct
               yylex.(*lexer).triples = $4.triples
               yylex.(*lexer).distinct = $1.distinct
               yylex.(*lexer).count = $1.count
             }
             ;

selectClause : SELECT varList
             {
                $$.varlist = $2.varlist
                $$.distinct = false
                $$.count = false
             }
             | SELECT DISTINCT varList
             {
                $$.varlist = $3.varlist
                $$.distinct = true
                $$.count = false
             }
             | COUNT varList
             {
                $$.varlist = $2.varlist
                $$.distinct = false
                $$.count = true
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
                $$.triples = []Filter{$1.triple}
             }
             | triple whereTriples
             {
                $$.triples = append($2.triples, $1.triple)
             }
             ;

triple       : term path term DOT
             {
                $$.triple = Filter{$1.val, $2.pred, $3.val}
             }
             | LBRACE term path term RBRACE DOT
             {
                $$.triple = Filter{$2.val, $3.pred, $4.val}
             }
             ;

path         : pathpart
             {
                $$.pred = $1.pred
             }
             | pathpart SLASH path
             {
                $$.pred = append($1.pred, $3.pred...)
             }
             ;

pathpart     : URI
             {
                $$.pred = []PathPattern{{Predicate: turtle.ParseURI($1.str), Pattern: PATTERN_SINGLE}}
             }
             | VAR
             {
                $$.pred = []PathPattern{{Predicate: turtle.ParseURI($1.str), Pattern: PATTERN_SINGLE}}
             }
             | URI PLUS
             {
                $$.pred = []PathPattern{{Predicate: turtle.ParseURI($1.str), Pattern: PATTERN_ONE_PLUS}}
             }
             | URI QUESTION
             {
                $$.pred = []PathPattern{{Predicate: turtle.ParseURI($1.str), Pattern: PATTERN_ZERO_ONE}}
             }
             | URI ASTERISK
             {
                $$.pred = []PathPattern{{Predicate: turtle.ParseURI($1.str), Pattern: PATTERN_ZERO_PLUS}}
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
    triples []Filter
    distinct bool
    count bool
    pos int
}

func newlexer(r io.Reader) *lexer {
	scanner := NewScanner(
		[]Definition{
            {Token: LBRACE,  Pattern: "\\{"},
            {Token: RBRACE,  Pattern: "\\}"},
            {Token: COMMA,  Pattern: "\\,"},
            {Token: SEMICOLON,  Pattern: ";"},
            {Token: DOT,  Pattern: "\\."},
            {Token: SELECT,  Pattern: "SELECT"},
            {Token: COUNT,  Pattern: "COUNT"},
            {Token: DISTINCT,  Pattern: "DISTINCT"},
            {Token: WHERE,  Pattern: "WHERE"},
            {Token: URI,  Pattern: "[a-zA-Z]+:[a-zA-Z0-9_\\-#%$@]+"},
            {Token: VAR,  Pattern: "\\?[a-zA-Z0-9_]+"},
            {Token: QUESTION,  Pattern: "\\?"},
            {Token: SLASH,  Pattern: "/"},
            {Token: PLUS,  Pattern: "\\+"},
            {Token: ASTERISK,  Pattern: "\\*"},
        })
	scanner.SetInput(r)
    return &lexer{
        scanner: scanner,
        pos: 0,
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
    l.pos += len(r.Value)
    return int(r.Token)
}

func (l *lexer) Error(s string) {
    l.error = fmt.Errorf("Error parsing: %s. Current line %d:%d. Recent token '%s'", s, l.scanner.lineNumber, l.pos, l.scanner.tokenizer.Text())
}

func TokenName(t Token) string {
    return yyTokname(int(t)-57342)
}
