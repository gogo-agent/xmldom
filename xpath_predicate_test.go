package xmldom

import (
	"strings"
	"testing"
)

func TestXPathPredicates(t *testing.T) {
	// Create a test XML document with more complex structure
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<library>
	<books>
		<book id="1" category="fiction" price="29.99">
			<title>The Great Novel</title>
			<author>John Doe</author>
			<published>2020</published>
		</book>
		<book id="2" category="non-fiction" price="39.99">
			<title>Learning Go</title>
			<author>Jane Smith</author>
			<published>2021</published>
		</book>
		<book id="3" category="fiction" price="19.99">
			<title>Mystery Story</title>
			<author>Bob Wilson</author>
			<published>2019</published>
		</book>
		<book id="4" category="science" price="49.99">
			<title>Advanced Physics</title>
			<author>Alice Brown</author>
			<published>2022</published>
		</book>
	</books>
	<magazines>
		<magazine id="101" type="monthly">Tech Today</magazine>
		<magazine id="102" type="weekly">Science Weekly</magazine>
		<magazine id="103" type="monthly">Art Review</magazine>
	</magazines>
</library>`

	// Parse the XML
	decoder := NewDecoder(strings.NewReader(xmlData))
	doc, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	testCases := []struct {
		name        string
		expression  string
		resultType  uint16
		expectedLen int // For node sets
		expectedStr string // For string results
		expectedNum float64 // For number results
		expectedBool bool   // For boolean results
		description string
	}{
		// Attribute comparison predicates
		{
			name:        "Book by exact ID",
			expression:  "//book[@id='1']",
			resultType:  XPATH_ORDERED_NODE_SNAPSHOT_TYPE,
			expectedLen: 1,
			description: "Select book with id='1'",
		},
		{
			name:        "Books by category",
			expression:  "//book[@category='fiction']",
			resultType:  XPATH_ORDERED_NODE_SNAPSHOT_TYPE,
			expectedLen: 2,
			description: "Select books with category='fiction'",
		},
		{
			name:        "Magazine by type",
			expression:  "//magazine[@type='monthly']",
			resultType:  XPATH_ORDERED_NODE_SNAPSHOT_TYPE,
			expectedLen: 2,
			description: "Select magazines with type='monthly'",
		},

		// Numeric comparison predicates
		{
			name:        "Expensive books",
			expression:  "//book[@price > 30]",
			resultType:  XPATH_ORDERED_NODE_SNAPSHOT_TYPE,
			expectedLen: 2,
			description: "Select books with price > 30",
		},
		{
			name:        "Cheap books",
			expression:  "//book[@price < 25]",
			resultType:  XPATH_ORDERED_NODE_SNAPSHOT_TYPE,
			expectedLen: 1,
			description: "Select books with price < 25",
		},

		// Positional predicates
		{
			name:        "First book",
			expression:  "//book[1]",
			resultType:  XPATH_FIRST_ORDERED_NODE_TYPE,
			expectedStr: "1", // Should have id='1'
			description: "Select first book",
		},
		{
			name:        "Second book",
			expression:  "//book[2]",
			resultType:  XPATH_FIRST_ORDERED_NODE_TYPE,
			expectedStr: "2", // Should have id='2'
			description: "Select second book",
		},
		{
			name:        "Last book using position",
			expression:  "//book[4]",
			resultType:  XPATH_FIRST_ORDERED_NODE_TYPE,
			expectedStr: "4", // Should have id='4'
			description: "Select fourth (last) book",
		},
		{
			name:        "Last book using last()",
			expression:  "//book[last()]",
			resultType:  XPATH_FIRST_ORDERED_NODE_TYPE,
			expectedStr: "4", // Should have id='4'
			description: "Select last book using last() function",
		},

		// Position comparison predicates
		{
			name:        "Books after first",
			expression:  "//book[position() > 1]",
			resultType:  XPATH_ORDERED_NODE_SNAPSHOT_TYPE,
			expectedLen: 3,
			description: "Select books after first position",
		},
		{
			name:        "First two books",
			expression:  "//book[position() <= 2]",
			resultType:  XPATH_ORDERED_NODE_SNAPSHOT_TYPE,
			expectedLen: 2,
			description: "Select first two books",
		},

		// Complex predicates combining attribute and position
		{
			name:        "First fiction book",
			expression:  "//book[@category='fiction'][1]",
			resultType:  XPATH_FIRST_ORDERED_NODE_TYPE,
			expectedStr: "1", // Should be the first fiction book with id='1'
			description: "Select first book with category='fiction'",
		},
		{
			name:        "Second fiction book",
			expression:  "//book[@category='fiction'][2]",
			resultType:  XPATH_FIRST_ORDERED_NODE_TYPE,
			expectedStr: "3", // Should be the second fiction book with id='3'
			description: "Select second book with category='fiction'",
		},

		// Text content predicates
		{
			name:        "Book by author name",
			expression:  "//book[author='Jane Smith']",
			resultType:  XPATH_FIRST_ORDERED_NODE_TYPE,
			expectedStr: "2", // Should have id='2'
			description: "Select book by author name",
		},
		{
			name:        "Book by title contains",
			expression:  "//book[contains(title, 'Go')]",
			resultType:  XPATH_FIRST_ORDERED_NODE_TYPE,
			expectedStr: "2", // Should have id='2'
			description: "Select book by title containing 'Go'",
		},

		// Counting with predicates
		{
			name:        "Count fiction books",
			expression:  "count(//book[@category='fiction'])",
			resultType:  XPATH_NUMBER_TYPE,
			expectedNum: 2,
			description: "Count books with category='fiction'",
		},
		{
			name:        "Count expensive books",
			expression:  "count(//book[@price > 35])",
			resultType:  XPATH_NUMBER_TYPE,
			expectedNum: 2,
			description: "Count books with price > 35",
		},

		// Boolean predicates
		{
			name:        "Has fiction books",
			expression:  "boolean(//book[@category='fiction'])",
			resultType:  XPATH_BOOLEAN_TYPE,
			expectedBool: true,
			description: "Check if fiction books exist",
		},
		{
			name:        "Has romance books",
			expression:  "boolean(//book[@category='romance'])",
			resultType:  XPATH_BOOLEAN_TYPE,
			expectedBool: false,
			description: "Check if romance books exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Evaluate XPath expression
			result, err := doc.Evaluate(tc.expression, doc.DocumentElement(), nil, tc.resultType, nil)
			if err != nil {
				t.Fatalf("XPath evaluation failed for %q: %v", tc.expression, err)
			}

			// Verify result type
			if result.ResultType() != tc.resultType {
				t.Errorf("Expected result type %d, got %d", tc.resultType, result.ResultType())
				return
			}

			// Verify result value based on type
			switch tc.resultType {
			case XPATH_ORDERED_NODE_SNAPSHOT_TYPE:
				count, err := result.SnapshotLength()
				if err != nil {
					t.Errorf("SnapshotLength() returned error: %v", err)
				} else if int(count) != tc.expectedLen {
					t.Errorf("Expected %d nodes, got %d for expression %q", tc.expectedLen, count, tc.expression)
				}

			case XPATH_FIRST_ORDERED_NODE_TYPE:
				node, err := result.SingleNodeValue()
				if err != nil {
					t.Errorf("SingleNodeValue() returned error: %v", err)
				} else if node == nil {
					t.Errorf("Expected node, got nil for expression %q", tc.expression)
				} else if tc.expectedStr != "" {
					// Check id attribute to verify we got the right node
					if elem, ok := node.(Element); ok {
						id := elem.GetAttribute("id")
						if string(id) != tc.expectedStr {
							t.Errorf("Expected node with id=%q, got id=%q for expression %q", tc.expectedStr, string(id), tc.expression)
						}
					}
				}

			case XPATH_NUMBER_TYPE:
				actual, err := result.NumberValue()
				if err != nil {
					t.Errorf("NumberValue() returned error: %v", err)
				} else if actual != tc.expectedNum {
					t.Errorf("Expected number value %f, got %f for expression %q", tc.expectedNum, actual, tc.expression)
				}

			case XPATH_BOOLEAN_TYPE:
				actual, err := result.BooleanValue()
				if err != nil {
					t.Errorf("BooleanValue() returned error: %v", err)
				} else if actual != tc.expectedBool {
					t.Errorf("Expected boolean value %t, got %t for expression %q", tc.expectedBool, actual, tc.expression)
				}
			}
		})
	}
}

func TestXPathPredicateEdgeCases(t *testing.T) {
	// Test edge cases for predicate evaluation
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<root>
	<items>
		<item value="0">Zero</item>
		<item value="">Empty</item>
		<item value="false">False String</item>
		<item>No Value Attribute</item>
	</items>
</root>`

	decoder := NewDecoder(strings.NewReader(xmlData))
	doc, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	testCases := []struct {
		name        string
		expression  string
		expectedLen int
		description string
	}{
		{
			name:        "Empty string attribute",
			expression:  "//item[@value='']",
			expectedLen: 1,
			description: "Should match item with empty value attribute",
		},
		{
			name:        "Zero string attribute",
			expression:  "//item[@value='0']",
			expectedLen: 1,
			description: "Should match item with '0' value attribute",
		},
		{
			name:        "Non-existent attribute",
			expression:  "//item[@missing]",
			expectedLen: 0,
			description: "Should not match items without the attribute",
		},
		{
			name:        "Out of range position",
			expression:  "//item[10]",
			expectedLen: 0,
			description: "Should not match anything for position beyond available nodes",
		},
		{
			name:        "Zero position",
			expression:  "//item[0]",
			expectedLen: 0,
			description: "Should not match anything for position 0 (XPath uses 1-based indexing)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := doc.Evaluate(tc.expression, doc.DocumentElement(), nil, XPATH_ORDERED_NODE_SNAPSHOT_TYPE, nil)
			if err != nil {
				t.Fatalf("XPath evaluation failed for %q: %v", tc.expression, err)
			}

			count, err := result.SnapshotLength()
			if err != nil {
				t.Errorf("SnapshotLength() returned error: %v", err)
			} else if int(count) != tc.expectedLen {
				t.Errorf("Expected %d nodes, got %d for expression %q (%s)", tc.expectedLen, count, tc.expression, tc.description)
			}
		})
	}
}
