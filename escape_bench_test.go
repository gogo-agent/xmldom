package xmldom_test

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/gogo-agent/xmldom"
)

// Benchmark data
var (
	benchSimple       = "Hello world!"
	benchWithSpecial  = `Hello <world> & "friends" 'everyone'!`
	benchMostlyText   = "This is a long text with just one < special character in the middle of lots of normal text that should be fast to process"
	benchManySpecials = strings.Repeat(`<>&"'`, 100)
	benchLargeText    = strings.Repeat("Lorem ipsum dolor sit amet, consectetur adipiscing elit. ", 1000)
	benchLargeMixed   = strings.Repeat(`Hello <world> & "friends" 'everyone'! `, 1000)
)

// Benchmarkxmldom.EscapeText_Simple benchmarks escaping simple text without special characters
func BenchmarkEscapeText_Simple(b *testing.B) {
	input := []byte(benchSimple)
	var buf bytes.Buffer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		xmldom.EscapeText(&buf, input)
	}
}

// Benchmarkxmldom.EscapeText_WithSpecial benchmarks escaping text with some special characters
func BenchmarkEscapeText_WithSpecial(b *testing.B) {
	input := []byte(benchWithSpecial)
	var buf bytes.Buffer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		xmldom.EscapeText(&buf, input)
	}
}

// Benchmarkxmldom.EscapeText_MostlyText benchmarks escaping mostly normal text
func BenchmarkEscapeText_MostlyText(b *testing.B) {
	input := []byte(benchMostlyText)
	var buf bytes.Buffer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		xmldom.EscapeText(&buf, input)
	}
}

// Benchmarkxmldom.EscapeText_ManySpecials benchmarks escaping text with many special characters
func BenchmarkEscapeText_ManySpecials(b *testing.B) {
	input := []byte(benchManySpecials)
	var buf bytes.Buffer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		xmldom.EscapeText(&buf, input)
	}
}

// Benchmarkxmldom.EscapeText_LargeText benchmarks escaping large amount of normal text
func BenchmarkEscapeText_LargeText(b *testing.B) {
	input := []byte(benchLargeText)
	var buf bytes.Buffer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		xmldom.EscapeText(&buf, input)
	}
}

// Benchmarkxmldom.EscapeText_LargeMixed benchmarks escaping large mixed content
func BenchmarkEscapeText_LargeMixed(b *testing.B) {
	input := []byte(benchLargeMixed)
	var buf bytes.Buffer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		xmldom.EscapeText(&buf, input)
	}
}

// BenchmarkEscapeString_Simple benchmarks the string convenience function
func BenchmarkEscapeString_Simple(b *testing.B) {
	input := benchSimple
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		xmldom.EscapeString(input)
	}
}

// BenchmarkEscapeString_WithSpecial benchmarks the string convenience function with specials
func BenchmarkEscapeString_WithSpecial(b *testing.B) {
	input := benchWithSpecial
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		xmldom.EscapeString(input)
	}
}

// Comparison benchmarks with encoding/xml

// BenchmarkXMLEscapeText_Simple benchmarks encoding/xml.EscapeText with simple text
func BenchmarkXMLEscapeText_Simple(b *testing.B) {
	input := []byte(benchSimple)
	var buf bytes.Buffer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		xml.EscapeText(&buf, input)
	}
}

// BenchmarkXMLEscapeText_WithSpecial benchmarks encoding/xml.EscapeText with special characters
func BenchmarkXMLEscapeText_WithSpecial(b *testing.B) {
	input := []byte(benchWithSpecial)
	var buf bytes.Buffer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		xml.EscapeText(&buf, input)
	}
}

// BenchmarkXMLEscapeText_MostlyText benchmarks encoding/xml.EscapeText with mostly normal text
func BenchmarkXMLEscapeText_MostlyText(b *testing.B) {
	input := []byte(benchMostlyText)
	var buf bytes.Buffer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		xml.EscapeText(&buf, input)
	}
}

// BenchmarkXMLEscapeText_ManySpecials benchmarks encoding/xml.EscapeText with many special characters
func BenchmarkXMLEscapeText_ManySpecials(b *testing.B) {
	input := []byte(benchManySpecials)
	var buf bytes.Buffer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		xml.EscapeText(&buf, input)
	}
}

// BenchmarkXMLEscapeText_LargeText benchmarks encoding/xml.EscapeText with large normal text
func BenchmarkXMLEscapeText_LargeText(b *testing.B) {
	input := []byte(benchLargeText)
	var buf bytes.Buffer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		xml.EscapeText(&buf, input)
	}
}

// BenchmarkXMLEscapeText_LargeMixed benchmarks encoding/xml.EscapeText with large mixed content
func BenchmarkXMLEscapeText_LargeMixed(b *testing.B) {
	input := []byte(benchLargeMixed)
	var buf bytes.Buffer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		xml.EscapeText(&buf, input)
	}
}

// Unescape benchmarks

// Benchmarkxmldom.UnescapeText_Simple benchmarks unescaping simple text
func BenchmarkUnescapeText_Simple(b *testing.B) {
	input := []byte("Hello world!")
	var buf bytes.Buffer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		xmldom.UnescapeText(&buf, input)
	}
}

// Benchmarkxmldom.UnescapeText_WithEntities benchmarks unescaping text with entities
func BenchmarkUnescapeText_WithEntities(b *testing.B) {
	input := []byte("Hello &lt;world&gt; &amp; &quot;friends&quot; &apos;everyone&apos;!")
	var buf bytes.Buffer
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		xmldom.UnescapeText(&buf, input)
	}
}

// Benchmarkxmldom.UnescapeString_WithEntities benchmarks the string convenience function for unescaping
func BenchmarkUnescapeString_WithEntities(b *testing.B) {
	input := "Hello &lt;world&gt; &amp; &quot;friends&quot; &apos;everyone&apos;!"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		xmldom.UnescapeString(input)
	}
}

// Round-trip benchmarks

// BenchmarkRoundTrip_EscapeUnescape benchmarks a complete round trip
func BenchmarkRoundTrip_EscapeUnescape(b *testing.B) {
	input := benchWithSpecial
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		escaped := xmldom.EscapeString(input)
		xmldom.UnescapeString(escaped)
	}
}
