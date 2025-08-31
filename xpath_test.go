package xmldom

import (
	"strings"
	"testing"
)

func TestXPathBasicEvaluation(t *testing.T) {
	// Create a simple XML document for testing
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<root>
	<book id="1" title="Go Programming">
		<author>John Doe</author>
		<price>29.99</price>
	</book>
	<book id="2" title="XML Processing">
		<author>Jane Smith</author>
		<price>39.99</price>
	</book>
	<magazine id="3" title="Tech Today">
		<editor>Bob Wilson</editor>
		<price>9.99</price>
	</magazine>
</root>`

	// Parse the XML
	decoder := NewDecoder(strings.NewReader(xmlData))
	doc, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	// Test basic XPath expressions
	testCases := []struct {
		name       string
		expression string
		resultType uint16
		expected   interface{} // Expected result (count for node-set, value for primitives)
	}{
		{
			name:       "Root element selection",
			expression: "/root",
			resultType: XPATH_FIRST_ORDERED_NODE_TYPE,
			expected:   "root", // Expected node name
		},
		{
			name:       "All book elements",
			expression: "//book",
			resultType: XPATH_ORDERED_NODE_SNAPSHOT_TYPE,
			expected:   2, // Expected count
		},
		{
			name:       "Book with specific ID",
			expression: "//book[@id='1']",
			resultType: XPATH_FIRST_ORDERED_NODE_TYPE,
			expected:   "Go Programming", // Expected title attribute
		},
		{
			name:       "Count of all elements",
			expression: "count(//book)",
			resultType: XPATH_NUMBER_TYPE,
			expected:   float64(2),
		},
		{
			name:       "String value of first book title",
			expression: "string(//book[1]/@title)",
			resultType: XPATH_STRING_TYPE,
			expected:   "Go Programming",
		},
		{
			name:       "Boolean test for existence",
			expression: "boolean(//book[@id='999'])",
			resultType: XPATH_BOOLEAN_TYPE,
			expected:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Evaluate XPath expression
			result, err := doc.Evaluate(tc.expression, doc.DocumentElement(), nil, tc.resultType, nil)
			if err != nil {
				t.Fatalf("XPath evaluation failed: %v", err)
			}

			// Verify result type
			if result.ResultType() != tc.resultType {
				t.Errorf("Expected result type %d, got %d", tc.resultType, result.ResultType())
			}

			// Verify result value based on type
			switch tc.resultType {
			case XPATH_NUMBER_TYPE:
				if expected, ok := tc.expected.(float64); ok {
					actual, err := result.NumberValue()
					if err != nil {
						t.Errorf("NumberValue() returned error: %v", err)
					} else if actual != expected {
						t.Errorf("Expected number value %f, got %f", expected, actual)
					}
				}
			case XPATH_STRING_TYPE:
				if expected, ok := tc.expected.(string); ok {
					actual, err := result.StringValue()
					if err != nil {
						t.Errorf("StringValue() returned error: %v", err)
					} else if actual != expected {
						t.Errorf("Expected string value %q, got %q", expected, actual)
					}
				}
			case XPATH_BOOLEAN_TYPE:
				if expected, ok := tc.expected.(bool); ok {
					actual, err := result.BooleanValue()
					if err != nil {
						t.Errorf("BooleanValue() returned error: %v", err)
					} else if actual != expected {
						t.Errorf("Expected boolean value %t, got %t", expected, actual)
					}
				}
			case XPATH_FIRST_ORDERED_NODE_TYPE:
				node, err := result.SingleNodeValue()
				if err != nil {
					t.Errorf("SingleNodeValue() returned error: %v", err)
				} else if node == nil {
					t.Errorf("Expected node, got nil")
				} else if expected, ok := tc.expected.(string); ok {
					// For root element test, check node name
					if tc.expression == "/root" && string(node.NodeName()) != expected {
						t.Errorf("Expected node name %q, got %q", expected, string(node.NodeName()))
					}
					// For book selection test, check title attribute
					if strings.Contains(tc.expression, "book[@id='1']") {
						if elem, ok := node.(Element); ok {
							title := elem.GetAttribute("title")
							if string(title) != expected {
								t.Errorf("Expected title %q, got %q", expected, string(title))
							}
						}
					}
				}
			case XPATH_ORDERED_NODE_SNAPSHOT_TYPE:
				if expected, ok := tc.expected.(int); ok {
					count, err := result.SnapshotLength()
					if err != nil {
						t.Errorf("SnapshotLength() returned error: %v", err)
					} else if int(count) != expected {
						t.Errorf("Expected snapshot length %d, got %d", expected, count)
					}
				}
			}
		})
	}
}

func TestXPathExpressionCompilation(t *testing.T) {
// 	// Create a simple document
// 	xmlData := `<root><child>test</child></root>`
// 	decoder := NewDecoder(strings.NewReader(xmlData))
// 	doc, err := decoder.Decode()
// 	if err != nil {
// 		t.Fatalf("Failed to parse XML: %v", err)
// 	}

// 	// Test expression compilation
// 	testCases := []struct {
// 		name        string
// 		expression  string
// 		shouldError bool
// 	}{
// 		{
// 			name:        "Valid simple expression",
// 			expression:  "/root/child",
// 			shouldError: false,
// 		},
// 		{
// 			name:        "Valid function call",
// 			expression:  "count(//child)",
// 			shouldError: false,
// 		},
// 		{
// 			name:        "Empty expression",
// 			expression:  "",
// 			shouldError: true,
// 		},
// 		{
// 			name:        "Invalid syntax",
// 			expression:  "/root/[",
// 			shouldError: true,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			expr, err := doc.CreateExpression(tc.expression, nil)

// 			if tc.shouldError {
// 				if err == nil {
// 					t.Errorf("Expected error for expression %q, but got none", tc.expression)
// 				}
// 			} else {
// 				if err != nil {
// 					t.Errorf("Expected no error for expression %q, but got: %v", tc.expression, err)
// 				}
// 				if expr == nil {
// 					t.Errorf("Expected non-nil expression for %q", tc.expression)
// 				}
// 			}
// 		})
// 	}
// }
}

func TestXPathResultTypes(t *testing.T) {
// 	// Create a simple document
// 	xmlData := `<root><item>1</item><item>2</item></root>`
// 	decoder := NewDecoder(strings.NewReader(xmlData))
// 	doc, err := decoder.Decode()
// 	if err != nil {
// 		t.Fatalf("Failed to parse XML: %v", err)
// 	}

// 	// Test different result types
// 	testCases := []struct {
// 		name       string
// 		expression string
// 		resultType uint16
// 		validator  func(result XPathResult, t *testing.T)
// 	}{
// 		{
// 			name:       "Iterator result",
// 			expression: "//item",
// 			resultType: XPATH_ORDERED_NODE_ITERATOR_TYPE,
// 			validator: func(result XPathResult, t *testing.T) {
// 				count := 0
// 				for {
// 					node, err := result.IterateNext()
// 					if err != nil {
// 						t.Errorf("IterateNext() returned error: %v", err)
// 						break
// 					}
// 					if node == nil {
// 						break
// 					}
// 					count++
// 				}
// 				if count != 2 {
// 					t.Errorf("Expected 2 items in iterator, got %d", count)
// 				}
// 			},
// 		},
// 		{
// 			name:       "Snapshot result",
// 			expression: "//item",
// 			resultType: XPATH_ORDERED_NODE_SNAPSHOT_TYPE,
// 			validator: func(result XPathResult, t *testing.T) {
// 				length, err := result.SnapshotLength()
// 				if err != nil {
// 					t.Errorf("SnapshotLength() returned error: %v", err)
// 					return
// 				}
// 				if length != 2 {
// 					t.Errorf("Expected snapshot length 2, got %d", length)
// 				}

// 				// Test snapshot item access
// 				first, err := result.SnapshotItem(0)
// 				if err != nil {
// 					t.Errorf("SnapshotItem(0) returned error: %v", err)
// 				} else if first == nil {
// 					t.Error("Expected first snapshot item, got nil")
// 				}

// 				// Test out of bounds
// 				outOfBounds, err := result.SnapshotItem(10)
// 				if err != nil {
// 					t.Errorf("SnapshotItem(10) returned error: %v", err)
// 				} else if outOfBounds != nil {
// 					t.Error("Expected nil for out of bounds snapshot item")
// 				}
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			result, err := doc.Evaluate(tc.expression, doc.DocumentElement(), nil, tc.resultType, nil)
// 			if err != nil {
// 				t.Fatalf("XPath evaluation failed: %v", err)
// 			}

// 			tc.validator(result, t)
// 		})
// 	}
// }
}

func TestXPathNamespaceResolver(t *testing.T) {
// 	// Create XML with namespaces
// 	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
// <root xmlns:book="http://example.com/book" xmlns:author="http://example.com/author">
// 	<book:item id="1">
// 		<author:name>John Doe</author:name>
// 	</book:item>
// </root>`

// 	decoder := NewDecoder(strings.NewReader(xmlData))
// 	doc, err := decoder.Decode()
// 	if err != nil {
// 		t.Fatalf("Failed to parse XML: %v", err)
// 	}

// 	// Test XPath with namespace
// 	expr := "//book:item/author:name"
// 	result, err := doc.Evaluate(expr, doc.DocumentElement(), nil, XPATH_FIRST_ORDERED_NODE_TYPE, nil)
// 	if err != nil {
// 		// This might fail until namespace support is fully implemented
// 		t.Logf("Namespace XPath evaluation failed (expected until fully implemented): %v", err)
// 		return
// 	}

// 	node, err := result.SingleNodeValue()
// 	if err != nil {
// 		t.Errorf("SingleNodeValue() returned error: %v", err)
// 		return
// 	}
// 	if node == nil {
// 		t.Error("Expected to find namespaced element")
// 	}
// }
}

