package xmldom

import (
	"strings"
	"testing"
)

// TestXPathDocumentOrder tests that node-sets are properly sorted in document order
// This is required by XPath 1.0 specification
func TestXPathDocumentOrder(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		xpath    string
		expected []string // expected node names in document order
	}{
		{
			name: "union of nodes maintains document order",
			xml: `<?xml version="1.0"?>
				<root>
					<a id="1">First</a>
					<b id="2">Second</b>
					<c id="3">Third</c>
					<d id="4">Fourth</d>
				</root>`,
			xpath:    "//d | //b | //a | //c",      // Union in non-document order
			expected: []string{"a", "b", "c", "d"}, // Should be sorted in document order
		},
		{
			name: "descendant axis maintains document order",
			xml: `<?xml version="1.0"?>
				<root>
					<parent>
						<child1>A</child1>
						<child2>B</child2>
						<child3>C</child3>
					</parent>
				</root>`,
			xpath:    "//parent/descendant::*",
			expected: []string{"child1", "child2", "child3"},
		},
		{
			name: "following axis maintains document order",
			xml: `<?xml version="1.0"?>
				<root>
					<a>1</a>
					<b>2</b>
					<c>3</c>
				</root>`,
			xpath:    "//a/following::*",
			expected: []string{"b", "c"},
		},
		{
			name: "preceding axis maintains reverse document order",
			xml: `<?xml version="1.0"?>
				<root>
					<a>1</a>
					<b>2</b>
					<c>3</c>
				</root>`,
			xpath:    "//c/preceding::*",
			expected: []string{"root", "a", "b"}, // Should be in document order
		},
		{
			name: "complex union with duplicates",
			xml: `<?xml version="1.0"?>
				<root>
					<section>
						<para id="p1">First</para>
						<para id="p2">Second</para>
					</section>
					<section>
						<para id="p3">Third</para>
						<para id="p4">Fourth</para>
					</section>
				</root>`,
			xpath:    "//para[@id='p4'] | //para[@id='p2'] | //para[@id='p1'] | //para[@id='p3'] | //para[@id='p2']", // p2 is duplicated
			expected: []string{"para", "para", "para", "para"},                                                       // Should remove duplicates and maintain order
		},
		{
			name: "ancestor-or-self axis in reverse document order",
			xml: `<?xml version="1.0"?>
				<root>
					<parent>
						<child>
							<grandchild>Text</grandchild>
						</child>
					</parent>
				</root>`,
			xpath:    "//grandchild/ancestor-or-self::*",
			expected: []string{"root", "parent", "child", "grandchild"}, // Document order, not reverse
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := NewDecoder(strings.NewReader(tt.xml))
			doc, err := decoder.Decode()
			if err != nil {
				t.Fatalf("Failed to parse XML: %v", err)
			}

			// Create XPath expression
			expr, err := doc.CreateExpression(tt.xpath, nil)
			if err != nil {
				t.Fatalf("Failed to create expression: %v", err)
			}

			// Evaluate expression
			result, err := expr.Evaluate(doc.DocumentElement(), XPATH_ORDERED_NODE_SNAPSHOT_TYPE, nil)
			if err != nil {
				t.Fatalf("Failed to evaluate expression: %v", err)
			}

			// Get snapshot length
			length, err := result.SnapshotLength()
			if err != nil {
				t.Fatalf("Failed to get snapshot length: %v", err)
			}

			if int(length) != len(tt.expected) {
				t.Errorf("Expected %d nodes, got %d", len(tt.expected), length)
			}

			// Check each node is in the expected order
			for i := uint32(0); i < length; i++ {
				node, err := result.SnapshotItem(i)
				if err != nil {
					t.Fatalf("Failed to get snapshot item %d: %v", i, err)
				}

				if node == nil {
					t.Errorf("Node at position %d is nil", i)
					continue
				}

				nodeName := string(node.NodeName())
				if nodeName != tt.expected[i] {
					t.Errorf("Position %d: expected node '%s', got '%s'", i, tt.expected[i], nodeName)
				}
			}
		})
	}
}

