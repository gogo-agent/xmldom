package xmldom

import (
	"strings"
	"testing"
)

func TestXPathNamespaceAttributeDebug(t *testing.T) {
	xml := `<?xml version="1.0"?>
	<root xmlns:foo="http://example.com/foo" xmlns:bar="http://example.com/bar">
		<child>Text</child>
	</root>`

	decoder := NewDecoder(strings.NewReader(xml))
	doc, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	// Get root element
	root := doc.DocumentElement()
	t.Logf("Root element: %s", root.NodeName())

	// Check attributes on root element
	attrs := root.Attributes()
	t.Logf("Number of attributes on root: %d", attrs.Length())

	for i := uint(0); i < attrs.Length(); i++ {
		attrNode := attrs.Item(i)
		if attr, ok := attrNode.(Attr); ok {
			t.Logf("Attribute %d: name='%s', value='%s', namespaceURI='%s', prefix='%s', localName='%s'",
				i, attr.NodeName(), attr.NodeValue(), attr.NamespaceURI(), attr.Prefix(), attr.LocalName())
		}
	}

	// Check if namespace is set on element
	t.Logf("Root namespace URI: '%s'", root.NamespaceURI())
	t.Logf("Root prefix: '%s'", root.Prefix())
}
