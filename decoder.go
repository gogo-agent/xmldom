package xmldom

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"golang.org/x/text/encoding/ianaindex"
)

// NewDecoderFromBytes creates a new Decoder from a byte slice, enabling position tracking
func NewDecoderFromBytes(data []byte) *Decoder {
	return NewDecoder(bytes.NewReader(data))
}

// findStartTagStartOffset scans backward from endOff-1 to find the '<' starting this start-tag
func findStartTagStartOffset(data []byte, endOff int64) int64 {
	if endOff > int64(len(data)) { endOff = int64(len(data)) }
	for i := endOff - 1; i >= 0; i-- {
		if data[i] == '<' {
			return i
		}
	}
	return 0
}

// findStartTagEndOffset scans forward from startOff to the matching '>' outside quotes
func findStartTagEndOffset(data []byte, startOff int64) int64 {
	inQuote := byte(0)
	for i := startOff; i < int64(len(data)); i++ {
		b := data[i]
		if inQuote != 0 {
			if b == inQuote {
				inQuote = 0
			}
			continue
		}
		if b == '\'' || b == '"' {
			inQuote = b
			continue
		}
		if b == '>' {
			return i
		}
	}
	return startOff
}

// findAttrNameOffset finds the absolute byte offset of the attribute local name within the start tag [startOff,endOff]
func findAttrNameOffset(data []byte, startOff, endOff int64, local string) int64 {
	if startOff < 0 { startOff = 0 }
	if endOff > int64(len(data)) { endOff = int64(len(data)) }
	seg := data[startOff:endOff]
	inQuote := byte(0)
	name := []byte(local)
	for i := 0; i+len(name) <= len(seg); i++ {
		b := seg[i]
		if inQuote != 0 {
			if b == inQuote { inQuote = 0 }
			continue
		}
		if b == '\'' || b == '"' {
			inQuote = b
			continue
		}
		if seg[i] == name[0] && bytes.HasPrefix(seg[i:], name) {
			// check simple boundary conditions: prev is whitespace or ':' or '<' or start
			prevOK := i == 0
			if i > 0 {
				p := seg[i-1]
				prevOK = p == ' ' || p == '\t' || p == ':' || p == '<' || p == '\n' || p == '\r'
			}
			// next should be '=' or whitespace or ':' (prefix already consumed) or '>'
			nextOK := true
			if i+len(name) < len(seg) {
				n := seg[i+len(name)]
				nextOK = n == '=' || n == ' ' || n == '\t' || n == ':' || n == '>'
			}
			if prevOK && nextOK {
				return startOff + int64(i)
			}
		}
	}
	return -1
}

// Decoder is a struct that decodes a DOM tree from an XML input stream.
//
// CDATA Section Limitation:
// The standard Go encoding/xml package does not differentiate between regular
// character data and CDATA sections. Both are reported as xml.CharData tokens.
// Therefore, this decoder will parse CDATA sections as Text nodes, not as
// CDATASection nodes.
//
// This means that XML like:
//
//	<script><![CDATA[var x = "<test>";]]></script>
//
// will be parsed as if it were:
//
//	<script>var x = "&lt;test&gt;";</script>
//
// The content is preserved correctly, but the CDATA structure is lost during
// parsing. When the document is serialized back to XML, the content will be
// escaped as regular character data.
//
// To work around this limitation for applications that require true CDATA
// support, create CDATASection nodes manually using Document.CreateCDATASection().
type Decoder struct {
	d             *xml.Decoder
	bufferedToken xml.Token

	// Position tracking
	sourceText []byte  // Original source text for line/column calculation
	lineStarts []int64 // Byte offsets where each line starts (1-based line numbering)
}

// DecoderOptions allows specifying decoder options.
type DecoderOptions struct {
	// CharsetReader, if non-nil, is used to decode XML input from non-UTF-8 character sets.
	CharsetReader func(charset string, input io.Reader) (io.Reader, error)
	// Strict defaults to true, requiring that XML input be well-formed.
	// If false, the decoder will make a best effort to parse malformed XML.
	Strict bool
	// Entity can be used to provide custom mappings for XML entities.
	Entity map[string]string
}

// NewDecoderWithOptions creates a new Decoder that reads from the given io.Reader
// and uses the provided options.
func NewDecoderWithOptions(r io.Reader, opts *DecoderOptions) *Decoder {
	d := xml.NewDecoder(r)
	if opts != nil {
		d.CharsetReader = opts.CharsetReader
		d.Strict = opts.Strict
		d.Entity = opts.Entity
	} else {
		d.Strict = true
	}

	if d.CharsetReader == nil {
		d.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
			e, err := ianaindex.IANA.Encoding(charset)
			if err != nil {
				return nil, fmt.Errorf("unsupported charset: %s", charset)
			}
			if e == nil {
				// This case can happen if the IANA name is known but the encoding is not available.
				// For example, the text repo may not include all encodings by default.
				return nil, fmt.Errorf("unsupported charset: %s", charset)
			}
			return e.NewDecoder().Reader(input), nil
		}
	}

