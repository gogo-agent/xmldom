package xmldom

import (
	"math"
	"strings"
	"testing"
)

// TestXPathNodeSetFunctions tests the newly implemented node-set functions
func TestXPathNodeSetFunctions(t *testing.T) {
	// Create test document with IDs and namespaces
	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<library xmlns:book="http://example.com/book">
	<book:book id="book1" isbn="978-1-234567-89-0">
		<title>The Go Programming Language</title>
		<author>Alan Donovan</author>
		<author>Brian Kernighan</author>
	</book:book>
	<book:book id="book2" isbn="978-0-987654-32-1">
		<title>Clean Code</title>
		<author>Robert Martin</author>
	</book:book>
	<magazine id="mag1">
		<title>Go Magazine</title>
		<issue>42</issue>
	</magazine>
</library>`

	doc, err := ParseFromString(xmlContent)
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	root := doc.DocumentElement()

	tests := []struct {
		name     string
		xpath    string
		expected interface{}
		wantErr  bool
	}{
		// id() function tests
		{
			name:     "id function with single ID",
			xpath:    "count(id('book1'))",
			expected: 1.0,
		},
		{
			name:     "id function with multiple IDs",
			xpath:    "count(id('book1 mag1'))",
			expected: 2.0,
		},
		{
			name:     "id function with non-existent ID",
			xpath:    "count(id('nonexistent'))",
			expected: 0.0,
		},
		{
			name:     "id function with empty string",
			xpath:    "count(id(''))",
			expected: 0.0,
		},

		// local-name() function tests
		{
			name:     "local-name with no argument (context node)",
			xpath:    "local-name()",
			expected: "library",
		},
		{
			name:     "local-name with node-set argument",
			xpath:    "local-name(//*[@id='book1'])",
			expected: "book",
		},
		{
			name:     "local-name with empty node-set",
			xpath:    "local-name(//nonexistent)",
			expected: "",
		},

		// namespace-uri() function tests
		{
			name:     "namespace-uri with no argument",
			xpath:    "namespace-uri()",
			expected: "",
		},
		{
			name:     "namespace-uri with namespaced element",
			xpath:    "namespace-uri(//*[@id='book1'])",
			expected: "http://example.com/book",
		},
		{
			name:     "namespace-uri with empty node-set",
			xpath:    "namespace-uri(//nonexistent)",
			expected: "",
		},

		// name() function tests
		{
			name:     "name with no argument (context node)",
			xpath:    "name()",
			expected: "library",
		},
		{
			name:     "name with namespaced element",
			xpath:    "name(//*[@id='book1'])",
			expected: "book", // Note: Decoder doesn't preserve namespace prefix
		},
		{
			name:     "name with empty node-set",
			xpath:    "name(//nonexistent)",
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := EvaluateXPath(doc, root, test.xpath)
			if test.wantErr {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			switch expected := test.expected.(type) {
			case float64:
				if numResult, ok := result.(xpathNumberValue); ok {
					if numResult.value != expected {
						t.Errorf("Expected %v, got %v", expected, numResult.value)
					}
				} else {
					t.Errorf("Expected number result, got %T", result)
				}
			case string:
				if strResult, ok := result.(xpathStringValue); ok {
					if strResult.value != expected {
						t.Errorf("Expected %q, got %q", expected, strResult.value)
					}
				} else {
					t.Errorf("Expected string result, got %T", result)
				}
			}
		})
	}
}

// TestXPathStringFunctions tests the newly implemented string functions
func TestXPathStringFunctions(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<data>
	<text>Hello, World!</text>
	<path>/usr/local/bin</path>
	<email>user@example.com</email>
	<empty></empty>
</data>`

	doc, err := ParseFromString(xmlContent)
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	root := doc.DocumentElement()

	tests := []struct {
		name     string
		xpath    string
		expected string
		wantErr  bool
	}{
		// substring-before() tests
		{
			name:     "substring-before basic case",
			xpath:    "substring-before('Hello, World!', ', ')",
			expected: "Hello",
		},
		{
			name:     "substring-before not found",
			xpath:    "substring-before('Hello World', 'xyz')",
			expected: "",
		},
		{
			name:     "substring-before empty substring",
			xpath:    "substring-before('Hello', '')",
			expected: "",
		},
		{
			name:     "substring-before with node content",
			xpath:    "substring-before(//email, '@')",
			expected: "user",
		},

		// substring-after() tests
		{
			name:     "substring-after basic case",
			xpath:    "substring-after('Hello, World!', ', ')",
			expected: "World!",
		},
		{
			name:     "substring-after not found",
			xpath:    "substring-after('Hello World', 'xyz')",
			expected: "",
		},
		{
			name:     "substring-after empty substring",
			xpath:    "substring-after('Hello', '')",
			expected: "Hello",
		},
		{
			name:     "substring-after with node content",
			xpath:    "substring-after(//email, '@')",
			expected: "example.com",
		},
		{
			name:     "substring-after with path separator",
			xpath:    "substring-after(//path, '/usr/')",
			expected: "local/bin",
		},

		// translate() tests
		{
			name:     "translate basic character replacement",
			xpath:    "translate('Hello', 'el', 'XY')",
			expected: "HXYYo",
		},
		{
			name:     "translate character removal",
			xpath:    "translate('Hello', 'el', 'X')",
			expected: "HXo", // 'e' -> 'X', 'l' removed (no corresponding char)
		},
		{
			name:     "translate no matching characters",
			xpath:    "translate('Hello', 'xyz', 'ABC')",
			expected: "Hello",
		},
		{
			name:     "translate empty string",
			xpath:    "translate('', 'abc', 'XYZ')",
			expected: "",
		},
		{
			name:     "translate Unicode characters",
			xpath:    "translate('Héllo', 'é', 'e')",
			expected: "Hello",
		},
		{
			name:     "translate complete removal",
			xpath:    "translate('Hello', 'Hello', '')",
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := EvaluateXPath(doc, root, test.xpath)
			if test.wantErr {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if strResult, ok := result.(xpathStringValue); ok {
				if strResult.value != test.expected {
					t.Errorf("Expected %q, got %q", test.expected, strResult.value)
				}
			} else {
				t.Errorf("Expected string result, got %T", result)
			}
		})
	}
}

