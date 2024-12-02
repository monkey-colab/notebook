package main

import (
	"bufio"
	"fmt"
	"strings"
	"unicode"
)

// TokenType defines the types of tokens in Turtle.
type TokenType int

const (
	TokenError TokenType = iota
	TokenEOF
	TokenIRI
	TokenPrefix
	TokenBase
	TokenLiteral
	TokenBlankNode
	TokenCollectionStart
	TokenCollectionEnd
	TokenComment
	TokenDot
	TokenComma
	TokenSemicolon
	TokenA
	TokenColon // Used for prefixed names
)

// Token represents a single parsed token.
type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

// Lexer represents the Turtle lexer.
type Lexer struct {
	input    *bufio.Reader
	line     int
	col      int
	buffered bool
	buffer   rune
}

// NewLexer creates a new Lexer instance.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input: bufio.NewReader(strings.NewReader(input)),
		line:  1,
		col:   0,
	}
}

// next reads the next character, handling buffering and tracking position.
func (l *Lexer) next() (rune, error) {
	if l.buffered {
		l.buffered = false
		return l.buffer, nil
	}

	ch, _, err := l.input.ReadRune()
	if err != nil {
		return 0, err
	}

	if ch == '\n' {
		l.line++
		l.col = 0
	} else {
		l.col++
	}
	return ch, nil
}

// backup buffers the last read character.
func (l *Lexer) backup(ch rune) {
	l.buffer = ch
	l.buffered = true
	l.col--
}

// Lex lexes the next token from the input.
// Lex lexes the next token from the input.
func (l *Lexer) Lex() Token {
	for {
		ch, err := l.next()
		if err != nil {
			return Token{Type: TokenEOF, Line: l.line, Col: l.col}
		}

		switch {
		case unicode.IsSpace(ch):
			continue
		case ch == '#': // Comment
			return l.lexComment()
		case ch == '.':
			return Token{Type: TokenDot, Value: ".", Line: l.line, Col: l.col}
		case ch == ',':
			return Token{Type: TokenComma, Value: ",", Line: l.line, Col: l.col}
		case ch == ';':
			return Token{Type: TokenSemicolon, Value: ";", Line: l.line, Col: l.col}
		case ch == '(': // Collection start
			return Token{Type: TokenCollectionStart, Value: "(", Line: l.line, Col: l.col}
		case ch == ')': // Collection end
			return Token{Type: TokenCollectionEnd, Value: ")", Line: l.line, Col: l.col}
		case ch == '<': // IRI
			return l.lexIRI()
		case ch == '"': // Literal
			return l.lexLiteral()
		case ch == '_': // Blank node
			return l.lexBlankNode()
		case unicode.IsLetter(ch): // Prefix or word
			l.backup(ch)
			return l.lexWord()
		default:
			return Token{Type: TokenError, Value: fmt.Sprintf("Unexpected character: %q", ch), Line: l.line, Col: l.col}
		}
	}
}

// lexLiteral supports strings, language tags, and data types.
func (l *Lexer) lexLiteral() Token {
	var sb strings.Builder
	for {
		ch, err := l.next()
		if err != nil || ch == '"' {
			break
		}
		sb.WriteRune(ch)
	}

	value := sb.String()
	token := Token{Type: TokenLiteral, Value: value, Line: l.line, Col: l.col}

	// Check for language tag or datatype
	ch, err := l.next()
	if err != nil {
		return token
	}

	if ch == '@' { // Language tag
		langTag, err := l.lexLanguageTag()
		if err != nil {
			return Token{Type: TokenError, Value: "Invalid language tag", Line: l.line, Col: l.col}
		}
		token.Value += "@" + langTag
	} else if ch == '^' { // Datatype
		if next, err := l.next(); err == nil && next == '^' {
			dataType := l.lexIRI()
			if dataType.Type == TokenIRI {
				token.Value += "^^" + dataType.Value
			} else {
				return Token{Type: TokenError, Value: "Invalid datatype", Line: l.line, Col: l.col}
			}
		} else {
			l.backup(ch)
		}
	} else {
		l.backup(ch)
	}

	return token
}

