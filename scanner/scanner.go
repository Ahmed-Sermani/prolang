package scanner

import (
	"strconv"

	"github.com/Ahmed-Sermani/prolang/parser/expressions"
	"github.com/Ahmed-Sermani/prolang/reporting"
)

const (
	// Single-character tokens.
	LEFT_PAREN = iota
	RIGHT_PAREN
	LEFT_BRACE
	RIGHT_BRACE
	COMMA
	DOT
	MINUS
	PLUS
	SEMICOLON
	SLASH
	STAR

	// One or two character tokens.
	BANG
	BANG_EQUAL
	EQUAL
	EQUAL_EQUAL
	GREATER
	GREATER_EQUAL
	LESS
	LESS_EQUAL

	// Literals.
	IDENTIFIER
	STRING
	NUMBER

	// Keywords.
	AND
	CLASS
	ELSE
	FALSE
	FUNC
	FOR
	IF
	NIL
	OR
	PRINT
	RETURN
	SUPER
	THIS
	TRUE
	LET
	WHILE
	EXTENDS

	EOF
)

// reserved keywords
var keywords = map[string]expressions.TokenType{
	"and":     AND,
	"class":   CLASS,
	"else":    ELSE,
	"false":   FALSE,
	"for":     FOR,
	"func":    FUNC,
	"if":      IF,
	"nil":     NIL,
	"or":      OR,
	"print":   PRINT,
	"return":  RETURN,
	"super":   SUPER,
	"this":    THIS,
	"true":    TRUE,
	"let":     LET,
	"while":   WHILE,
	"extends": EXTENDS,
}

type Scanner struct {
	source string
	tokens []expressions.Token
	// start points to the first character in the lexeme being scanned
	start int
	// current points at the character currently being considered
	current int
	// line tracks what source line current is on so can produce tokens with their location.
	line int
}

func New(source string) *Scanner {
	return &Scanner{
		source: source,
		line:   1,
	}
}

func (scanner *Scanner) ScanTokens() []expressions.Token {
	for !scanner.isAtEnd() {
		scanner.start = scanner.current
		scanner.scanToken()
	}

	// appending EOF token at the end of the scan
	scanner.tokens = append(scanner.tokens, expressions.Token{
		Kind:    EOF,
		Lexeme:  "",
		Literal: expressions.Literal{Value: nil},
		Line:    scanner.line,
	})
	return scanner.tokens
}

func (scanner *Scanner) isAtEnd() bool {
	return scanner.current >= len(scanner.source)
}

func (scanner *Scanner) scanToken() {

	switch c := scanner.advance(); c {
	case '(':
		scanner.addToken(LEFT_PAREN, expressions.Literal{Value: nil})
	case ')':
		scanner.addToken(RIGHT_PAREN, expressions.Literal{Value: nil})
	case '{':
		scanner.addToken(LEFT_BRACE, expressions.Literal{Value: nil})
	case '}':
		scanner.addToken(RIGHT_BRACE, expressions.Literal{Value: nil})
	case ',':
		scanner.addToken(COMMA, expressions.Literal{Value: nil})
	case '.':
		scanner.addToken(DOT, expressions.Literal{Value: nil})
	case '-':
		scanner.addToken(MINUS, expressions.Literal{Value: nil})
	case '+':
		scanner.addToken(PLUS, expressions.Literal{Value: nil})
	case ';':
		scanner.addToken(SEMICOLON, expressions.Literal{Value: nil})
	case '*':
		scanner.addToken(STAR, expressions.Literal{Value: nil})

	// two stages match
	case '!':
		if scanner.match('=') {
			scanner.addToken(BANG_EQUAL, expressions.Literal{Value: nil})
		} else {
			scanner.addToken(BANG, expressions.Literal{Value: nil})
		}
	case '=':
		if scanner.match('=') {
			scanner.addToken(EQUAL_EQUAL, expressions.Literal{Value: nil})
		} else {
			scanner.addToken(EQUAL, expressions.Literal{Value: nil})
		}
	case '<':
		if scanner.match('=') {
			scanner.addToken(LESS_EQUAL, expressions.Literal{Value: nil})
		} else {
			scanner.addToken(LESS, expressions.Literal{Value: nil})
		}
	case '>':
		if scanner.match('=') {
			scanner.addToken(GREATER_EQUAL, expressions.Literal{Value: nil})
		} else {
			scanner.addToken(GREATER, expressions.Literal{Value: nil})
		}

	case '/':

		if scanner.match('/') {
			// comment goes until the end of the line
			for scanner.peek() != '\n' && !scanner.isAtEnd() {
				// consume the comment
				scanner.advance()
			}
		} else {
			scanner.addToken(SLASH, expressions.Literal{Value: nil})
		}

	// ignored bytes
	case ' ':
	case '\t':
	case '\r':

	case '\n':
		scanner.line++

	// literals
	case '"':
		scanner.parseStringLiteral()
	default:
		// matching digits
		if isDigit(c) {
			scanner.parseNumberLiteral()
			// matching identifers
		} else if isAlpha(c) {
			scanner.parseIdentifer()
		} else {
			reporting.ReportError(scanner.line, "Unexpected character.")
		}
	}

}

