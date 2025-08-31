# XPath Implementation TODO List

## Progress Summary

### ✅ COMPLETED TASKS

1. **Analyze existing DOM implementation and architecture**
   - Reviewed the current XML DOM implementation to understand existing structures, interfaces, and patterns
   - Identified the Document interface and its current methods
   - Examined existing tree traversal mechanisms and element lookup capabilities
   - Studied the current indexing strategies used for efficient DOM operations
   - Documented key integration points where XPath evaluation hooks into the existing codebase

2. **Design XPath 1.0 core data structures and interfaces**
   - Defined the core XPath types: string, number, boolean, and node-set
   - Created interfaces for XPath expressions following the DOM Living Standard
   - Designed the XPathResult interface with appropriate result types (ANY_TYPE, NUMBER_TYPE, STRING_TYPE, BOOLEAN_TYPE, etc.)
   - Defined the XPathEvaluator interface that extends the Document
   - Planned the internal AST representation for parsed XPath expressions

3. **Extend Document interface with Evaluate method**
   - Added Evaluate method to the Document interface matching browser APIs
   - Implemented XPathNSResolver for namespace resolution
   - Created factory methods following Go naming conventions (NewXPathResult for pointers, MakeXPathResult for values)
   - Ensured the API matches the DOM Living Standard specification
   - Integrated with existing Document implementation without breaking changes

4. **Fix Document Evaluate method return handling**
   - Updated `core.go` to properly handle the two return values from `expr.Evaluate()`
   - Fixed compilation errors related to method signature mismatches

5. **Update XPath result method signatures**
   - Modified all `xpathResult` methods to return `(value, error)` tuples:
     - `BooleanValue() (bool, error)`
     - `NumberValue() (float64, error)`
     - `StringValue() (string, error)`
     - `SingleNodeValue() (Node, error)`
     - `IterateNext() (Node, error)`
     - `SnapshotLength() (uint32, error)`
     - `SnapshotItem(index uint32) (Node, error)`
   - Updated all concrete XPath result implementations to match interface

6. **Update most XPath tests**
   - Fixed majority of test method calls to handle new error return values
   - Tests now properly check for errors and handle the new signatures

### ✅ COMPLETED TASKS (CONTINUED)

7. **Fix remaining XPath test method calls**
   - **STATUS**: ✅ COMPLETED
   - All XPath test functions have been uncommented and properly configured
   - All method calls now properly handle `(value, error)` return patterns:
     - `result.SingleNodeValue()` calls updated with error handling
     - `result.NumberValue()`, `result.StringValue()`, `result.BooleanValue()` calls updated
     - `result.SnapshotLength()`, `result.SnapshotItem()`, `result.IterateNext()` calls updated
   - Tests compile successfully and run (though fail as expected since evaluation engine not implemented)
   - Error handling follows Go conventions consistently throughout

### 🚧 PENDING TASKS

8. **Implement XPath expression parser**
   - Create a lexer to tokenize XPath 1.0 expressions
   - Implement a recursive descent parser for XPath grammar
   - Build an Abstract Syntax Tree (AST) representation of XPath expressions
   - Handle all XPath 1.0 syntax including axes, node tests, predicates, and functions
   - Ensure proper error handling and reporting for invalid expressions

9. **Implement XPath evaluation engine**
   - Create evaluator that walks the AST and executes against the DOM tree
   - Leverage existing DOM tree traversal methods for axis navigation
   - Reuse existing element lookup indexes for optimized searches
   - Implement all XPath 1.0 axes (child, descendant, parent, ancestor, following-sibling, etc.)
   - Handle node tests (element names, wildcards, node types)
   - Implement predicate evaluation with proper context handling

10. **Implement XPath 1.0 core functions**
    - Implement node-set functions: last(), position(), count(), id(), local-name(), namespace-uri(), name()
    - Implement string functions: string(), concat(), starts-with(), contains(), substring(), string-length(), normalize-space(), translate()
    - Implement boolean functions: boolean(), not(), true(), false(), lang()
    - Implement number functions: number(), sum(), floor(), ceiling(), round()
    - Ensure all functions follow XPath 1.0 specification behavior

11. **Optimize performance using existing DOM indexes**
    - Identify opportunities to use existing element indexes for descendant queries
    - Implement caching for compiled XPath expressions
    - Optimize common patterns like //elementName using existing lookup tables
    - Avoid double parsing by directly operating on the existing DOM tree
    - Use stack allocation where possible to minimize heap pressure

12. **Implement comprehensive test suite**
    - Create unit tests for the XPath parser covering all syntax variations
    - Test each XPath axis with various node configurations
    - Verify all XPath 1.0 functions with edge cases
    - Test integration with the existing DOM implementation
    - Ensure namespace handling works correctly
    - Use gtimeout for all test executions
    - Achieve >90% test coverage with 100% passing tests

13. **Add OpenTelemetry instrumentation**
    - Instrument XPath expression parsing with OTEL spans
    - Add metrics for evaluation performance
    - Track cache hit rates for compiled expressions
    - Monitor memory usage for large result sets
    - Ensure tracing integrates with existing OTEL setup

14. **Create documentation and examples**
    - Document the public API with godoc comments
    - Create examples showing common XPath usage patterns
    - Document performance characteristics and optimization tips
    - Explain how the implementation leverages existing DOM structures
    - Include migration guide for users familiar with browser XPath APIs

## Current Status

✅ **Build Status**: PASSING  
✅ **Tests Status**: All XPath tests are uncommented and properly configured with error handling  
📁 **Files Modified**: `core.go`, `xpath.go`, `xpath_test.go`  
🧪 **Test Results**: XPath tests compile and run (fail as expected - evaluation engine not implemented)  

## Next Steps

1. **IMMEDIATE**: Begin implementing the XPath expression parser (Task #8)
2. **SHORT TERM**: Implement the evaluation engine and core functions  
3. **MEDIUM TERM**: Performance optimization and comprehensive testing
4. **LONG TERM**: OpenTelemetry instrumentation and documentation

## Architecture Notes

- The XPath implementation integrates cleanly with the existing DOM structure
- Error handling follows Go conventions with `(value, error)` return patterns
- The interface design matches the DOM Living Standard for browser compatibility
- All XPath result types are properly implemented with error propagation