// TestXPathNumberFunctions tests the newly implemented number functions
func TestXPathNumberFunctions(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<data>
	<prices>
		<item price="10.50">Book</item>
		<item price="25.99">Magazine</item>
		<item price="5.00">Newspaper</item>
		<item price="invalid">Invalid</item>
	</prices>
	<numbers>
		<value>3.7</value>
		<value>-2.3</value>
		<value>8.5</value>
		<value>0</value>
	</numbers>
</data>`

	doc, err := ParseFromString(xmlContent)
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	root := doc.DocumentElement()

	tests := []struct {
		name     string
		xpath    string
		expected float64
		wantErr  bool
	}{
		// sum() function tests
		{
			name:     "sum of simple numbers",
			xpath:    "sum(//numbers/value)",
			expected: 9.9, // 3.7 + (-2.3) + 8.5 + 0
		},
		{
			name:     "sum with invalid numbers (NaN)",
			xpath:    "sum(//prices/item/@price)",
			expected: math.NaN(), // includes "invalid" which becomes NaN
		},
		{
			name:     "sum of empty node-set",
			xpath:    "sum(//nonexistent)",
			expected: 0.0,
		},
		{
			name:     "sum of first two values",
			xpath:    "sum(//numbers/value[1]) + sum(//numbers/value[2])", 
			expected: 1.4, // 3.7 + (-2.3)
		},

		// floor() function tests
		{
			name:     "floor of positive number",
			xpath:    "floor(3.7)",
			expected: 3.0,
		},
		{
			name:     "floor of negative number",
			xpath:    "floor(-2.3)",
			expected: -3.0,
		},
		{
			name:     "floor of integer",
			xpath:    "floor(5)",
			expected: 5.0,
		},
		{
			name:     "floor of zero",
			xpath:    "floor(0)",
			expected: 0.0,
		},
		{
			name:     "floor of NaN",
			xpath:    "floor(number('invalid'))",
			expected: math.NaN(),
		},
		{
			name:     "floor of positive infinity",
			xpath:    "floor(1 div 0)",
			expected: math.Inf(1),
		},

		// ceiling() function tests
		{
			name:     "ceiling of positive number",
			xpath:    "ceiling(3.2)",
			expected: 4.0,
		},
		{
			name:     "ceiling of negative number",
			xpath:    "ceiling(-2.7)",
			expected: -2.0,
		},
		{
			name:     "ceiling of integer",
			xpath:    "ceiling(5)",
			expected: 5.0,
		},
		{
			name:     "ceiling of NaN",
			xpath:    "ceiling(number('invalid'))",
			expected: math.NaN(),
		},

		// round() function tests
		{
			name:     "round positive number up",
			xpath:    "round(3.7)",
			expected: 4.0,
		},
		{
			name:     "round positive number down",
			xpath:    "round(3.2)",
			expected: 3.0,
		},
		{
			name:     "round positive half up",
			xpath:    "round(3.5)",
			expected: 4.0, // XPath 1.0: round away from zero
		},
		{
			name:     "round negative number toward zero",
			xpath:    "round(-2.3)",
			expected: -2.0,
		},
		{
			name:     "round negative number away from zero",
			xpath:    "round(-2.7)",
			expected: -3.0,
		},
		{
			name:     "round negative half away from zero",
			xpath:    "round(-2.5)",
			expected: -3.0, // XPath 1.0: round away from zero
		},
		{
			name:     "round zero",
			xpath:    "round(0)",
			expected: 0.0,
		},
		{
			name:     "round NaN",
			xpath:    "round(number('invalid'))",
			expected: math.NaN(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := EvaluateXPath(doc, root, test.xpath)
			if test.wantErr {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if numResult, ok := result.(xpathNumberValue); ok {
				if math.IsNaN(test.expected) {
					if !math.IsNaN(numResult.value) {
						t.Errorf("Expected NaN, got %v", numResult.value)
					}
				} else if math.IsInf(test.expected, 0) {
					if !math.IsInf(numResult.value, 0) || math.Signbit(numResult.value) != math.Signbit(test.expected) {
						t.Errorf("Expected %v, got %v", test.expected, numResult.value)
					}
				} else if math.Abs(numResult.value-test.expected) > 0.001 { // Allow small floating point differences
					t.Errorf("Expected %v, got %v", test.expected, numResult.value)
				}
			} else {
				t.Errorf("Expected number result, got %T", result)
			}
		})
	}
}

// TestXPathBooleanFunctions tests the newly implemented boolean functions
func TestXPathBooleanFunctions(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<document xml:lang="en">
	<section xml:lang="en-US">
		<paragraph>English content</paragraph>
		<note xml:lang="fr">French note</note>
	</section>
	<chapter xml:lang="de-DE">
		<title>German chapter</title>
		<subsection>
			<text>Inherits German</text>
		</subsection>
	</chapter>
	<appendix>
		<content>No language specified</content>
	</appendix>
</document>`

	doc, err := ParseFromString(xmlContent)
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	tests := []struct {
		name        string
		contextPath string // XPath to find context node
		langTest    string // Language to test
		expected    bool
	}{
		// lang() function tests
		{
			name:        "lang exact match",
			contextPath: "//section",
			langTest:    "en-US",
			expected:    true,
		},
		{
			name:        "lang subtag match",
			contextPath: "//section",
			langTest:    "en",
			expected:    true, // "en" should match "en-US"
		},
		{
			name:        "lang inherited match",
			contextPath: "//section/paragraph",
			langTest:    "en-US",
			expected:    true, // Inherits from parent
		},
		{
			name:        "lang inherited subtag match",
			contextPath: "//section/paragraph",
			langTest:    "en",
			expected:    true, // "en" matches inherited "en-US"
		},
		{
			name:        "lang no match",
			contextPath: "//section",
			langTest:    "fr",
			expected:    false,
		},
		{
			name:        "lang case insensitive",
			contextPath: "//chapter",
			langTest:    "DE-de",
			expected:    true, // Should match "de-DE" case-insensitively
		},
		{
			name:        "lang overridden by child",
			contextPath: "//note",
			langTest:    "en",
			expected:    false, // note has xml:lang="fr", not "en"
		},
		{
			name:        "lang overridden exact match",
			contextPath: "//note",
			langTest:    "fr",
			expected:    true,
		},
		{
			name:        "lang inherited deep",
			contextPath: "//subsection/text",
			langTest:    "de",
			expected:    true, // Inherits "de-DE" from ancestor
		},
		{
			name:        "lang no attribute in hierarchy",
			contextPath: "//appendix/content",
			langTest:    "en",
			expected:    true, // Inherits from document root xml:lang="en"
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// First find the context node
			contextResult, err := EvaluateXPath(doc, doc.DocumentElement(), test.contextPath)
			if err != nil {
				t.Fatalf("Failed to find context node: %v", err)
			}

			nodeSet, ok := contextResult.(xpathNodeSetValue)
			if !ok || len(nodeSet.nodes) == 0 {
				t.Fatalf("Context path %q did not return any nodes", test.contextPath)
			}

			contextNode := nodeSet.nodes[0]

			// Now test the lang() function
			langXPath := "lang('" + test.langTest + "')"
			result, err := EvaluateXPath(doc, contextNode, langXPath)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if boolResult, ok := result.(xpathBooleanValue); ok {
				if boolResult.value != test.expected {
					t.Errorf("Expected %v, got %v", test.expected, boolResult.value)
				}
			} else {
				t.Errorf("Expected boolean result, got %T", result)
			}
		})
	}
}

