package xmldom

import (
	"bytes"
	"encoding/xml"
	"strings"
)

// Unmarshal parses XML-encoded data and stores the result in the value pointed to by v.
// This function delegates to Go's standard xml.Unmarshal for struct unmarshaling.
func Unmarshal(data []byte, v interface{}) error {
	return xml.Unmarshal(data, v)
}

// UnmarshalDOM parses XML-encoded data and returns a DOM Document.
// This creates a DOM tree that can be manipulated using the xmldom API.
func UnmarshalDOM(data []byte) (Document, error) {
	decoder := NewDecoder(strings.NewReader(string(data)))
	return decoder.Decode()
}

// Marshal returns the XML encoding of v.
// This function handles both DOM Documents and regular structs.
func Marshal(v interface{}) ([]byte, error) {
	// Check if v is a DOM Document
	if doc, ok := v.(Document); ok {
		return marshalDOM(doc)
	}
	// Check if v is a DOM Element
	if elem, ok := v.(Element); ok {
		return marshalElement(elem)
	}
	// Check if v is any DOM Node
	if node, ok := v.(Node); ok {
		return marshalNode(node)
	}
	// For non-DOM objects, delegate to Go's standard xml.Marshal
	return xml.Marshal(v)
}

// marshalDOM serializes a DOM Document to XML
func marshalDOM(doc Document) ([]byte, error) {
	var buf bytes.Buffer

	// Write XML declaration
	buf.WriteString(`<?xml version="1.0"?>`)

	// Serialize the document element
	root := doc.DocumentElement()
	if root == nil {
		return buf.Bytes(), nil // Empty document
	}

	if err := serializeElement(&buf, root, false); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// marshalElement serializes a DOM Element to XML (without XML declaration)
func marshalElement(elem Element) ([]byte, error) {
	var buf bytes.Buffer
	if err := serializeElement(&buf, elem, false); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// marshalNode serializes any DOM Node to XML (without XML declaration)
func marshalNode(node Node) ([]byte, error) {
	var buf bytes.Buffer
	if err := serializeNode(&buf, node); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// serializeElement serializes an Element and its children to XML
func serializeElement(buf *bytes.Buffer, elem Element, skipRoot bool) error {
	if !skipRoot {
		// Write opening tag
		buf.WriteString("<")
		buf.WriteString(string(elem.TagName()))

		// Write attributes
		attrs := elem.Attributes()
		if attrs != nil {
			for i := uint(0); i < attrs.Length(); i++ {
				attr := attrs.Item(i)
				if attr != nil && attr.NodeType() == ATTRIBUTE_NODE {
					if attrNode, ok := attr.(Attr); ok {
						buf.WriteString(" ")
						buf.WriteString(string(attrNode.Name()))
						buf.WriteString(`="`)
						buf.WriteString(EscapeString(string(attrNode.Value())))
						buf.WriteString(`"`)
					}
				}
			}
		}

		// Check if element has children
		hasChildren := elem.HasChildNodes()
		if !hasChildren {
			// For SCXML conformance, always use explicit opening/closing tags
			// instead of self-closing tags for empty elements
			buf.WriteString("></")
			buf.WriteString(string(elem.TagName()))
			buf.WriteString(">")
			return nil
		}

		buf.WriteString(">")
	}

	// Serialize children
	for child := elem.FirstChild(); child != nil; child = child.NextSibling() {
		if err := serializeNode(buf, child); err != nil {
			return err
		}
	}

	if !skipRoot {
		// Write closing tag
		buf.WriteString("</")
		buf.WriteString(string(elem.TagName()))
		buf.WriteString(">")
	}

	return nil
}

// serializeNode serializes any DOM node to XML
func serializeNode(buf *bytes.Buffer, node Node) error {
	switch node.NodeType() {
	case ELEMENT_NODE:
		if elem, ok := node.(Element); ok {
			return serializeElement(buf, elem, false)
		}
	case TEXT_NODE:
		if text, ok := node.(Text); ok {
			buf.WriteString(EscapeString(string(text.Data())))
		}
	case COMMENT_NODE:
		if comment, ok := node.(Comment); ok {
			buf.WriteString("<!--")
			buf.WriteString(string(comment.Data()))
			buf.WriteString("-->")
		}
	case CDATA_SECTION_NODE:
		if cdata, ok := node.(CDATASection); ok {
			buf.WriteString("<![CDATA[")
			buf.WriteString(string(cdata.Data()))
			buf.WriteString("]]>")
		}
	case PROCESSING_INSTRUCTION_NODE:
		if pi, ok := node.(ProcessingInstruction); ok {
			buf.WriteString("<?")
			buf.WriteString(string(pi.Target()))
			if data := string(pi.Data()); data != "" {
				buf.WriteString(" ")
				buf.WriteString(data)
			}
			buf.WriteString("?>")
		}
		// Skip other node types for now
	}
	return nil
}
