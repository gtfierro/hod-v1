%{
package query

import (
    "io"
    turtle "github.com/gtfierro/hod/goraptor"
    "fmt"
)

%}


%union{
    str string
    val turtle.URI
    selectvar SelectVar
    pred []PathPattern
    multipred [][]PathPattern
    triples []Filter
    orclauses []OrClause
    varlist []SelectVar
    links []Link
    distinct bool
    count bool
    partial bool
    selectAllLinks bool
}

%token SELECT COUNT DISTINCT WHERE OR UNION PARTIAL
%token COMMA LBRACE RBRACE LPAREN RPAREN DOT SEMICOLON SLASH PLUS QUESTION ASTERISK BAR
%token LINK VAR URI LBRACK RBRACK

%%

query        : selectClause WHERE LBRACE whereTriples RBRACE SEMICOLON
             {
               yylex.(*lexer).varlist = $1.varlist
               yylex.(*lexer).distinct = $1.distinct
               yylex.(*lexer).triples = $4.triples
               yylex.(*lexer).distinct = $1.distinct
               yylex.(*lexer).partial = $1.partial
               yylex.(*lexer).count = $1.count
               yylex.(*lexer).orclauses = $4.orclauses
             }
             ;

selectClause : SELECT varList
             {
                $$.varlist = $2.varlist
                $$.distinct = false
                $$.count = false
                $$.partial = false
             }
             | SELECT DISTINCT varList
             {
                $$.varlist = $3.varlist
                $$.distinct = true
                $$.count = false
                $$.partial = false
             }
             | SELECT PARTIAL varList
             {
                $$.varlist = $3.varlist
                $$.distinct = false
                $$.count = false
                $$.partial = true
             }
             | COUNT varList
             {
                $$.varlist = $2.varlist
                $$.distinct = false
                $$.count = true
                $$.partial = false
             }
             | COUNT PARTIAL varList
             {
                $$.varlist = $3.varlist
                $$.distinct = false
                $$.count = true
                $$.partial = true
             }
             ;

varList      : var
             {
                $$.varlist = []SelectVar{$1.selectvar}
             }
             | var varList
             {
                $$.varlist = append([]SelectVar{$1.selectvar}, $2.varlist...)
             }
             ;

var          : VAR LBRACK linkList RBRACK
             {
                if $3.selectAllLinks {
                $$.selectvar = SelectVar{Var: turtle.ParseURI($1.str), AllLinks: true}
                } else {
                $$.selectvar = SelectVar{Var: turtle.ParseURI($1.str), AllLinks: false, Links: $3.links}
                }
             }
             | VAR
             {
                $$.selectvar = SelectVar{Var: turtle.ParseURI($1.str), AllLinks: false}
             }
             ;

linkList     : LINK
             {
                $$.links = []Link{{Name: $1.str}}
             }
             | ASTERISK
             {
                $$.selectAllLinks = true
             }
             | LINK COMMA linkList
             {
                $$.links = append([]Link{{Name: $1.str}}, $3.links...)
             }
             | ASTERISK COMMA linkList
             {
                $$.selectAllLinks = true
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
                triple := Filter{Subject: $1.val, Object: $3.val}
                if len($2.multipred) > 0 {
                    var recurse func(preds [][]PathPattern) OrClause
                    recurse = func(preds [][]PathPattern) OrClause {
                        triple.Path = preds[0]
                        first := OrClause{RightTerms: []Filter{triple}}
                        if len(preds) > 1 {
                          first.LeftOr = []OrClause{recurse(preds[1:])}
                        }
                        return first
                    }
                    var cur = recurse($2.multipred)
                    $$.orclauses = []OrClause{cur}
                } else {
                    triple.Path = $2.pred
                    $$.triples = []Filter{triple}
                }
             }
             | LBRACE term path term RBRACE
             {
                triple := Filter{Subject: $2.val, Object: $4.val}
                if len($3.multipred) > 0 {
                    var recurse func(preds [][]PathPattern) OrClause
                    recurse = func(preds [][]PathPattern) OrClause {
                        triple.Path = preds[0]
                        first := OrClause{RightTerms: []Filter{triple}}
                        if len(preds) > 1 {
                          first.LeftOr = []OrClause{recurse(preds[1:])}
                        }
                        return first
                    }
                    var cur = recurse($3.multipred)
                    $$.orclauses = []OrClause{cur}
                } else {
                    triple.Path = $3.pred
                    $$.triples = []Filter{triple}
                }
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
             | compound UNION whereTriples
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
             | pathatom BAR path
             {
                if len($1.multipred) > 0 {
                  $$.multipred = append($$.multipred, $1.multipred...)
                } else {
                  $$.multipred = append($$.multipred, $1.pred)
                }
                if len($3.multipred) > 0 {
                  $$.multipred = append($$.multipred, $3.multipred...)
                } else {
                  $$.multipred = append($$.multipred, $3.pred)
                }
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
             | LPAREN path RPAREN
             {
                $$.pred = $2.pred
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
    varlist []SelectVar
    triples []Filter
    orclauses []OrClause
    distinct bool
    count bool
    partial bool
    pos int
}

func newlexer(r io.Reader) *lexer {
	scanner := NewScanner(
		[]Definition{
            {Token: LBRACE,  Pattern: "\\{"},
            {Token: RBRACE,  Pattern: "\\}"},
            {Token: LPAREN,  Pattern: "\\("},
            {Token: RPAREN,  Pattern: "\\)"},
            {Token: LBRACK,  Pattern: "\\["},
            {Token: RBRACK,  Pattern: "\\]"},
            {Token: COMMA,  Pattern: "\\,"},
            {Token: SEMICOLON,  Pattern: ";"},
            {Token: DOT,  Pattern: "\\."},
            {Token: BAR,  Pattern: "\\|"},
            {Token: SELECT,  Pattern: "SELECT"},
            {Token: COUNT,  Pattern: "COUNT"},
            {Token: DISTINCT,  Pattern: "DISTINCT"},
            {Token: WHERE,  Pattern: "WHERE"},
            {Token: OR,  Pattern: "OR"},
            {Token: UNION,  Pattern: "UNION"},
            {Token: PARTIAL,  Pattern: "PARTIAL"},
            {Token: URI,  Pattern: "[a-zA-Z]+:[a-zA-Z0-9_\\-#%$@]+"},
            {Token: VAR,  Pattern: "\\?[a-zA-Z0-9_]+"},
            {Token: LINK,  Pattern: "[a-zA-Z][a-zA-Z0-9_-]*"},
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
