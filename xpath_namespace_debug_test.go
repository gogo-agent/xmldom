package xmldom

import (
	"strings"
	"testing"
)

func TestXPathNamespaceAxisDebug(t *testing.T) {
	xml := `<?xml version="1.0"?>
	<root xmlns:foo="http://example.com/foo" xmlns:bar="http://example.com/bar">
		<child>Text</child>
	</root>`

	decoder := NewDecoder(strings.NewReader(xml))
	doc, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	// Test parsing the namespace axis expression
	parser := NewXPathParser()
	ast, err := parser.Parse("/root/namespace::*")
	if err != nil {
		t.Fatalf("Failed to parse XPath expression: %v", err)
	}

	t.Logf("AST created successfully: %T", ast)

	// Create XPath expression
	expr, err := doc.CreateExpression("/root/namespace::*", nil)
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

	t.Logf("Namespace nodes found: %d", length)

	// We expect at least 3 (xml, foo, bar)
	if length < 3 {
		t.Errorf("Expected at least 3 namespace nodes, got %d", length)
	}

	// List all namespace nodes
	for i := uint32(0); i < length; i++ {
		node, err := result.SnapshotItem(i)
		if err != nil {
			t.Fatalf("Failed to get snapshot item %d: %v", i, err)
		}
		if node != nil {
			t.Logf("Namespace %d: prefix='%s', value='%s'", i, node.NodeName(), node.NodeValue())
		}
	}
}

// Test that namespace nodes can be matched by name
func TestXPathNamespaceAxisByName(t *testing.T) {
	xml := `<?xml version="1.0"?>
	<doc xmlns:test="http://test.com">
		<elem/>
	</doc>`

	decoder := NewDecoder(strings.NewReader(xml))
	doc, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	// Test matching specific namespace by name
	expr, err := doc.CreateExpression("//elem/namespace::test", nil)
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

	t.Logf("Namespace nodes matching 'test': %d", length)

	if length != 1 {
		t.Errorf("Expected 1 namespace node matching 'test', got %d", length)
	}
}
