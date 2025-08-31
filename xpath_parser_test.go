package xmldom

import (
	"strings"
	"testing"
)

func TestXPathLexer(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []XPathToken
	}{
		{
			name:  "Simple path",
			input: "/root",
			expected: []XPathToken{
				{Type: TokenSlash, Value: "/", Position: 0},
				{Type: TokenName, Value: "root", Position: 1},
				{Type: TokenEOF, Value: "", Position: 5},
			},
		},
		{
			name:  "Descendant path",
			input: "//element",
			expected: []XPathToken{
				{Type: TokenDoubleSlash, Value: "//", Position: 0},
				{Type: TokenName, Value: "element", Position: 2},
				{Type: TokenEOF, Value: "", Position: 9},
			},
		},
		{
			name:  "Attribute selection",
			input: "@id",
			expected: []XPathToken{
				{Type: TokenAt, Value: "@", Position: 0},
				{Type: TokenName, Value: "id", Position: 1},
				{Type: TokenEOF, Value: "", Position: 3},
			},
		},
		{
			name:  "Predicate with number",
			input: "item[1]",
			expected: []XPathToken{
				{Type: TokenName, Value: "item", Position: 0},
				{Type: TokenLeftBracket, Value: "[", Position: 4},
				{Type: TokenNumber, Value: "1", Position: 5},
				{Type: TokenRightBracket, Value: "]", Position: 6},
				{Type: TokenEOF, Value: "", Position: 7},
			},
		},
		{
			name:  "String literal",
			input: "'hello world'",
			expected: []XPathToken{
				{Type: TokenString, Value: "hello world", Position: 0},
				{Type: TokenEOF, Value: "", Position: 13},
			},
		},
		{
			name:  "Comparison operators",
			input: "a = b != c",
			expected: []XPathToken{
				{Type: TokenName, Value: "a", Position: 0},
				{Type: TokenEq, Value: "=", Position: 2},
				{Type: TokenName, Value: "b", Position: 4},
				{Type: TokenNeq, Value: "!=", Position: 6},
				{Type: TokenName, Value: "c", Position: 9},
				{Type: TokenEOF, Value: "", Position: 10},
			},
		},
		{
			name:  "Axis specifier",
			input: "child::element",
			expected: []XPathToken{
				{Type: TokenAxis, Value: "child::", Position: 0},
				{Type: TokenName, Value: "element", Position: 7},
				{Type: TokenEOF, Value: "", Position: 14},
			},
		},
		{
			name:  "Function call",
			input: "count(//item)",
			expected: []XPathToken{
				{Type: TokenName, Value: "count", Position: 0},
				{Type: TokenLeftParen, Value: "(", Position: 5},
				{Type: TokenDoubleSlash, Value: "//", Position: 6},
				{Type: TokenName, Value: "item", Position: 8},
				{Type: TokenRightParen, Value: ")", Position: 12},
				{Type: TokenEOF, Value: "", Position: 13},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lexer := NewXPathLexer(tc.input)
			
			for i, expected := range tc.expected {
				token := lexer.NextToken()
				if token.Type != expected.Type {
					t.Errorf("Token %d: expected type %s, got %s", i, tokenTypeName(expected.Type), tokenTypeName(token.Type))
				}
				if token.Value != expected.Value {
					t.Errorf("Token %d: expected value %q, got %q", i, expected.Value, token.Value)
				}
				if token.Position != expected.Position {
					t.Errorf("Token %d: expected position %d, got %d", i, expected.Position, token.Position)
				}
			}
		})
	}
}

