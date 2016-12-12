//line lang.y:2
package lang

import __yyfmt__ "fmt"

//line lang.y:2
import (
	"fmt"
	turtle "github.com/gtfierro/hod/goraptor"
	"io"
)

//line lang.y:13
type yySymType struct {
	yys       int
	str       string
	val       turtle.URI
	pred      []PathPattern
	multipred [][]PathPattern
	triples   []Filter
	orclauses []OrClause
	varlist   []turtle.URI
	distinct  bool
	count     bool
	partial   bool
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
const VAR = 57365
const URI = 57366

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
	"VAR",
	"URI",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line lang.y:236

const eof = 0

type lexer struct {
	scanner   *Scanner
	error     error
	varlist   []turtle.URI
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

const yyNprod = 28
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 57

var yyAct = [...]int{

	17, 19, 25, 20, 21, 22, 7, 29, 35, 9,
	8, 11, 36, 33, 21, 22, 28, 27, 45, 24,
	48, 32, 30, 9, 9, 23, 12, 34, 37, 38,
	39, 52, 40, 41, 45, 43, 44, 5, 46, 47,
	42, 6, 26, 49, 50, 51, 10, 3, 4, 13,
	14, 15, 31, 16, 18, 2, 1,
}
var yyPact = [...]int{

	43, -1000, 30, 0, 1, 14, -1000, -14, -14, -14,
	-1000, -14, -9, -1000, -1000, -1000, -1000, 12, -9, -7,
	-9, -1000, -1000, -4, -1000, -19, -10, 9, -1000, -7,
	-7, 27, -1000, -1000, 2, -7, -7, -1000, -1000, -1000,
	5, -19, -1000, -9, -9, -1000, -1000, -1000, -1000, 18,
	-1000, -1000, -1000,
}
var yyPgo = [...]int{

	0, 56, 55, 0, 41, 54, 1, 2, 52, 42,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 2, 2, 4, 4, 3,
	3, 5, 5, 5, 8, 8, 8, 7, 7, 7,
	9, 9, 9, 9, 9, 9, 6, 6,
}
var yyR2 = [...]int{

	0, 6, 2, 3, 3, 2, 3, 1, 2, 1,
	2, 4, 5, 3, 1, 3, 3, 1, 3, 3,
	1, 1, 2, 2, 2, 3, 1, 1,
}
var yyChk = [...]int{

	-1000, -1, -2, 4, 5, 7, -4, 6, 10, 23,
	-4, 10, 12, -4, -4, -4, -4, -3, -5, -6,
	12, 23, 24, 13, -3, -7, -9, 24, 23, 14,
	-6, -8, -3, 17, -6, 18, 22, 19, 20, 21,
	-7, -7, 13, 8, 9, 16, -7, -7, 15, -6,
	-3, -3, 13,
}
var yyDef = [...]int{

	0, -2, 0, 0, 0, 0, 2, 0, 0, 7,
	5, 0, 0, 3, 4, 8, 6, 0, 9, 0,
	0, 26, 27, 0, 10, 0, 17, 20, 21, 0,
	0, 0, 14, 1, 0, 0, 0, 22, 23, 24,
	0, 0, 13, 0, 0, 11, 18, 19, 25, 0,
	15, 16, 12,
}
var yyTok1 = [...]int{

	1,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24,
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
		//line lang.y:33
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
		//line lang.y:45
		{
			yyVAL.varlist = yyDollar[2].varlist
			yyVAL.distinct = false
			yyVAL.count = false
			yyVAL.partial = false
		}
	case 3:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:52
		{
			yyVAL.varlist = yyDollar[3].varlist
			yyVAL.distinct = true
			yyVAL.count = false
			yyVAL.partial = false
		}
	case 4:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:59
		{
			yyVAL.varlist = yyDollar[3].varlist
			yyVAL.distinct = false
			yyVAL.count = false
			yyVAL.partial = true
		}
	case 5:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:66
		{
			yyVAL.varlist = yyDollar[2].varlist
			yyVAL.distinct = false
			yyVAL.count = true
			yyVAL.partial = false
		}
	case 6:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:73
		{
			yyVAL.varlist = yyDollar[3].varlist
			yyVAL.distinct = false
			yyVAL.count = true
			yyVAL.partial = true
		}
	case 7:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:82
		{
			yyVAL.varlist = []turtle.URI{turtle.ParseURI(yyDollar[1].str)}
		}
	case 8:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:86
		{
			yyVAL.varlist = append([]turtle.URI{turtle.ParseURI(yyDollar[1].str)}, yyDollar[2].varlist...)
		}
	case 9:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:92
		{
			if len(yyDollar[1].orclauses) > 0 {
				yyVAL.orclauses = yyDollar[1].orclauses
			} else {
				yyVAL.triples = yyDollar[1].triples
			}
		}
	case 10:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:100
		{
			yyVAL.triples = append(yyDollar[2].triples, yyDollar[1].triples...)
			yyVAL.orclauses = append(yyDollar[2].orclauses, yyDollar[1].orclauses...)
		}
	case 11:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line lang.y:107
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
	case 12:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line lang.y:127
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
	case 13:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:147
		{
			if len(yyDollar[2].orclauses) > 0 {
				yyVAL.orclauses = yyDollar[2].orclauses
			} else {
				yyVAL.triples = yyDollar[2].triples
			}
		}
	case 14:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:157
		{
			yyVAL.triples = yyDollar[1].triples
			yyVAL.orclauses = yyDollar[1].orclauses
		}
	case 15:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:162
		{
			yyVAL.orclauses = []OrClause{{LeftOr: yyDollar[3].orclauses,
				LeftTerms:  yyDollar[3].triples,
				RightOr:    yyDollar[1].orclauses,
				RightTerms: yyDollar[1].triples}}
		}
	case 16:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:169
		{
			yyVAL.orclauses = []OrClause{{LeftOr: yyDollar[3].orclauses,
				LeftTerms:  yyDollar[3].triples,
				RightOr:    yyDollar[1].orclauses,
				RightTerms: yyDollar[1].triples}}
		}
	case 17:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:179
		{
			yyVAL.pred = yyDollar[1].pred
		}
	case 18:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:183
		{
			yyVAL.pred = append(yyDollar[1].pred, yyDollar[3].pred...)
		}
	case 19:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:187
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
	case 20:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:202
		{
			yyVAL.pred = []PathPattern{{Predicate: turtle.ParseURI(yyDollar[1].str), Pattern: PATTERN_SINGLE}}
		}
	case 21:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:206
		{
			yyVAL.pred = []PathPattern{{Predicate: turtle.ParseURI(yyDollar[1].str), Pattern: PATTERN_SINGLE}}
		}
	case 22:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:210
		{
			yyVAL.pred = []PathPattern{{Predicate: turtle.ParseURI(yyDollar[1].str), Pattern: PATTERN_ONE_PLUS}}
		}
	case 23:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:214
		{
			yyVAL.pred = []PathPattern{{Predicate: turtle.ParseURI(yyDollar[1].str), Pattern: PATTERN_ZERO_ONE}}
		}
	case 24:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:218
		{
			yyVAL.pred = []PathPattern{{Predicate: turtle.ParseURI(yyDollar[1].str), Pattern: PATTERN_ZERO_PLUS}}
		}
	case 25:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:222
		{
			yyVAL.pred = yyDollar[2].pred
		}
	case 26:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:228
		{
			yyVAL.val = turtle.ParseURI(yyDollar[1].str)
		}
	case 27:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:232
		{
			yyVAL.val = turtle.ParseURI(yyDollar[1].str)
		}
	}
	goto yystack /* stack new state and value */
}
