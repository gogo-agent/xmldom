# XML DOM

A high-performance, W3C-compliant XML DOM implementation for Go with XPath support and comprehensive XML processing capabilities.

## Features

* **W3C DOM Compliance**: Full implementation of W3C DOM Level 2 Core specification
* **High Performance**: Optimized for speed and memory efficiency
* **XPath Support**: Complete XPath 1.0 implementation with functions and predicates
* **Streaming Parser**: Memory-efficient streaming XML parser
* **Thread Safe**: Concurrent access support with proper synchronization
* **Comprehensive**: Full XML 1.0 specification support including namespaces
* **Testing**: Extensive test suite using W3C XML conformance tests

## Installation

```bash
go get github.com/gogo-agent/xmldom
```

## Quick Start

### Parsing XML

```go
package main

import (
    "fmt"
    "log"
    "strings"
    
    "github.com/gogo-agent/xmldom"
)

func main() {
    xml := `<?xml version="1.0"?>
    <bookstore>
        <book id="1">
            <title>Go Programming</title>
            <author>John Doe</author>
            <price>29.99</price>
        </book>
        <book id="2">
            <title>XML Processing</title>
            <author>Jane Smith</author>
            <price>39.99</price>
        </book>
    </bookstore>`
    
    // Parse XML
    doc, err := xmldom.ParseString(xml)
    if err != nil {
        log.Fatal(err)
    }
    
    // Access elements
    root := doc.DocumentElement()
    fmt.Printf("Root element: %s\n", root.TagName())
    
    // Get all book elements
    books := root.GetElementsByTagName("book")
    fmt.Printf("Found %d books\n", books.Length())
    
    // Iterate through books
    for i := 0; i < books.Length(); i++ {
        book := books.Item(i).(*xmldom.Element)
        id := book.GetAttribute("id")
        title := book.GetElementsByTagName("title").Item(0).TextContent()
        fmt.Printf("Book %s: %s\n", id, title)
    }
}
```

### Creating XML Documents

```go
// Create new document
doc := xmldom.NewDocument()

// Create root element
root := doc.CreateElement("catalog")
doc.AppendChild(root)

// Create and add book element
book := doc.CreateElement("book")
book.SetAttribute("id", "123")
root.AppendChild(book)

// Add title
title := doc.CreateElement("title")
titleText := doc.CreateTextNode("Learning XML")
title.AppendChild(titleText)
book.AppendChild(title)

// Serialize to XML
xml := doc.String()
fmt.Println(xml)
```

## XPath Support

### Basic XPath Queries

```go
// Parse document
doc, err := xmldom.ParseString(xmlContent)
if err != nil {
    log.Fatal(err)
}

// Create XPath context
ctx := xmldom.NewXPathContext(doc)

// Simple path expressions
books, err := ctx.Evaluate("//book")
if err != nil {
    log.Fatal(err)
}

// With predicates
expensiveBooks, err := ctx.Evaluate("//book[price > 30]")
if err != nil {
    log.Fatal(err)
}

// Attribute selection
bookIds, err := ctx.Evaluate("//book/@id")
if err != nil {
    log.Fatal(err)
}
```

### XPath Functions

```go
// String functions
titleCount, err := ctx.Evaluate("count(//title)")
if err != nil {
    log.Fatal(err)
}

// Text content
firstTitle, err := ctx.Evaluate("string(//book[1]/title)")
if err != nil {
    log.Fatal(err)
}

// Boolean expressions
hasBooks, err := ctx.Evaluate("boolean(//book)")
if err != nil {
    log.Fatal(err)
}

// Position-based selection
lastBook, err := ctx.Evaluate("//book[last()]")
if err != nil {
    log.Fatal(err)
}
```

### Advanced XPath

```go
// Complex predicates
result, err := ctx.Evaluate("//book[author='John Doe' and price < 40]/title")

// Axes
result, err := ctx.Evaluate("//title/following-sibling::author")

// Functions with predicates
result, err := ctx.Evaluate("//book[contains(title, 'XML')]")

// Multiple conditions
result, err := ctx.Evaluate("//book[@id='1' or @id='3']/title")
```

## DOM Manipulation

### Node Operations

```go
// Create elements
element := doc.CreateElement("section")
element.SetAttribute("class", "content")

// Create text nodes
textNode := doc.CreateTextNode("Hello, World!")

// Create comments
comment := doc.CreateComment("This is a comment")

// Create CDATA sections
cdata := doc.CreateCDATASection("Some <markup> content")

// Append nodes
parent.AppendChild(element)
element.AppendChild(textNode)
```

### Tree Manipulation

```go
// Insert nodes
parent.InsertBefore(newNode, referenceNode)

// Replace nodes
parent.ReplaceChild(newNode, oldNode)

// Remove nodes
parent.RemoveChild(childNode)

// Clone nodes
clonedNode := originalNode.CloneNode(true) // deep clone
```

### Attribute Operations

```go
element := doc.CreateElement("img")

// Set attributes
element.SetAttribute("src", "image.jpg")
element.SetAttribute("alt", "Description")

// Get attributes
src := element.GetAttribute("src")
hasAlt := element.HasAttribute("alt")

// Remove attributes
element.RemoveAttribute("alt")

// Get all attributes
attrs := element.Attributes()
for i := 0; i < attrs.Length(); i++ {
    attr := attrs.Item(i).(*xmldom.Attr)
    fmt.Printf("%s = %s\n", attr.Name(), attr.Value())
}
```

## Streaming Parser

For large XML documents, use the streaming parser:

```go
import "github.com/gogo-agent/xmldom/decoder"

// Create streaming decoder
decoder := xmldom.NewDecoder(reader)

// Parse incrementally
for {
    token, err := decoder.Token()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    
    // Process token
    switch t := token.(type) {
    case xmldom.StartElement:
        fmt.Printf("Start element: %s\n", t.Name.Local)
    case xmldom.EndElement:
        fmt.Printf("End element: %s\n", t.Name.Local)
    case xmldom.CharData:
        fmt.Printf("Text: %s\n", string(t))
    }
}
```

## Namespace Support

```go
xml := `<?xml version="1.0"?>
<root xmlns:book="http://example.com/book">
    <book:catalog>
        <book:item id="1">Programming Guide</book:item>
    </book:catalog>
</root>`

doc, err := xmldom.ParseString(xml)
if err != nil {
    log.Fatal(err)
}

// Namespace-aware operations
elements := doc.GetElementsByTagNameNS("http://example.com/book", "item")
```

## Performance

* **Memory Efficient**: Optimized memory usage for large documents
* **Fast Parsing**: High-performance XML parsing with minimal allocations
* **Concurrent Safe**: Thread-safe operations where specified
* **Streaming Support**: Process large documents without loading entirely into memory

## W3C Compliance

This implementation is tested against the official W3C XML conformance test suite, ensuring compatibility with standard XML processing expectations.

## Testing

```bash
# Run all tests
go test ./...

# Run conformance tests
go test -run TestConformance

# Run XPath tests  
go test -run TestXPath

# Run benchmarks
go test -bench=.
```

## License

This project is part of the gogo-agent ecosystem.
