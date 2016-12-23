//line lang.y:2
package query

import __yyfmt__ "fmt"

//line lang.y:2
import (
	"fmt"
	turtle "github.com/gtfierro/hod/goraptor"
	"io"
)

//line lang.y:13
type yySymType struct {
	yys            int
	str            string
	val            turtle.URI
	selectvar      SelectVar
	pred           []PathPattern
	multipred      [][]PathPattern
	triples        []Filter
	orclauses      []OrClause
	varlist        []SelectVar
	links          []Link
	distinct       bool
	count          bool
	partial        bool
	selectAllLinks bool
}

const SELECT = 57346
const COUNT = 57347
const DISTINCT = 57348
const WHERE = 57349
const OR = 57350
const UNION = 57351
const PARTIAL = 57352
const COMMA = 57353
const LBRACE = 57354
const RBRACE = 57355
const LPAREN = 57356
const RPAREN = 57357
const DOT = 57358
const SEMICOLON = 57359
const SLASH = 57360
const PLUS = 57361
const QUESTION = 57362
const ASTERISK = 57363
const BAR = 57364
const LINK = 57365
const VAR = 57366
const URI = 57367
const LBRACK = 57368
const RBRACK = 57369

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"SELECT",
	"COUNT",
	"DISTINCT",
	"WHERE",
	"OR",
	"UNION",
	"PARTIAL",
	"COMMA",
	"LBRACE",
	"RBRACE",
	"LPAREN",
	"RPAREN",
	"DOT",
	"SEMICOLON",
	"SLASH",
	"PLUS",
	"QUESTION",
	"ASTERISK",
	"BAR",
	"LINK",
	"VAR",
	"URI",
	"LBRACK",
	"RBRACK",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line lang.y:271

const eof = 0

type lexer struct {
	scanner   *Scanner
	error     error
	varlist   []SelectVar
	triples   []Filter
	orclauses []OrClause
	distinct  bool
	count     bool
	partial   bool
	pos       int
}

func newlexer(r io.Reader) *lexer {
	scanner := NewScanner(
		[]Definition{
			{Token: LBRACE, Pattern: "\\{"},
			{Token: RBRACE, Pattern: "\\}"},
			{Token: LPAREN, Pattern: "\\("},
			{Token: RPAREN, Pattern: "\\)"},
			{Token: LBRACK, Pattern: "\\["},
			{Token: RBRACK, Pattern: "\\]"},
			{Token: COMMA, Pattern: "\\,"},
			{Token: SEMICOLON, Pattern: ";"},
			{Token: DOT, Pattern: "\\."},
			{Token: BAR, Pattern: "\\|"},
			{Token: SELECT, Pattern: "SELECT"},
			{Token: COUNT, Pattern: "COUNT"},
			{Token: DISTINCT, Pattern: "DISTINCT"},
			{Token: WHERE, Pattern: "WHERE"},
			{Token: OR, Pattern: "OR"},
			{Token: UNION, Pattern: "UNION"},
			{Token: PARTIAL, Pattern: "PARTIAL"},
			{Token: URI, Pattern: "[a-zA-Z]+:[a-zA-Z0-9_\\-#%$@]+"},
			{Token: VAR, Pattern: "\\?[a-zA-Z0-9_]+"},
			{Token: LINK, Pattern: "[a-zA-Z][a-zA-Z0-9_-]*"},
			{Token: QUESTION, Pattern: "\\?"},
			{Token: SLASH, Pattern: "/"},
			{Token: PLUS, Pattern: "\\+"},
			{Token: ASTERISK, Pattern: "\\*"},
		})
	scanner.SetInput(r)
	return &lexer{
		scanner: scanner,
		pos:     0,
	}
}

