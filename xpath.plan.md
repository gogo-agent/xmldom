# XPath 1.0 Implementation Plan for xmldom

## Overview

This document outlines the plan to implement XPath 1.0 support for the existing `xmldom` package. The implementation will follow the [DOM Living Standard XPath interfaces](https://dom.spec.whatwg.org/#xpath) and leverage the existing DOM implementation to avoid double parsing and maximize performance.

## Current DOM Infrastructure Analysis

### Existing Assets to Leverage

From analyzing `core.go`, we have a robust DOM implementation with:

1. **Complete DOM tree structure**: `Node`, `Element`, `Document` interfaces with full tree navigation
2. **Element lookup indexes**: `document.idMap` for fast ID-based lookups
3. **Live NodeList implementation**: With update mechanisms for DOM mutations
4. **Tree traversal utilities**: `NodeIterator` and `TreeWalker` implementations
5. **Namespace support**: Full namespace URI and prefix handling
6. **Concurrent access safety**: RWMutex protection for DOM operations

### Key Integration Points

- **Document interface**: Already implements factory methods and tree operations
- **Element traversal**: Existing `GetElementsByTagName`, `GetElementsByTagNameNS` methods
- **Node iteration**: Current `NodeIterator` can be leveraged for axis traversal
- **Live collections**: Pattern for maintaining up-to-date result sets

## XPath 1.0 Implementation Strategy

### Phase 1: Core Interfaces and Data Structures

#### XPath Value Types

```go
// XPath 1.0 fundamental types
type XPathValue interface {
    Type() XPathType
    String() string
    Number() float64
    Boolean() bool
    NodeSet() []Node
}

type XPathType uint8

const (
    XPathTypeString XPathType = iota
    XPathTypeNumber
    XPathTypeBoolean
    XPathTypeNodeSet
)
```

#### XPathResult Interface (DOM Living Standard)

```go
// XPathResult constants matching DOM spec
const (
    XPATH_ANY_TYPE                   uint16 = 0
    XPATH_NUMBER_TYPE               uint16 = 1
    XPATH_STRING_TYPE               uint16 = 2
    XPATH_BOOLEAN_TYPE              uint16 = 3
    XPATH_UNORDERED_NODE_ITERATOR_TYPE uint16 = 4
    XPATH_ORDERED_NODE_ITERATOR_TYPE   uint16 = 5
    XPATH_UNORDERED_NODE_SNAPSHOT_TYPE uint16 = 6
    XPATH_ORDERED_NODE_SNAPSHOT_TYPE   uint16 = 7
    XPATH_ANY_UNORDERED_NODE_TYPE     uint16 = 8
    XPATH_FIRST_ORDERED_NODE_TYPE     uint16 = 9
)

type XPathResult interface {
    ResultType() uint16
    NumberValue() float64
    StringValue() string
    BooleanValue() bool
    SingleNodeValue() Node
    InvalidIteratorState() bool
    SnapshotLength() uint
    IterateNext() Node
    SnapshotItem(index uint) Node
}
```

#### Document Extension

```go
// Extend Document interface following DOM Living Standard
type XPathEvaluator interface {
    // Core evaluation method matching browser API
    Evaluate(expression string, contextNode Node, resolver XPathNSResolver, 
             type_ uint16, result XPathResult) XPathResult
    
    // Expression compilation for reuse
    CreateExpression(expression string, resolver XPathNSResolver) (XPathExpression, error)
    
    // Namespace resolver factory (legacy compatibility)
    CreateNSResolver(nodeResolver Node) Node
}

// XPathNSResolver callback for namespace resolution
type XPathNSResolver interface {
    LookupNamespaceURI(prefix string) string
}

// Compiled expression for reuse
type XPathExpression interface {
    Evaluate(contextNode Node, type_ uint16, result XPathResult) XPathResult
}
```

### Phase 2: Expression Parsing

#### Lexer Implementation

```go
type XPathTokenType int

const (
    TOKEN_NAME XPathTokenType = iota
    TOKEN_STRING
    TOKEN_NUMBER
    TOKEN_OPERATOR
    TOKEN_AXIS
    TOKEN_FUNCTION
    TOKEN_LEFT_PAREN
    TOKEN_RIGHT_PAREN
    // ... other token types
)

type XPathToken struct {
    Type  XPathTokenType
    Value string
    Pos   int
}

type XPathLexer struct {
    input string
    pos   int
    // Use sync.OnceValue for lazy initialization
}
```

#### AST Node Types

```go
type XPathNode interface {
    Type() XPathNodeType
    Evaluate(ctx *XPathContext) XPathValue
}

type XPathNodeType uint8

const (
    NodeTypePath XPathNodeType = iota
    NodeTypeAxis
    NodeTypeFunction
    NodeTypeLiteral
    NodeTypeBinary
    NodeTypeUnary
    NodeTypePredicate
)

// Leverage existing DOM traversal for axis evaluation
type AxisNode struct {
    Axis      XPathAxis
    NodeTest  NodeTest
    Predicates []XPathNode
}

type XPathAxis uint8

const (
    AxisChild XPathAxis = iota
    AxisDescendant
    AxisParent
    AxisAncestor
    AxisFollowingSibling
    AxisPrecedingSibling
    AxisFollowing
    AxisPreceding
    AxisAttribute
    AxisNamespace
    AxisSelf
    AxisDescendantOrSelf
    AxisAncestorOrSelf
)
```

### Phase 3: Evaluation Engine

#### Context Management

```go
type XPathContext struct {
    ContextNode Node
    ContextSize int
    ContextPosition int
    VariableBindings map[string]XPathValue
    FunctionLibrary map[string]XPathFunction
    NamespaceResolver XPathNSResolver
    Document Document // Access to existing DOM indexes
}
```

#### Axis Implementation Strategy

Leverage existing DOM tree traversal:

1. **Child axis**: Use `Node.FirstChild()` and `Node.NextSibling()`
2. **Descendant axis**: Utilize existing `NodeIterator` with `SHOW_ELEMENT`
3. **Attribute axis**: Access `Node.Attributes()` NamedNodeMap
4. **ID lookups**: Use `Document.GetElementById()` for optimized `id()` function

#### Performance Optimizations

1. **Index utilization**: Use `document.idMap` for ID-based queries
2. **Live result caching**: Extend existing live NodeList pattern
3. **Expression compilation**: Cache parsed ASTs using `sync.OnceValue`
4. **Memory allocation**: Prefer stack allocation following Go rules

### Phase 4: Core Function Library

#### Node-set Functions
- `last()`: Return context size
- `position()`: Return context position  
- `count(node-set)`: Count nodes in set
- `id(object)`: Leverage `Document.GetElementById()`
- `local-name(node-set?)`: Use existing `Node.LocalName()`
- `namespace-uri(node-set?)`: Use existing `Node.NamespaceURI()`
- `name(node-set?)`: Use existing `Node.NodeName()`

#### String Functions
- `string(object?)`: Convert to string representation
- `concat(string, string, string*)`: Concatenate strings
- `starts-with(string, string)`: String prefix check
- `contains(string, string)`: Substring search
- `substring-before(string, string)`: Extract before delimiter
- `substring-after(string, string)`: Extract after delimiter
- `substring(string, number, number?)`: Extract substring
- `string-length(string?)`: String length
- `normalize-space(string?)`: Normalize whitespace
- `translate(string, string, string)`: Character translation

#### Boolean Functions
- `boolean(object)`: Convert to boolean
- `not(boolean)`: Logical negation
- `true()`: Boolean true constant
- `false()`: Boolean false constant
- `lang(string)`: Language test using `xml:lang`

#### Number Functions
- `number(object?)`: Convert to number
- `sum(node-set)`: Sum numeric values
- `floor(number)`: Math floor
- `ceiling(number)`: Math ceiling
- `round(number)`: Math round

### Phase 5: Integration with Existing DOM

#### Document Interface Extension

```go
// Add XPath evaluation to existing Document interface
func (d *document) Evaluate(expression string, contextNode Node, 
                          resolver XPathNSResolver, type_ uint16, 
                          result XPathResult) XPathResult {
    d.mu.RLock()
    defer d.mu.RUnlock()
    
    // Compile expression (with caching)
    expr := d.compileExpression(expression, resolver)
    
    // Evaluate using existing DOM tree
    return expr.Evaluate(contextNode, type_, result)
}
```

#### Leverage Existing Indexes

```go
// Optimize descendant queries using existing element lookup
func (e *evaluator) evaluateDescendantAxis(ctx *XPathContext, nodeTest NodeTest) []Node {
    if nodeTest.IsElementName() {
        // Use existing GetElementsByTagName for performance
        elements := ctx.Document.GetElementsByTagName(nodeTest.Name())
        return filterDescendants(ctx.ContextNode, elements)
    }
    
    // Fallback to tree traversal for complex node tests
    return e.traverseDescendants(ctx, nodeTest)
}
```

#### Namespace Integration

```go
// Leverage existing namespace support
func (r *defaultNSResolver) LookupNamespaceURI(prefix string) string {
    if r.contextNode != nil {
        return string(r.contextNode.LookupNamespaceURI(DOMString(prefix)))
    }
    return ""
}
```

### Phase 6: Error Handling and Edge Cases

#### Error Types

```go
type XPathError struct {
    Type    XPathErrorType
    Message string
    Position int
}

type XPathErrorType uint8

const (
    XPathSyntaxError XPathErrorType = iota
    XPathTypeError
    XPathFunctionError
    XPathAxisError
)
```

#### Following Go Rules

1. **Error handling**: Return `(result, error)` tuples
2. **No testify**: Use standard `testing` package
3. **Factory naming**: `NewXPathResult()` for pointers, `MakeXPathContext()` for values
4. **Logging**: Use `slog` for debugging output
5. **Metrics**: OpenTelemetry instrumentation for performance monitoring

### Phase 7: Testing Strategy

#### Unit Tests Structure

```
xpath_test.go           - Main API tests
xpath_parser_test.go    - Parser unit tests  
xpath_functions_test.go - Function library tests
xpath_axes_test.go      - Axis evaluation tests
xpath_integration_test.go - DOM integration tests
xpath_performance_test.go - Performance benchmarks
```

#### Test Coverage Requirements

- **Parser tests**: All XPath 1.0 syntax variations
- **Function tests**: All core functions with edge cases
- **Axis tests**: Each axis with various DOM configurations
- **Integration tests**: Real XML documents with complex queries
- **Performance tests**: Memory usage and execution time benchmarks
- **Namespace tests**: Namespace resolution and conflicts
- **Error handling**: Invalid expressions and type errors

#### Testing with gtimeout

```go
func TestXPathEvaluation(t *testing.T) {
    cmd := exec.Command("gtimeout", "5s", "go", "test", "-v", "./...")
    output, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("Tests failed: %v\n%s", err, output)
    }
}
```

### Phase 8: Performance Considerations

#### Memory Management

1. **Stack allocation**: Use value types where possible
2. **Pool reuse**: Object pools for frequent allocations
3. **Lazy evaluation**: Only compute results when accessed
4. **Result streaming**: Iterator pattern for large node sets

#### Caching Strategy

```go
// Expression cache using sync.OnceValue
var expressionCache = sync.Map{} // string -> *compiledExpression

func (d *document) compileExpression(expr string, resolver XPathNSResolver) *compiledExpression {
    key := expr + "|" + resolverKey(resolver)
    
    if cached, ok := expressionCache.Load(key); ok {
        return cached.(*compiledExpression)
    }
    
    compiled := parseAndCompile(expr, resolver)
    expressionCache.Store(key, compiled)
    return compiled
}
```

#### Index Optimization

1. **Element name queries**: Use existing `GetElementsByTagName`
2. **ID queries**: Leverage `document.idMap` 
3. **Descendant searches**: Smart pruning based on element hierarchy
4. **Attribute lookups**: Direct access via `NamedNodeMap`

### Phase 9: OpenTelemetry Integration

#### Instrumentation Points

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func (d *document) Evaluate(expression string, contextNode Node, 
                          resolver XPathNSResolver, type_ uint16, 
                          result XPathResult) XPathResult {
    
    ctx, span := otel.Tracer("xmldom").Start(context.Background(), "xpath.evaluate")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("xpath.expression", expression),
        attribute.String("xpath.context_node", contextNode.NodeName()),
        attribute.Int("xpath.result_type", int(type_)),
    )
    
    // ... evaluation logic
}
```

#### Metrics Collection

1. **Expression parsing time**: Duration metrics
2. **Evaluation performance**: Execution time by expression complexity
3. **Cache hit rates**: Compiled expression cache effectiveness
4. **Memory usage**: Result set sizes and memory pressure
5. **Error rates**: Parsing and evaluation failures

### Phase 10: Documentation and Examples

#### API Documentation

Complete godoc documentation covering:

1. **XPath 1.0 compliance**: Supported features and limitations
2. **Performance characteristics**: When to use different result types
3. **Integration patterns**: How XPath leverages existing DOM structures
4. **Error handling**: Common error scenarios and recovery
5. **Namespace handling**: Best practices for namespace resolution

#### Usage Examples

```go
// Basic element selection
doc := parseXMLDocument(xmlData)
result := doc.Evaluate("//book[@isbn]", doc.DocumentElement(), nil, XPATH_ORDERED_NODE_SNAPSHOT_TYPE, nil)
for i := 0; i < result.SnapshotLength(); i++ {
    book := result.SnapshotItem(i)
    fmt.Printf("Book: %s\n", book.TextContent())
}

