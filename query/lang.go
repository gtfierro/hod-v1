//line lang.y:2
package query

import __yyfmt__ "fmt"

//line lang.y:2
import (
	"fmt"
	turtle "github.com/gtfierro/hod/goraptor"
	"io"
	"strconv"
)

//line lang.y:14
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
	limit          int64
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
const LIMIT = 57353
const COMMA = 57354
const LBRACE = 57355
const RBRACE = 57356
const LPAREN = 57357
const RPAREN = 57358
const DOT = 57359
const SEMICOLON = 57360
const SLASH = 57361
const PLUS = 57362
const QUESTION = 57363
const ASTERISK = 57364
const BAR = 57365
const LINK = 57366
const VAR = 57367
const URI = 57368
const FULLURI = 57369
const LBRACK = 57370
const RBRACK = 57371
const NUMBER = 57372

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
	"LIMIT",
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
	"FULLURI",
	"LBRACK",
	"RBRACK",
	"NUMBER",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line lang.y:298

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
	limit     int64
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
			{Token: SELECT, Pattern: "(SELECT)|(select)"},
			{Token: COUNT, Pattern: "COUNT"},
			{Token: DISTINCT, Pattern: "DISTINCT"},
			{Token: WHERE, Pattern: "WHERE"},
			{Token: OR, Pattern: "OR"},
			{Token: UNION, Pattern: "UNION"},
			{Token: PARTIAL, Pattern: "PARTIAL"},
			{Token: LIMIT, Pattern: "LIMIT"},
			{Token: NUMBER, Pattern: "[0-9]+"},
			{Token: URI, Pattern: "[a-zA-Z_]+:[a-zA-Z0-9_\\-#%$@]+"},
			{Token: VAR, Pattern: "\\?[a-zA-Z0-9_]+"},
			{Token: LINK, Pattern: "[a-zA-Z][a-zA-Z0-9_-]*"},
			{Token: QUESTION, Pattern: "\\?"},
			{Token: SLASH, Pattern: "/"},
			{Token: PLUS, Pattern: "\\+"},
			{Token: ASTERISK, Pattern: "\\*"},
			{Token: FULLURI, Pattern: "<[^<>\"{}|^`\\\\]*>"},
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

const yyNprod = 38
const yyPrivate = 57344

var yyTokenNames []string
var yyStates []string

const yyLast = 79

