package jsonpath

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// https://talks.golang.org/2011/lex.slide

type lexer struct {
	input   []rune
	start   int
	pos     int
	items   chan jsonpathSym
	lastSym jsonpathSym
}

type jsonpathSym interface {
	Lexeme() string
	identifier() int
}

type singleCh struct{ ch int }
type ident struct{ val string }
type number struct{ val float64 }
type str struct{ val string }

type lte struct{}
type gte struct{}
type eq struct{}
type neq struct{}
type and struct{}
type or struct{}
type not struct{}

type keyword struct{ which int }

type errSym struct {
	msg   string
	begin int
	end   int
}

func (s singleCh) Lexeme() string { return string(s.ch) }
func (s ident) Lexeme() string    { return s.val }
func (s number) Lexeme() string   { return fmt.Sprintf("%v", s.val) }
func (s str) Lexeme() string      { return fmt.Sprintf("'%v'", s.val) }

func (s lte) Lexeme() string { return "<=" }
func (s gte) Lexeme() string { return ">=" }
func (s eq) Lexeme() string  { return "==" }
func (s neq) Lexeme() string { return "!=" }
func (s and) Lexeme() string { return "&&" }
func (s or) Lexeme() string  { return "||" }
func (s not) Lexeme() string { return "!" }

func (s keyword) Lexeme() string { return invertedKeywords[s.which] }

func (s errSym) Lexeme() string { return s.msg }

func (s singleCh) identifier() int { return s.ch }
func (s ident) identifier() int    { return IDENT }
func (s number) identifier() int   { return NUMBER }
func (s str) identifier() int      { return STR }

func (s lte) identifier() int { return LTE }
func (s gte) identifier() int { return GTE }
func (s eq) identifier() int  { return EQ }
func (s neq) identifier() int { return NEQ }
func (s and) identifier() int { return AND }
func (s or) identifier() int  { return OR }
func (s not) identifier() int { return UNOT }

func (s keyword) identifier() int { return s.which }

func (s errSym) identifier() int { return -1 }

const eof = 0

func validFirstIdentifierChar(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '$'
}

func validIdentifierChar(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '$'
}

func (l *lexer) currentRune() rune {
	if l.pos >= len(l.input) {
		return eof
	}
	return l.input[l.pos]
}

func (l *lexer) peek() rune {
	if l.pos >= len(l.input) {
		return eof
	}
	return l.input[l.pos]
}

func (l *lexer) peekBack() rune {
	if l.pos-1 < 0 {
		return eof
	}
	return l.input[l.pos-1]
}

func (l *lexer) advance(i int) {
	l.pos += i
}

func (l *lexer) scooch(i int) {
	l.start += i
}

func (l *lexer) gobble(i int) {
	if l.start < l.pos {
		panic("can only gobble without advancing")
	}
	l.pos += i
	l.start = l.pos
}

func (l *lexer) run() {
	for state := startState; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) emit(sym jsonpathSym) {
	l.start = l.pos
	l.lastSym = sym
	l.items <- sym
}

func (l *lexer) err(msg string, args ...interface{}) {
	l.emit(errSym{fmt.Sprintf(msg, args...), l.start, l.pos})
}

func (l *lexer) current() string {
	return string(l.input[l.start:l.pos])
}

type stateFn func(*lexer) stateFn

var opChars = map[rune]struct{}{
	'<': struct{}{},
	'>': struct{}{},
	'|': struct{}{},
	'!': struct{}{},
	'&': struct{}{},
	'=': struct{}{},
	'+': struct{}{},
	'-': struct{}{},
	'*': struct{}{},
	'/': struct{}{},
	',': struct{}{},
	'%': struct{}{},
	'$': struct{}{},
	'.': struct{}{},
	'@': struct{}{},
	'?': struct{}{},
	'[': struct{}{},
	']': struct{}{},
	'(': struct{}{},
	')': struct{}{},
}

func startState(l *lexer) stateFn {
	for unicode.IsSpace(l.peek()) {
		l.gobble(1)
	}
	ch := l.peek()

	switch {
	case ch == '$':
		return identifierState
	case validFirstIdentifierChar(ch):
		if l.lastSym == (singleCh{'.'}) {
			return identifierFollowingDotState
		}
		return keywordState
	case unicode.IsDigit(ch):
		return numberState
	case ch == '\'':
		parseString(l, '\'')
		return startState
	case ch == '"':
		parseString(l, '"')
		return startState
	}

	if _, ok := opChars[ch]; ok {
		return opState
	}

	return nil
}

func escapable(r rune) bool {
	switch r {
	case 'n':
	default:
		return false
	}
	return true
}

func reescape(s string, quoteChar rune) string {
	// Overestimate.
	result := make([]rune, 0, len(s))
	// TODO: this conversion should be unnecessary
	runes := []rune(s)

	for i := 0; i < len(runes); i++ {
		if runes[i] == '\\' {
			i++
			switch runes[i] {
			case 'n':
				result = append(result, '\n')
			case quoteChar:
				result = append(result, quoteChar)
			}
		} else {
			result = append(result, runes[i])
		}
	}
	return string(result)
}

