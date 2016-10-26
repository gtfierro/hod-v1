//line grammar.y:2
package turtle

import __yyfmt__ "fmt"

//line grammar.y:2
import (
	"fmt"
	"io"
)

//line grammar.y:11
type ttlSymType struct {
	yys int
	str string
}

const IRIREF = 57346
const PNAME_NS = 57347
const PNAME_LN = 57348
const BLANK_NODE_LABEL = 57349
const LANGTAG = 57350
const INTEGER = 57351
const DECIMAL = 57352
const DOUBLE = 57353
const EXPONENT = 57354
const STRING_LITERAL_QUOTE = 57355
const STRING_LITERAL_SINGLE_QUOTE = 57356
const STRING_LITERAL_LONG_SINGLE_QUOTE = 57357
const STRING_LITERAL_LONG_QUOTE = 57358
const WS = 57359
const ANON = 57360
const PN_CHARS_BASE = 57361
const PN_CHARS_U = 57362
const PN_CHARS = 57363
const PN_PREFIX = 57364
const PN_LOCAL = 57365
const PLX = 57366
const PERCENT = 57367
const HEX = 57368
const PN_LOCAL_ESC = 57369
const LANGLE = 57370
const RANGLE = 57371
const LBRACK = 57372
const RBRACK = 57373
const LPAREN = 57374
const RPAREN = 57375
const DOUBLECARAT = 57376
const SEMICOLON = 57377
const TRUE = 57378
const FALSE = 57379
const DOT = 57380
const ATPREFIX = 57381
const ATBASE = 57382
const BASE = 57383
const PREFIX = 57384
const COMMA = 57385
const A = 57386

var ttlToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"IRIREF",
	"PNAME_NS",
	"PNAME_LN",
	"BLANK_NODE_LABEL",
	"LANGTAG",
	"INTEGER",
	"DECIMAL",
	"DOUBLE",
	"EXPONENT",
	"STRING_LITERAL_QUOTE",
	"STRING_LITERAL_SINGLE_QUOTE",
	"STRING_LITERAL_LONG_SINGLE_QUOTE",
	"STRING_LITERAL_LONG_QUOTE",
	"WS",
	"ANON",
	"PN_CHARS_BASE",
	"PN_CHARS_U",
	"PN_CHARS",
	"PN_PREFIX",
	"PN_LOCAL",
	"PLX",
	"PERCENT",
	"HEX",
	"PN_LOCAL_ESC",
	"LANGLE",
	"RANGLE",
	"LBRACK",
	"RBRACK",
	"LPAREN",
	"RPAREN",
	"DOUBLECARAT",
	"SEMICOLON",
	"TRUE",
	"FALSE",
	"DOT",
	"ATPREFIX",
	"ATBASE",
	"BASE",
	"PREFIX",
	"COMMA",
	"A",
}
var ttlStatenames = [...]string{}

const ttlEofCode = 1
const ttlErrCode = 2
const ttlInitialStackSize = 16

//line grammar.y:156

const eof = 0

type lexer struct {
	scanner      *Scanner
	tokens       []Token
	values       [][]byte
	error        error
	iri          string
	namespaces   map[string]string
	bnodeLabels  map[string]string
	curSubject   string
	curPredicate string
}