var yyAct = [...]int{

	19, 21, 27, 24, 22, 32, 59, 23, 25, 26,
	40, 17, 12, 10, 58, 63, 23, 25, 26, 36,
	29, 31, 28, 39, 37, 34, 60, 10, 7, 35,
	25, 26, 8, 30, 45, 48, 49, 50, 13, 42,
	34, 34, 51, 52, 56, 57, 67, 10, 41, 60,
	34, 34, 61, 62, 64, 65, 66, 46, 6, 54,
	55, 47, 5, 11, 44, 53, 14, 15, 16, 3,
	4, 18, 33, 38, 20, 9, 43, 2, 1,
}
var yyPact = [...]int{

	65, -1000, 55, 22, 2, 25, -1000, -12, -12, -12,
	-17, -1000, -12, -9, -1000, -1000, -1000, -2, -1000, 19,
	-9, 4, -9, -1000, -1000, -1000, -1000, -19, 36, 27,
	53, -1000, -18, 38, 15, -1000, 4, 4, 51, -1000,
	-1000, -2, -2, -4, -24, 9, 4, 4, -1000, -1000,
	-1000, -1, -18, -1000, -9, -9, -1000, -1000, -1000, -1000,
	-1000, -1000, -1000, -1000, 32, -1000, -1000, -1000,
}
var yyPgo = [...]int{

	0, 78, 77, 0, 76, 58, 75, 2, 74, 1,
	5, 73, 72, 3,
}
var yyR1 = [...]int{

	0, 1, 2, 2, 2, 2, 2, 4, 4, 5,
	5, 6, 6, 7, 7, 7, 7, 3, 3, 8,
	8, 8, 11, 11, 11, 10, 10, 10, 13, 13,
	12, 12, 12, 12, 12, 12, 9, 9,
}
var yyR2 = [...]int{

	0, 7, 2, 3, 3, 2, 3, 0, 2, 1,
	2, 4, 1, 1, 1, 3, 3, 1, 2, 4,
	5, 3, 1, 3, 3, 1, 3, 3, 1, 1,
	1, 1, 2, 2, 2, 3, 1, 1,
}
var yyChk = [...]int{

	-1000, -1, -2, 4, 5, 7, -5, 6, 10, -6,
	25, -5, 10, 13, -5, -5, -5, 28, -5, -3,
	-8, -9, 13, 25, -13, 26, 27, -7, 24, 22,
	14, -3, -10, -12, -13, 25, 15, -9, -11, -3,
	29, 12, 12, -4, 11, -9, 19, 23, 20, 21,
	22, -10, -10, 14, 8, 9, -7, -7, 18, 30,
	17, -10, -10, 16, -9, -3, -3, 14,
}
var yyDef = [...]int{

	0, -2, 0, 0, 0, 0, 2, 0, 0, 9,
	12, 5, 0, 0, 3, 4, 10, 0, 6, 0,
	17, 0, 0, 36, 37, 28, 29, 0, 13, 14,
	7, 18, 0, 25, 30, 31, 0, 0, 0, 22,
	11, 0, 0, 0, 0, 0, 0, 0, 32, 33,
	34, 0, 0, 21, 0, 0, 15, 16, 1, 8,
	19, 26, 27, 35, 0, 23, 24, 20,
}
var yyTok1 = [...]int{

	1,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30,
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
		yyDollar = yyS[yypt-7 : yypt+1]
		//line lang.y:38
		{
			yylex.(*lexer).varlist = yyDollar[1].varlist
			yylex.(*lexer).distinct = yyDollar[1].distinct
			yylex.(*lexer).triples = yyDollar[4].triples
			yylex.(*lexer).distinct = yyDollar[1].distinct
			yylex.(*lexer).partial = yyDollar[1].partial
			yylex.(*lexer).count = yyDollar[1].count
			yylex.(*lexer).orclauses = yyDollar[4].orclauses
			yylex.(*lexer).limit = yyDollar[6].limit
		}
	case 2:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:51
		{
			yyVAL.varlist = yyDollar[2].varlist
			yyVAL.distinct = false
			yyVAL.count = false
			yyVAL.partial = false
		}
	case 3:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:58
		{
			yyVAL.varlist = yyDollar[3].varlist
			yyVAL.distinct = true
			yyVAL.count = false
			yyVAL.partial = false
		}
	case 4:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:65
		{
			yyVAL.varlist = yyDollar[3].varlist
			yyVAL.distinct = false
			yyVAL.count = false
			yyVAL.partial = true
		}
	case 5:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:72
		{
			yyVAL.varlist = yyDollar[2].varlist
			yyVAL.distinct = false
			yyVAL.count = true
			yyVAL.partial = false
		}
	case 6:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:79
		{
			yyVAL.varlist = yyDollar[3].varlist
			yyVAL.distinct = false
			yyVAL.count = true
			yyVAL.partial = true
		}
	case 7:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line lang.y:88
		{
			yyVAL.limit = 0
		}
	case 8:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:92
		{
			num, err := strconv.ParseInt(yyDollar[2].str, 10, 64)
			if err != nil {
				yylex.Error(err.Error())
			}
			yyVAL.limit = num
		}
	case 9:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:102
		{
			yyVAL.varlist = []SelectVar{yyDollar[1].selectvar}
		}
	case 10:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:106
		{
			yyVAL.varlist = append([]SelectVar{yyDollar[1].selectvar}, yyDollar[2].varlist...)
		}
	case 11:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line lang.y:112
		{
			if yyDollar[3].selectAllLinks {
				yyVAL.selectvar = SelectVar{Var: turtle.ParseURI(yyDollar[1].str), AllLinks: true}
			} else {
				yyVAL.selectvar = SelectVar{Var: turtle.ParseURI(yyDollar[1].str), AllLinks: false, Links: yyDollar[3].links}
			}
		}
	case 12:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:120
		{
			yyVAL.selectvar = SelectVar{Var: turtle.ParseURI(yyDollar[1].str), AllLinks: false}
		}
	case 13:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:126
		{
			yyVAL.links = []Link{{Name: yyDollar[1].str}}
		}
	case 14:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:130
		{
			yyVAL.selectAllLinks = true
		}
	case 15:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:134
		{
			yyVAL.links = append([]Link{{Name: yyDollar[1].str}}, yyDollar[3].links...)
		}
	case 16:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:138
		{
			yyVAL.selectAllLinks = true
		}
	case 17:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:144
		{
			if len(yyDollar[1].orclauses) > 0 {
				yyVAL.orclauses = yyDollar[1].orclauses
			} else {
				yyVAL.triples = yyDollar[1].triples
			}
		}
	case 18:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:152
		{
			yyVAL.triples = append(yyDollar[2].triples, yyDollar[1].triples...)
			yyVAL.orclauses = append(yyDollar[2].orclauses, yyDollar[1].orclauses...)
		}
	case 19:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line lang.y:159
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
	case 20:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line lang.y:179
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
	case 21:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:199
		{
			if len(yyDollar[2].orclauses) > 0 {
				yyVAL.orclauses = yyDollar[2].orclauses
			} else {
				yyVAL.triples = yyDollar[2].triples
			}
		}
	case 22:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:209
		{
			yyVAL.triples = yyDollar[1].triples
			yyVAL.orclauses = yyDollar[1].orclauses
		}
	case 23:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:214
		{
			yyVAL.orclauses = []OrClause{{LeftOr: yyDollar[3].orclauses,
				LeftTerms:  yyDollar[3].triples,
				RightOr:    yyDollar[1].orclauses,
				RightTerms: yyDollar[1].triples}}
		}
	case 24:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:221
		{
			yyVAL.orclauses = []OrClause{{LeftOr: yyDollar[3].orclauses,
				LeftTerms:  yyDollar[3].triples,
				RightOr:    yyDollar[1].orclauses,
				RightTerms: yyDollar[1].triples}}
		}
	case 25:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:231
		{
			yyVAL.pred = yyDollar[1].pred
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:235
		{
			yyVAL.pred = append(yyDollar[1].pred, yyDollar[3].pred...)
		}
	case 27:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:239
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
	case 28:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:254
		{
			yyVAL.str = yyDollar[1].str
		}
	case 29:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:258
		{
			yyVAL.str = yyDollar[1].str[1 : len(yyDollar[1].str)-1]
		}
	case 30:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:264
		{
			yyVAL.pred = []PathPattern{{Predicate: turtle.ParseURI(yyDollar[1].str), Pattern: PATTERN_SINGLE}}
		}
	case 31:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:268
		{
			yyVAL.pred = []PathPattern{{Predicate: turtle.ParseURI(yyDollar[1].str), Pattern: PATTERN_SINGLE}}
		}
	case 32:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:272
		{
			yyVAL.pred = []PathPattern{{Predicate: turtle.ParseURI(yyDollar[1].str), Pattern: PATTERN_ONE_PLUS}}
		}
	case 33:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:276
		{
			yyVAL.pred = []PathPattern{{Predicate: turtle.ParseURI(yyDollar[1].str), Pattern: PATTERN_ZERO_ONE}}
		}
	case 34:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line lang.y:280
		{
			yyVAL.pred = []PathPattern{{Predicate: turtle.ParseURI(yyDollar[1].str), Pattern: PATTERN_ZERO_PLUS}}
		}
	case 35:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line lang.y:284
		{
			yyVAL.pred = yyDollar[2].pred
		}
	case 36:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:290
		{
			yyVAL.val = turtle.ParseURI(yyDollar[1].str)
		}
	case 37:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line lang.y:294
		{
			yyVAL.val = turtle.ParseURI(yyDollar[1].str)
		}
	}
	goto yystack /* stack new state and value */
}