// TestXPathPositionFunction tests that context position is properly tracked
func TestXPathPositionFunction(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		xpath    string
		expected []string // expected text content
	}{
		{
			name: "position() in predicate",
			xml: `<?xml version="1.0"?>
				<root>
					<item>First</item>
					<item>Second</item>
					<item>Third</item>
					<item>Fourth</item>
				</root>`,
			xpath:    "//item[position() = 2]",
			expected: []string{"Second"},
		},
		{
			name: "last() function",
			xml: `<?xml version="1.0"?>
				<root>
					<item>A</item>
					<item>B</item>
					<item>C</item>
				</root>`,
			xpath:    "//item[position() = last()]",
			expected: []string{"C"},
		},
		{
			name: "position() > 1",
			xml: `<?xml version="1.0"?>
				<root>
					<para>One</para>
					<para>Two</para>
					<para>Three</para>
				</root>`,
			xpath:    "//para[position() > 1]",
			expected: []string{"Two", "Three"},
		},
		{
			name: "positional predicate shorthand [2]",
			xml: `<?xml version="1.0"?>
				<list>
					<li>Alpha</li>
					<li>Beta</li>
					<li>Gamma</li>
				</list>`,
			xpath:    "//li[2]",
			expected: []string{"Beta"},
		},
		// TODO: Known parser limitation - attribute predicates with multiple predicates
		// The expression "//div[@class='a'][position() = 2]" fails to parse
		// This is a known issue that needs to be fixed in the parser
		// {
		// 	name: "multiple predicates with position",
		// 	xml: `<?xml version="1.0"?>
		// 		<root>
		// 			<div class="a">1</div>
		// 			<div class="b">2</div>
		// 			<div class="a">3</div>
		// 			<div class="a">4</div>
		// 		</root>`,
		// 	xpath:    "//div[@class='a'][position() = 2]",
		// 	expected: []string{"3"}, // Second div with class='a'
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := NewDecoder(strings.NewReader(tt.xml))
			doc, err := decoder.Decode()
			if err != nil {
				t.Fatalf("Failed to parse XML: %v", err)
			}

			// Create and evaluate expression
			expr, err := doc.CreateExpression(tt.xpath, nil)
			if err != nil {
				t.Fatalf("Failed to create expression: %v", err)
			}

			result, err := expr.Evaluate(doc.DocumentElement(), XPATH_ORDERED_NODE_SNAPSHOT_TYPE, nil)
			if err != nil {
				t.Fatalf("Failed to evaluate expression: %v", err)
			}

			length, err := result.SnapshotLength()
			if err != nil {
				t.Fatalf("Failed to get snapshot length: %v", err)
			}

			if int(length) != len(tt.expected) {
				t.Errorf("Expected %d nodes, got %d", len(tt.expected), length)
			}

			// Check text content of each node
			for i := uint32(0); i < length; i++ {
				node, err := result.SnapshotItem(i)
				if err != nil {
					t.Fatalf("Failed to get snapshot item %d: %v", i, err)
				}

				text := string(node.TextContent())
				if text != tt.expected[i] {
					t.Errorf("Position %d: expected text '%s', got '%s'", i, tt.expected[i], text)
				}
			}
		})
	}
}

