package xmldom

import (
	"encoding/xml"
	"fmt"
	"io"
)

// Encoder writes DOM nodes as XML to an output stream.
type Encoder struct {
	e *xml.Encoder
	w io.Writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	enc := &Encoder{
		e: xml.NewEncoder(w),
		w: w,
	}
	enc.e.Indent("", "  ")
	return enc
}

// SetIndent sets the indentation for the encoder.
// The prefix is written at the beginning of each line except the first.
// The indent string is written for each level of indentation.
func (enc *Encoder) SetIndent(prefix, indent string) {
	enc.e.Indent(prefix, indent)
}

// Encode writes the XML encoding of node to the stream.
func (enc *Encoder) Encode(node Node) error {
	if node.NodeType() == DOCUMENT_NODE {
		doc := node.(Document)
		if doc.Doctype() != nil {
			if err := enc.encodeDoctype(doc.Doctype()); err != nil {
				return err
			}
		}
	}

	if err := enc.encodeNode(node); err != nil {
		return err
	}

	return enc.e.Flush()
}

func (enc *Encoder) encodeNode(node Node) error {
	if node == nil {
		return nil
	}

	switch node.NodeType() {
	case ELEMENT_NODE:
		return enc.encodeElement(node.(Element))

	case TEXT_NODE:
		return enc.e.EncodeToken(xml.CharData(node.NodeValue()))

	case COMMENT_NODE:
		return enc.e.EncodeToken(xml.Comment(node.NodeValue()))

	case CDATA_SECTION_NODE:
		// CDATA sections must be written manually since Go's xml.Encoder
		// doesn't provide a CDATA token type and would escape the content
		// if we used xml.CharData
		_, err := fmt.Fprintf(enc.w, "<![CDATA[%s]]>", string(node.NodeValue()))
		return err

	case PROCESSING_INSTRUCTION_NODE:
		pi := node.(ProcessingInstruction)
		return enc.e.EncodeToken(xml.ProcInst{
			Target: string(pi.Target()),
			Inst:   []byte(pi.Data()),
		})

	case DOCUMENT_NODE:
		// Encode children
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			if err := enc.encodeNode(child); err != nil {
				return err
			}
		}

	case DOCUMENT_FRAGMENT_NODE:
		// Encode children
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			if err := enc.encodeNode(child); err != nil {
				return err
			}
		}
	}

	return nil
}

func (enc *Encoder) encodeElement(elem Element) error {
	// Create start element
	start := xml.StartElement{
		Name: xml.Name{
			Space: string(elem.NamespaceURI()),
			Local: string(elem.LocalName()),
		},
	}

	// Add attributes
	if attrs := elem.Attributes(); attrs != nil {
		for i := uint(0); i < attrs.Length(); i++ {
			attr := attrs.Item(i)
			if attr != nil && attr.NodeType() == ATTRIBUTE_NODE {
				a := attr.(Attr)
				start.Attr = append(start.Attr, xml.Attr{
					Name: xml.Name{
						Space: string(attr.NamespaceURI()),
						Local: string(a.LocalName()),
					},
					Value: string(a.Value()),
				})
			}
		}
	}

	// Encode start element
	if err := enc.e.EncodeToken(start); err != nil {
		return err
	}

	// Encode children
	for child := elem.FirstChild(); child != nil; child = child.NextSibling() {
		if err := enc.encodeNode(child); err != nil {
			return err
		}
	}

	// Encode end element
	return enc.e.EncodeToken(xml.EndElement{Name: start.Name})
}

func (enc *Encoder) encodeDoctype(doctype DocumentType) error {
	// XML encoder doesn't support DOCTYPE directly, write as string
	// This is a simplified approach
	docStr := "<!DOCTYPE " + string(doctype.Name())
	if doctype.PublicId() != "" {
		docStr += " PUBLIC \"" + string(doctype.PublicId()) + "\""
		if doctype.SystemId() != "" {
			docStr += " \"" + string(doctype.SystemId()) + "\""
		}
	} else if doctype.SystemId() != "" {
		docStr += " SYSTEM \"" + string(doctype.SystemId()) + "\""
	}
	docStr += ">"

	// Write directly as a comment (workaround for XML encoder limitation)
	return enc.e.EncodeToken(xml.Comment(docStr))
}
