%{
package turtle

import (
    "io"
    "fmt"
)

%}

%union{
    str string
}

%token IRIREF PNAME_NS PNAME_LN BLANK_NODE_LABEL LANGTAG
%token INTEGER DECIMAL DOUBLE EXPONENT
%token STRING_LITERAL_QUOTE STRING_LITERAL_SINGLE_QUOTE STRING_LITERAL_LONG_SINGLE_QUOTE STRING_LITERAL_LONG_QUOTE
%token WS ANON PN_CHARS_BASE PN_CHARS_U PN_CHARS PN_PREFIX PN_LOCAL
%token PLX PERCENT HEX PN_LOCAL_ESC
%token LANGLE RANGLE LBRACK RBRACK LPAREN RPAREN DOUBLECARAT SEMICOLON TRUE FALSE DOT ATPREFIX ATBASE BASE PREFIX COMMA A

%%

turtleDoc   :   statementList
            ;

statementList : statement
              | statement statementList
              ;

statement   :   directive
            |   triples DOT
            ;

directive   : prefixID
            | base
            | sparqlPrefix
            | sparqlBase
            ;

prefixID    : ATPREFIX PNAME_NS IRIREF DOT
            {
                //fmt.Println("> PREFIX ID:",$1,$2,$3,$4)
            }
            ;

base        : ATBASE IRIREF DOT
            {
                //fmt.Println("> BASE:",$1,$2,$3)
            }
            ;

sparqlBase  : BASE IRIREF
            ;

sparqlPrefix : PREFIX PNAME_NS IRIREF
             ;

triples      : subject predicateObjectList
                    {
                        fmt.Println("> 1")
                    }
             | blankNodePropertyList
                    {
                        fmt.Println("> 2")
                    }
             | blankNodePropertyList predicateObjectList
                    {
                        fmt.Println("> 3")
                    }
             ;

predicateObjectList : verb objectList
                    {
                        fmt.Println("> 1")
                    }
                    | verb objectList SEMICOLON
                    {
                        fmt.Println("> 2")
                    }
                    | verb objectList SEMICOLON predicateObjectList
                    {
                        fmt.Println("> 3")
                    }
                    ;

verb        : predicate
            | A
            ;

subject     : iri
            | BlankNode
            | collection
            ;

predicate   : iri
            ;

object      : iri
            | BlankNode
            | collection
            | blankNodePropertyList
            | literal
            ;

literal     : RDFLiteral
            | NumericLiteral
            | BooleanLiteral
            ;

objectList  : object
            | object COMMA objectList
            ;


blankNodePropertyList : LBRACK predicateObjectList RBRACK
                      ;

collection  : LPAREN objectSequence RPAREN
            ;

objectSequence : object
               | object objectSequence
               ;

NumericLiteral : INTEGER
               | DECIMAL
               | DOUBLE
               ;

RDFLiteral      : String LANGTAG
                | String DOUBLECARAT iri
                ;

BooleanLiteral  : TRUE
                | FALSE
                ;

String          : STRING_LITERAL_QUOTE
                | STRING_LITERAL_SINGLE_QUOTE
                | STRING_LITERAL_LONG_SINGLE_QUOTE
                | STRING_LITERAL_LONG_QUOTE
                ;

iri             : IRIREF
                | PrefixedName
                ;

PrefixedName    : PNAME_LN
                | PNAME_NS
                ;

BlankNode       : BLANK_NODE_LABEL
                | ANON
                ;
%%

const eof = 0
type lexer struct {
    scanner         *Scanner
    tokens          []Token
    values          [][]byte
    error           error
    iri             string
    namespaces      map[string]string
    bnodeLabels     map[string]string
    curSubject      string
    curPredicate    string
}