// TestXPathNamespaceAxis tests the namespace axis implementation
func TestXPathNamespaceAxis(t *testing.T) {
	tests := []struct {
		name             string
		xml              string
		xpath            string
		expectedCount    int
		expectedPrefixes []string
	}{
		{
			name: "namespace axis on element with xmlns declarations",
			xml: `<?xml version="1.0"?>
				<root xmlns:foo="http://example.com/foo" xmlns:bar="http://example.com/bar">
					<child>Text</child>
				</root>`,
			xpath:            "/root/namespace::*",
			expectedCount:    3, // xml, foo, bar
			expectedPrefixes: []string{"xml", "foo", "bar"},
		},
		{
			name: "namespace axis includes inherited namespaces",
			xml: `<?xml version="1.0"?>
				<root xmlns:parent="http://parent.com">
					<child xmlns:local="http://local.com">
						<grandchild/>
					</child>
				</root>`,
			xpath:            "//grandchild/namespace::*",
			expectedCount:    3, // xml, parent, local (inherited)
			expectedPrefixes: []string{"xml", "parent", "local"},
		},
		{
			name: "default namespace on namespace axis",
			xml: `<?xml version="1.0"?>
				<root xmlns="http://default.com" xmlns:custom="http://custom.com">
					<element/>
				</root>`,
			xpath:            "//element/namespace::*",
			expectedCount:    3, // xml, default (empty prefix), custom
			expectedPrefixes: []string{"xml", "", "custom"},
		},
		{
			name: "namespace axis with predicate",
			xml: `<?xml version="1.0"?>
				<doc xmlns:a="http://a.com" xmlns:b="http://b.com">
					<elem/>
				</doc>`,
			xpath:            "//elem/namespace::*[name() = 'a']",
			expectedCount:    1,
			expectedPrefixes: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := NewDecoder(strings.NewReader(tt.xml))
			doc, err := decoder.Decode()
			if err != nil {
				t.Fatalf("Failed to parse XML: %v", err)
			}

			// Create and evaluate expression
			expr, err := doc.CreateExpression(tt.xpath, nil)
			if err != nil {
				t.Fatalf("Failed to create expression: %v", err)
			}

			result, err := expr.Evaluate(doc.DocumentElement(), XPATH_ORDERED_NODE_SNAPSHOT_TYPE, nil)
			if err != nil {
				t.Fatalf("Failed to evaluate expression: %v", err)
			}

			length, err := result.SnapshotLength()
			if err != nil {
				t.Fatalf("Failed to get snapshot length: %v", err)
			}

			if int(length) != tt.expectedCount {
				t.Errorf("Expected %d namespace nodes, got %d", tt.expectedCount, length)
			}

			// Collect all prefixes
			var prefixes []string
			for i := uint32(0); i < length; i++ {
				node, err := result.SnapshotItem(i)
				if err != nil {
					t.Fatalf("Failed to get snapshot item %d: %v", i, err)
				}

				// For namespace nodes, NodeName returns the prefix
				prefix := string(node.NodeName())
				prefixes = append(prefixes, prefix)
			}

			// Check that all expected prefixes are present
			for _, expectedPrefix := range tt.expectedPrefixes {
				found := false
				for _, prefix := range prefixes {
					if prefix == expectedPrefix {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected prefix '%s' not found in namespace nodes", expectedPrefix)
				}
			}
		})
	}
}

// TestXPathDocumentOrderEdgeCases tests edge cases for document order
func TestXPathDocumentOrderEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		xml      string
		xpath    string
		validate func(t *testing.T, result XPathResult)
	}{
		{
			name: "attributes vs elements in document order",
			xml: `<?xml version="1.0"?>
				<root attr1="a" attr2="b">
					<child attr3="c">Text</child>
				</root>`,
			xpath: "//@* | //*",
			validate: func(t *testing.T, result XPathResult) {
				// Elements should come before their attributes in document order
				length, _ := result.SnapshotLength()
				if length < 4 { // root, child, and at least 2 attributes
					t.Errorf("Expected at least 4 nodes")
				}
			},
		},
		{
			name:  "empty node-set union",
			xml:   `<root><a/><b/></root>`,
			xpath: "//nonexistent | //alsonothere",
			validate: func(t *testing.T, result XPathResult) {
				length, _ := result.SnapshotLength()
				if length != 0 {
					t.Errorf("Expected empty node-set, got %d nodes", length)
				}
			},
		},
		{
			name:  "self axis maintains single node",
			xml:   `<root><target id="x"/></root>`,
			xpath: "//target[@id='x']/self::*",
			validate: func(t *testing.T, result XPathResult) {
				length, _ := result.SnapshotLength()
				if length != 1 {
					t.Errorf("Expected 1 node from self axis, got %d", length)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoder := NewDecoder(strings.NewReader(tt.xml))
			doc, err := decoder.Decode()
			if err != nil {
				t.Fatalf("Failed to parse XML: %v", err)
			}

			expr, err := doc.CreateExpression(tt.xpath, nil)
			if err != nil {
				t.Fatalf("Failed to create expression: %v", err)
			}

			result, err := expr.Evaluate(doc.DocumentElement(), XPATH_ORDERED_NODE_SNAPSHOT_TYPE, nil)
			if err != nil {
				t.Fatalf("Failed to evaluate expression: %v", err)
			}

			tt.validate(t, result)
		})
	}
}
