package xmldom

import (
	"io"
	"strings"
)

// EscapeText writes to w the properly escaped XML equivalent of plain text data.
// This function implements XML 1.0 Fifth Edition, Section 2.4 character escaping rules
// and DOM Level 2+ Core specification requirements for character data handling.
//
// For compatibility with encoding/xml.EscapeText, the following characters are escaped:
// - '<' as &lt;
// - '>' as &gt; (required in text content for compatibility)
// - '&' as &amp;
// - '"' as &#34; (numeric character reference for compatibility)
// - '\'' as &#39; (numeric character reference for compatibility)
// - '\t' as &#x9; (numeric character reference for tab)
// - '\n' as &#xA; (numeric character reference for newline)
// - '\r' as &#xD; (numeric character reference for carriage return)
//
// Characters outside the valid XML character range are replaced with the Unicode replacement character.
// This function provides full compatibility with encoding/xml.EscapeText while ensuring
// DOM specification compliance for character data handling.
func EscapeText(w io.Writer, s []byte) error {
	var esc []byte
	last := 0
	for i, c := range s {
		switch c {
		case '<':
			esc = []byte("&lt;")
		case '>':
			esc = []byte("&gt;")
		case '&':
			esc = []byte("&amp;")
		case '"':
			esc = []byte("&#34;")
		case '\'':
			esc = []byte("&#39;")
		case '\t':
			esc = []byte("&#x9;")
		case '\n':
			esc = []byte("&#xA;")
		case '\r':
			esc = []byte("&#xD;")
		default:
			// Handle invalid XML characters (control characters except tab, newline, carriage return)
			if c < 0x20 && c != 0x09 && c != 0x0A && c != 0x0D {
				// Replace invalid characters with Unicode replacement character (U+FFFD)
				// In UTF-8, this is the 3-byte sequence: 0xEF 0xBF 0xBD
				esc = []byte("\uFFFD")
			} else {
				continue
			}
		}
		if _, err := w.Write(s[last:i]); err != nil {
			return err
		}
		if _, err := w.Write(esc); err != nil {
			return err
		}
		last = i + 1
	}
	_, err := w.Write(s[last:])
	return err
}

// EscapeString returns the properly escaped XML equivalent of plain text data.
// This is a convenience function that wraps EscapeText for string input/output.
//
// Per XML 1.0 Fifth Edition Section 2.4 and DOM Level 2+ Core specification,
// this escapes all XML special characters to ensure valid character data representation.
func EscapeString(s string) string {
	var b strings.Builder
	if err := EscapeText(&b, []byte(s)); err != nil {
		// strings.Builder.Write never returns an error, so this should never happen
		panic("unexpected error from strings.Builder.Write: " + err.Error())
	}
	return b.String()
}

// UnescapeText decodes XML character entity references in text data.
// This function reverses the escaping performed by EscapeText.
//
// It handles both named entity references and numeric character references:
// - &lt; -> '<'
// - &gt; -> '>'
// - &amp; -> '&'
// - &quot; or &#34; -> '"'
// - &apos; or &#39; -> '\''
// - &#x9; -> '\t'
// - &#xA; -> '\n'
// - &#xD; -> '\r'
//
// This function is provided for completeness and round-trip compatibility.
func UnescapeText(w io.Writer, s []byte) error {
	last := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '&' {
			if _, err := w.Write(s[last:i]); err != nil {
				return err
			}
			
			// Find the end of the entity reference
			end := i + 1
			for end < len(s) && s[end] != ';' {
				end++
			}
			
			if end >= len(s) {
				// No closing semicolon found, write as-is
				if _, err := w.Write(s[i:i+1]); err != nil {
					return err
				}
				last = i + 1
				continue
			}
			
			entity := string(s[i+1 : end])
			var replacement byte
			var replacementStr string
			handled := false
			
			switch entity {
			case "lt":
				replacement = '<'
				handled = true
			case "gt":
				replacement = '>'
				handled = true
			case "amp":
				replacement = '&'
				handled = true
			case "quot", "#34":
				replacement = '"'
				handled = true
			case "apos", "#39":
				replacement = '\''
				handled = true
			case "#x9":
				replacement = '\t'
				handled = true
			case "#xA":
				replacement = '\n'
				handled = true
			case "#xD":
				replacement = '\r'
				handled = true
			default:
				// Check for Unicode replacement character
				if entity == "#xFFFD" {
					replacementStr = "\uFFFD"
					handled = true
				}
			}
			
			if handled {
				if replacementStr != "" {
					if _, err := w.Write([]byte(replacementStr)); err != nil {
						return err
					}
				} else {
					if _, err := w.Write([]byte{replacement}); err != nil {
						return err
					}
				}
			} else {
				// Unknown entity, write as-is
				if _, err := w.Write(s[i:end+1]); err != nil {
					return err
				}
			}
			
			last = end + 1
			i = end
		}
	}
	_, err := w.Write(s[last:])
	return err
}

// UnescapeString returns the decoded XML equivalent of escaped text data.
// This is a convenience function that wraps UnescapeText for string input/output.
func UnescapeString(s string) string {
	var b strings.Builder
	if err := UnescapeText(&b, []byte(s)); err != nil {
		// strings.Builder.Write never returns an error, so this should never happen
		panic("unexpected error from strings.Builder.Write: " + err.Error())
	}
	return b.String()
}