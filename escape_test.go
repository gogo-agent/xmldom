package xmldom_test

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/gogo-agent/xmldom"
)

func TestEscapeText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no special characters",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "less than",
			input:    "a < b",
			expected: "a &lt; b",
		},
		{
			name:     "greater than",
			input:    "a > b",
			expected: "a &gt; b",
		},
		{
			name:     "ampersand",
			input:    "fish & chips",
			expected: "fish &amp; chips",
		},
		{
			name:     "double quote",
			input:    `say "hello"`,
			expected: "say &#34;hello&#34;",
		},
		{
			name:     "single quote",
			input:    "don't",
			expected: "don&#39;t",
		},
		{
			name:     "all special characters",
			input:    `<>&"'`,
			expected: "&lt;&gt;&amp;&#34;&#39;",
		},
		{
			name:     "mixed content",
			input:    `Hello <name>John & "Jane"</name>`,
			expected: "Hello &lt;name&gt;John &amp; &#34;Jane&#34;&lt;/name&gt;",
		},
		{
			name:     "tab and newline",
			input:    "line1\nline2\ttab",
			expected: "line1&#xA;line2&#x9;tab",
		},
		{
			name:     "control characters",
			input:    "control chars: \x01\x02\x03",
			expected: "control chars: \uFFFD\uFFFD\uFFFD",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only special characters",
			input:    "<<>>&&\"\"''",
			expected: "&lt;&lt;&gt;&gt;&amp;&amp;&#34;&#34;&#39;&#39;",
		},
		{
			name:     "unicode characters with specials",
			input:    "Hello 世界 < test",
			expected: "Hello 世界 &lt; test",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := xmldom.EscapeText(&buf, []byte(tc.input))
			if err != nil {
				t.Fatalf("EscapeText failed: %v", err)
			}
			result := buf.String()
			if result != tc.expected {
				t.Errorf("EscapeText(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestEscapeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple case",
			input:    "hello & world",
			expected: "hello &amp; world",
		},
		{
			name:     "all specials",
			input:    `<>&"'`,
			expected: "&lt;&gt;&amp;&#34;&#39;",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := xmldom.EscapeString(tc.input)
			if result != tc.expected {
				t.Errorf("EscapeString(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestUnescapeText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no entities",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "lt entity",
			input:    "a &lt; b",
			expected: "a < b",
		},
		{
			name:     "gt entity",
			input:    "a &gt; b",
			expected: "a > b",
		},
		{
			name:     "amp entity",
			input:    "fish &amp; chips",
			expected: "fish & chips",
		},
		{
			name:     "quot entity",
			input:    "say &quot;hello&quot;",
			expected: `say "hello"`,
		},
		{
			name:     "numeric quot entity",
			input:    "say &#34;hello&#34;",
			expected: `say "hello"`,
		},
		{
			name:     "apos entity",
			input:    "don&apos;t",
			expected: "don't",
		},
		{
			name:     "numeric apos entity",
			input:    "don&#39;t",
			expected: "don't",
		},
		{
			name:     "tab and newline entities",
			input:    "line1&#xA;line2&#x9;tab",
			expected: "line1\nline2\ttab",
		},
		{
			name:     "all entities",
			input:    "&lt;&gt;&amp;&quot;&apos;",
			expected: `<>&"'`,
		},
		{
			name:     "all numeric entities",
			input:    "&lt;&gt;&amp;&#34;&#39;",
			expected: `<>&"'`,
		},
		{
			name:     "mixed content",
			input:    "Hello &lt;name&gt;John &amp; &#34;Jane&#34;&lt;/name&gt;",
			expected: `Hello <name>John & "Jane"</name>`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "unknown entity",
			input:    "test &unknown; more",
			expected: "test &unknown; more",
		},
		{
			name:     "incomplete entity",
			input:    "test &amp more",
			expected: "test &amp more",
		},
		{
			name:     "multiple unknown entities",
			input:    "&foo; &bar; &baz;",
			expected: "&foo; &bar; &baz;",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := xmldom.UnescapeText(&buf, []byte(tc.input))
			if err != nil {
				t.Fatalf("UnescapeText failed: %v", err)
			}
			result := buf.String()
			if result != tc.expected {
				t.Errorf("UnescapeText(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestUnescapeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple case",
			input:    "hello &amp; world",
			expected: "hello & world",
		},
		{
			name:     "all entities",
			input:    "&lt;&gt;&amp;&#34;&#39;",
			expected: `<>&"'`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := xmldom.UnescapeString(tc.input)
			if result != tc.expected {
				t.Errorf("UnescapeString(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestEscapeRoundTrip ensures that escaping and unescaping are inverse operations
func TestEscapeRoundTrip(t *testing.T) {
	testStrings := []string{
		"hello world",
		`<tag attr="value">content & more</tag>`,
		`Mix of "quotes" and 'apostrophes' with <brackets> & ampersands`,
		"",
		"<<<>>>",
		"&&&",
		`"""`,
		"'''",
		"Hello 世界 < > & \" ' test",
	}

	for _, original := range testStrings {
		t.Run("roundtrip_"+original, func(t *testing.T) {
			escaped := xmldom.EscapeString(original)
			unescaped := xmldom.UnescapeString(escaped)
			if unescaped != original {
				t.Errorf("Round trip failed: %q -> %q -> %q", original, escaped, unescaped)
			}
		})
	}
}

// TestCompatibilityWithEncodingXML ensures our implementation behaves like encoding/xml.EscapeText
func TestCompatibilityWithEncodingXML(t *testing.T) {
	testStrings := []string{
		"hello world",
		"a < b",
		"a > b", 
		"fish & chips",
		`say "hello"`,
		"don't",
		`<>&"'`,
		`Hello <name>John & "Jane"</name>`,
		"",
		"Hello 世界 < test",
		"line1\nline2\ttab",
		"control chars: \x01\x02\x03",
	}

	for _, input := range testStrings {
		t.Run("compat_"+input, func(t *testing.T) {
			// Test against encoding/xml
			var xmlBuf bytes.Buffer
			xml.EscapeText(&xmlBuf, []byte(input))
			xmlResult := xmlBuf.String()

			// Test our implementation
			var ourBuf bytes.Buffer
			err := xmldom.EscapeText(&ourBuf, []byte(input))
			if err != nil {
				t.Fatalf("EscapeText failed: %v", err)
			}
			ourResult := ourBuf.String()

			if ourResult != xmlResult {
				t.Errorf("Compatibility mismatch for %q:\nencoding/xml: %q\nxmldom:       %q", 
					input, xmlResult, ourResult)
			}
		})
	}
}

// TestErrorHandling tests edge cases and error conditions
func TestErrorHandling(t *testing.T) {
	// Test with nil writer should panic or return error
	t.Run("nil_writer", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic with nil writer")
			}
		}()
		xmldom.EscapeText(nil, []byte("test"))
	})
}

// Test large input performance and correctness
func TestLargeInput(t *testing.T) {
	// Create a large string with special characters
	var large strings.Builder
	pattern := `Hello <world> & "friends" 'everyone'! `
	for i := 0; i < 1000; i++ {
		large.WriteString(pattern)
	}
	input := large.String()

	result := xmldom.EscapeString(input)
	
	// Verify it round-trips correctly
	unescaped := xmldom.UnescapeString(result)
	if unescaped != input {
		t.Error("Large input round-trip failed")
	}

	// Verify it matches encoding/xml for the same input
	var xmlBuf bytes.Buffer
	xml.EscapeText(&xmlBuf, []byte(input))
	xmlResult := xmlBuf.String()
	
	if result != xmlResult {
		t.Error("Large input compatibility with encoding/xml failed")
	}
}