// TestXPathFunctionErrorCases tests error conditions for all new functions
func TestXPathFunctionErrorCases(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<data>
	<item>test</item>
</data>`

	doc, err := ParseFromString(xmlContent)
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	root := doc.DocumentElement()

	errorTests := []struct {
		name     string
		xpath    string
		errorMsg string
	}{
		// Wrong argument types
		{
			name:     "local-name with non-node-set",
			xpath:    "local-name('string')",
			errorMsg: "local-name() argument must be a node-set",
		},
		{
			name:     "namespace-uri with non-node-set",
			xpath:    "namespace-uri(123)",
			errorMsg: "namespace-uri() argument must be a node-set",
		},
		{
			name:     "name with non-node-set",
			xpath:    "name(true())",
			errorMsg: "name() argument must be a node-set",
		},
		{
			name:     "sum with non-node-set",
			xpath:    "sum('not a node-set')",
			errorMsg: "sum() argument must be a node-set",
		},
		{
			name:     "count with non-node-set",
			xpath:    "count(123)",
			errorMsg: "count() argument must be a node-set",
		},

		// Wrong argument counts would be caught by the function framework
		// These are tested implicitly through the argument validation
	}

	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			_, err := EvaluateXPath(doc, root, test.xpath)
			if err == nil {
				t.Errorf("Expected error for %q, but got none", test.xpath)
				return
			}

			if !strings.Contains(err.Error(), test.errorMsg) {
				t.Errorf("Expected error containing %q, got %q", test.errorMsg, err.Error())
			}
		})
	}
}

// TestXPathFunctionEdgeCases tests edge cases and XPath 1.0 compliance
func TestXPathFunctionEdgeCases(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<data>
	<unicode>Héllo Wörld 你好</unicode>
	<numbers>
		<positive>3.14159</positive>
		<negative>-2.71828</negative>
		<zero>0</zero>
		<large>999999999.999</large>
	</numbers>
	<special>
		<empty></empty>
		<whitespace>   </whitespace>
		<mixed>123abc</mixed>
	</special>
</data>`

	doc, err := ParseFromString(xmlContent)
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	root := doc.DocumentElement()

	tests := []struct {
		name     string
		xpath    string
		check    func(result XPathValue) bool
		desc     string
	}{
		// Unicode handling in string functions
		{
			name:  "translate with Unicode",
			xpath: "translate(//unicode, 'ö你', 'o你')",
			check: func(result XPathValue) bool {
				return result.(xpathStringValue).value == "Héllo World 你好"
			},
			desc: "Unicode characters should be handled correctly",
		},

		// Special numeric values
		{
			name:  "sum with mixed valid/invalid numbers",
			xpath: "sum(//numbers/value) + sum(//special/mixed)",
			check: func(result XPathValue) bool {
				// sum(//numbers/value) = 3.14159 + (-2.71828) + 0 = 0.42331
				// sum(//special/mixed) = NaN (because "123abc" -> NaN)
				// 0.42331 + NaN = NaN
				return math.IsNaN(result.(xpathNumberValue).value)
			},
			desc: "NaN should propagate in arithmetic operations",
		},

		// Rounding edge cases
		{
			name:  "round edge case positive half",
			xpath: "round(2.5)",
			check: func(result XPathValue) bool {
				return result.(xpathNumberValue).value == 3.0
			},
			desc: "XPath 1.0 rounds 0.5 away from zero",
		},
		{
			name:  "round edge case negative half",
			xpath: "round(-2.5)",
			check: func(result XPathValue) bool {
				return result.(xpathNumberValue).value == -3.0
			},
			desc: "XPath 1.0 rounds -0.5 away from zero",
		},

		// String function edge cases
		{
			name:  "substring-before with overlapping patterns",
			xpath: "substring-before('ababab', 'abab')",
			check: func(result XPathValue) bool {
				return result.(xpathStringValue).value == ""
			},
			desc: "Should find first occurrence only",
		},

		// Empty and whitespace handling
		{
			name:  "translate empty replacement string",
			xpath: "translate('abc', 'abc', '')",
			check: func(result XPathValue) bool {
				return result.(xpathStringValue).value == ""
			},
			desc: "Characters should be removed when no replacement provided",
		},

		// Multiple ID handling
		{
			name:  "id function with whitespace-separated IDs",
			xpath: "count(id('   nonexistent1    nonexistent2   '))",
			check: func(result XPathValue) bool {
				return result.(xpathNumberValue).value == 0.0
			},
			desc: "ID function should handle whitespace correctly",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := EvaluateXPath(doc, root, test.xpath)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !test.check(result) {
				t.Errorf("Test failed: %s. Result: %v", test.desc, result)
			}
		})
	}
}

