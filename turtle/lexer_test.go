package turtle

import (
	"regexp"
	"strings"
	"testing"
)

func TestRegexp(t *testing.T) {
	for _, test := range []struct {
		input   string
		regex   string
		matches bool
	}{
		{
			"0xc0",
			"[0xc0-0xd6]",
			true,
		},
	} {
		r := regexp.MustCompile(test.regex)
		if r.MatchString(test.input) != test.matches {
			if test.matches {
				t.Errorf("Regexp %s should match %s but does not", test.regex, test.input)
			} else {
				t.Errorf("Regexp %s should NOT match %s but does", test.regex, test.input)
			}
		}
	}
}

func TestLexer(t *testing.T) {
	for _, test := range []struct {
		input  string
		tokens []Token
	}{
		{
			"@prefix somePrefix: <http://www.perceive.net/schemas/relationship/> .",
			[]Token{ATPREFIX, PNAME_NS, IRIREF, DOT},
		},
		{
			"@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .",
			[]Token{ATPREFIX, PNAME_NS, IRIREF, DOT},
		},
		{
			"<#green-goblin>",
			[]Token{IRIREF},
		},
		{
			"foaf:Person",
			[]Token{PNAME_LN},
		},
		{
			"\"spiderman\"",
			[]Token{STRING_LITERAL_QUOTE},
		},
		{
			"'hello' 'world'",
			[]Token{STRING_LITERAL_SINGLE_QUOTE, STRING_LITERAL_SINGLE_QUOTE},
		},
		{
			"\"RDF/XML Syntax Specification (Revised)\"",
			[]Token{STRING_LITERAL_QUOTE},
		},
		{
			"dc:title \"RDF/XML Syntax Specification (Revised)\" ;\n ex:editor [ ex:fullname \"Dave Beckett\"; ] .",
			[]Token{PNAME_LN, STRING_LITERAL_QUOTE, SEMICOLON, PNAME_LN, LBRACK, PNAME_LN, STRING_LITERAL_QUOTE, SEMICOLON, RBRACK, DOT},
		},
	} {
		input := strings.NewReader(test.input)
		l := newlexer(input)
		results := l.scanner.Tokenize()
		if len(results) != len(test.tokens) {
			wantlist := []string{}
			gotlist := []string{}
			for _, tok := range test.tokens {
				wantlist = append(wantlist, TokenName(tok))
			}
			for _, r := range results {
				gotlist = append(gotlist, TokenName(r.Token))
			}
			t.Errorf("Got tokens \n%+v but wanted \n%+v", gotlist, wantlist)
			continue
		}
		for i, tok := range test.tokens {
			if tok != results[i].Token {
				t.Errorf("Mismatched token. Got %s, wanted %s", TokenName(results[i].Token), TokenName(tok))
			}
		}

	}
}