// consumes the next character in the source and returns it.
func (scanner *Scanner) advance() byte {
	c := scanner.source[scanner.current]
	scanner.current++
	return c
}

// peeks on the current byte without incrementing current
func (scanner *Scanner) peek() byte {
	if scanner.isAtEnd() {
		return '\x00'
	}
	return scanner.source[scanner.current]
}

// grabs the text of the current lexeme and creates a new token for it.
func (scanner *Scanner) addToken(kind expressions.TokenType, literal expressions.Literal) {
	text := scanner.source[scanner.start:scanner.current]
	scanner.tokens = append(scanner.tokens, expressions.Token{
		Kind:    kind,
		Lexeme:  text,
		Literal: literal,
		Line:    scanner.line,
	})
}

// matches the current character if the match succeeded the incremant current
func (scanner *Scanner) match(expected byte) bool {
	if scanner.isAtEnd() {
		return false
	}

	if scanner.source[scanner.current] != expected {
		return false
	}

	scanner.current++
	return true
}

// peaks on the next byte
func (scanner *Scanner) lookahead() byte {
	if scanner.current+1 >= len(scanner.source) {
		return '\x00'
	}
	return scanner.source[scanner.current+1]
}

// parses string literal and add its token. supports multiline strings
func (scanner *Scanner) parseStringLiteral() {
	for scanner.peek() != '"' && !scanner.isAtEnd() {
		if scanner.peek() == '\n' {
			scanner.line++
		}
		scanner.advance()
	}

	if scanner.isAtEnd() {
		reporting.ReportError(scanner.line, "Unterminated string.")
		return
	}

	// consume the closing "
	scanner.advance()

	// matching the string without the beginnering the closing "
	value := scanner.source[scanner.start+1 : scanner.current-1]
	scanner.addToken(STRING, expressions.Literal{Value: value})
}

// consume as many digits as it can find for the integer part of the literal.
// Then it look for a fractional part, which is a decimal point (.) followed by at least one digit.
// If the fractional part exits, it consume as many digits as we can find.
func (scanner *Scanner) parseNumberLiteral() {
	for isDigit(scanner.peek()) {
		// consume numbers
		scanner.advance()
	}

	// handle the fractional part
	if scanner.peek() == '.' && isDigit(scanner.lookahead()) {
		// consume .
		scanner.advance()

		for isDigit(scanner.peek()) {
			// consume fractions
			scanner.advance()
		}
	}

	// parse the number as float64
	value, err := strconv.ParseFloat(scanner.source[scanner.start:scanner.current], 64)
	if err != nil {
		reporting.ReportError(scanner.line, "Invalid Float Value")
	}
	scanner.addToken(NUMBER, expressions.Literal{Value: value})
}

func (scanner *Scanner) parseIdentifer() {
	for isAlphaNumeric(scanner.peek()) {
		scanner.advance()
	}

	// after we scan an identifier, check to see if it matches anything of the keywords.
	// otherwise it's an identifer
	identifer := scanner.source[scanner.start:scanner.current]
	kind, ok := keywords[identifer]
	if !ok {
		kind = IDENTIFIER
	}
	scanner.addToken(kind, expressions.Literal{Value: nil})
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isAlphaNumeric(c byte) bool {
	return isAlpha(c) || isDigit(c)
}
