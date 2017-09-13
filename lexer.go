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

type tru struct{}
type fals struct{}
type null struct{}
type strict struct{}
type lax struct{}
type last struct{}
type to struct{}
type exists struct{}
type likeRegex struct{}
type flag struct{}
type starts struct{}
type with struct{}
type is struct{}
type unknown struct{}

type funcType struct{}
type funcSize struct{}
type funcDouble struct{}
type funcCeiling struct{}
type funcFloor struct{}
type funcAbs struct{}
type funcDatetime struct{}
type funcKeyvalue struct{}

type errSym struct{ msg string }

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

func (s tru) Lexeme() string       { return "true" }
func (s fals) Lexeme() string      { return "false" }
func (s null) Lexeme() string      { return "null" }
func (s strict) Lexeme() string    { return "strict" }
func (s lax) Lexeme() string       { return "lax" }
func (s last) Lexeme() string      { return "last" }
func (s to) Lexeme() string        { return "to" }
func (s exists) Lexeme() string    { return "exists" }
func (s likeRegex) Lexeme() string { return "like_regex" }
func (s flag) Lexeme() string      { return "flag" }
func (s starts) Lexeme() string    { return "starts" }
func (s with) Lexeme() string      { return "with" }
func (s is) Lexeme() string        { return "is" }
func (s unknown) Lexeme() string   { return "unknown" }

func (s funcType) Lexeme() string     { return "type" }
func (s funcSize) Lexeme() string     { return "size" }
func (s funcDouble) Lexeme() string   { return "double" }
func (s funcCeiling) Lexeme() string  { return "ceiling" }
func (s funcFloor) Lexeme() string    { return "floor" }
func (s funcAbs) Lexeme() string      { return "abs" }
func (s funcDatetime) Lexeme() string { return "datetime" }
func (s funcKeyvalue) Lexeme() string { return "keyvalue" }

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

func (s tru) identifier() int       { return TRUE }
func (s fals) identifier() int      { return FALSE }
func (s null) identifier() int      { return NULL }
func (s strict) identifier() int    { return STRICT }
func (s lax) identifier() int       { return LAX }
func (s last) identifier() int      { return LAST }
func (s to) identifier() int        { return TO }
func (s exists) identifier() int    { return EXISTS }
func (s likeRegex) identifier() int { return LIKE_REGEX }
func (s flag) identifier() int      { return FLAG }
func (s starts) identifier() int    { return STARTS }
func (s with) identifier() int      { return WITH }
func (s is) identifier() int        { return IS }
func (s unknown) identifier() int   { return UNKNOWN }

func (s funcType) identifier() int     { return FUNC_TYPE }
func (s funcSize) identifier() int     { return FUNC_SIZE }
func (s funcDouble) identifier() int   { return FUNC_DOUBLE }
func (s funcCeiling) identifier() int  { return FUNC_CEILING }
func (s funcFloor) identifier() int    { return FUNC_FLOOR }
func (s funcAbs) identifier() int      { return FUNC_ABS }
func (s funcDatetime) identifier() int { return FUNC_DATETIME }
func (s funcKeyvalue) identifier() int { return FUNC_KEYVALUE }

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
	l.emit(errSym{fmt.Sprintf(msg, args...)})
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
	l.gobble(1)
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

var keywords = map[string]jsonpathSym{
	"true":       tru{},
	"false":      fals{},
	"null":       null{},
	"strict":     strict{},
	"lax":        lax{},
	"last":       last{},
	"to":         to{},
	"exists":     exists{},
	"like_regex": likeRegex{},
	"flag":       flag{},
	"starts":     starts{},
	"with":       with{},
	"is":         is{},
	"unknown":    unknown{},
}

func keywordState(l *lexer) stateFn {
	for validIdentifierChar(l.peek()) {
		l.advance(1)
	}
	if sym, ok := keywords[l.current()]; ok {
		l.emit(sym)
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

var funcs = map[string]jsonpathSym{
	"type":     funcType{},
	"size":     funcSize{},
	"double":   funcDouble{},
	"ceiling":  funcCeiling{},
	"floor":    funcFloor{},
	"abs":      funcAbs{},
	"datetime": funcDatetime{},
	"keyvalue": funcKeyvalue{},
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
			l.emit(n)
		} else {
			l.err("invalid function \"%s\"", name)
		}
	}
	return startState
}

type tokenStream struct {
	expr  jsonPathNode
	items chan jsonpathSym
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
		lval.val = Number{val: n.val}
	case ident:
		lval.str = n.val
	case str:
		lval.str = n.val
	}
	return next.identifier()
}

func (t *tokenStream) Error(e string) {
	panic(e)
}

func tokens(input string) *tokenStream {
	return &tokenStream{
		items: lex(input),
	}
}

func lex(input string) chan jsonpathSym {
	c := make(chan jsonpathSym)
	l := lexer{
		input: []rune(input),
		start: 0,
		pos:   0,
		items: c,
	}
	go l.run()
	return c
}