func newlexer(r io.Reader) *lexer {
	scanner := NewScanner(
		[]Definition{
			{Token: LBRACK, Pattern: "\\["},
			{Token: RBRACK, Pattern: "\\]"},
			{Token: LPAREN, Pattern: "\\("},
			{Token: RPAREN, Pattern: "\\)"},
			{Token: DOUBLECARAT, Pattern: "\\^\\^"},
			{Token: SEMICOLON, Pattern: ";"},
			{Token: TRUE, Pattern: "true"},
			{Token: FALSE, Pattern: "false"},
			{Token: DOT, Pattern: "\\."},
			{Token: ATPREFIX, Pattern: "@prefix"},
			{Token: ATBASE, Pattern: "@base"},
			{Token: BASE, Pattern: "base"},
			{Token: PREFIX, Pattern: "prefix"},
			{Token: COMMA, Pattern: "comma"},
			{Token: A, Pattern: "a"},
			{Token: IRIREF, Pattern: "<(([^<>\"{}|^`])|(\\\\u([0-9]|[A-F]|[a-f]){4})|(\\\\U([0-9]|[A-F]|[a-f]){8}))*>"},
			{Token: PNAME_LN, Pattern: "([A-Z]|[a-z]|[0-9]{1,6})((([A-Z]|[a-z]|[0-9]{1,6}|_|-|\\.)*([A-Z]|[a-z]|[0-9]{1,6}|_|-|\\.))?)?:([A-Z]|[a-z]|[0-9]{1,6}|_|:|(\\\\(_|~|\\.|-|!|$|&|'|(|)|\\*|\\+|,|;|=|/|\\?|#|@|%)))(([A-Z]|[a-z]|[0-9]{1,6}|_|-|\\.|:|(%([0-9]|[A-F]|[a-f])))*([A-Z]|[a-z]|[0-9]{1,6}|_|-|:|(%([0-9]|[A-F]|[a-f])))?)"},
			{Token: PNAME_NS, Pattern: "([A-Z]|[a-z]|[0-9]{1,6})((([A-Z]|[a-z]|[0-9]{1,6}|_|-|\\.)*([A-Z]|[a-z]|[0-9]{1,6}|_|-|\\.))?)?:"},
			{Token: BLANK_NODE_LABEL, Pattern: "_:([A-Z]|[a-z]|[0-9]{1,6}_)(([A-Z]|[a-z]|[0-9]{1,6}|_|-|\\.)*[A-Z]|[a-z]|[0-9]{1,6}|_|-)?"},
			{Token: LANGTAG, Pattern: "@[a-zA-Z]+(-[a-zA-Z0-9]+)*"},
			{Token: INTEGER, Pattern: "[+-]?[0-9]+"},
			{Token: DECIMAL, Pattern: "[+-]?[0-9]*\\.[0-9]+"},
			{Token: DOUBLE, Pattern: "[+-]?(([0-9]+\\.[0-9]*([eE][+-]?[0-9]+))|(\\.[0-9]+([eE][+-]?[0-9]+))|([0-9]+([eE][+-]?[0-9]+)))"},
			{Token: EXPONENT, Pattern: "([eE][+-]?[0-9]+)"},
			{Token: STRING_LITERAL_QUOTE, Pattern: "\"[^\"]*\""},
			{Token: STRING_LITERAL_SINGLE_QUOTE, Pattern: "'[^']*'"},
			{Token: PN_CHARS_BASE, Pattern: "[A-Z]|[a-z]|[0-9]{1,6}"},
			{Token: PN_CHARS_U, Pattern: "[A-Z]|[a-z]|[0-9]{1,6}|_"},
			{Token: PN_CHARS, Pattern: "([A-Z]|[a-z]|[0-9]{1,6}|_|-)"},
			{Token: PN_LOCAL, Pattern: "([A-Z]|[a-z]|[0-9]{1,6}|_|:|(\\\\(_|~|\\.|-|!|$|&|'|(|)|\\*|\\+|,|;|=|/|\\?|#|@|%)))(([A-Z]|[a-z]|[0-9]{1,6}|_|-|\\.|:|(%([0-9]|[A-F]|[a-f])))*([A-Z]|[a-z]|[0-9]{1,6}|_|-|:|(%([0-9]|[A-F]|[a-f])))?)"},
			{Token: PLX, Pattern: "%([0-9]|[A-F]|[a-f])"},
			{Token: PN_LOCAL_ESC, Pattern: "\\\\(_|~|\\.|-|!|$|&|'|(|)|\\*|\\+|,|;|=|/|\\?|#|@|%)"},
			{Token: LANGLE, Pattern: "<"},
			{Token: RANGLE, Pattern: ">"},
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
			fmt.Println("ERROR", string(r.Value))
		}
		return eof
	}
	fmt.Printf("%s: %s\n", TokenName(r.Token), string(r.Value))
	l.tokens = append(l.tokens, r.Token)
	l.values = append(l.values, r.Value)
	return int(r.Token)
}