decoder := &Decoder{
		d: d,
	}

	// Capture full source for position tracking by buffering the reader
	if bytesReader, ok := r.(*bytes.Reader); ok {
		pos, _ := bytesReader.Seek(0, io.SeekCurrent)
		bytesReader.Seek(0, io.SeekStart)
		sourceText, _ := io.ReadAll(bytesReader)
		bytesReader.Seek(pos, io.SeekStart)
		decoder.sourceText = sourceText
	} else {
		// Read and re-wrap the reader so xml.Decoder reads from our buffer
		if data, err := io.ReadAll(r); err == nil {
			decoder.sourceText = data
			decoder.d = xml.NewDecoder(bytes.NewReader(data))
		}
	}
	decoder.buildLineIndex()

	return decoder
}

// NewDecoder creates a new Decoder that reads from the given io.Reader
// with default options.
func NewDecoder(r io.Reader) *Decoder {
	return NewDecoderWithOptions(r, nil)
}

// ParsingError represents an error that occurred during XML parsing.
type ParsingError struct {
	// The underlying error from the xml package.
	Err error
}

func (e *ParsingError) Error() string {
	return fmt.Sprintf("XML parsing error: %v", e.Err)
}

func (d *Decoder) nextToken() (xml.Token, error) {
	if d.bufferedToken != nil {
		token := d.bufferedToken
		d.bufferedToken = nil
		return token, nil
	}
	return d.d.Token()
}

func (d *Decoder) peekToken() (xml.Token, error) {
	if d.bufferedToken != nil {
		return d.bufferedToken, nil
	}
	token, err := d.d.Token()
	if err != nil {
		return nil, err
	}
	d.bufferedToken = token
	return token, nil
}

func isValidXMLChar(r rune) bool {
	return r == 0x9 || r == 0xA || r == 0xD ||
		(r >= 0x20 && r <= 0xD7FF) ||
		(r >= 0xE000 && r <= 0xFFFD) ||
		(r >= 0x10000 && r <= 0x10FFFF)
}

// calculateLineColumn calculates the line and column number for a given byte offset
func (d *Decoder) calculateLineColumn(offset int64) (line, column int) {
	if len(d.sourceText) == 0 || offset < 0 || offset > int64(len(d.sourceText)) {
		return 0, 0
	}
	if len(d.lineStarts) == 0 {
		d.buildLineIndex()
	}
	// Binary search for the greatest line start <= offset
	lo, hi := 0, len(d.lineStarts)-1
	for lo <= hi {
		mid := (lo + hi) / 2
		if d.lineStarts[mid] <= offset {
			lo = mid + 1
		} else {
			hi = mid - 1
		}
	}
	lineIdx := hi
	if lineIdx < 0 { lineIdx = 0 }
	line = lineIdx + 1 // 1-based line number
	colStart := d.lineStarts[lineIdx]
	column = int(offset-colStart) + 1
	return line, column
}

func (d *Decoder) buildLineIndex() {
	if len(d.sourceText) == 0 { return }
	d.lineStarts = d.lineStarts[:0]
	// line 1 starts at offset 0
	d.lineStarts = append(d.lineStarts, 0)
	for i := 0; i < len(d.sourceText); i++ {
		if d.sourceText[i] == '\n' {
			// next line starts at i+1
			d.lineStarts = append(d.lineStarts, int64(i+1))
		}
	}
}

