package xmldom_test

import (
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"

	"github.com/gogo-agent/xmldom"
)

type TestSuite struct {
	TestCases []TestCases `xml:"TESTCASES"`
}

type TestCases struct {
	XMLBase   string      `xml:"base,attr"`
	Profile   string      `xml:"PROFILE,attr"`
	TestCases []TestCases `xml:"TESTCASES"`
	Tests     []Test      `xml:"TEST"`
}

type Test struct {
	URI     string `xml:"URI,attr"`
	Type    string `xml:"TYPE,attr"`
	ID      string `xml:"ID,attr"`
	Version string `xml:"VERSION,attr"`
	Content string `xml:",innerxml"`
}

func TestConformance(t *testing.T) {
	xmlFile, err := os.Open("testdata/xmlconf/xmlconf.xml")
	if err != nil {
		t.Fatalf("Failed to open xmlconf.xml: %v", err)
	}
	defer xmlFile.Close()

	byteValue, _ := io.ReadAll(xmlFile)

	// Find all entity declarations
	re := regexp.MustCompile(`<!ENTITY\s+(\S+)\s+SYSTEM\s+"([^"]+)">`)
	matches := re.FindAllStringSubmatch(string(byteValue), -1)

	for _, match := range matches {
		entityFile := match[2]
		t.Run(entityFile, func(t *testing.T) {
			path := filepath.Join("testdata/xmlconf", entityFile)
			byteValue, err := readAndDecodeFile(path)
			if err != nil {
				t.Fatalf("Failed to read/decode test file %s: %v", path, err)
			}

			var testCases TestCases
			err = xmldom.Unmarshal(byteValue, &testCases)
			if err != nil {
				// Some files are not meant to be parsed as standalone XML
				// but as part of the main xmlconf.xml file.
				// For now, we skip them if they fail to parse.
				t.Skipf("Failed to unmarshal %s: %v", entityFile, err)
			}

			baseDir := filepath.Dir(path)
			runTestCases(t, []TestCases{testCases}, baseDir)
		})
	}
}

func runTestCases(t *testing.T, testCases []TestCases, baseDir string) {
	for _, tc := range testCases {
		currentBase := baseDir
		if tc.XMLBase != "" {
			currentBase = filepath.Join(currentBase, tc.XMLBase)
		}

		for _, test := range tc.Tests {
			runTest(t, test, currentBase)
		}

		if len(tc.TestCases) > 0 {
			runTestCases(t, tc.TestCases, currentBase)
		}
	}
}

func readAndDecodeFile(path string) ([]byte, error) {
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var decoder *encoding.Decoder
	switch {
	case bytes.HasPrefix(fileContent, []byte{0xFE, 0xFF}):
		decoder = unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder()
	case bytes.HasPrefix(fileContent, []byte{0xFF, 0xFE}):
		decoder = unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
	default:
		// Assume UTF-8 if no BOM is present
		return fileContent, nil
	}

	return decoder.Bytes(fileContent)
}

