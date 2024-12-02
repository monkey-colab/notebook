package main

import (
	"errors"
	"fmt"
	"strings"
)

type Triple struct {
	Subject   string
	Predicate string
	Object    string
}

// Parser represents the Turtle parser with prefix and base management.
type Parser struct {
	lexer    *Lexer
	tokens   []Token
	pos      int
	prefixes map[string]string // Map of prefixes to IRI namespaces
	base     string            // Current base IRI
	triples  []Triple          // Parsed triples
}

// NewParser creates a new Parser instance.
func NewParser(input string) *Parser {
	return &Parser{
		lexer:    NewLexer(input),
		prefixes: make(map[string]string),
		base:     "",
	}
}

// Parse parses the Turtle input into RDF triples.
func (p *Parser) Parse() ([]Triple, error) {
	for {
		token := p.nextToken()
		if token.Type == TokenEOF {
			break
		}

		// Handle directives like @prefix and @base
		if token.Type == TokenPrefix && token.Value == "@prefix" {
			if err := p.parsePrefix(); err != nil {
				return nil, err
			}
			continue
		}
		if token.Type == TokenPrefix && token.Value == "@base" {
			if err := p.parseBase(); err != nil {
				return nil, err
			}
			continue
		}

		p.backup()
		triple, err := p.parseTriple()
		if err != nil {
			return nil, err
		}
		p.triples = append(p.triples, triple)
	}
	return p.triples, nil
}

// parsePrefix handles @prefix declarations.
func (p *Parser) parsePrefix() error {
	prefixToken := p.nextToken()
	if prefixToken.Type != TokenPrefix {
		return errors.New("expected prefix declaration")
	}
	iriToken := p.nextToken()
	if iriToken.Type != TokenIRI {
		return errors.New("expected IRI after prefix declaration")
	}
	if endToken := p.nextToken(); endToken.Type != TokenDot {
		return errors.New("expected '.' after prefix declaration")
	}

	prefix := strings.TrimSuffix(prefixToken.Value, ":")
	p.prefixes[prefix] = iriToken.Value
	return nil
}

// parseBase handles @base declarations.
func (p *Parser) parseBase() error {
	iriToken := p.nextToken()
	if iriToken.Type != TokenIRI {
		return errors.New("expected IRI after @base")
	}
	if endToken := p.nextToken(); endToken.Type != TokenDot {
		return errors.New("expected '.' after @base declaration")
	}
	p.base = iriToken.Value
	return nil
}

// parseTriple parses a single Turtle triple.
func (p *Parser) parseTriple() (Triple, error) {
	subject, err := p.parseSubject()
	if err != nil {
		return Triple{}, err
	}

	predicate, err := p.parsePredicate()
	if err != nil {
		return Triple{}, err
	}

	object, err := p.parseObject()
	if err != nil {
		return Triple{}, err
	}

	// Expect a dot to terminate the triple.
	if token := p.nextToken(); token.Type != TokenDot {
		return Triple{}, fmt.Errorf("expected '.', got %v", token)
	}

	return Triple{Subject: subject, Predicate: predicate, Object: object}, nil
}

// parseSubject parses the subject of a triple.
func (p *Parser) parseSubject() (string, error) {
	token := p.nextToken()
	switch token.Type {
	case TokenIRI:
		return p.expandIRI(token.Value), nil
	case TokenBlankNode:
		return token.Value, nil
	default:
		return "", fmt.Errorf("invalid subject: %v", token)
	}
}

// parsePredicate parses the predicate of a triple.
func (p *Parser) parsePredicate() (string, error) {
	token := p.nextToken()
	switch token.Type {
	case TokenIRI:
		return p.expandIRI(token.Value), nil
	case TokenA:
		return "rdf:type", nil // Shorthand for rdf:type
	default:
		return "", fmt.Errorf("invalid predicate: %v", token)
	}
}

// parseObject parses the object of a triple.
func (p *Parser) parseObject() (string, error) {
	token := p.nextToken()
	switch token.Type {
	case TokenIRI:
		return p.expandIRI(token.Value), nil
	case TokenLiteral:
		return token.Value, nil
	case TokenBlankNode:
		return token.Value, nil
	case TokenCollectionStart:
		return p.parseCollection()
	default:
		return "", fmt.Errorf("invalid object: %v", token)
	}
}

// parseCollection parses a collection as an object.
func (p *Parser) parseCollection() (string, error) {
	var items []string
	for {
		token := p.nextToken()
		if token.Type == TokenCollectionEnd {
			break
		}
		p.backup()
		item, err := p.parseObject()
		if err != nil {
			return "", err
		}
		items = append(items, item)
	}
	return fmt.Sprintf("(%v)", strings.Join(items, " ")), nil
}

// expandIRI expands a prefixed IRI or applies the base IRI.
func (p *Parser) expandIRI(value string) string {
	if strings.HasPrefix(value, ":") {
		return p.base + value[1:]
	}
	if strings.Contains(value, ":") {
		parts := strings.SplitN(value, ":", 2)
		prefix, local := parts[0], parts[1]
		if iri, ok := p.prefixes[prefix]; ok {
			return iri + local
		}
	}
	return value // Already an absolute IRI
}

// nextToken fetches the next token from the lexer.
func (p *Parser) nextToken() Token {
	if p.pos < len(p.tokens) {
		token := p.tokens[p.pos]
		p.pos++
		return token
	}

	token := p.lexer.Lex()
	p.tokens = append(p.tokens, token)
	p.pos++
	return token
}

// backup steps back one token.
func (p *Parser) backup() {
	if p.pos > 0 {
		p.pos--
	}
}