func TestXPathParser(t *testing.T) {
	testCases := []struct {
		name        string
		expression  string
		shouldError bool
		nodeType    XPathNodeType // Expected root node type
	}{
		{
			name:        "Simple element name",
			expression:  "element",
			shouldError: false,
			nodeType:    XPathNodeTypeAxis,
		},
		{
			name:        "Absolute path",
			expression:  "/root",
			shouldError: false,
			nodeType:    XPathNodeTypePath,
		},
		{
			name:        "Descendant path",
			expression:  "//element",
			shouldError: false,
			nodeType:    XPathNodeTypePath,
		},
		{
			name:        "String literal",
			expression:  "'hello'",
			shouldError: false,
			nodeType:    XPathNodeTypeLiteral,
		},
		{
			name:        "Number literal",
			expression:  "42",
			shouldError: false,
			nodeType:    XPathNodeTypeLiteral,
		},
		{
			name:        "Function call",
			expression:  "count(//item)",
			shouldError: false,
			nodeType:    XPathNodeTypeFunction,
		},
		{
			name:        "Binary operation",
			expression:  "a = b",
			shouldError: false,
			nodeType:    XPathNodeTypeBinary,
		},
		{
			name:        "Complex expression",
			expression:  "//book[@id='1']/title",
			shouldError: false,
			nodeType:    XPathNodeTypePath,
		},
		{
			name:        "Empty expression",
			expression:  "",
			shouldError: true,
		},
		{
			name:        "Parenthesized expression",
			expression:  "(1 + 2)",
			shouldError: false,
			nodeType:    XPathNodeTypeBinary,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewXPathParser()
			node, err := parser.Parse(tc.expression)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for expression %q, but got none", tc.expression)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for expression %q: %v", tc.expression, err)
				return
			}

			if node == nil {
				t.Errorf("Expected non-nil node for expression %q", tc.expression)
				return
			}

			if node.Type() != tc.nodeType {
				t.Errorf("Expected node type %d for expression %q, got %d", tc.nodeType, tc.expression, node.Type())
			}
		})
	}
}

func TestXPathParserErrors(t *testing.T) {
	testCases := []struct {
		name       string
		expression string
	}{
		{
			name:       "Unterminated string",
			expression: "'unterminated",
		},
		{
			name:       "Invalid character after bang",
			expression: "!x",
		},
		{
			name:       "Unexpected character",
			expression: "element#invalid",
		},
		{
			name:       "Missing closing parenthesis",
			expression: "count(//item",
		},
		{
			name:       "Missing closing bracket",
			expression: "item[1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewXPathParser()
			_, err := parser.Parse(tc.expression)

			if err == nil {
				t.Errorf("Expected error for invalid expression %q, but got none", tc.expression)
			}
		})
	}
}

func TestXPathNodeTestMatching(t *testing.T) {
	// Create a simple test document
	xmlData := `<root><element>text</element><other/></root>`
	decoder := NewDecoder(strings.NewReader(xmlData))
	doc, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Failed to parse test XML: %v", err)
	}

	rootElement := doc.DocumentElement()
	firstChild := rootElement.FirstChild() // should be <element>

	testCases := []struct {
		name     string
		nodeTest *xpathNodeTest
		node     Node
		expected bool
	}{
		{
			name:     "Element name match",
			nodeTest: &xpathNodeTest{name: "element", nodeType: "element"},
			node:     firstChild,
			expected: true,
		},
		{
			name:     "Element name mismatch",
			nodeTest: &xpathNodeTest{name: "other", nodeType: "element"},
			node:     firstChild,
			expected: false,
		},
		{
			name:     "Wildcard match",
			nodeTest: &xpathNodeTest{name: "*", nodeType: "element"},
			node:     firstChild,
			expected: true,
		},
		{
			name:     "Node() match",
			nodeTest: &xpathNodeTest{name: "node()", nodeType: "node()"},
			node:     firstChild,
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &XPathContext{
				ContextNode: tc.node,
				Document:    doc,
			}
			result := tc.nodeTest.Matches(tc.node, ctx)
			if result != tc.expected {
				t.Errorf("Expected %t for node test %q on node %q, got %t",
					tc.expected, tc.nodeTest.Name(), tc.node.NodeName(), result)
			}
		})
	}
}
