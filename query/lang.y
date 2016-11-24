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
    triples []Filter
    orclauses []OrClause
    varlist []turtle.URI
    distinct bool
    count bool
}

%token SELECT COUNT DISTINCT WHERE OR
%token COMMA LBRACE RBRACE LPAREN RPAREN DOT SEMICOLON SLASH PLUS QUESTION ASTERISK
%token VAR URI

%%

query        : selectClause WHERE LBRACE whereTriples RBRACE SEMICOLON
             {
               yylex.(*lexer).varlist = $1.varlist
               yylex.(*lexer).distinct = $1.distinct
               yylex.(*lexer).triples = $4.triples
               yylex.(*lexer).distinct = $1.distinct
               yylex.(*lexer).count = $1.count
               yylex.(*lexer).orclauses = $4.orclauses
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
                if len($1.orclauses) > 0 {
                  $$.orclauses = $1.orclauses
                } else {
                  $$.triples = $1.triples
                }
             }
             | triple whereTriples
             {
                $$.triples = append($2.triples, $1.triples...)
                $$.orclauses = append($2.orclauses, $1.orclauses...)
             }
             ;

triple       : term path term DOT
             {
                $$.triples = []Filter{{$1.val, $2.pred, $3.val}}
             }
             | LBRACE compound RBRACE
             {
                if len($2.orclauses) > 0 {
                  $$.orclauses = $2.orclauses
                } else {
                  $$.triples = $2.triples
                }
             }
             ;

compound     : whereTriples
             {
                $$.triples = $1.triples
                $$.orclauses = $1.orclauses
             }
             | compound OR whereTriples
             {
                $$.orclauses = []OrClause{{LeftOr: $3.orclauses, 
                                            LeftTerms: $3.triples, 
                                            RightOr: $1.orclauses,
                                            RightTerms: $1.triples}}
             }
             ;


path         : pathatom
             {
                $$.pred = $1.pred
             }
             | pathatom SLASH path
             {
                $$.pred = append($1.pred, $3.pred...)
             }
             ;

pathatom     : URI
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
    orclauses []OrClause
    distinct bool
    count bool
    pos int
}

func newlexer(r io.Reader) *lexer {
	scanner := NewScanner(
		[]Definition{
            {Token: LBRACE,  Pattern: "\\{"},
            {Token: RBRACE,  Pattern: "\\}"},
            {Token: LPAREN,  Pattern: "\\("},
            {Token: RPAREN,  Pattern: "\\)"},
            {Token: COMMA,  Pattern: "\\,"},
            {Token: SEMICOLON,  Pattern: ";"},
            {Token: DOT,  Pattern: "\\."},
            {Token: SELECT,  Pattern: "SELECT"},
            {Token: COUNT,  Pattern: "COUNT"},
            {Token: DISTINCT,  Pattern: "DISTINCT"},
            {Token: WHERE,  Pattern: "WHERE"},
            {Token: OR,  Pattern: "OR"},
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