// lexUnquotedLiteral handles booleans and numbers.
func (l *Lexer) lexUnquotedLiteral() Token {
	var sb strings.Builder
	ch, _ := l.next()
	if ch == '-' || ch == '+' {
		sb.WriteRune(ch)
		ch, _ = l.next()
	}
	if !unicode.IsDigit(ch) && ch != 't' && ch != 'f' {
		return Token{Type: TokenError, Value: "Invalid unquoted literal", Line: l.line, Col: l.col}
	}
	l.backup(ch)

	// Parse boolean or numeric literal
	for {
		ch, err := l.next()
		if err != nil || unicode.IsSpace(ch) || strings.ContainsRune(".,;()[]<>", ch) {
			l.backup(ch)
			break
		}
		sb.WriteRune(ch)
	}

	value := sb.String()
	token := Token{Type: TokenLiteral, Value: value, Line: l.line, Col: l.col}

	// Assign datatype
	if value == "true" || value == "false" {
		token.Value += "^^<http://www.w3.org/2001/XMLSchema#boolean>"
	} else if strings.Contains(value, ".") || strings.ContainsAny(value, "eE") {
		token.Value += "^^<http://www.w3.org/2001/XMLSchema#decimal>"
	} else {
		token.Value += "^^<http://www.w3.org/2001/XMLSchema#integer>"
	}
	return token
}

// lexBlankNode handles blank nodes.
func (l *Lexer) lexBlankNode() Token {
	var sb strings.Builder
	sb.WriteRune('_') // Start with "_"
	ch, err := l.next()
	if err != nil || ch != ':' {
		return Token{Type: TokenError, Value: "Invalid blank node syntax", Line: l.line, Col: l.col}
	}
	sb.WriteRune(ch)

	for {
		ch, err := l.next()
		if err != nil || unicode.IsSpace(ch) || strings.ContainsRune(".,;()[]<>", ch) {
			l.backup(ch)
			break
		}
		sb.WriteRune(ch)
	}

	return Token{Type: TokenBlankNode, Value: sb.String(), Line: l.line, Col: l.col}
}

// Add boolean recognition to `lexWord`
func (l *Lexer) lexWord() Token {
	var sb strings.Builder
	for {
		ch, err := l.next()
		if err != nil || unicode.IsSpace(ch) || strings.ContainsRune(".,;()[]<>", ch) {
			l.backup(ch)
			break
		}
		sb.WriteRune(ch)
	}
	word := sb.String()
	if word == "true" || word == "false" {
		return Token{
			Type:  TokenLiteral,
			Value: word + "^^<http://www.w3.org/2001/XMLSchema#boolean>",
			Line:  l.line,
			Col:   l.col,
		}
	}
	return Token{Type: TokenBlankNode, Value: word, Line: l.line, Col: l.col}
}

func (l *Lexer) lexComment() Token {
	var sb strings.Builder
	for {
		ch, err := l.next()
		if err != nil || ch == '\n' {
			break
		}
		sb.WriteRune(ch)
	}
	return Token{Type: TokenComment, Value: sb.String(), Line: l.line, Col: l.col}
}

func (l *Lexer) lexIRI() Token {
	var sb strings.Builder
	for {
		ch, err := l.next()
		if err != nil || ch == '>' {
			break
		}
		sb.WriteRune(ch)
	}
	return Token{Type: TokenIRI, Value: sb.String(), Line: l.line, Col: l.col}
}

// lexLanguageTag lexes a valid language tag (e.g., "en", "en-US").
func (l *Lexer) lexLanguageTag() (string, error) {
	var sb strings.Builder
	for {
		ch, err := l.next()
		if err != nil || unicode.IsSpace(ch) || strings.ContainsRune(".,;()[]<>", ch) {
			l.backup(ch)
			break
		}
		if !unicode.IsLetter(ch) && ch != '-' {
			return "", fmt.Errorf("invalid character in language tag: %c", ch)
		}
		sb.WriteRune(ch)
	}
	return sb.String(), nil
}

func (l *Lexer) lexDirective() Token {
	var sb strings.Builder
	for {
		ch, err := l.next()
		if err != nil || unicode.IsSpace(ch) {
			break
		}
		sb.WriteRune(ch)
	}
	switch sb.String() {
	case "prefix":
		return Token{Type: TokenPrefix, Value: "@prefix", Line: l.line, Col: l.col}
	case "base":
		return Token{Type: TokenBase, Value: "@base", Line: l.line, Col: l.col}
	default:
		return Token{Type: TokenError, Value: sb.String(), Line: l.line, Col: l.col}
	}
}