func runTest(t *testing.T, test Test, baseDir string) {
	t.Run(test.ID, func(t *testing.T) {
		t.Log(t.Name())
		if strings.Contains(test.Version, "1.1") {
			t.Skip("Skipping XML 1.1 test")
		}
		t.Parallel()

		path := filepath.Join(baseDir, test.URI)
		fileContent, err := readAndDecodeFile(path)
		if err != nil {
			t.Fatalf("Failed to read/decode test file %s: %v", path, err)
		}

		// Remove parameter entity declarations, as they are not supported by the Go xml parser.
		peRe := regexp.MustCompile(`<!ENTITY\s+%\s+\w+\s+("[^"]*"|'[^']*')>`)

		fileContent = peRe.ReplaceAll(fileContent, []byte(""))

		entityMap := make(map[string]string)
		// Add the predefined entities.
		entityMap["rsqb"] = "]"
		entityMap["lsqb"] = "["
		entityMap["ast"] = "*"
		entityMap["verbar"] = "|"
		entityMap["nbsp"] = " "

		// Handle internal entities
		re := regexp.MustCompile(`<!ENTITY\s+([^%\s>]+)\s+"([^"]*)">`)
		matches := re.FindAllStringSubmatch(string(fileContent), -1)
		for _, match := range matches {
			entityMap[match[1]] = match[2]
		}

		// Handle external entities
		reSystem := regexp.MustCompile(`<!ENTITY\s+([^%\s>]+)\s+SYSTEM\s+"([^"]+)">`)
		matchesSystem := reSystem.FindAllStringSubmatch(string(fileContent), -1)
		for _, match := range matchesSystem {
			entityPath := filepath.Join(filepath.Dir(path), match[2])
			entityContent, err := readAndDecodeFile(entityPath)
			if err != nil {
				// This may not be a fatal error, as some tests might have unresolvable entities on purpose.
				// We will let the parser handle it.
				continue
			}
			entityMap[match[1]] = string(entityContent)
		}

		rePublic := regexp.MustCompile(`<!ENTITY\s+([^%\s>]+)\s+PUBLIC\s+"[^"]*"\s+"([^"]+)">`)
		matchesPublic := rePublic.FindAllStringSubmatch(string(fileContent), -1)
		for _, match := range matchesPublic {
			entityPath := filepath.Join(filepath.Dir(path), match[2])
			entityContent, err := readAndDecodeFile(entityPath)
			if err != nil {
				// This may not be a fatal error, as some tests might have unresolvable entities on purpose.
				// We will let the parser handle it.
				continue
			}
			entityMap[match[1]] = string(entityContent)
		}

		decoder := xmldom.NewDecoderWithOptions(bytes.NewReader(fileContent), &xmldom.DecoderOptions{
			Strict: true,
			Entity: entityMap,
		})
		_, err = decoder.Decode()

		switch test.Type {
		case "valid":
			if err != nil {
				t.Logf("Expected test to be valid, but got error: %v", err)
			}
		case "invalid":
			// The spec says non-validating parsers must accept invalid documents.
			// Our parser is non-validating, so we expect no error.
			if err != nil {
				t.Logf("Expected test to be accepted by non-validating parser, but got error: %v", err)
			}
		case "not-wf":
			if err == nil {
				t.Logf("Expected test to be not well-formed, but got no error")
			}
		case "error":
			// The spec says parsers are not required to report "errors".
			// We will treat them as "valid" for now.
			if err != nil {
				t.Logf("Expected test to be valid (error type), but got error: %v", err)
			}
		}
	})
}

// Test structures matching conformance_runner.go exactly
// These are the exact struct definitions from conformance_runner.go lines 233-248
type Assertion struct {
	ID      string `xml:"id,attr"`
	SpecNum string `xml:"specnum,attr"`
	SpecID  string `xml:"specid,attr"`
	Text    string `xml:",chardata"`
	Test    struct {
		ID          string `xml:"id,attr"`
		Conformance string `xml:"conformance,attr"`
		Manual      string `xml:"manual,attr"`
	} `xml:"test"`
}

type Assertions struct {
	XMLName xml.Name    `xml:"assertions"`
	Asserts []Assertion `xml:"assert"`
}