// Decode reads the XML from the input stream and returns a Document.
//
// See the Decoder struct documentation for important notes about CDATA sections.
func (d *Decoder) Decode() (Document, error) {
	impl := NewDOMImplementation()
	doc, err := impl.CreateDocument("", "", nil)
	if err != nil {
		return nil, &ParsingError{Err: err}
	}
	docImpl := doc.(*document)

	stack := []Node{doc}

	for {
		token, err := d.nextToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, &ParsingError{Err: err}
		}

		parent := stack[len(stack)-1]

		switch t := token.(type) {
		case xml.StartElement:
			// Validate element namespace rules
			if t.Name.Space == "xmlns" {
				return nil, &ParsingError{Err: fmt.Errorf("elements cannot use xmlns prefix")}
			}

			// Create a new element
			elem, err := doc.CreateElementNS(DOMString(t.Name.Space), DOMString(t.Name.Local))
			if err != nil {
				return nil, &ParsingError{Err: err}
			}

// Store position information at start-of-start-tag
			if elemImpl, ok := elem.(*element); ok {
				endOff := d.d.InputOffset()
				startOff := findStartTagStartOffset(d.sourceText, endOff)
				line, col := d.calculateLineColumn(startOff)
				elemImpl.sourcePosition = position{
					Line:   line,
					Column: col,
					Offset: startOff,
				}
			}

			// Determine start/end of start tag for attribute position mapping
			endOff := findStartTagEndOffset(d.sourceText, getInternalNode(elem).sourcePosition.Offset)

			// Copy attributes with namespace validation
			for _, attr := range t.Attr {
				// Validate namespace prefix rules during parsing
				if attr.Name.Space == "xmlns" {
					// This is a namespace declaration: xmlns:prefix="..."
					prefix := attr.Name.Local
					if prefix == "" {
						return nil, &ParsingError{Err: fmt.Errorf("empty prefix in namespace declaration")}
					}
					if prefix == "xmlns" {
						return nil, &ParsingError{Err: fmt.Errorf("cannot declare xmlns prefix")}
					}
					if prefix == "xml" && attr.Value != "http://www.w3.org/XML/1998/namespace" {
						return nil, &ParsingError{Err: fmt.Errorf("xml prefix must be bound to http://www.w3.org/XML/1998/namespace")}
					}
					if attr.Value == "http://www.w3.org/XML/1998/namespace" && prefix != "xml" {
						return nil, &ParsingError{Err: fmt.Errorf("http://www.w3.org/XML/1998/namespace can only be bound to xml prefix")}
					}
				} else if attr.Name.Space == "" && attr.Name.Local == "xmlns" {
					// Default namespace declaration: xmlns="..."
					if attr.Value == "http://www.w3.org/2000/xmlns/" {
						return nil, &ParsingError{Err: fmt.Errorf("cannot bind default namespace to xmlns namespace")}
					}
				} else if attr.Name.Space == "" && strings.HasPrefix(attr.Name.Local, "xmlns:") {
					// Malformed namespace declaration: xmlns:=""
					if attr.Name.Local == "xmlns:" {
						return nil, &ParsingError{Err: fmt.Errorf("empty prefix in namespace declaration")}
					}
					// Extract the prefix after xmlns:
					prefix := attr.Name.Local[6:] // Remove "xmlns:" prefix
					if prefix == "xmlns" {
						return nil, &ParsingError{Err: fmt.Errorf("cannot declare xmlns prefix")}
					}
					if prefix == "xml" && attr.Value != "http://www.w3.org/XML/1998/namespace" {
						return nil, &ParsingError{Err: fmt.Errorf("xml prefix must be bound to http://www.w3.org/XML/1998/namespace")}
					}
					if attr.Value == "http://www.w3.org/XML/1998/namespace" && prefix != "xml" {
						return nil, &ParsingError{Err: fmt.Errorf("http://www.w3.org/XML/1998/namespace can only be bound to xml prefix")}
					}
				}

				err := elem.SetAttributeNS(DOMString(attr.Name.Space), DOMString(attr.Name.Local), DOMString(attr.Value))
				if err != nil {
					return nil, &ParsingError{Err: err}
				}

				// Set position information for the attribute at the attribute name start, if found
				if attrNode := elem.GetAttributeNodeNS(DOMString(attr.Name.Space), DOMString(attr.Name.Local)); attrNode != nil {
					if attrImpl := getInternalNode(attrNode); attrImpl != nil {
						startOff := getInternalNode(elem).sourcePosition.Offset
						attrStart := findAttrNameOffset(d.sourceText, startOff, endOff, string(attr.Name.Local))
						if attrStart >= 0 {
							line, col := d.calculateLineColumn(attrStart)
							attrImpl.sourcePosition = position{Line: line, Column: col, Offset: attrStart}
						} else {
							// Fallback to element position
							attrImpl.sourcePosition = getInternalNode(elem).sourcePosition
						}
					}
				}
			}

			// Append the new element to the parent
			parent.AppendChild(elem)

			// Peek at the next token to see if it's a matching end element.
			nextToken, err := d.peekToken()
			if err != nil {
				if err == io.EOF {
					// This is a valid case for the last element in the document.
				} else {
					return nil, &ParsingError{Err: err}
				}
			}

			if end, ok := nextToken.(xml.EndElement); ok && end.Name == t.Name {
				// This is a self-closing element. Consume the end token.
				_, _ = d.nextToken()
			} else {
				// This is a regular start element. Push it onto the stack.
				stack = append(stack, elem)
			}

			if docImpl.documentElement == nil {
				docImpl.documentElement = elem
			}
		case xml.EndElement:
			stack = stack[:len(stack)-1]
		case xml.CharData:
			for _, r := range string(t) {
				if !isValidXMLChar(r) {
					return nil, &ParsingError{Err: fmt.Errorf("invalid character 0x%x in CharData", r)}
				}
			}
			text := doc.CreateTextNode(DOMString(t))

			// Store position information (start of text token if possible)
			if textNode := getInternalNode(text); textNode != nil {
				offEnd := d.d.InputOffset()
				// Best effort: map to end offset; for exact start, a raw slice length is needed.
				line, col := d.calculateLineColumn(offEnd)
				textNode.sourcePosition = position{Line: line, Column: col, Offset: offEnd}
			}

			parent.AppendChild(text)
		case xml.Comment:
			commentText := DOMString(t)
			for _, r := range commentText {
				if !isValidXMLChar(r) {
					return nil, &ParsingError{Err: fmt.Errorf("invalid character 0x%x in comment", r)}
				}
			}
			if strings.Contains(string(commentText), "--") {
				return nil, &ParsingError{Err: fmt.Errorf("comment contains '--'")}
			}
			comment := doc.CreateComment(commentText)

			// Store position information (at end of comment token)
			if commentNode := getInternalNode(comment); commentNode != nil {
				offEnd := d.d.InputOffset()
				line, col := d.calculateLineColumn(offEnd)
				commentNode.sourcePosition = position{Line: line, Column: col, Offset: offEnd}
			}

			parent.AppendChild(comment)
		case xml.ProcInst:
			// The Go XML parser reports the XML declaration as a ProcInst with target "xml".
			// We need to ignore this, as it's not a real processing instruction.
			if strings.EqualFold(t.Target, "xml") {
				continue
			}
			for _, r := range string(t.Inst) {
				if !isValidXMLChar(r) {
					return nil, &ParsingError{Err: fmt.Errorf("invalid character 0x%x in processing instruction", r)}
				}
			}
			pi, err := doc.CreateProcessingInstruction(DOMString(t.Target), DOMString(t.Inst))
			if err != nil {
				return nil, &ParsingError{Err: err}
			}

			// Store position information (at end of PI token)
			if piNode := getInternalNode(pi); piNode != nil {
				offEnd := d.d.InputOffset()
				line, col := d.calculateLineColumn(offEnd)
				piNode.sourcePosition = position{Line: line, Column: col, Offset: offEnd}
			}

			parent.AppendChild(pi)
		case xml.Directive:
			s := string(t)
			// Check for invalid characters in the directive itself
			for _, r := range s {
				if !isValidXMLChar(r) {
					return nil, &ParsingError{Err: fmt.Errorf("invalid character 0x%x in directive", r)}
				}
			}
			if strings.HasPrefix(s, "DOCTYPE") {
				s = strings.TrimSpace(s[len("DOCTYPE"):])

				var name, publicId, systemId string

				// Extract name
				nameEnd := strings.IndexAny(s, " \t\n\r[")
				if nameEnd == -1 {
					name = s
					s = ""
				} else {
					name = s[:nameEnd]
					s = strings.TrimSpace(s[nameEnd:])
				}

				if strings.HasPrefix(s, "PUBLIC") {
					s = strings.TrimSpace(s[len("PUBLIC"):])
					if len(s) > 0 && s[0] == '"' {
						end := strings.Index(s[1:], "\"")
						if end != -1 {
							publicId = s[1 : end+1]
							s = strings.TrimSpace(s[end+2:])
							if len(s) > 0 && s[0] == '"' {
								end = strings.Index(s[1:], "\"")
								if end != -1 {
									systemId = s[1 : end+1]
								}
							}
						}
					}
				} else if strings.HasPrefix(s, "SYSTEM") {
					s = strings.TrimSpace(s[len("SYSTEM"):])
					if len(s) > 0 && s[0] == '"' {
						end := strings.Index(s[1:], "\"")
						if end != -1 {
							systemId = s[1 : end+1]
						}
					}
				}

				doctype, err := doc.Implementation().CreateDocumentType(DOMString(name), DOMString(publicId), DOMString(systemId))
				if err != nil {
					return nil, &ParsingError{Err: err}
				}

				// Store position information for DocumentType (at end of directive)
				if dtNode := getInternalNode(doctype); dtNode != nil {
					offEnd := d.d.InputOffset()
					line, col := d.calculateLineColumn(offEnd)
					dtNode.sourcePosition = position{Line: line, Column: col, Offset: offEnd}
				}

				if docImpl, ok := doc.(*document); ok {
					docImpl.doctype = doctype
				}
			}
		}
	}

	return doc, nil
}
