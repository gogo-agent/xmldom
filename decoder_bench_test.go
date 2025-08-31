package xmldom_test

import (
	"strings"
	"testing"

	"github.com/gogo-agent/xmldom"
)

func BenchmarkDecode_Small(b *testing.B) {
	xmlStr := `<root><child>text</child></root>`
	reader := strings.NewReader(xmlStr)
	for i := 0; i < b.N; i++ {
		reader.Seek(0, 0)
		decoder := xmldom.NewDecoder(reader)
		_, err := decoder.Decode()
		if err != nil {
			b.Fatalf("Decode() failed: %v", err)
		}
	}
}

func BenchmarkDecode_Medium(b *testing.B) {
	xmlStr := generateXML(10, 3)
	reader := strings.NewReader(xmlStr)
	for i := 0; i < b.N; i++ {
		reader.Seek(0, 0)
		decoder := xmldom.NewDecoder(reader)
		_, err := decoder.Decode()
		if err != nil {
			b.Fatalf("Decode() failed: %v", err)
		}
	}
}

func BenchmarkDecode_Large(b *testing.B) {
	xmlStr := generateXML(20, 4)
	reader := strings.NewReader(xmlStr)
	for i := 0; i < b.N; i++ {
		reader.Seek(0, 0)
		decoder := xmldom.NewDecoder(reader)
		_, err := decoder.Decode()
		if err != nil {
			b.Fatalf("Decode() failed: %v", err)
		}
	}
}

func generateXML(width, depth int) string {
	var sb strings.Builder
	sb.WriteString("<root>")
	generateXMLElement(&sb, width, depth)
	sb.WriteString("</root>")
	return sb.String()
}

func generateXMLElement(sb *strings.Builder, width, depth int) {
	if depth <= 0 {
		return
	}
	for i := 0; i < width; i++ {
		sb.WriteString("<child>")
		if depth > 1 {
			generateXMLElement(sb, width, depth-1)
		}
		sb.WriteString("</child>")
	}
}
