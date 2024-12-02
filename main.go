package main

import (
	"fmt"
	"log"
)

func main() {
	data := `
@prefix ex: <http://example.org/> .
@prefix xsd: <http://www.w3.org/2001/XMLSchema#> .

# This is a simple RDF example with nested lists, blank nodes, and datatypes
ex:subject ex:predicate _:blankNode1 .

# Example of a blank node with nested lists
_:blankNode1 ex:hasValue (
    "Nested Value 1"^^xsd:string
    _:blankNode2
    ( "Inner List"^^xsd:string _:blankNode3 )
) .

# Another blank node
_:blankNode2 ex:hasValue (
    "Another value"^^xsd:string
) .

# Boolean literals and numeric literals
ex:subject ex:hasStatus true .
ex:subject ex:hasAge 25 .

# Literal with datatype
ex:subject ex:hasName "John Doe"^^xsd:string .

# More complex nested structures with datatypes and blank nodes
ex:object ex:hasDetails (
    _:blankNode4
    "Some literal"^^xsd:int
    _:blankNode5
    (
        _:blankNode6
        "Deep nested value"^^xsd:decimal
    )
) .

# More blank nodes with properties
_:blankNode4 ex:hasProp "Value for BlankNode4" .
_:blankNode5 ex:hasProp "Value for BlankNode5" .

# Boolean example with nested structure
ex:anotherSubject ex:hasStatus false .
	`
	lexer := NewLexer(data)
	for {
		token := lexer.Lex()
		if token.Type == TokenEOF {
			break
		}
		fmt.Printf("Token: %+v\n", token)
	}

	parser := NewParser(data)
	triples, err := parser.Parse()
	if err != nil {
		log.Fatalf("Error parsing Turtle: %v", err)
	}

	for _, triple := range triples {
		fmt.Printf("Triple: %+v\n", triple)
	}
}