// Using compiled expressions for performance
expr := doc.CreateExpression("//author[position() > 1]", nil)
result := expr.Evaluate(doc.DocumentElement(), XPATH_UNORDERED_NODE_ITERATOR_TYPE, nil)
for node := result.IterateNext(); node != nil; node = result.IterateNext() {
    fmt.Printf("Author: %s\n", node.TextContent())
}

// Namespace-aware queries
resolver := NewNamespaceResolver(map[string]string{
    "book": "http://example.com/book",
    "author": "http://example.com/author",
})
result := doc.Evaluate("//book:title | //author:name", doc, resolver, XPATH_ORDERED_NODE_SNAPSHOT_TYPE, nil)
```

## Implementation Timeline

### Milestones

1. **Week 1**: Core interfaces and data structures
2. **Week 2**: Expression parser and AST
3. **Week 3**: Basic evaluation engine and axes
4. **Week 4**: Function library implementation  
5. **Week 5**: Document integration and optimization
6. **Week 6**: Comprehensive testing and benchmarks
7. **Week 7**: OpenTelemetry integration and monitoring
8. **Week 8**: Documentation and performance tuning

## Success Criteria

1. **XPath 1.0 compliance**: Pass W3C XPath 1.0 test suite
2. **Performance**: Faster than creating new DOM parser for simple queries
3. **Memory efficiency**: Stack allocation where possible, minimal heap pressure
4. **Test coverage**: >90% coverage with 100% passing tests
5. **API compatibility**: Match DOM Living Standard XPath interfaces
6. **Integration**: Seamless integration with existing DOM implementation
7. **Documentation**: Complete API documentation with examples
8. **Monitoring**: Full OpenTelemetry instrumentation

## Benefits of This Approach

1. **No double parsing**: Direct evaluation on existing DOM tree
2. **Index leverage**: Use existing element lookup optimizations  
3. **Memory efficiency**: Reuse existing Node objects and structures
4. **API familiarity**: Standard browser XPath API for easy adoption
5. **Performance**: Optimized for common XPath patterns using DOM indexes
6. **Maintainability**: Clean separation between parsing and evaluation
7. **Extensibility**: Easy to add XPath 2.0+ features in the future

This implementation plan provides a robust foundation for adding XPath 1.0 support while maximizing the value of the existing DOM implementation and following all specified Go development practices.