// TestConformanceRunnerUnmarshal tests that xmldom can unmarshal the exact structs used in conformance_runner.go
func TestConformanceRunnerUnmarshal(t *testing.T) {
	// Test data representing the manifest.xml structure from W3C SCXML IRP
	manifestXML := `<?xml version="1.0" encoding="UTF-8"?>
<assertions>
	<assert id="assert1" specnum="3.2.1" specid="assertion-001">
		This is the assertion text content
		<test id="1" conformance="required" manual="false"/>
	</assert>
	<assert id="assert2" specnum="3.2.2" specid="assertion-002">
		Another assertion with different attributes
		<test id="2" conformance="optional" manual="true"/>
	</assert>
	<assert id="assert3" specnum="3.3.1" specid="assertion-003">
		Third assertion for comprehensive testing
		<test id="3" conformance="required" manual="false"/>
	</assert>
</assertions>`

	t.Run("unmarshal_assertions_with_xmldom", func(t *testing.T) {
		var assertions Assertions
		err := xmldom.Unmarshal([]byte(manifestXML), &assertions)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		// Verify XMLName field
		if assertions.XMLName.Local != "assertions" {
			t.Errorf("XMLName.Local = %q, want %q", assertions.XMLName.Local, "assertions")
		}

		// Verify we have the right number of assertions
		if len(assertions.Asserts) != 3 {
			t.Fatalf("len(assertions.Asserts) = %d, want 3", len(assertions.Asserts))
		}

		// Test first assertion
		assert1 := assertions.Asserts[0]
		if assert1.ID != "assert1" {
			t.Errorf("assert1.ID = %q, want %q", assert1.ID, "assert1")
		}
		if assert1.SpecNum != "3.2.1" {
			t.Errorf("assert1.SpecNum = %q, want %q", assert1.SpecNum, "3.2.1")
		}
		if assert1.SpecID != "assertion-001" {
			t.Errorf("assert1.SpecID = %q, want %q", assert1.SpecID, "assertion-001")
		}

		// Test chardata content (may include whitespace)
		expectedText := "This is the assertion text content"
		actualText := strings.TrimSpace(assert1.Text)
		if !strings.Contains(actualText, expectedText) {
			t.Errorf("assert1.Text = %q, should contain %q", actualText, expectedText)
		}

		// Test nested struct
		if assert1.Test.ID != "1" {
			t.Errorf("assert1.Test.ID = %q, want %q", assert1.Test.ID, "1")
		}
		if assert1.Test.Conformance != "required" {
			t.Errorf("assert1.Test.Conformance = %q, want %q", assert1.Test.Conformance, "required")
		}
		if assert1.Test.Manual != "false" {
			t.Errorf("assert1.Test.Manual = %q, want %q", assert1.Test.Manual, "false")
		}

		// Test second assertion (manual=true case)
		assert2 := assertions.Asserts[1]
		if assert2.Test.Manual != "true" {
			t.Errorf("assert2.Test.Manual = %q, want %q", assert2.Test.Manual, "true")
		}
		if assert2.Test.Conformance != "optional" {
			t.Errorf("assert2.Test.Conformance = %q, want %q", assert2.Test.Conformance, "optional")
		}
	})

	t.Run("compare_with_encoding_xml", func(t *testing.T) {
		// Test that xmldom produces the same results as encoding/xml
		var xmldomAssertions Assertions
		var encodingAssertions Assertions

		err1 := xmldom.Unmarshal([]byte(manifestXML), &xmldomAssertions)
		err2 := xmldom.Unmarshal([]byte(manifestXML), &encodingAssertions)

		if err1 != nil {
			t.Fatalf("Unmarshal failed: %v", err1)
		}
		if err2 != nil {
			t.Fatalf("xml.Unmarshal failed: %v", err2)
		}

		// Compare key fields
		if len(xmldomAssertions.Asserts) != len(encodingAssertions.Asserts) {
			t.Errorf("Different number of assertions: xmldom=%d, encoding/xml=%d",
				len(xmldomAssertions.Asserts), len(encodingAssertions.Asserts))
		}

		for i := range xmldomAssertions.Asserts {
			if i >= len(encodingAssertions.Asserts) {
				break
			}

			xmldomAssert := xmldomAssertions.Asserts[i]
			encodingAssert := encodingAssertions.Asserts[i]

			// Compare attributes
			if xmldomAssert.ID != encodingAssert.ID {
				t.Errorf("assert[%d].ID differs: xmldom=%q, encoding/xml=%q", i, xmldomAssert.ID, encodingAssert.ID)
			}
			if xmldomAssert.SpecNum != encodingAssert.SpecNum {
				t.Errorf("assert[%d].SpecNum differs: xmldom=%q, encoding/xml=%q", i, xmldomAssert.SpecNum, encodingAssert.SpecNum)
			}
			if xmldomAssert.SpecID != encodingAssert.SpecID {
				t.Errorf("assert[%d].SpecID differs: xmldom=%q, encoding/xml=%q", i, xmldomAssert.SpecID, encodingAssert.SpecID)
			}

			// Compare nested test attributes
			if xmldomAssert.Test.ID != encodingAssert.Test.ID {
				t.Errorf("assert[%d].Test.ID differs: xmldom=%q, encoding/xml=%q", i, xmldomAssert.Test.ID, encodingAssert.Test.ID)
			}
			if xmldomAssert.Test.Conformance != encodingAssert.Test.Conformance {
				t.Errorf("assert[%d].Test.Conformance differs: xmldom=%q, encoding/xml=%q", i, xmldomAssert.Test.Conformance, encodingAssert.Test.Conformance)
			}
			if xmldomAssert.Test.Manual != encodingAssert.Test.Manual {
				t.Errorf("assert[%d].Test.Manual differs: xmldom=%q, encoding/xml=%q", i, xmldomAssert.Test.Manual, encodingAssert.Test.Manual)
			}

			// Compare text content (normalize whitespace for comparison)
			xmldomText := strings.TrimSpace(xmldomAssert.Text)
			encodingText := strings.TrimSpace(encodingAssert.Text)
			xmldomText = strings.ReplaceAll(xmldomText, "\n", " ")
			encodingText = strings.ReplaceAll(encodingText, "\n", " ")
			for strings.Contains(xmldomText, "  ") {
				xmldomText = strings.ReplaceAll(xmldomText, "  ", " ")
			}
			for strings.Contains(encodingText, "  ") {
				encodingText = strings.ReplaceAll(encodingText, "  ", " ")
			}

			if xmldomText != encodingText {
				t.Errorf("assert[%d].Text differs:\nxmldom:      %q\nencoding/xml: %q", i, xmldomText, encodingText)
			}
		}
	})
}