func newlexer(r io.Reader) *lexer {
	scanner := NewScanner(
		[]Definition{
            {Token: LBRACK,                         Pattern: "\\["},
            {Token: RBRACK,                         Pattern: "\\]"},
            {Token: LPAREN,                         Pattern: "\\("},
            {Token: RPAREN,                         Pattern: "\\)"},
            {Token: DOUBLECARAT,                    Pattern: "\\^\\^"},
            {Token: SEMICOLON,                      Pattern: ";"},
            {Token: TRUE,                           Pattern: "true"},
            {Token: FALSE,                          Pattern: "false"},
            {Token: DOT,                            Pattern: "\\."},
            {Token: ATPREFIX,                       Pattern: "@prefix"},
            {Token: ATBASE,                         Pattern: "@base"},
            {Token: BASE,                           Pattern: "base"},
            {Token: PREFIX,                         Pattern: "prefix"},
            {Token: COMMA,                          Pattern: "comma"},
            {Token: A,                              Pattern: "a"},
            {Token: IRIREF,                         Pattern: "<(([^<>\"{}|^`])|(\\\\u([0-9]|[A-F]|[a-f]){4})|(\\\\U([0-9]|[A-F]|[a-f]){8}))*>"},
            {Token: PNAME_LN,                       Pattern: "([A-Z]|[a-z]|[0-9]{1,6})((([A-Z]|[a-z]|[0-9]{1,6}|_|-|\\.)*([A-Z]|[a-z]|[0-9]{1,6}|_|-|\\.))?)?:([A-Z]|[a-z]|[0-9]{1,6}|_|:|(\\\\(_|~|\\.|-|!|$|&|'|(|)|\\*|\\+|,|;|=|/|\\?|#|@|%)))(([A-Z]|[a-z]|[0-9]{1,6}|_|-|\\.|:|(%([0-9]|[A-F]|[a-f])))*([A-Z]|[a-z]|[0-9]{1,6}|_|-|:|(%([0-9]|[A-F]|[a-f])))?)"},
            {Token: PNAME_NS,                       Pattern: "([A-Z]|[a-z]|[0-9]{1,6})((([A-Z]|[a-z]|[0-9]{1,6}|_|-|\\.)*([A-Z]|[a-z]|[0-9]{1,6}|_|-|\\.))?)?:"},
            {Token: BLANK_NODE_LABEL,               Pattern: "_:([A-Z]|[a-z]|[0-9]{1,6}_)(([A-Z]|[a-z]|[0-9]{1,6}|_|-|\\.)*[A-Z]|[a-z]|[0-9]{1,6}|_|-)?"},
            {Token: LANGTAG,                        Pattern: "@[a-zA-Z]+(-[a-zA-Z0-9]+)*"},
            {Token: INTEGER,                        Pattern: "[+-]?[0-9]+"},
            {Token: DECIMAL,                        Pattern: "[+-]?[0-9]*\\.[0-9]+"},
            {Token: DOUBLE,                         Pattern: "[+-]?(([0-9]+\\.[0-9]*([eE][+-]?[0-9]+))|(\\.[0-9]+([eE][+-]?[0-9]+))|([0-9]+([eE][+-]?[0-9]+)))"},
            {Token: EXPONENT,                       Pattern: "([eE][+-]?[0-9]+)"},
            {Token: STRING_LITERAL_QUOTE,           Pattern: "\"[^\"]*\""},
            {Token: STRING_LITERAL_SINGLE_QUOTE,    Pattern: "'[^']*'"},
            {Token: PN_CHARS_BASE,                  Pattern: "[A-Z]|[a-z]|[0-9]{1,6}"},
            {Token: PN_CHARS_U,                     Pattern: "[A-Z]|[a-z]|[0-9]{1,6}|_"},
            {Token: PN_CHARS,                       Pattern: "([A-Z]|[a-z]|[0-9]{1,6}|_|-)"},
            {Token: PN_LOCAL,                       Pattern: "([A-Z]|[a-z]|[0-9]{1,6}|_|:|(\\\\(_|~|\\.|-|!|$|&|'|(|)|\\*|\\+|,|;|=|/|\\?|#|@|%)))(([A-Z]|[a-z]|[0-9]{1,6}|_|-|\\.|:|(%([0-9]|[A-F]|[a-f])))*([A-Z]|[a-z]|[0-9]{1,6}|_|-|:|(%([0-9]|[A-F]|[a-f])))?)"},
            {Token: PLX,                            Pattern: "%([0-9]|[A-F]|[a-f])"},
            {Token: PN_LOCAL_ESC,                   Pattern: "\\\\(_|~|\\.|-|!|$|&|'|(|)|\\*|\\+|,|;|=|/|\\?|#|@|%)"},
            {Token: LANGLE,                         Pattern: "<"},
            {Token: RANGLE,                         Pattern: ">"},
		})
	scanner.SetInput(r)
    return &lexer{
        scanner: scanner,
    }
}

func (l *lexer) Lex(lval *ttlSymType) int {
    r := l.scanner.Next()
    if r.Token == Error {
        if len(r.Value) > 0 {
            fmt.Println("ERROR",string(r.Value))
        }
        return eof
    }
    fmt.Printf("%s: %s\n",TokenName(r.Token), string(r.Value))
    l.tokens = append(l.tokens, r.Token)
    l.values = append(l.values, r.Value)
    return int(r.Token)
}

func (l *lexer) Error(s string) {
    l.error = fmt.Errorf(s)
}

func TokenName(t Token) string {
    return ttlTokname(int(t)-57342)
}
