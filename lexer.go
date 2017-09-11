package jsonpath

import (
	"fmt"
	"strconv"
	"unicode"
)

// https://talks.golang.org/2011/lex.slide

type lexer struct {
	input []rune
	start int
	pos   int
	items chan jsonpathSym
}

type jsonpathSym interface {
	Lexeme() string
	identifier() int
}

type dollar struct{}
type dot struct{}
type leftsq struct{}
type rightsq struct{}
type leftparen struct{}
type rightparen struct{}
type star struct{}
type at struct{}
type word struct{ val string }
type number struct{ val float64 }
type str struct{ val string }

type lt struct{}
type lte struct{}
type gt struct{}
type gte struct{}
type eq struct{}
type and struct{}
type or struct{}
type plus struct{}
type minus struct{}
type times struct{}
type div struct{}

type errSym struct{ msg string }

func (s dollar) Lexeme() string     { return "$" }
func (s dot) Lexeme() string        { return "." }
func (s leftsq) Lexeme() string     { return "[" }
func (s rightsq) Lexeme() string    { return "]" }
func (s leftparen) Lexeme() string  { return "(" }
func (s rightparen) Lexeme() string { return ")" }
func (s star) Lexeme() string       { return "*" }
func (s at) Lexeme() string         { return "@" }
func (s word) Lexeme() string       { return s.val }
func (s number) Lexeme() string     { return fmt.Sprintf("%v", s.val) }
func (s str) Lexeme() string        { return fmt.Sprintf("'%v'", s.val) }

func (s lt) Lexeme() string    { return "<" }
func (s lte) Lexeme() string   { return "<=" }
func (s gt) Lexeme() string    { return ">" }
func (s gte) Lexeme() string   { return ">=" }
func (s eq) Lexeme() string    { return "==" }
func (s and) Lexeme() string   { return "&&" }
func (s or) Lexeme() string    { return "||" }
func (s plus) Lexeme() string  { return "+" }
func (s minus) Lexeme() string { return "-" }
func (s times) Lexeme() string { return "*" }
func (s div) Lexeme() string   { return "/" }

func (s errSym) Lexeme() string { return s.msg }

func (s dollar) identifier() int     { return DOLLAR }
func (s dot) identifier() int        { return DOT }
func (s leftsq) identifier() int     { return LEFTSQ }
func (s rightsq) identifier() int    { return RIGHTSQ }
func (s leftparen) identifier() int  { return LEFTPAREN }
func (s rightparen) identifier() int { return RIGHTPAREN }
func (s star) identifier() int       { return STAR }
func (s at) identifier() int         { return AT }
func (s word) identifier() int       { return WORD }
func (s number) identifier() int     { return NUMBER }
func (s str) identifier() int        { return STR }

func (s lt) identifier() int    { return LT }
func (s lte) identifier() int   { return LTE }
func (s gt) identifier() int    { return GT }
func (s gte) identifier() int   { return GTE }
func (s eq) identifier() int    { return EQ }
func (s and) identifier() int   { return AND }
func (s or) identifier() int    { return OR }
func (s plus) identifier() int  { return PLUS }
func (s minus) identifier() int { return MINUS }
func (s times) identifier() int { return TIMES }
func (s div) identifier() int   { return DIV }

func (s errSym) identifier() int { return -1 }

const eof = 0

func validFirstIdentifierChar(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func validIdentifierChar(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_'
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
	l.items <- sym
}

func (l *lexer) err(msg string) {
	l.emit(errSym{msg})
}

func (l *lexer) current() string {
	return string(l.input[l.start:l.pos])
}

type stateFn func(*lexer) stateFn

var singleChars = map[rune]jsonpathSym{
	'$': dollar{},
	'.': dot{},
	'@': at{},
	'*': star{},
	'[': leftsq{},
	']': rightsq{},
	'(': leftparen{},
	')': rightparen{},
}

var opChars = map[rune]struct{}{
	'<': struct{}{},
	'>': struct{}{},
	'|': struct{}{},
	'&': struct{}{},
	'=': struct{}{},
	'+': struct{}{},
	'-': struct{}{},
}

func startState(l *lexer) stateFn {
	for unicode.IsSpace(l.peek()) {
		l.gobble(1)
	}
	ch := l.peek()

	sym, ok := singleChars[ch]
	if ok {
		l.advance(1)
		l.emit(sym)
		return startState
	}

	if _, ok := opChars[ch]; ok {
		return opState
	}

	switch {
	case validFirstIdentifierChar(ch):
		return identifierState
	case unicode.IsDigit(ch):
		return numberState
	case ch == '\'':
		parseString(l, '\'')
		return startState
	case ch == '"':
		parseString(l, '"')
		return startState
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
				l.err("invalid escape sequence")
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
	parsed, err := strconv.ParseFloat(l.current(), 64)
	if err != nil {
		panic("PANIC!!!")
	}
	l.emit(number{parsed})
	return startState
}

func opState(l *lexer) stateFn {
	switch l.peek() {
	case '<':
		l.advance(1)
		if l.peek() == '=' {
			l.advance(1)
			l.emit(lte{})
		} else {
			l.emit(lt{})
		}
	case '>':
		l.advance(1)
		if l.peek() == '=' {
			l.advance(1)
			l.emit(gte{})
		} else {
			l.emit(gt{})
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
	case '+':
		l.advance(1)
		l.emit(plus{})
	case '-':
		l.advance(1)
		l.emit(plus{})
	case '*':
		l.advance(1)
		l.emit(plus{})
	case '/':
		l.advance(1)
		l.emit(plus{})
	}
	return startState
}

func identifierState(l *lexer) stateFn {
	for validIdentifierChar(l.peek()) {
		l.advance(1)
	}
	l.emit(word{l.current()})
	return startState
}

type tokenStream struct {
	items chan jsonpathSym
}

func (t *tokenStream) Lex(lval *yySymType) int {
	next := <-t.items
	if next == nil {
		return yyEofCode
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