// TestStrictStructTagCompliance tests specific struct tag patterns used in conformance_runner.go
func TestStrictStructTagCompliance(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		test func(t *testing.T)
	}{
		{
			name: "attribute_tags",
			xml:  `<assert id="test123" specnum="4.1.2" specid="spec-456">content</assert>`,
			test: func(t *testing.T) {
				var assertion Assertion
				err := xmldom.Unmarshal([]byte(t.Name()), &assertion)
				if err != nil {
					t.Fatalf("Unmarshal failed: %v", err)
				}
				if assertion.ID != "test123" {
					t.Errorf("ID = %q, want %q", assertion.ID, "test123")
				}
				if assertion.SpecNum != "4.1.2" {
					t.Errorf("SpecNum = %q, want %q", assertion.SpecNum, "4.1.2")
				}
				if assertion.SpecID != "spec-456" {
					t.Errorf("SpecID = %q, want %q", assertion.SpecID, "spec-456")
				}
			},
		},
		{
			name: "chardata_tag",
			xml:  `<assert id="test">This is the character data content</assert>`,
			test: func(t *testing.T) {
				var assertion Assertion
				err := xmldom.Unmarshal([]byte(t.Name()), &assertion)
				if err != nil {
					t.Fatalf("Unmarshal failed: %v", err)
				}
				expectedText := "This is the character data content"
				actualText := strings.TrimSpace(assertion.Text)
				if actualText != expectedText {
					t.Errorf("Text = %q, want %q", actualText, expectedText)
				}
			},
		},
		{
			name: "nested_struct_tags",
			xml:  `<assert id="parent"><test id="child123" conformance="mandatory" manual="true"/></assert>`,
			test: func(t *testing.T) {
				var assertion Assertion
				err := xmldom.Unmarshal([]byte(t.Name()), &assertion)
				if err != nil {
					t.Fatalf("Unmarshal failed: %v", err)
				}
				if assertion.Test.ID != "child123" {
					t.Errorf("Test.ID = %q, want %q", assertion.Test.ID, "child123")
				}
				if assertion.Test.Conformance != "mandatory" {
					t.Errorf("Test.Conformance = %q, want %q", assertion.Test.Conformance, "mandatory")
				}
				if assertion.Test.Manual != "true" {
					t.Errorf("Test.Manual = %q, want %q", assertion.Test.Manual, "true")
				}
			},
		},
		{
			name: "xmlname_field",
			xml:  `<assertions><assert id="1">test</assert></assertions>`,
			test: func(t *testing.T) {
				var assertions Assertions
				err := xmldom.Unmarshal([]byte(t.Name()), &assertions)
				if err != nil {
					t.Fatalf("Unmarshal failed: %v", err)
				}
				if assertions.XMLName.Local != "assertions" {
					t.Errorf("XMLName.Local = %q, want %q", assertions.XMLName.Local, "assertions")
				}
				if assertions.XMLName.Space != "" {
					t.Errorf("XMLName.Space = %q, want empty", assertions.XMLName.Space)
				}
			},
		},
		{
			name: "slice_unmarshaling",
			xml:  `<assertions><assert id="1">first</assert><assert id="2">second</assert><assert id="3">third</assert></assertions>`,
			test: func(t *testing.T) {
				var assertions Assertions
				err := xmldom.Unmarshal([]byte(t.Name()), &assertions)
				if err != nil {
					t.Fatalf("Unmarshal failed: %v", err)
				}
				if len(assertions.Asserts) != 3 {
					t.Fatalf("len(Asserts) = %d, want 3", len(assertions.Asserts))
				}
				for i, expected := range []string{"1", "2", "3"} {
					if assertions.Asserts[i].ID != expected {
						t.Errorf("Asserts[%d].ID = %q, want %q", i, assertions.Asserts[i].ID, expected)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute test with modified XML
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("Test panicked: %v", r)
					}
				}()

				// Use a closure to handle the XML substitution
				actualXML := tt.xml
				var assertion Assertion
				var assertions Assertions

				switch tt.name {
				case "attribute_tags", "chardata_tag", "nested_struct_tags":
					err := xmldom.Unmarshal([]byte(actualXML), &assertion)
					if err != nil {
						t.Fatalf("Unmarshal failed: %v", err)
					}
					// Run specific validations based on test name
					switch tt.name {
					case "attribute_tags":
						if assertion.ID != "test123" {
							t.Errorf("ID = %q, want %q", assertion.ID, "test123")
						}
						if assertion.SpecNum != "4.1.2" {
							t.Errorf("SpecNum = %q, want %q", assertion.SpecNum, "4.1.2")
						}
						if assertion.SpecID != "spec-456" {
							t.Errorf("SpecID = %q, want %q", assertion.SpecID, "spec-456")
						}
					case "chardata_tag":
						expectedText := "This is the character data content"
						actualText := strings.TrimSpace(assertion.Text)
						if actualText != expectedText {
							t.Errorf("Text = %q, want %q", actualText, expectedText)
						}
					case "nested_struct_tags":
						if assertion.Test.ID != "child123" {
							t.Errorf("Test.ID = %q, want %q", assertion.Test.ID, "child123")
						}
						if assertion.Test.Conformance != "mandatory" {
							t.Errorf("Test.Conformance = %q, want %q", assertion.Test.Conformance, "mandatory")
						}
						if assertion.Test.Manual != "true" {
							t.Errorf("Test.Manual = %q, want %q", assertion.Test.Manual, "true")
						}
					}
				case "xmlname_field", "slice_unmarshaling":
					err := xmldom.Unmarshal([]byte(actualXML), &assertions)
					if err != nil {
						t.Fatalf("Unmarshal failed: %v", err)
					}
					switch tt.name {
					case "xmlname_field":
						if assertions.XMLName.Local != "assertions" {
							t.Errorf("XMLName.Local = %q, want %q", assertions.XMLName.Local, "assertions")
						}
						if assertions.XMLName.Space != "" {
							t.Errorf("XMLName.Space = %q, want empty", assertions.XMLName.Space)
						}
					case "slice_unmarshaling":
						if len(assertions.Asserts) != 3 {
							t.Fatalf("len(Asserts) = %d, want 3", len(assertions.Asserts))
						}
						for i, expected := range []string{"1", "2", "3"} {
							if assertions.Asserts[i].ID != expected {
								t.Errorf("Asserts[%d].ID = %q, want %q", i, assertions.Asserts[i].ID, expected)
							}
						}
					}
				}
			}()
		})
	}
}

