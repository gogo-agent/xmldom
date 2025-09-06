package xmldom

import (
	"fmt"
	"strings"
	"testing"
)

// BenchmarkXPathParsing benchmarks XPath expression parsing without cache
func BenchmarkXPathParsing(b *testing.B) {
	expressions := []string{
		"//book[@id='1']",
		"/root/child::element[@attr='value']",
		"//node()[position() > 3 and position() < 10]",
		"//item[last() - 1]",
		"//*[@id='test' or @class='example']",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expr := expressions[i%len(expressions)]
		parser := NewXPathParser()
		_, err := parser.Parse(expr)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkXPathParsingWithCache benchmarks XPath expression parsing with cache
func BenchmarkXPathParsingWithCache(b *testing.B) {
	// Pre-warm the cache
	expressions := []string{
		"//book[@id='1']",
		"/root/child::element[@attr='value']",
		"//node()[position() > 3 and position() < 10]",
		"//item[last() - 1]",
		"//*[@id='test' or @class='example']",
	}

	for _, expr := range expressions {
		parser := NewXPathParser()
		ast, _ := parser.Parse(expr)
		setCachedExpression(expr, ast)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expr := expressions[i%len(expressions)]
		if _, found := getCachedExpression(expr); !found {
			parser := NewXPathParser()
			ast, err := parser.Parse(expr)
			if err != nil {
				b.Fatal(err)
			}
			setCachedExpression(expr, ast)
		}
	}
}

// BenchmarkXPathEvaluation benchmarks full XPath evaluation
func BenchmarkXPathEvaluation(b *testing.B) {
	xml := `<?xml version="1.0"?>
	<library>
		<book id="1" genre="fiction">
			<title>The Great Novel</title>
			<author>John Doe</author>
			<price>29.99</price>
		</book>
		<book id="2" genre="science">
			<title>Quantum Physics</title>
			<author>Jane Smith</author>
			<price>39.99</price>
		</book>
		<book id="3" genre="fiction">
			<title>Another Story</title>
			<author>Bob Wilson</author>
			<price>24.99</price>
		</book>
	</library>`

	decoder := NewDecoder(strings.NewReader(xml))
	doc, err := decoder.Decode()
	if err != nil {
		b.Fatal(err)
	}

	expressions := []string{
		"//book[@genre='fiction']",
		"//book[price > 25]",
		"//author",
		"/library/book[position() = 2]",
		"//book[@id='1' or @id='3']",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expr := expressions[i%len(expressions)]
		xpathExpr, err := doc.CreateExpression(expr, nil)
		if err != nil {
			b.Fatal(err)
		}

		result, err := xpathExpr.Evaluate(doc.DocumentElement(), XPATH_ORDERED_NODE_SNAPSHOT_TYPE, nil)
		if err != nil {
			b.Fatal(err)
		}

		// Access results to ensure evaluation happens
		result.SnapshotLength()
	}
}

// BenchmarkXPathDocumentOrder benchmarks document order sorting
func BenchmarkXPathDocumentOrder(b *testing.B) {
	// Create a document with many nodes
	var xmlBuilder strings.Builder
	xmlBuilder.WriteString(`<?xml version="1.0"?><root>`)
	for i := 0; i < 100; i++ {
		xmlBuilder.WriteString(fmt.Sprintf(`<item id="%d">Value %d</item>`, i, i))
	}
	xmlBuilder.WriteString(`</root>`)

	decoder := NewDecoder(strings.NewReader(xmlBuilder.String()))
	doc, err := decoder.Decode()
	if err != nil {
		b.Fatal(err)
	}

	// Union expression that results in many nodes needing sorting
	expr, err := doc.CreateExpression("//item[@id > '50'] | //item[@id < '20']", nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := expr.Evaluate(doc.DocumentElement(), XPATH_ORDERED_NODE_SNAPSHOT_TYPE, nil)
		if err != nil {
			b.Fatal(err)
		}

		// Access all nodes to ensure sorting happens
		length, _ := result.SnapshotLength()
		for j := uint32(0); j < length; j++ {
			result.SnapshotItem(j)
		}
	}
}

// BenchmarkXPathComplexPredicate benchmarks complex predicate evaluation
func BenchmarkXPathComplexPredicate(b *testing.B) {
	// Create a document with nested structure
	xml := `<?xml version="1.0"?>
	<root>
		<section id="1">
			<para>First paragraph in section 1</para>
			<para>Second paragraph in section 1</para>
			<para>Third paragraph in section 1</para>
		</section>
		<section id="2">
			<para>First paragraph in section 2</para>
			<para>Second paragraph in section 2</para>
		</section>
		<section id="3">
			<para>Only paragraph in section 3</para>
		</section>
	</root>`

	decoder := NewDecoder(strings.NewReader(xml))
	doc, err := decoder.Decode()
	if err != nil {
		b.Fatal(err)
	}

	// Complex expression with multiple predicates
	expr, err := doc.CreateExpression("//section[@id > '1']/para[position() = last() or position() = 1]", nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := expr.Evaluate(doc.DocumentElement(), XPATH_ORDERED_NODE_SNAPSHOT_TYPE, nil)
		if err != nil {
			b.Fatal(err)
		}

		// Access results
		result.SnapshotLength()
	}
}

// BenchmarkXPathNamespaceAxis benchmarks namespace axis performance
func BenchmarkXPathNamespaceAxis(b *testing.B) {
	xml := `<?xml version="1.0"?>
	<root xmlns:a="http://a.com" xmlns:b="http://b.com" xmlns:c="http://c.com">
		<child xmlns:d="http://d.com">
			<grandchild xmlns:e="http://e.com"/>
		</child>
	</root>`

	decoder := NewDecoder(strings.NewReader(xml))
	doc, err := decoder.Decode()
	if err != nil {
		b.Fatal(err)
	}

	expr, err := doc.CreateExpression("//*/namespace::*", nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := expr.Evaluate(doc.DocumentElement(), XPATH_ORDERED_NODE_SNAPSHOT_TYPE, nil)
		if err != nil {
			b.Fatal(err)
		}

		// Access all namespace nodes
		length, _ := result.SnapshotLength()
		for j := uint32(0); j < length; j++ {
			result.SnapshotItem(j)
		}
	}
}

// BenchmarkXPathConcurrent benchmarks concurrent XPath evaluation
func BenchmarkXPathConcurrent(b *testing.B) {
	xml := `<?xml version="1.0"?>
	<root>
		<item>1</item>
		<item>2</item>
		<item>3</item>
	</root>`

	decoder := NewDecoder(strings.NewReader(xml))
	doc, err := decoder.Decode()
	if err != nil {
		b.Fatal(err)
	}

	expr, err := doc.CreateExpression("//item", nil)
	if err != nil {
		b.Fatal(err)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result, err := expr.Evaluate(doc.DocumentElement(), XPATH_ORDERED_NODE_SNAPSHOT_TYPE, nil)
			if err != nil {
				b.Fatal(err)
			}
			result.SnapshotLength()
		}
	})
}