func (l *lexer) Error(s string) {
	l.error = fmt.Errorf(s)
}

func TokenName(t Token) string {
	return ttlTokname(int(t) - 57342)
}

//line yacctab:1
var ttlExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const ttlNprod = 57
const ttlPrivate = 57344

var ttlTokenNames []string
var ttlStates []string

const ttlLast = 96

var ttlAct = [...]int{

	33, 16, 41, 60, 16, 71, 29, 20, 26, 25,
	72, 63, 40, 28, 70, 66, 65, 37, 34, 35,
	20, 26, 25, 68, 64, 42, 39, 62, 44, 18,
	38, 42, 18, 61, 20, 26, 25, 22, 36, 51,
	52, 53, 42, 56, 57, 58, 59, 32, 23, 69,
	20, 26, 25, 22, 67, 43, 17, 45, 11, 17,
	19, 11, 24, 21, 23, 2, 54, 55, 50, 27,
	73, 49, 42, 48, 61, 75, 19, 74, 24, 47,
	46, 31, 30, 10, 9, 12, 13, 15, 14, 8,
	7, 6, 5, 4, 3, 1,
}
var ttlPact = [...]int{

	46, -1000, -1000, 46, -1000, -25, -1000, -1000, -1000, -1000,
	3, 3, 14, 34, 12, 26, -1000, -1000, -1000, 3,
	-1000, -1000, -1000, -1000, 30, -1000, -1000, -1000, -1000, -1000,
	30, -1000, -1000, -1000, -1000, 23, -27, 20, -1000, -15,
	-18, 30, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	15, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000, -1000,
	-21, -38, -28, -1000, -1000, -1000, -1000, -1000, -1000, 16,
	3, 30, -1000, -1000, -1000, -1000,
}
var ttlPgo = [...]int{

	0, 95, 65, 94, 93, 92, 91, 90, 89, 84,
	83, 6, 57, 82, 3, 81, 0, 55, 28, 2,
	80, 79, 73, 71, 12, 68, 63,
}
var ttlR1 = [...]int{

	0, 1, 2, 2, 3, 3, 4, 4, 4, 4,
	6, 7, 9, 8, 5, 5, 5, 11, 11, 11,
	13, 13, 10, 10, 10, 15, 19, 19, 19, 19,
	19, 20, 20, 20, 14, 14, 12, 18, 24, 24,
	22, 22, 22, 21, 21, 23, 23, 25, 25, 25,
	25, 16, 16, 26, 26, 17, 17,
}
var ttlR2 = [...]int{

	0, 1, 1, 2, 1, 2, 1, 1, 1, 1,
	4, 3, 2, 3, 2, 1, 2, 2, 3, 4,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 3, 3, 3, 1, 2,
	1, 1, 1, 2, 3, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1,
}
var ttlChk = [...]int{

	-1000, -1, -2, -3, -4, -5, -6, -7, -8, -9,
	-10, -12, 39, 40, 42, 41, -16, -17, -18, 30,
	4, -26, 7, 18, 32, 6, 5, -2, 38, -11,
	-13, -15, 44, -16, -11, 5, 4, 5, 4, -11,
	-24, -19, -16, -17, -18, -12, -20, -21, -22, -23,
	-25, 9, 10, 11, 36, 37, 13, 14, 15, 16,
	-14, -19, 4, 38, 4, 31, 33, -24, 8, 34,
	35, 43, 38, -16, -11, -14,
}
var ttlDef = [...]int{

	0, -2, 1, 2, 4, 0, 6, 7, 8, 9,
	0, 15, 0, 0, 0, 0, 22, 23, 24, 0,
	51, 52, 55, 56, 0, 53, 54, 3, 5, 14,
	0, 20, 21, 25, 16, 0, 0, 0, 12, 0,
	0, 38, 26, 27, 28, 29, 30, 31, 32, 33,
	0, 40, 41, 42, 45, 46, 47, 48, 49, 50,
	17, 34, 0, 11, 13, 36, 37, 39, 43, 0,
	18, 0, 10, 44, 19, 35,
}
var ttlTok1 = [...]int{

	1,
}
var ttlTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21,
	22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 38, 39, 40, 41,
	42, 43, 44,
}
var ttlTok3 = [...]int{
	0,
}

var ttlErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	ttlDebug        = 0
	ttlErrorVerbose = false
)

type ttlLexer interface {
	Lex(lval *ttlSymType) int
	Error(s string)
}

type ttlParser interface {
	Parse(ttlLexer) int
	Lookahead() int
}

type ttlParserImpl struct {
	lval  ttlSymType
	stack [ttlInitialStackSize]ttlSymType
	char  int
}

func (p *ttlParserImpl) Lookahead() int {
	return p.char
}

func ttlNewParser() ttlParser {
	return &ttlParserImpl{}
}

const ttlFlag = -1000

func ttlTokname(c int) string {
	if c >= 1 && c-1 < len(ttlToknames) {
		if ttlToknames[c-1] != "" {
			return ttlToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func ttlStatname(s int) string {
	if s >= 0 && s < len(ttlStatenames) {
		if ttlStatenames[s] != "" {
			return ttlStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func ttlErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !ttlErrorVerbose {
		return "syntax error"
	}

	for _, e := range ttlErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + ttlTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := ttlPact[state]
	for tok := TOKSTART; tok-1 < len(ttlToknames); tok++ {
		if n := base + tok; n >= 0 && n < ttlLast && ttlChk[ttlAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if ttlDef[state] == -2 {
		i := 0
		for ttlExca[i] != -1 || ttlExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; ttlExca[i] >= 0; i += 2 {
			tok := ttlExca[i]
			if tok < TOKSTART || ttlExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if ttlExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += ttlTokname(tok)
	}
	return res
}

func ttllex1(lex ttlLexer, lval *ttlSymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = ttlTok1[0]
		goto out
	}
	if char < len(ttlTok1) {
		token = ttlTok1[char]
		goto out
	}
	if char >= ttlPrivate {
		if char < ttlPrivate+len(ttlTok2) {
			token = ttlTok2[char-ttlPrivate]
			goto out
		}
	}
	for i := 0; i < len(ttlTok3); i += 2 {
		token = ttlTok3[i+0]
		if token == char {
			token = ttlTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = ttlTok2[1] /* unknown char */
	}
	if ttlDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", ttlTokname(token), uint(char))
	}
	return char, token
}

func ttlParse(ttllex ttlLexer) int {
	return ttlNewParser().Parse(ttllex)
}

func (ttlrcvr *ttlParserImpl) Parse(ttllex ttlLexer) int {
	var ttln int
	var ttlVAL ttlSymType
	var ttlDollar []ttlSymType
	_ = ttlDollar // silence set and not used
	ttlS := ttlrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	ttlstate := 0
	ttlrcvr.char = -1
	ttltoken := -1 // ttlrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		ttlstate = -1
		ttlrcvr.char = -1
		ttltoken = -1
	}()
	ttlp := -1
	goto ttlstack

ret0:
	return 0

ret1:
	return 1

ttlstack:
	/* put a state and value onto the stack */
	if ttlDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", ttlTokname(ttltoken), ttlStatname(ttlstate))
	}

	ttlp++
	if ttlp >= len(ttlS) {
		nyys := make([]ttlSymType, len(ttlS)*2)
		copy(nyys, ttlS)
		ttlS = nyys
	}
	ttlS[ttlp] = ttlVAL
	ttlS[ttlp].yys = ttlstate

ttlnewstate:
	ttln = ttlPact[ttlstate]
	if ttln <= ttlFlag {
		goto ttldefault /* simple state */
	}
	if ttlrcvr.char < 0 {
		ttlrcvr.char, ttltoken = ttllex1(ttllex, &ttlrcvr.lval)
	}
	ttln += ttltoken
	if ttln < 0 || ttln >= ttlLast {
		goto ttldefault
	}
	ttln = ttlAct[ttln]
	if ttlChk[ttln] == ttltoken { /* valid shift */
		ttlrcvr.char = -1
		ttltoken = -1
		ttlVAL = ttlrcvr.lval
		ttlstate = ttln
		if Errflag > 0 {
			Errflag--
		}
		goto ttlstack
	}

ttldefault:
	/* default state action */
	ttln = ttlDef[ttlstate]
	if ttln == -2 {
		if ttlrcvr.char < 0 {
			ttlrcvr.char, ttltoken = ttllex1(ttllex, &ttlrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if ttlExca[xi+0] == -1 && ttlExca[xi+1] == ttlstate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			ttln = ttlExca[xi+0]
			if ttln < 0 || ttln == ttltoken {
				break
			}
		}
		ttln = ttlExca[xi+1]
		if ttln < 0 {
			goto ret0
		}
	}
	if ttln == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			ttllex.Error(ttlErrorMessage(ttlstate, ttltoken))
			Nerrs++
			if ttlDebug >= 1 {
				__yyfmt__.Printf("%s", ttlStatname(ttlstate))
				__yyfmt__.Printf(" saw %s\n", ttlTokname(ttltoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for ttlp >= 0 {
				ttln = ttlPact[ttlS[ttlp].yys] + ttlErrCode
				if ttln >= 0 && ttln < ttlLast {
					ttlstate = ttlAct[ttln] /* simulate a shift of "error" */
					if ttlChk[ttlstate] == ttlErrCode {
						goto ttlstack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if ttlDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", ttlS[ttlp].yys)
				}
				ttlp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if ttlDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", ttlTokname(ttltoken))
			}
			if ttltoken == ttlEofCode {
				goto ret1
			}
			ttlrcvr.char = -1
			ttltoken = -1
			goto ttlnewstate /* try again in the same state */
		}
	}

	/* reduction by production ttln */
	if ttlDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", ttln, ttlStatname(ttlstate))
	}

	ttlnt := ttln
	ttlpt := ttlp
	_ = ttlpt // guard against "declared and not used"

	ttlp -= ttlR2[ttln]
	// ttlp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if ttlp+1 >= len(ttlS) {
		nyys := make([]ttlSymType, len(ttlS)*2)
		copy(nyys, ttlS)
		ttlS = nyys
	}
	ttlVAL = ttlS[ttlp+1]

	/* consult goto table to find next state */
	ttln = ttlR1[ttln]
	ttlg := ttlPgo[ttln]
	ttlj := ttlg + ttlS[ttlp].yys + 1

	if ttlj >= ttlLast {
		ttlstate = ttlAct[ttlg]
	} else {
		ttlstate = ttlAct[ttlj]
		if ttlChk[ttlstate] != -ttln {
			ttlstate = ttlAct[ttlg]
		}
	}
	// dummy call; replaced with literal code
	switch ttlnt {

	case 10:
		ttlDollar = ttlS[ttlpt-4 : ttlpt+1]
		//line grammar.y:42
		{
			//fmt.Println("> PREFIX ID:",$1,$2,$3,$4)
		}
	case 11:
		ttlDollar = ttlS[ttlpt-3 : ttlpt+1]
		//line grammar.y:48
		{
			//fmt.Println("> BASE:",$1,$2,$3)
		}
	case 14:
		ttlDollar = ttlS[ttlpt-2 : ttlpt+1]
		//line grammar.y:60
		{
			fmt.Println("> 1")
		}
	case 15:
		ttlDollar = ttlS[ttlpt-1 : ttlpt+1]
		//line grammar.y:64
		{
			fmt.Println("> 2")
		}
	case 16:
		ttlDollar = ttlS[ttlpt-2 : ttlpt+1]
		//line grammar.y:68
		{
			fmt.Println("> 3")
		}
	case 17:
		ttlDollar = ttlS[ttlpt-2 : ttlpt+1]
		//line grammar.y:74
		{
			fmt.Println("> 1")
		}
	case 18:
		ttlDollar = ttlS[ttlpt-3 : ttlpt+1]
		//line grammar.y:78
		{
			fmt.Println("> 2")
		}
	case 19:
		ttlDollar = ttlS[ttlpt-4 : ttlpt+1]
		//line grammar.y:82
		{
			fmt.Println("> 3")
		}
	}
	goto ttlstack /* stack new state and value */
}