// TestXPathFunctionIntegration tests functions working together
func TestXPathFunctionIntegration(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<catalog xml:lang="en">
	<products>
		<book id="b1" category="fiction">
			<title>The Hobbit</title>
			<price>12.99</price>
		</book>
		<book id="b2" category="tech">
			<title>Go Programming</title>
			<price>45.00</price>
		</book>
		<magazine id="m1" category="tech">
			<title>Tech Monthly</title>
			<price>8.50</price>
		</magazine>
	</products>
</catalog>`

	doc, err := ParseFromString(xmlContent)
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	root := doc.DocumentElement()

	tests := []struct {
		name     string
		xpath    string
		expected interface{}
	}{
		// Complex expressions using multiple functions
		{
			name:     "count with id() selection",
			xpath:    "count(id('b1'))",
			expected: 1.0, // Should find one book
		},
		{
			name:     "translate with substring functions",
			xpath:    "translate(substring-before(//book[1]/title, ' '), 'Th', 'th')",
			expected: "the", // "The" -> "the"
		},
		{
			name:     "round sum with conditional selection",
			xpath:    "round(sum(//book[@category='tech']/price))",
			expected: 45.0,
		},
		{
			name:     "local-name of id-selected element",
			xpath:    "local-name(id('m1'))",
			expected: "magazine",
		},
		{
			name:     "lang test with complex path",
			xpath:    "lang('en') and contains(//book[1]/title, 'Hobbit')",
			expected: true,
		},
		{
			name:     "ceiling of average price",
			xpath:    "ceiling(sum(//book/price) div count(//book))",
			expected: 29.0, // ceiling((12.99 + 45.00) / 2) = ceiling(28.995) = 29
		},
		{
			name:     "complex string manipulation",
			xpath:    "substring-after(translate(//book[1]/title, ' ', '_'), 'The_')",
			expected: "Hobbit", // "The Hobbit" -> "The_Hobbit" -> "Hobbit"
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := EvaluateXPath(doc, root, test.xpath)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			switch expected := test.expected.(type) {
			case float64:
				if numResult, ok := result.(xpathNumberValue); ok {
					if math.Abs(numResult.value-expected) > 0.001 { // Allow small floating point differences
						t.Errorf("Expected %v, got %v", expected, numResult.value)
					}
				} else {
					t.Errorf("Expected number result, got %T", result)
				}
			case string:
				if strResult, ok := result.(xpathStringValue); ok {
					if strResult.value != expected {
						t.Errorf("Expected %q, got %q", expected, strResult.value)
					}
				} else {
					t.Errorf("Expected string result, got %T", result)
				}
			case bool:
				if boolResult, ok := result.(xpathBooleanValue); ok {
					if boolResult.value != expected {
						t.Errorf("Expected %v, got %v", expected, boolResult.value)
					}
				} else {
					t.Errorf("Expected boolean result, got %T", result)
				}
			}
		})
	}
}

// TestXPathFunctionArgumentValidation tests argument count validation
func TestXPathFunctionArgumentValidation(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<data><item>test</item></data>`

	doc, err := ParseFromString(xmlContent)
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	root := doc.DocumentElement()

	// Test cases for argument count validation
	invalidArgTests := []struct {
		name  string
		xpath string
	}{
		// Too few arguments
		{"id() no args", "id()"},
		{"substring-before() one arg", "substring-before('test')"},
		{"substring-after() one arg", "substring-after('test')"},
		{"translate() two args", "translate('test', 'abc')"},
		{"sum() no args", "sum()"},
		{"floor() no args", "floor()"},
		{"ceiling() no args", "ceiling()"},
		{"round() no args", "round()"},
		{"lang() no args", "lang()"},

		// Too many arguments  
		{"id() two args", "id('a', 'b')"},
		{"local-name() two args", "local-name(//item, //item)"},
		{"namespace-uri() two args", "namespace-uri(//item, //item)"},
		{"name() two args", "name(//item, //item)"},
		{"substring-before() three args", "substring-before('a', 'b', 'c')"},
		{"substring-after() three args", "substring-after('a', 'b', 'c')"},
		{"translate() four args", "translate('a', 'b', 'c', 'd')"},
		{"sum() two args", "sum(//item, //item)"},
		{"floor() two args", "floor(1, 2)"},
		{"ceiling() two args", "ceiling(1, 2)"},
		{"round() two args", "round(1, 2)"},
		{"lang() two args", "lang('en', 'us')"},
	}

	for _, test := range invalidArgTests {
		t.Run(test.name, func(t *testing.T) {
			_, err := EvaluateXPath(doc, root, test.xpath)
			if err == nil {
				t.Errorf("Expected error for %q, but got none", test.xpath)
			}
			// Check that error message mentions argument count
			if !strings.Contains(err.Error(), "requires") && !strings.Contains(err.Error(), "arguments") {
				t.Errorf("Expected argument count error, got %q", err.Error())
			}
		})
	}
}