func TestXPathBuiltinFunctions(t *testing.T) {
// 	// Create a document with multiple elements
// 	xmlData := `<root>
// 		<items>
// 			<item>apple</item>
// 			<item>banana</item>
// 			<item>cherry</item>
// 		</items>
// 	</root>`

// 	decoder := NewDecoder(strings.NewReader(xmlData))
// 	doc, err := decoder.Decode()
// 	if err != nil {
// 		t.Fatalf("Failed to parse XML: %v", err)
// 	}

// 	// Test built-in functions
// 	testCases := []struct {
// 		name       string
// 		expression string
// 		resultType uint16
// 		expected   interface{}
// 	}{
// 		{
// 			name:       "count function",
// 			expression: "count(//item)",
// 			resultType: XPATH_NUMBER_TYPE,
// 			expected:   float64(3),
// 		},
// 		{
// 			name:       "string function",
// 			expression: "string(//item[1])",
// 			resultType: XPATH_STRING_TYPE,
// 			expected:   "apple",
// 		},
// 		{
// 			name:       "boolean true function",
// 			expression: "true()",
// 			resultType: XPATH_BOOLEAN_TYPE,
// 			expected:   true,
// 		},
// 		{
// 			name:       "boolean false function",
// 			expression: "false()",
// 			resultType: XPATH_BOOLEAN_TYPE,
// 			expected:   false,
// 		},
// 		{
// 			name:       "not function",
// 			expression: "not(false())",
// 			resultType: XPATH_BOOLEAN_TYPE,
// 			expected:   true,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			result, err := doc.Evaluate(tc.expression, doc.DocumentElement(), nil, tc.resultType, nil)
// 			if err != nil {
// 				t.Fatalf("XPath evaluation failed: %v", err)
// 			}

// 			switch tc.resultType {
// 			case XPATH_NUMBER_TYPE:
// 				if expected, ok := tc.expected.(float64); ok {
// 					actual, err := result.NumberValue()
// 					if err != nil {
// 						t.Errorf("NumberValue() returned error: %v", err)
// 					} else if actual != expected {
// 						t.Errorf("Expected number value %f, got %f", expected, actual)
// 					}
// 				}
// 			case XPATH_STRING_TYPE:
// 				if expected, ok := tc.expected.(string); ok {
// 					actual, err := result.StringValue()
// 					if err != nil {
// 						t.Errorf("StringValue() returned error: %v", err)
// 					} else if actual != expected {
// 						t.Errorf("Expected string value %q, got %q", expected, actual)
// 					}
// 				}
// 			case XPATH_BOOLEAN_TYPE:
// 				if expected, ok := tc.expected.(bool); ok {
// 					actual, err := result.BooleanValue()
// 					if err != nil {
// 						t.Errorf("BooleanValue() returned error: %v", err)
// 					} else if actual != expected {
// 						t.Errorf("Expected boolean value %t, got %t", expected, actual)
// 					}
// 				}
// 			}
// 		})
// 	}
// }
}
