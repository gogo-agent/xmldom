package xmldom_test

import (
	"strings"
	"testing"

	"github.com/gogo-agent/xmldom"
)

func TestDecode_Simple(t *testing.T) {
	xml := `<root><child>text</child></root>`
	decoder := xmldom.NewDecoder(strings.NewReader(xml))
	doc, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}

	if doc == nil {
		t.Fatal("Decode() returned a nil document")
	}

	root := doc.DocumentElement()
	if root == nil {
		t.Fatal("DocumentElement is nil")
	}

	if root.NodeName() != "root" {
		t.Errorf("Expected root node name to be 'root', got '%s'", root.NodeName())
	}

	if root.ChildNodes().Length() != 1 {
		t.Fatalf("Expected root to have 1 child, got %d", root.ChildNodes().Length())
	}

	child := root.ChildNodes().Item(0)
	if child.NodeName() != "child" {
		t.Errorf("Expected child node name to be 'child', got '%s'", child.NodeName())
	}

	if child.ChildNodes().Length() != 1 {
		t.Fatalf("Expected child to have 1 child, got %d", child.ChildNodes().Length())
	}

	text := child.ChildNodes().Item(0)
	if text.NodeType() != xmldom.TEXT_NODE {
		t.Errorf("Expected a text node, got %d", text.NodeType())
	}

	if text.NodeValue() != "text" {
		t.Errorf("Expected text node value to be 'text', got '%s'", text.NodeValue())
	}
}

func TestDecode_RoundTrip(t *testing.T) {
	xmlStr := `<root a="1" b="2"><child>text</child></root>`
	decoder := xmldom.NewDecoder(strings.NewReader(xmlStr))
	doc, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}

	var buf strings.Builder
	encoder := xmldom.NewEncoder(&buf)
	err = encoder.Encode(doc)
	if err != nil {
		t.Fatalf("Encode() failed: %v", err)
	}

	// This is a very basic comparison. A more robust comparison would
	// involve parsing both XMLs and comparing the DOM trees.
	// For now, we just compare the strings.
	// Note that attribute order is not guaranteed to be preserved.
	// A better comparison would be to parse the output and check for equivalence.
	// For this test, we will just check for the presence of the elements and attributes.

	got := buf.String()
	if !strings.Contains(got, "<root") {
		t.Error("Expected to find '<root' in the output")
	}
	if !strings.Contains(got, `a="1"`) {
		t.Error(`Expected to find 'a="1"' in the output`)
	}
	if !strings.Contains(got, `b="2"`) {
		t.Error(`Expected to find 'b="2"' in the output`)
	}
	if !strings.Contains(got, "<child>text</child>") {
		t.Error("Expected to find '<child>text</child>' in the output")
	}
}

func TestDecode_Namespaces(t *testing.T) {
	xmlStr := `<root xmlns="http://example.com/default" xmlns:p="http://example.com/prefixed"><p:child>text</p:child></root>`
	decoder := xmldom.NewDecoder(strings.NewReader(xmlStr))
	doc, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}

	root := doc.DocumentElement()
	if root.NamespaceURI() != "http://example.com/default" {
		t.Errorf("Expected root namespace URI to be 'http://example.com/default', got '%s'", root.NamespaceURI())
	}
	if root.NodeName() != "root" {
		t.Errorf("Expected root node name to be 'root', got '%s'", root.NodeName())
	}

	child := root.FirstChild()
	if child.NodeType() != xmldom.ELEMENT_NODE {
		t.Fatalf("Expected an element node, got %d", child.NodeType())
	}
	elem := child.(xmldom.Element)
	if elem.NamespaceURI() != "http://example.com/prefixed" {
		t.Errorf("Expected child namespace URI to be 'http://example.com/prefixed', got '%s'", elem.NamespaceURI())
	}
	// Note: encoding/xml does not preserve prefixes, so NodeName is just the local name.
	if elem.NodeName() != "child" {
		t.Errorf("Expected child node name to be 'child', got '%s'", elem.NodeName())
	}
	if elem.LocalName() != "child" {
		t.Errorf("Expected child local name to be 'child', got '%s'", elem.LocalName())
	}
}

func TestDecode_Doctype(t *testing.T) {
	xmlStr := `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd"><html></html>`
	decoder := xmldom.NewDecoder(strings.NewReader(xmlStr))
	doc, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}

	doctype := doc.Doctype()
	if doctype == nil {
		t.Fatal("Doctype is nil")
	}

	if doctype.Name() != "html" {
		t.Errorf("Expected doctype name to be 'html', got '%s'", doctype.Name())
	}
	if doctype.PublicId() != "-//W3C//DTD XHTML 1.0 Transitional//EN" {
		t.Errorf("Expected public id to be '-//W3C//DTD XHTML 1.0 Transitional//EN', got '%s'", doctype.PublicId())
	}
	if doctype.SystemId() != "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd" {
		t.Errorf("Expected system id to be 'http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd', got '%s'", doctype.SystemId())
	}
}