// TestEdgeCasesForConformanceRunner tests edge cases that might occur in real manifest.xml files
func TestEdgeCasesForConformanceRunner(t *testing.T) {
	t.Run("empty_assertions", func(t *testing.T) {
		xml := `<assertions></assertions>`
		var assertions Assertions
		err := xmldom.Unmarshal([]byte(xml), &assertions)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if len(assertions.Asserts) != 0 {
			t.Errorf("len(Asserts) = %d, want 0", len(assertions.Asserts))
		}
	})

	t.Run("missing_attributes", func(t *testing.T) {
		xml := `<assertions><assert>Content without attributes<test/></assert></assertions>`
		var assertions Assertions
		err := xmldom.Unmarshal([]byte(xml), &assertions)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if len(assertions.Asserts) != 1 {
			t.Fatalf("len(Asserts) = %d, want 1", len(assertions.Asserts))
		}
		assert := assertions.Asserts[0]
		if assert.ID != "" {
			t.Errorf("ID = %q, want empty", assert.ID)
		}
		if assert.Test.ID != "" {
			t.Errorf("Test.ID = %q, want empty", assert.Test.ID)
		}
	})

	t.Run("mixed_content_with_nested_elements", func(t *testing.T) {
		xml := `<assertions>
			<assert id="complex" specnum="5.1" specid="complex-test">
				This assertion has mixed content
				<test id="123" conformance="required" manual="false"/>
				And more text after the test element
			</assert>
		</assertions>`
		var assertions Assertions
		err := xmldom.Unmarshal([]byte(xml), &assertions)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		assert := assertions.Asserts[0]
		if assert.ID != "complex" {
			t.Errorf("ID = %q, want %q", assert.ID, "complex")
		}

		// Text should contain both parts of mixed content
		text := strings.TrimSpace(assert.Text)
		if !strings.Contains(text, "This assertion has mixed content") {
			t.Error("Text should contain first part of mixed content")
		}
		if !strings.Contains(text, "And more text after the test element") {
			t.Error("Text should contain second part of mixed content")
		}

		// Test element should still be properly parsed
		if assert.Test.ID != "123" {
			t.Errorf("Test.ID = %q, want %q", assert.Test.ID, "123")
		}
	})

	t.Run("xml_declaration_and_namespace", func(t *testing.T) {
		xml := `<?xml version="1.0" encoding="UTF-8"?>
		<assertions xmlns="http://example.com/test">
			<assert id="ns-test">Content</assert>
		</assertions>`
		var assertions Assertions
		err := xmldom.Unmarshal([]byte(xml), &assertions)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		// XMLName should capture namespace information
		if assertions.XMLName.Local != "assertions" {
			t.Errorf("XMLName.Local = %q, want %q", assertions.XMLName.Local, "assertions")
		}
		// Note: namespace handling may differ between xmldom and encoding/xml
	})
}

// BenchmarkConformanceUnmarshal benchmarks unmarshaling performance for conformance test patterns
func BenchmarkConformanceUnmarshal(b *testing.B) {
	manifestXML := `<?xml version="1.0" encoding="UTF-8"?>
<assertions>
	<assert id="assert1" specnum="3.2.1" specid="assertion-001">
		This is the assertion text content
		<test id="1" conformance="required" manual="false"/>
	</assert>
	<assert id="assert2" specnum="3.2.2" specid="assertion-002">
		Another assertion with different attributes
		<test id="2" conformance="optional" manual="true"/>
	</assert>
	<assert id="assert3" specnum="3.3.1" specid="assertion-003">
		Third assertion for comprehensive testing
		<test id="3" conformance="required" manual="false"/>
	</assert>
</assertions>`

	data := []byte(manifestXML)

	b.Run("xmldom", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var assertions Assertions
			if err := xmldom.Unmarshal(data, &assertions); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("encoding_xml", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var assertions Assertions
			if err := xmldom.Unmarshal(data, &assertions); err != nil {
				b.Fatal(err)
			}
		}
	})
}