// ParseFromString parses XML from a string into a Document for tests
func ParseFromString(xml string) (Document, error) {
	dec := NewDecoder(strings.NewReader(xml))
	return dec.Decode()
}

// EvaluateXPath evaluates an XPath expression against a context node and
// returns a concrete XPathValue matching the underlying type, to keep
// existing tests working with the Document.Evaluate API.
func EvaluateXPath(doc Document, contextNode Node, xpath string) (XPathValue, error) {
	// Use ANY type so the evaluator selects the most appropriate result
	res, err := doc.Evaluate(xpath, contextNode, nil, XPATH_ANY_TYPE, nil)
	if err != nil {
		return nil, err
	}

	switch res.ResultType() {
	case XPATH_STRING_TYPE:
		s, err := res.StringValue()
		if err != nil {
			return nil, err
		}
		return NewXPathStringValue(s), nil
	case XPATH_NUMBER_TYPE:
		n, err := res.NumberValue()
		if err != nil {
			return nil, err
		}
		return NewXPathNumberValue(n), nil
	case XPATH_BOOLEAN_TYPE:
		b, err := res.BooleanValue()
		if err != nil {
			return nil, err
		}
		return NewXPathBooleanValue(b), nil
	case XPATH_UNORDERED_NODE_ITERATOR_TYPE, XPATH_ORDERED_NODE_ITERATOR_TYPE:
		var nodes []Node
		for {
			node, err := res.IterateNext()
			if err != nil {
				return nil, err
			}
			if node == nil {
				break
			}
			nodes = append(nodes, node)
		}
		return NewXPathNodeSetValue(nodes), nil
	case XPATH_UNORDERED_NODE_SNAPSHOT_TYPE, XPATH_ORDERED_NODE_SNAPSHOT_TYPE:
		l, err := res.SnapshotLength()
		if err != nil {
			return nil, err
		}
		var nodes []Node
		for i := uint32(0); i < l; i++ {
			n, err := res.SnapshotItem(i)
			if err != nil {
				return nil, err
			}
			if n != nil {
				nodes = append(nodes, n)
			}
		}
		return NewXPathNodeSetValue(nodes), nil
	case XPATH_ANY_UNORDERED_NODE_TYPE, XPATH_FIRST_ORDERED_NODE_TYPE:
		n, err := res.SingleNodeValue()
		if err != nil {
			return nil, err
		}
		if n == nil {
			return NewXPathNodeSetValue(nil), nil
		}
		return NewXPathNodeSetValue([]Node{n}), nil
	default:
		// Fallback: try to coerce into basic types
		if s, err := res.StringValue(); err == nil {
			return NewXPathStringValue(s), nil
		}
		if n, err := res.NumberValue(); err == nil {
			return NewXPathNumberValue(n), nil
		}
		if b, err := res.BooleanValue(); err == nil {
			return NewXPathBooleanValue(b), nil
		}
		return nil, NewXPathException("TYPE_ERR", "Unsupported XPath result type")
	}
}