func (l *lexer) Lex(lval *yySymType) int {
	r := l.scanner.Next()
	if r.Token == Error {
		if len(r.Value) > 0 {
			fmt.Println("ERROR", string(r.Value))
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
	return yyTokname(int(t) - 57342)
}

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyNprod = 34
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 67

var yyAct = [...]int{

	19, 21, 30, 25, 22, 38, 23, 24, 34, 17,
	7, 12, 10, 41, 8, 58, 23, 24, 33, 32,
	27, 29, 26, 37, 35, 10, 28, 55, 10, 45,
	46, 47, 42, 62, 13, 43, 55, 48, 49, 44,
	51, 52, 40, 53, 54, 50, 56, 57, 39, 6,
	5, 59, 60, 61, 11, 3, 4, 14, 15, 16,
	31, 36, 18, 20, 9, 2, 1,
}
var yyPact = [...]int{

	51, -1000, 43, 4, 1, 22, -1000, -12, -12, -12,
	-17, -1000, -12, -8, -1000, -1000, -1000, -1, -1000, 13,
	-8, -6, -8, -1000, -1000, -22, 37, 31, -4, -1000,
	-18, 17, 10, -1000, -6, -6, 32, -1000, -1000, -1,
	-1, -1000, 11, -6, -6, -1000, -1000, -1000, 0, -18,
	-1000, -8, -8, -1000, -1000, -1000, -1000, -1000, -1000, 20,
	-1000, -1000, -1000,
}
var yyPgo = [...]int{

	0, 66, 65, 0, 49, 64, 3, 63, 1, 2,
	61, 60,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 2, 2, 4, 4, 5,
	5, 6, 6, 6, 6, 3, 3, 7, 7, 7,
	10, 10, 10, 9, 9, 9, 11, 11, 11, 11,
	11, 11, 8, 8,
}
var yyR2 = [...]int{

	0, 6, 2, 3, 3, 2, 3, 1, 2, 4,
	1, 1, 1, 3, 3, 1, 2, 4, 5, 3,
	1, 3, 3, 1, 3, 3, 1, 1, 2, 2,
	2, 3, 1, 1,
}
var yyChk = [...]int{

	-1000, -1, -2, 4, 5, 7, -4, 6, 10, -5,
	24, -4, 10, 12, -4, -4, -4, 26, -4, -3,
	-7, -8, 12, 24, 25, -6, 23, 21, 13, -3,
	-9, -11, 25, 24, 14, -8, -10, -3, 27, 11,
	11, 17, -8, 18, 22, 19, 20, 21, -9, -9,
	13, 8, 9, -6, -6, 16, -9, -9, 15, -8,
	-3, -3, 13,
}
var yyDef = [...]int{

	0, -2, 0, 0, 0, 0, 2, 0, 0, 7,
	10, 5, 0, 0, 3, 4, 8, 0, 6, 0,
	15, 0, 0, 32, 33, 0, 11, 12, 0, 16,
	0, 23, 26, 27, 0, 0, 0, 20, 9, 0,
	0, 1, 0, 0, 0, 28, 29, 30, 0, 0,
	19, 0, 0, 13, 14, 17, 24, 25, 31, 0,
	21, 22, 18,
}
var yyTok1 = [...]int{

	1,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27,
}
var yyTok3 = [...]int{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := yyPact[state]
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && yyChk[yyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || yyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := yyExca[i]
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		token = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = yyTok3[i+0]
		if token == char {
			token = yyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-6 : yypt+1]
		//line lang.y:36
		{
			yylex.(*lexer).varlist = yyDollar[1].varlist
			yylex.(*lexer).distinct = yyDollar[1].distinct
			yylex.(*lexer).triples = yyDollar[4].triples
			yylex.(*lexer).distinct = yyDollar[1].distinct
			yylex.(*lexer).partial = yyDollar[1].partial
			yylex.(*lexer).count = yyDollar[1].count
			yylex.(*lexer).orclauses = yyDollar[4].orclauses
		}
	case 2:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:48
		{
			yyVAL.varlist = yyDollar[2].varlist
			yyVAL.distinct = false
			yyVAL.count = false
			yyVAL.partial = false
		}
	case 3:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:55
		{
			yyVAL.varlist = yyDollar[3].varlist
			yyVAL.distinct = true
			yyVAL.count = false
			yyVAL.partial = false
		}
	case 4:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:62
		{
			yyVAL.varlist = yyDollar[3].varlist
			yyVAL.distinct = false
			yyVAL.count = false
			yyVAL.partial = true
		}
	case 5:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:69
		{
			yyVAL.varlist = yyDollar[2].varlist
			yyVAL.distinct = false
			yyVAL.count = true
			yyVAL.partial = false
		}
	case 6:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:76
		{
			yyVAL.varlist = yyDollar[3].varlist
			yyVAL.distinct = false
			yyVAL.count = true
			yyVAL.partial = true
		}
	case 7:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:85
		{
			yyVAL.varlist = []SelectVar{yyDollar[1].selectvar}
		}
	case 8:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:89
		{
			yyVAL.varlist = append([]SelectVar{yyDollar[1].selectvar}, yyDollar[2].varlist...)
		}
	case 9:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line lang.y:95
		{
			if yyDollar[3].selectAllLinks {
				yyVAL.selectvar = SelectVar{Var: turtle.ParseURI(yyDollar[1].str), AllLinks: true}
			} else {
				yyVAL.selectvar = SelectVar{Var: turtle.ParseURI(yyDollar[1].str), AllLinks: false, Links: yyDollar[3].links}
			}
		}
	case 10:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:103
		{
			yyVAL.selectvar = SelectVar{Var: turtle.ParseURI(yyDollar[1].str), AllLinks: false}
		}
	case 11:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:109
		{
			yyVAL.links = []Link{{Name: yyDollar[1].str}}
		}
	case 12:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:113
		{
			yyVAL.selectAllLinks = true
		}
	case 13:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:117
		{
			yyVAL.links = append([]Link{{Name: yyDollar[1].str}}, yyDollar[3].links...)
		}
	case 14:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:121
		{
			yyVAL.selectAllLinks = true
		}
	case 15:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:127
		{
			if len(yyDollar[1].orclauses) > 0 {
				yyVAL.orclauses = yyDollar[1].orclauses
			} else {
				yyVAL.triples = yyDollar[1].triples
			}
		}
	case 16:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:135
		{
			yyVAL.triples = append(yyDollar[2].triples, yyDollar[1].triples...)
			yyVAL.orclauses = append(yyDollar[2].orclauses, yyDollar[1].orclauses...)
		}
	case 17:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line lang.y:142
		{
			triple := Filter{Subject: yyDollar[1].val, Object: yyDollar[3].val}
			if len(yyDollar[2].multipred) > 0 {
				var recurse func(preds [][]PathPattern) OrClause
				recurse = func(preds [][]PathPattern) OrClause {
					triple.Path = preds[0]
					first := OrClause{RightTerms: []Filter{triple}}
					if len(preds) > 1 {
						first.LeftOr = []OrClause{recurse(preds[1:])}
					}
					return first
				}
				var cur = recurse(yyDollar[2].multipred)
				yyVAL.orclauses = []OrClause{cur}
			} else {
				triple.Path = yyDollar[2].pred
				yyVAL.triples = []Filter{triple}
			}
		}
	case 18:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line lang.y:162
		{
			triple := Filter{Subject: yyDollar[2].val, Object: yyDollar[4].val}
			if len(yyDollar[3].multipred) > 0 {
				var recurse func(preds [][]PathPattern) OrClause
				recurse = func(preds [][]PathPattern) OrClause {
					triple.Path = preds[0]
					first := OrClause{RightTerms: []Filter{triple}}
					if len(preds) > 1 {
						first.LeftOr = []OrClause{recurse(preds[1:])}
					}
					return first
				}
				var cur = recurse(yyDollar[3].multipred)
				yyVAL.orclauses = []OrClause{cur}
			} else {
				triple.Path = yyDollar[3].pred
				yyVAL.triples = []Filter{triple}
			}
		}
	case 19:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:182
		{
			if len(yyDollar[2].orclauses) > 0 {
				yyVAL.orclauses = yyDollar[2].orclauses
			} else {
				yyVAL.triples = yyDollar[2].triples
			}
		}
	case 20:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:192
		{
			yyVAL.triples = yyDollar[1].triples
			yyVAL.orclauses = yyDollar[1].orclauses
		}
	case 21:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:197
		{
			yyVAL.orclauses = []OrClause{{LeftOr: yyDollar[3].orclauses,
				LeftTerms:  yyDollar[3].triples,
				RightOr:    yyDollar[1].orclauses,
				RightTerms: yyDollar[1].triples}}
		}
	case 22:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:204
		{
			yyVAL.orclauses = []OrClause{{LeftOr: yyDollar[3].orclauses,
				LeftTerms:  yyDollar[3].triples,
				RightOr:    yyDollar[1].orclauses,
				RightTerms: yyDollar[1].triples}}
		}
	case 23:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:214
		{
			yyVAL.pred = yyDollar[1].pred
		}
	case 24:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:218
		{
			yyVAL.pred = append(yyDollar[1].pred, yyDollar[3].pred...)
		}
	case 25:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:222
		{
			if len(yyDollar[1].multipred) > 0 {
				yyVAL.multipred = append(yyVAL.multipred, yyDollar[1].multipred...)
			} else {
				yyVAL.multipred = append(yyVAL.multipred, yyDollar[1].pred)
			}
			if len(yyDollar[3].multipred) > 0 {
				yyVAL.multipred = append(yyVAL.multipred, yyDollar[3].multipred...)
			} else {
				yyVAL.multipred = append(yyVAL.multipred, yyDollar[3].pred)
			}
		}
	case 26:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:237
		{
			yyVAL.pred = []PathPattern{{Predicate: turtle.ParseURI(yyDollar[1].str), Pattern: PATTERN_SINGLE}}
		}
	case 27:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:241
		{
			yyVAL.pred = []PathPattern{{Predicate: turtle.ParseURI(yyDollar[1].str), Pattern: PATTERN_SINGLE}}
		}
	case 28:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:245
		{
			yyVAL.pred = []PathPattern{{Predicate: turtle.ParseURI(yyDollar[1].str), Pattern: PATTERN_ONE_PLUS}}
		}
	case 29:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:249
		{
			yyVAL.pred = []PathPattern{{Predicate: turtle.ParseURI(yyDollar[1].str), Pattern: PATTERN_ZERO_ONE}}
		}
	case 30:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:253
		{
			yyVAL.pred = []PathPattern{{Predicate: turtle.ParseURI(yyDollar[1].str), Pattern: PATTERN_ZERO_PLUS}}
		}
	case 31:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:257
		{
			yyVAL.pred = yyDollar[2].pred
		}
	case 32:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:263
		{
			yyVAL.val = turtle.ParseURI(yyDollar[1].str)
		}
	case 33:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:267
		{
			yyVAL.val = turtle.ParseURI(yyDollar[1].str)
		}
	}
	goto yystack /* stack new state and value */
}