func parseString(l *lexer, quoteChar rune) stateFn {
	l.advance(1)
	ch := l.peek()
	for ch != quoteChar {
		if ch == eof {
			l.err("unterminated string")
			return nil
		}
		if ch == '\\' {
			l.advance(1)
			if l.peek() != quoteChar && !escapable(l.peek()) {
				l.err("invalid escape sequence \"\\%s\"", string(l.peek()))
				return nil
			}
		}
		l.advance(1)
		ch = l.peek()
	}
	l.scooch(1)
	l.emit(str{reescape(l.current(), quoteChar)})
	l.gobble(1)
	return startState
}

func numberState(l *lexer) stateFn {
	for unicode.IsDigit(l.peek()) {
		l.advance(1)
	}
	if l.peek() == '.' {
		l.advance(1)
		for unicode.IsDigit(l.peek()) {
			l.advance(1)
		}
	}
	if l.peek() == 'e' {
		l.advance(1)
		for unicode.IsDigit(l.peek()) {
			l.advance(1)
		}
	}
	parsed, err := strconv.ParseFloat(l.current(), 64)
	if err != nil {
		panic("PANIC!!!")
	}
	l.emit(number{parsed})
	return startState
}

func opState(l *lexer) stateFn {
	ch := l.peek()
	switch ch {
	case '!':
		l.advance(1)
		if l.peek() == '=' {
			l.advance(1)
			l.emit(neq{})
		} else {
			l.emit(not{})
		}
	case '=':
		l.advance(1)
		if l.peek() == '=' {
			l.advance(1)
			l.emit(eq{})
		} else {
			l.err("use == instead of =")
		}
	case '<':
		l.advance(1)
		if l.peek() == '=' {
			l.advance(1)
			l.emit(lte{})
		} else if l.peek() == '>' {
			l.advance(1)
			l.emit(neq{})
		} else {
			l.emit(singleCh{'<'})
		}
	case '>':
		l.advance(1)
		if l.peek() == '=' {
			l.advance(1)
			l.emit(gte{})
		} else {
			l.emit(singleCh{'>'})
		}
	case '&':
		l.advance(1)
		if l.peek() == '&' {
			l.advance(1)
			l.emit(and{})
		} else {
			l.err("& must be followed by &")
			return nil
		}
	case '|':
		l.advance(1)
		if l.peek() == '|' {
			l.advance(1)
			l.emit(or{})
		} else {
			l.err("| must be followed by |")
			return nil
		}
	default:
		l.advance(1)
		l.emit(singleCh{ch: int(ch)})
	}
	return startState
}

var keywords = map[string]int{
	"true":       TRUE,
	"false":      FALSE,
	"null":       NULL,
	"strict":     STRICT,
	"lax":        LAX,
	"last":       LAST,
	"to":         TO,
	"exists":     EXISTS,
	"like_regex": LIKE_REGEX,
	"flag":       FLAG,
	"starts":     STARTS,
	"with":       WITH,
	"is":         IS,
	"unknown":    UNKNOWN,
}

var funcs = map[string]int{
	"type":     FUNC_TYPE,
	"size":     FUNC_SIZE,
	"double":   FUNC_DOUBLE,
	"ceiling":  FUNC_CEILING,
	"floor":    FUNC_FLOOR,
	"abs":      FUNC_ABS,
	"datetime": FUNC_DATETIME,
	"keyvalue": FUNC_KEYVALUE,
}

var invertedKeywords = map[int]string{}

func init() {
	for k, v := range keywords {
		invertedKeywords[v] = k
	}
	for k, v := range funcs {
		invertedKeywords[v] = k
	}
}

func keywordState(l *lexer) stateFn {
	for validIdentifierChar(l.peek()) {
		l.advance(1)
	}
	if sym, ok := keywords[l.current()]; ok {
		l.emit(keyword{sym})
	} else {
		l.err("unrecognized keyword \"%s\"", l.current())
	}
	return startState
}

func identifierState(l *lexer) stateFn {
	for validIdentifierChar(l.peek()) {
		l.advance(1)
	}
	l.emit(ident{l.current()})
	return startState
}

func identifierFollowingDotState(l *lexer) stateFn {
	for validIdentifierChar(l.peek()) {
		l.advance(1)
	}
	for unicode.IsSpace(l.peek()) {
		l.advance(1)
	}
	if l.peek() != '(' {
		l.emit(ident{l.current()})
	} else {
		name := strings.TrimRight(l.current(), " ")
		if n, ok := funcs[name]; ok {
			l.emit(keyword{n})
		} else {
			l.err("invalid function \"%s\"", name)
		}
	}
	return startState
}

type tokenStream struct {
	expr  jsonPathNode
	items chan jsonpathSym
	err   error
	lexer *lexer
}

func (t *tokenStream) Lex(lval *yySymType) int {
	next := <-t.items
	if next == nil {
		return 0
	}
	switch n := next.(type) {
	case singleCh:
		return n.ch
	case number:
		lval.val = NumberExpr{val: n.val}
	case ident:
		lval.str = n.val
	case str:
		lval.str = n.val
	}
	return next.identifier()
}

func (t *tokenStream) Error(e string) {
	t.err = fmt.Errorf(e)
}

func tokens(input string) *tokenStream {
	lexer, items := lex(input)
	return &tokenStream{
		lexer: lexer,
		items: items,
	}
}

func lex(input string) (*lexer, chan jsonpathSym) {
	c := make(chan jsonpathSym)
	l := lexer{
		input: []rune(input),
		start: 0,
		pos:   0,
		items: c,
	}
	go l.run()
	return &l, c
}
