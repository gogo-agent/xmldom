package xmldom

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/groupcache/lru"
)

// XPath result type constants matching DOM Living Standard
// https://dom.spec.whatwg.org/#xpath
const (
	XPATH_ANY_TYPE                     uint16 = 0
	XPATH_NUMBER_TYPE                  uint16 = 1
	XPATH_STRING_TYPE                  uint16 = 2
	XPATH_BOOLEAN_TYPE                 uint16 = 3
	XPATH_UNORDERED_NODE_ITERATOR_TYPE uint16 = 4
	XPATH_ORDERED_NODE_ITERATOR_TYPE   uint16 = 5
	XPATH_UNORDERED_NODE_SNAPSHOT_TYPE uint16 = 6
	XPATH_ORDERED_NODE_SNAPSHOT_TYPE   uint16 = 7
	XPATH_ANY_UNORDERED_NODE_TYPE      uint16 = 8
	XPATH_FIRST_ORDERED_NODE_TYPE      uint16 = 9
)

// XPathValue represents the fundamental XPath 1.0 data types
// All XPath expressions evaluate to one of these four types
type XPathValue interface {
	Type() XPathValueType
	String() string
	Number() float64
	Boolean() bool
	NodeSet() []Node
}

// XPathValueType represents the four fundamental XPath 1.0 types
type XPathValueType uint8

const (
	XPathValueTypeString XPathValueType = iota
	XPathValueTypeNumber
	XPathValueTypeBoolean
	XPathValueTypeNodeSet
)

// XPathResult interface matching DOM Living Standard
// Provides access to XPath evaluation results in various formats
type XPathResult interface {
	// Result type and basic values
	ResultType() uint16
	NumberValue() (float64, error)
	StringValue() (string, error)
	BooleanValue() (bool, error)
	SingleNodeValue() (Node, error)

	// Iterator state and operations
	InvalidIteratorState() bool
	IterateNext() (Node, error)

	// Snapshot operations
	SnapshotLength() (uint32, error)
	SnapshotItem(index uint32) (Node, error)
}

// XPathExpression represents a compiled XPath expression
// Following DOM Living Standard for performance optimization
type XPathExpression interface {
	// Evaluate the compiled expression against a context node
	Evaluate(contextNode Node, resultType uint16, result XPathResult) (XPathResult, error)
	// SetVariableBindings sets variable bindings for the expression
	SetVariableBindings(bindings map[string]XPathValue)
}

// XPathNSResolver provides namespace resolution for XPath expressions
// Callback interface matching DOM Living Standard
type XPathNSResolver interface {
	LookupNamespaceURI(prefix string) string
}

// XPathEvaluatorBase defines the core XPath evaluation methods
// This will be mixed into the Document interface
type XPathEvaluatorBase interface {
	// Create a compiled XPath expression for reuse
	CreateExpression(expression string, resolver XPathNSResolver) (XPathExpression, error)

	// Create namespace resolver (legacy compatibility)
	CreateNSResolver(nodeResolver Node) Node

	// Evaluate XPath expression directly
	Evaluate(expression string, contextNode Node, resolver XPathNSResolver,
		resultType uint16, result XPathResult) (XPathResult, error)
}

// Internal XPath data structures

// xpathStringValue implements XPathValue for string results
type xpathStringValue struct {
	value string
}

func (v xpathStringValue) Type() XPathValueType { return XPathValueTypeString }
func (v xpathStringValue) String() string       { return v.value }
func (v xpathStringValue) Number() float64      { return stringToNumber(v.value) }
func (v xpathStringValue) Boolean() bool        { return v.value != "" }
func (v xpathStringValue) NodeSet() []Node      { return nil }

// xpathNumberValue implements XPathValue for numeric results
type xpathNumberValue struct {
	value float64
}

func (v xpathNumberValue) Type() XPathValueType { return XPathValueTypeNumber }
func (v xpathNumberValue) String() string       { return numberToString(v.value) }
func (v xpathNumberValue) Number() float64      { return v.value }
func (v xpathNumberValue) Boolean() bool        { return v.value != 0 && !isNaN(v.value) }
func (v xpathNumberValue) NodeSet() []Node      { return nil }

// xpathBooleanValue implements XPathValue for boolean results
type xpathBooleanValue struct {
	value bool
}

func (v xpathBooleanValue) Type() XPathValueType { return XPathValueTypeBoolean }
func (v xpathBooleanValue) String() string       { return booleanToString(v.value) }
func (v xpathBooleanValue) Number() float64      { return booleanToNumber(v.value) }
func (v xpathBooleanValue) Boolean() bool        { return v.value }
func (v xpathBooleanValue) NodeSet() []Node      { return nil }

// xpathNodeSetValue implements XPathValue for node-set results
type xpathNodeSetValue struct {
	nodes []Node
}

func (v xpathNodeSetValue) Type() XPathValueType { return XPathValueTypeNodeSet }
func (v xpathNodeSetValue) String() string       { return nodeSetToString(v.nodes) }
func (v xpathNodeSetValue) Number() float64      { return stringToNumber(v.String()) }
func (v xpathNodeSetValue) Boolean() bool        { return len(v.nodes) > 0 }
func (v xpathNodeSetValue) NodeSet() []Node      { return v.nodes }

// XPathResult implementation

// xpathResult implements the XPathResult interface
type xpathResult struct {
	resultType   uint16
	value        XPathValue
	iterator     *xpathNodeIterator
	snapshot     []Node
	invalidState bool
	mu           sync.RWMutex // Protect iterator state
}

func (r *xpathResult) ResultType() uint16 {
	return r.resultType
}

func (r *xpathResult) NumberValue() (float64, error) {
	if r.resultType != XPATH_NUMBER_TYPE {
		return 0, NewXPathException("TYPE_ERR", "Result is not a number")
	}
	return r.value.Number(), nil
}

func (r *xpathResult) StringValue() (string, error) {
	if r.resultType != XPATH_STRING_TYPE {
		return "", NewXPathException("TYPE_ERR", "Result is not a string")
	}
	return r.value.String(), nil
}

func (r *xpathResult) BooleanValue() (bool, error) {
	if r.resultType != XPATH_BOOLEAN_TYPE {
		return false, NewXPathException("TYPE_ERR", "Result is not a boolean")
	}
	return r.value.Boolean(), nil
}

func (r *xpathResult) SingleNodeValue() (Node, error) {
	switch r.resultType {
	case XPATH_ANY_UNORDERED_NODE_TYPE, XPATH_FIRST_ORDERED_NODE_TYPE:
		if nodes := r.value.NodeSet(); len(nodes) > 0 {
			return nodes[0], nil
		}
		return nil, nil
	}
	return nil, NewXPathException("TYPE_ERR", "Result is not a single node")
}

func (r *xpathResult) InvalidIteratorState() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.invalidState
}

func (r *xpathResult) IterateNext() (Node, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.invalidState {
		return nil, NewXPathException("INVALID_STATE_ERR", "Iterator is in invalid state")
	}

	switch r.resultType {
	case XPATH_UNORDERED_NODE_ITERATOR_TYPE, XPATH_ORDERED_NODE_ITERATOR_TYPE:
		if r.iterator != nil {
			return r.iterator.nextNode(), nil
		}
	}
	return nil, NewXPathException("TYPE_ERR", "Result is not an iterator")
}

func (r *xpathResult) SnapshotLength() (uint32, error) {
	switch r.resultType {
	case XPATH_UNORDERED_NODE_SNAPSHOT_TYPE, XPATH_ORDERED_NODE_SNAPSHOT_TYPE:
		return uint32(len(r.snapshot)), nil
	}
	return 0, NewXPathException("TYPE_ERR", "Result is not a snapshot")
}

func (r *xpathResult) SnapshotItem(index uint32) (Node, error) {
	switch r.resultType {
	case XPATH_UNORDERED_NODE_SNAPSHOT_TYPE, XPATH_ORDERED_NODE_SNAPSHOT_TYPE:
		if int(index) < len(r.snapshot) {
			return r.snapshot[index], nil
		}
		return nil, nil
	}
	return nil, NewXPathException("TYPE_ERR", "Result is not a snapshot")
}

// xpathNodeIterator provides iteration over node sets
type xpathNodeIterator struct {
	nodes []Node
	index int
}

func (it *xpathNodeIterator) nextNode() Node {
	if it.index >= len(it.nodes) {
		return nil
	}
	node := it.nodes[it.index]
	it.index++
	return node
}

// Factory functions following Go naming conventions

// NewXPathResult creates a new XPathResult (pointer type)
func NewXPathResult(resultType uint16, value XPathValue) XPathResult {
	result := &xpathResult{
		resultType: resultType,
		value:      value,
	}

	// Configure based on result type
	switch resultType {
	case XPATH_UNORDERED_NODE_ITERATOR_TYPE, XPATH_ORDERED_NODE_ITERATOR_TYPE:
		if nodes := value.NodeSet(); nodes != nil {
			result.iterator = &xpathNodeIterator{nodes: nodes}
		}
	case XPATH_UNORDERED_NODE_SNAPSHOT_TYPE, XPATH_ORDERED_NODE_SNAPSHOT_TYPE:
		result.snapshot = value.NodeSet()
	}

	return result
}

// NewXPathStringValue creates a string XPathValue
func NewXPathStringValue(s string) XPathValue {
	return xpathStringValue{value: s}
}

// NewXPathNumberValue creates a number XPathValue
func NewXPathNumberValue(n float64) XPathValue {
	return xpathNumberValue{value: n}
}

// NewXPathBooleanValue creates a boolean XPathValue
func NewXPathBooleanValue(b bool) XPathValue {
	return xpathBooleanValue{value: b}
}

// NewXPathNodeSetValue creates a node-set XPathValue
func NewXPathNodeSetValue(nodes []Node) XPathValue {
	return xpathNodeSetValue{nodes: nodes}
}

// XPath AST node types for internal representation

// XPathNode represents a node in the XPath AST
type XPathNode interface {
	Type() XPathNodeType
	Evaluate(ctx *XPathContext) (XPathValue, error)
}

// XPathNodeType represents different types of AST nodes
type XPathNodeType uint8

const (
	XPathNodeTypePath XPathNodeType = iota
	XPathNodeTypeAxis
	XPathNodeTypeFunction
	XPathNodeTypeLiteral
	XPathNodeTypeBinary
	XPathNodeTypeUnary
	XPathNodeTypePredicate
	XPathNodeTypeFilter
)

// XPathContext provides evaluation context for XPath expressions
type XPathContext struct {
	ContextNode       Node
	ContextSize       int
	ContextPosition   int
	VariableBindings  map[string]XPathValue
	FunctionLibrary   map[string]XPathFunction
	NamespaceResolver XPathNSResolver
	Document          Document // Access to existing DOM indexes and operations

	// Context for tracing and cancellation
	Context context.Context
}

// XPathFunction represents an XPath function implementation
type XPathFunction interface {
	Call(ctx *XPathContext, args []XPathValue) (XPathValue, error)
	MinArgs() int
	MaxArgs() int // -1 for unlimited
}

// XPath axis types matching XPath 1.0 specification
type XPathAxis uint8

const (
	XPathAxisChild XPathAxis = iota
	XPathAxisDescendant
	XPathAxisParent
	XPathAxisAncestor
	XPathAxisFollowingSibling
	XPathAxisPrecedingSibling
	XPathAxisFollowing
	XPathAxisPreceding
	XPathAxisAttribute
	XPathAxisNamespace
	XPathAxisSelf
	XPathAxisDescendantOrSelf
	XPathAxisAncestorOrSelf
)

// XPathNodeTest represents node test conditions
type XPathNodeTest interface {
	Matches(node Node, ctx *XPathContext) bool
	Name() string
	IsWildcard() bool
}

// XPath operators for binary/unary expressions
type XPathOperator uint8

const (
	XPathOperatorOr XPathOperator = iota
	XPathOperatorAnd
	XPathOperatorEq
	XPathOperatorNeq
	XPathOperatorLt
	XPathOperatorLte
	XPathOperatorGt
	XPathOperatorGte
	XPathOperatorPlus
	XPathOperatorMinus
	XPathOperatorMultiply
	XPathOperatorDiv
	XPathOperatorMod
	XPathOperatorUnion
	XPathOperatorUnaryMinus
)

// XPath expression cache for performance using groupcache/lru
var (
	exprCache   *lru.Cache
	exprCacheMu sync.RWMutex
)

// Initialize the cache with a capacity of 1000 expressions
func init() {
	exprCache = lru.New(1000)
}

// getCachedExpression retrieves a cached expression from the LRU cache
func getCachedExpression(expr string) (XPathNode, bool) {
	exprCacheMu.RLock()
	defer exprCacheMu.RUnlock()

	if ast, ok := exprCache.Get(expr); ok {
		if node, valid := ast.(XPathNode); valid {
			return node, true
		}
	}

	return nil, false
}

// setCachedExpression stores a parsed expression in the LRU cache
func setCachedExpression(expr string, ast XPathNode) {
	exprCacheMu.Lock()
	defer exprCacheMu.Unlock()

	exprCache.Add(expr, ast)
}

// sortNodesInDocumentOrder sorts a slice of nodes in document order
// This is required by XPath 1.0 specification for all node-set results
func sortNodesInDocumentOrder(nodes []Node) {
	if len(nodes) <= 1 {
		return
	}

	sort.Slice(nodes, func(i, j int) bool {
		// Use CompareDocumentPosition to determine order
		// DOCUMENT_POSITION_FOLLOWING means j comes after i in document order
		// So if j is following i, then i < j (i comes before j), return true
		position := nodes[i].CompareDocumentPosition(nodes[j])
		return position&DOCUMENT_POSITION_FOLLOWING != 0
	})
}

// xpathNamespaceNode represents a namespace node in XPath
// These are virtual nodes that don't exist in the DOM tree
type xpathNamespaceNode struct {
	prefix       string
	namespaceURI string
	ownerElement Element
}

// Implement Node interface for namespace nodes
func (n *xpathNamespaceNode) NodeType() uint16                                   { return 13 } // Custom type for namespace nodes
func (n *xpathNamespaceNode) NodeName() DOMString                                { return DOMString(n.prefix) }
func (n *xpathNamespaceNode) NodeValue() DOMString                               { return DOMString(n.namespaceURI) }
func (n *xpathNamespaceNode) SetNodeValue(value DOMString) error                 { return nil }
func (n *xpathNamespaceNode) TextContent() DOMString                             { return DOMString(n.namespaceURI) }
func (n *xpathNamespaceNode) SetTextContent(content DOMString)                   {}
func (n *xpathNamespaceNode) ParentNode() Node                                   { return n.ownerElement }
func (n *xpathNamespaceNode) ParentElement() Element                             { return n.ownerElement }
func (n *xpathNamespaceNode) FirstChild() Node                                   { return nil }
func (n *xpathNamespaceNode) LastChild() Node                                    { return nil }
func (n *xpathNamespaceNode) PreviousSibling() Node                              { return nil }
func (n *xpathNamespaceNode) NextSibling() Node                                  { return nil }
func (n *xpathNamespaceNode) ChildNodes() NodeList                               { return nil }
func (n *xpathNamespaceNode) HasChildNodes() bool                                { return false }
func (n *xpathNamespaceNode) OwnerDocument() Document                            { return n.ownerElement.OwnerDocument() }
func (n *xpathNamespaceNode) LocalName() DOMString                               { return DOMString(n.prefix) }
func (n *xpathNamespaceNode) NamespaceURI() DOMString                            { return "" }
func (n *xpathNamespaceNode) Prefix() DOMString                                  { return "" }
func (n *xpathNamespaceNode) SetPrefix(prefix DOMString) error                   { return nil }
func (n *xpathNamespaceNode) BaseURI() DOMString                                 { return "" }
func (n *xpathNamespaceNode) IsConnected() bool                                  { return n.ownerElement.IsConnected() }
func (n *xpathNamespaceNode) AppendChild(child Node) (Node, error)               { return nil, nil }
func (n *xpathNamespaceNode) InsertBefore(newChild, refChild Node) (Node, error) { return nil, nil }
func (n *xpathNamespaceNode) ReplaceChild(newChild, oldChild Node) (Node, error) { return nil, nil }
func (n *xpathNamespaceNode) RemoveChild(child Node) (Node, error)               { return nil, nil }
func (n *xpathNamespaceNode) Normalize()                                         {}
func (n *xpathNamespaceNode) CloneNode(deep bool) Node                           { return n }
func (n *xpathNamespaceNode) IsEqualNode(other Node) bool {
	if otherNS, ok := other.(*xpathNamespaceNode); ok {
		return n.prefix == otherNS.prefix && n.namespaceURI == otherNS.namespaceURI
	}
	return false
}
func (n *xpathNamespaceNode) IsSameNode(other Node) bool {
	otherNS, ok := other.(*xpathNamespaceNode)
	return ok && n == otherNS
}
func (n *xpathNamespaceNode) Contains(other Node) bool { return false }
func (n *xpathNamespaceNode) CompareDocumentPosition(other Node) uint16 {
	// Namespace nodes come after their owner element
	if n.ownerElement == other {
		return DOCUMENT_POSITION_CONTAINED_BY | DOCUMENT_POSITION_FOLLOWING
	}
	return n.ownerElement.CompareDocumentPosition(other)
}
func (n *xpathNamespaceNode) Position() (line, column int, offset int64) {
	return n.ownerElement.Position()
}
func (n *xpathNamespaceNode) LookupNamespaceURI(prefix DOMString) DOMString         { return "" }
func (n *xpathNamespaceNode) LookupPrefix(namespaceURI DOMString) DOMString         { return "" }
func (n *xpathNamespaceNode) IsDefaultNamespace(namespaceURI DOMString) bool        { return false }
func (n *xpathNamespaceNode) Attributes() NamedNodeMap                              { return nil }
func (n *xpathNamespaceNode) HasAttributes() bool                                   { return false }
func (n *xpathNamespaceNode) IsSupported(feature DOMString, version DOMString) bool { return false }
func (n *xpathNamespaceNode) GetRootNode() Node                                     { return n.ownerElement.GetRootNode() }

// Concrete AST node implementations

// xpathPathNode represents path expressions like '/root/book' or '//book[@id="1"]'
type xpathPathNode struct {
	steps []XPathNode
}

func (n xpathPathNode) Type() XPathNodeType { return XPathNodeTypePath }

func (n xpathPathNode) Evaluate(ctx *XPathContext) (XPathValue, error) {
	// Start with the context node as a single-node set
	currentNodes := []Node{ctx.ContextNode}

	// Apply each step in sequence
	for _, step := range n.steps {
		var nextNodes []Node

		// Apply the step to each node in the current node set
		for position, node := range currentNodes {
			// Create new context for this step with proper position tracking
			// XPath positions are 1-based, not 0-based
			stepCtx := &XPathContext{
				ContextNode:       node,
				ContextSize:       len(currentNodes),
				ContextPosition:   position + 1, // Convert to 1-based position
				VariableBindings:  ctx.VariableBindings,
				FunctionLibrary:   ctx.FunctionLibrary,
				NamespaceResolver: ctx.NamespaceResolver,
				Document:          ctx.Document,
				Context:           ctx.Context,
			}

			// Evaluate the step
			result, err := step.Evaluate(stepCtx)
			if err != nil {
				return nil, err
			}

			// Add resulting nodes to the next node set
			if nodeSet, ok := result.(xpathNodeSetValue); ok {
				nextNodes = append(nextNodes, nodeSet.nodes...)
			} else {
				// Non-node-set results are not valid for location steps
				return nil, NewXPathError(XPathErrorTypeType, "location step must return node-set", 0)
			}
		}

		// Remove duplicates and maintain document order
		currentNodes = n.removeDuplicatesAndSort(nextNodes)
	}

	return NewXPathNodeSetValue(currentNodes), nil
}

// removeDuplicatesAndSort removes duplicate nodes and sorts them in document order
// This implements proper document order sorting as required by XPath 1.0 spec
func (n xpathPathNode) removeDuplicatesAndSort(nodes []Node) []Node {
	if len(nodes) == 0 {
		return nodes
	}

	// Use a map to track seen nodes for deduplication
	seen := make(map[Node]bool)
	var uniqueNodes []Node

	for _, node := range nodes {
		if !seen[node] {
			uniqueNodes = append(uniqueNodes, node)
			seen[node] = true
		}
	}

	// Sort nodes in document order using CompareDocumentPosition
	// XPath 1.0 requires node-sets to always be in document order
	sortNodesInDocumentOrder(uniqueNodes)

	return uniqueNodes
}

// xpathRootNode represents the document root for absolute paths
type xpathRootNode struct{}

func (n xpathRootNode) Type() XPathNodeType { return XPathNodeTypePath }

func (n xpathRootNode) Evaluate(ctx *XPathContext) (XPathValue, error) {
	// For XPath root node, we need to return the document node, not the document element
	// The subsequent axis steps will then be able to find the document element as a child

	// Find the document node by traversing up
	document := ctx.ContextNode
	for document.ParentNode() != nil {
		document = document.ParentNode()
	}

	return NewXPathNodeSetValue([]Node{document}), nil
}

// xpathAxisNode represents axis steps like child::element or descendant::*
type xpathAxisNode struct {
	axis       XPathAxis
	nodeTest   XPathNodeTest
	predicates []XPathNode
}

func (n xpathAxisNode) Type() XPathNodeType { return XPathNodeTypeAxis }

func (n xpathAxisNode) Evaluate(ctx *XPathContext) (XPathValue, error) {
	nodes := n.evaluateAxis(ctx.ContextNode, ctx)

	// Apply predicates to the node set
	if len(n.predicates) > 0 {
		// Apply each predicate sequentially to the current node set
		currentNodes := nodes
		for _, predicate := range n.predicates {
			var filteredNodes []Node
			for i, node := range currentNodes {
				// Create new context for predicate evaluation
				predCtx := &XPathContext{
					ContextNode:       node,
					ContextSize:       len(currentNodes), // Size of current node set
					ContextPosition:   i + 1,             // Position within current node set (1-based)
					VariableBindings:  ctx.VariableBindings,
					FunctionLibrary:   ctx.FunctionLibrary,
					NamespaceResolver: ctx.NamespaceResolver,
					Document:          ctx.Document,
					Context:           ctx.Context,
				}

				result, err := predicate.Evaluate(predCtx)
				if err != nil {
					return nil, err
				}

				// XPath 1.0 predicate evaluation rules:
				// If the result is a number, it's a positional predicate
				// Otherwise, convert to boolean
				if n.evaluatePredicate(result, i+1) {
					filteredNodes = append(filteredNodes, node)
				}
			}
			// Update current nodes for the next predicate
			currentNodes = filteredNodes
		}
		nodes = currentNodes
	}

	return NewXPathNodeSetValue(nodes), nil
}

// evaluateAxis performs DOM traversal based on the axis type
func (n xpathAxisNode) evaluateAxis(contextNode Node, ctx *XPathContext) []Node {
	var nodes []Node

	switch n.axis {
	case XPathAxisSelf:
		if n.nodeTest.Matches(contextNode, ctx) {
			nodes = append(nodes, contextNode)
		}

	case XPathAxisChild:
		for child := contextNode.FirstChild(); child != nil; child = child.NextSibling() {
			if n.nodeTest.Matches(child, ctx) {
				nodes = append(nodes, child)
			}
		}

	case XPathAxisParent:
		if parent := contextNode.ParentNode(); parent != nil {
			if n.nodeTest.Matches(parent, ctx) {
				nodes = append(nodes, parent)
			}
		}

	case XPathAxisDescendant:
		n.traverseDescendants(contextNode, false, ctx, &nodes)

	case XPathAxisDescendantOrSelf:
		if n.nodeTest.Matches(contextNode, ctx) {
			nodes = append(nodes, contextNode)
		}
		n.traverseDescendants(contextNode, false, ctx, &nodes)

	case XPathAxisAncestor:
		for ancestor := contextNode.ParentNode(); ancestor != nil; ancestor = ancestor.ParentNode() {
			if n.nodeTest.Matches(ancestor, ctx) {
				nodes = append(nodes, ancestor)
			}
		}

	case XPathAxisAncestorOrSelf:
		if n.nodeTest.Matches(contextNode, ctx) {
			nodes = append(nodes, contextNode)
		}
		for ancestor := contextNode.ParentNode(); ancestor != nil; ancestor = ancestor.ParentNode() {
			if n.nodeTest.Matches(ancestor, ctx) {
				nodes = append(nodes, ancestor)
			}
		}

	case XPathAxisFollowingSibling:
		for sibling := contextNode.NextSibling(); sibling != nil; sibling = sibling.NextSibling() {
			if n.nodeTest.Matches(sibling, ctx) {
				nodes = append(nodes, sibling)
			}
		}

	case XPathAxisPrecedingSibling:
		for sibling := contextNode.PreviousSibling(); sibling != nil; sibling = sibling.PreviousSibling() {
			if n.nodeTest.Matches(sibling, ctx) {
				nodes = append(nodes, sibling)
			}
		}

	case XPathAxisAttribute:
		if elem, ok := contextNode.(Element); ok {
			attrs := elem.Attributes()
			for i := uint(0); i < attrs.Length(); i++ {
				attr := attrs.Item(i)
				if n.nodeTest.Matches(attr, ctx) {
					nodes = append(nodes, attr)
				}
			}
		}

	case XPathAxisNamespace:
		// XPath 1.0 namespace axis - returns namespace nodes for the context element
		// Only elements have namespace nodes
		if elem, ok := contextNode.(Element); ok {
			nodes = n.getNamespaceNodes(elem, ctx)
			// Debug: log how many namespace nodes were found
			// fmt.Printf("DEBUG: namespace axis found %d nodes for element %s\n", len(nodes), elem.NodeName())
		}

	case XPathAxisFollowing:
		// XPath 1.0 following axis - all nodes after context node in document order
		n.traverseFollowing(contextNode, ctx, &nodes)

	case XPathAxisPreceding:
		// XPath 1.0 preceding axis - all nodes before context node in document order
		n.traversePreceding(contextNode, ctx, &nodes)
	}

	return nodes
}

// traverseDescendants recursively traverses descendant nodes
func (n xpathAxisNode) traverseDescendants(node Node, includeSelf bool, ctx *XPathContext, nodes *[]Node) {
	if includeSelf && n.nodeTest.Matches(node, ctx) {
		*nodes = append(*nodes, node)
	}

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if n.nodeTest.Matches(child, ctx) {
			*nodes = append(*nodes, child)
		}
		n.traverseDescendants(child, false, ctx, nodes)
	}
}

// traverseFollowing traverses all nodes that come after the context node in document order
func (n xpathAxisNode) traverseFollowing(contextNode Node, ctx *XPathContext, nodes *[]Node) {
	// Find the document root to start traversal
	document := contextNode
	for document.ParentNode() != nil {
		document = document.ParentNode()
	}

	// Start traversal from document root, collecting nodes after context node
	n.collectFollowing(document, contextNode, ctx, nodes, false)
}

// collectFollowing recursively collects nodes that come after the context node
func (n xpathAxisNode) collectFollowing(current Node, contextNode Node, ctx *XPathContext, nodes *[]Node, foundContext bool) bool {
	found := foundContext

	// Check if we've found the context node
	if current == contextNode {
		found = true
		return found
	}

	// If we've found the context node, collect descendants
	if found {
		if n.nodeTest.Matches(current, ctx) {
			*nodes = append(*nodes, current)
		}
	}

	// Recursively traverse children
	for child := current.FirstChild(); child != nil; child = child.NextSibling() {
		if n.collectFollowing(child, contextNode, ctx, nodes, found) {
			found = true
		}
	}

	return found
}

// traversePreceding traverses all nodes that come before the context node in document order
func (n xpathAxisNode) traversePreceding(contextNode Node, ctx *XPathContext, nodes *[]Node) {
	// Find the document root to start traversal
	document := contextNode
	for document.ParentNode() != nil {
		document = document.ParentNode()
	}

	// Start traversal from document root, collecting nodes before context node
	n.collectPreceding(document, contextNode, ctx, nodes)
}

// collectPreceding recursively collects nodes that come before the context node
func (n xpathAxisNode) collectPreceding(current Node, contextNode Node, ctx *XPathContext, nodes *[]Node) bool {
	// If we've reached the context node, stop collecting
	if current == contextNode {
		return true
	}

	// Collect this node if it matches
	if n.nodeTest.Matches(current, ctx) {
		*nodes = append(*nodes, current)
	}

	// Recursively traverse children
	for child := current.FirstChild(); child != nil; child = child.NextSibling() {
		if n.collectPreceding(child, contextNode, ctx, nodes) {
			return true // Stop when we reach context node
		}
	}

	return false
}

// getNamespaceNodes returns namespace nodes for an element
// XPath 1.0 namespace axis includes all namespace declarations in scope
func (n xpathAxisNode) getNamespaceNodes(elem Element, ctx *XPathContext) []Node {
	var nodes []Node
	namespaces := make(map[string]string) // prefix -> URI mapping

	// Always include the xml namespace (implicit in all documents)
	namespaces["xml"] = "http://www.w3.org/XML/1998/namespace"

	// Walk up the tree collecting namespace declarations
	current := Node(elem)
	for current != nil {
		if currentElem, ok := current.(Element); ok {
			// Check for namespace declarations on this element
			attrs := currentElem.Attributes()
			for i := uint(0); i < attrs.Length(); i++ {
				attrNode := attrs.Item(i)
				// Cast to Attr interface to access Value method
				if attr, ok := attrNode.(Attr); ok {
					attrName := string(attr.NodeName())
					nsURI := string(attr.NamespaceURI())

					// Check for namespace declarations
					// They have namespaceURI == "xmlns" OR name starts with "xmlns"
					if nsURI == "xmlns" {
						// This is a namespace declaration
						prefix := string(attr.LocalName())
						if _, exists := namespaces[prefix]; !exists {
							namespaces[prefix] = string(attr.Value())
						}
					} else if attrName == "xmlns" {
						// Default namespace declaration
						if _, exists := namespaces[""]; !exists {
							namespaces[""] = string(attr.Value())
						}
					} else if strings.HasPrefix(attrName, "xmlns:") {
						// Prefixed namespace declaration (fallback for parsers that don't set nsURI)
						prefix := attrName[6:] // Remove "xmlns:" prefix
						if _, exists := namespaces[prefix]; !exists {
							namespaces[prefix] = string(attr.Value())
						}
					}
				}
			}

			// Also check the element's own namespace
			if ns := currentElem.NamespaceURI(); ns != "" {
				if prefix := currentElem.Prefix(); prefix != "" {
					if _, exists := namespaces[string(prefix)]; !exists {
						namespaces[string(prefix)] = string(ns)
					}
				}
			}
		}
		current = current.ParentNode()
	}

	// Create namespace nodes for all collected namespaces
	for prefix, uri := range namespaces {
		nsNode := &xpathNamespaceNode{
			prefix:       prefix,
			namespaceURI: uri,
			ownerElement: elem,
		}

		// Apply node test to filter namespace nodes
		if n.nodeTest.Matches(nsNode, ctx) {
			nodes = append(nodes, nsNode)
		}
	}

	return nodes
}

// evaluatePredicate evaluates a predicate result according to XPath 1.0 rules
// Returns true if the predicate matches for the given position
func (n xpathAxisNode) evaluatePredicate(result XPathValue, position int) bool {
	// XPath 1.0 predicate evaluation rules:
	// - If the result is a number, compare it with the current position
	// - Otherwise, convert to boolean

	switch result.Type() {
	case XPathValueTypeNumber:
		// Positional predicate: [1], [2], [last()], etc.
		numValue := result.Number()
		// Convert to integer and compare with 1-based position
		// XPath uses 1-based indexing
		if math.IsNaN(numValue) || math.IsInf(numValue, 0) {
			return false
		}
		predPosition := int(math.Round(numValue))
		return predPosition == position

	default:
		// Boolean predicate: [@id='1'], [author='John'], etc.
		return booleanValueOf(result)
	}
}

// xpathLiteralNode represents string and number literals
type xpathLiteralNode struct {
	value XPathValue
}

func (n xpathLiteralNode) Type() XPathNodeType { return XPathNodeTypeLiteral }

func (n xpathLiteralNode) Evaluate(ctx *XPathContext) (XPathValue, error) {
	return n.value, nil
}

// xpathBinaryOpNode represents binary operations like +, =, and, or
type xpathBinaryOpNode struct {
	operator    XPathOperator
	left, right XPathNode
}

func (n xpathBinaryOpNode) Type() XPathNodeType { return XPathNodeTypeBinary }

func (n xpathBinaryOpNode) Evaluate(ctx *XPathContext) (XPathValue, error) {
	// Evaluate left and right operands
	leftValue, err := n.left.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	rightValue, err := n.right.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	switch n.operator {
	// Logical operators
	case XPathOperatorOr:
		leftBool := booleanValueOf(leftValue)
		if leftBool {
			return NewXPathBooleanValue(true), nil
		}
		rightBool := booleanValueOf(rightValue)
		return NewXPathBooleanValue(rightBool), nil

	case XPathOperatorAnd:
		leftBool := booleanValueOf(leftValue)
		if !leftBool {
			return NewXPathBooleanValue(false), nil
		}
		rightBool := booleanValueOf(rightValue)
		return NewXPathBooleanValue(rightBool), nil

	// Equality operators
	case XPathOperatorEq:
		return NewXPathBooleanValue(n.compareValues(leftValue, rightValue, "==")), nil

	case XPathOperatorNeq:
		return NewXPathBooleanValue(!n.compareValues(leftValue, rightValue, "==")), nil

	// Relational operators
	case XPathOperatorLt:
		return NewXPathBooleanValue(n.compareValues(leftValue, rightValue, "<")), nil

	case XPathOperatorLte:
		return NewXPathBooleanValue(n.compareValues(leftValue, rightValue, "<=")), nil

	case XPathOperatorGt:
		return NewXPathBooleanValue(n.compareValues(leftValue, rightValue, ">")), nil

	case XPathOperatorGte:
		return NewXPathBooleanValue(n.compareValues(leftValue, rightValue, ">=")), nil

	// Arithmetic operators
	case XPathOperatorPlus:
		leftNum := numberValueOf(leftValue)
		rightNum := numberValueOf(rightValue)
		return NewXPathNumberValue(leftNum + rightNum), nil

	case XPathOperatorMinus:
		leftNum := numberValueOf(leftValue)
		rightNum := numberValueOf(rightValue)
		return NewXPathNumberValue(leftNum - rightNum), nil

	case XPathOperatorMultiply:
		leftNum := numberValueOf(leftValue)
		rightNum := numberValueOf(rightValue)
		return NewXPathNumberValue(leftNum * rightNum), nil

	case XPathOperatorDiv:
		leftNum := numberValueOf(leftValue)
		rightNum := numberValueOf(rightValue)
		if rightNum == 0 {
			// XPath 1.0: division by zero should result in positive or negative infinity
			if leftNum >= 0 {
				return NewXPathNumberValue(math.Inf(1)), nil // Positive infinity
			} else {
				return NewXPathNumberValue(math.Inf(-1)), nil // Negative infinity
			}
		}
		return NewXPathNumberValue(leftNum / rightNum), nil

	case XPathOperatorMod:
		leftNum := numberValueOf(leftValue)
		rightNum := numberValueOf(rightValue)
		if rightNum == 0 {
			// XPath 1.0: mod by zero should result in NaN
			return NewXPathNumberValue(math.NaN()), nil
		}
		// XPath mod operator has specific rules for negative numbers
		result := leftNum - rightNum*float64(int(leftNum/rightNum))
		return NewXPathNumberValue(result), nil

	// Union operator
	case XPathOperatorUnion:
		leftNodes := []Node{}
		rightNodes := []Node{}

		if leftValue.Type() == XPathValueTypeNodeSet {
			leftNodes = leftValue.NodeSet()
		}
		if rightValue.Type() == XPathValueTypeNodeSet {
			rightNodes = rightValue.NodeSet()
		}

		// Combine node sets and remove duplicates
		unionNodes := make([]Node, 0, len(leftNodes)+len(rightNodes))
		nodeSet := make(map[Node]bool)

		for _, node := range leftNodes {
			if !nodeSet[node] {
				unionNodes = append(unionNodes, node)
				nodeSet[node] = true
			}
		}
		for _, node := range rightNodes {
			if !nodeSet[node] {
				unionNodes = append(unionNodes, node)
				nodeSet[node] = true
			}
		}

		// XPath 1.0 requires union results to be in document order
		sortNodesInDocumentOrder(unionNodes)

		return NewXPathNodeSetValue(unionNodes), nil

	default:
		return nil, NewXPathError(XPathErrorTypeType, "unsupported binary operator", 0)
	}
}

// compareValues compares two XPathValue instances according to XPath 1.0 rules
func (n xpathBinaryOpNode) compareValues(left, right XPathValue, op string) bool {
	// XPath 1.0 comparison rules are complex - simplified implementation

	// If both are node-sets, iterate through combinations
	if left.Type() == XPathValueTypeNodeSet && right.Type() == XPathValueTypeNodeSet {
		leftNodes := left.NodeSet()
		rightNodes := right.NodeSet()

		for _, lNode := range leftNodes {
			for _, rNode := range rightNodes {
				leftStr := string(lNode.TextContent())
				rightStr := string(rNode.TextContent())
				if n.compareStrings(leftStr, rightStr, op) {
					return true
				}
			}
		}
		return false
	}

	// If one is a node-set, handle according to XPath 1.0 rules
	if left.Type() == XPathValueTypeNodeSet {
		leftNodes := left.NodeSet()
		// Empty node-set never equals anything (XPath 1.0 spec)
		if len(leftNodes) == 0 {
			return false
		}
		// Compare each node's string value with the right operand
		rightStr := stringValueOf(right)
		for _, node := range leftNodes {
			var nodeStr string
			if node.NodeType() == ATTRIBUTE_NODE {
				if attr, ok := node.(Attr); ok {
					nodeStr = string(attr.Value())
				}
			} else {
				nodeStr = string(node.TextContent())
			}
			if n.compareStrings(nodeStr, rightStr, op) {
				return true
			}
		}
		return false
	}
	if right.Type() == XPathValueTypeNodeSet {
		rightNodes := right.NodeSet()
		// Empty node-set never equals anything (XPath 1.0 spec)
		if len(rightNodes) == 0 {
			return false
		}
		// Compare left operand with each node's string value
		leftStr := stringValueOf(left)
		for _, node := range rightNodes {
			var nodeStr string
			if node.NodeType() == ATTRIBUTE_NODE {
				if attr, ok := node.(Attr); ok {
					nodeStr = string(attr.Value())
				}
			} else {
				nodeStr = string(node.TextContent())
			}
			if n.compareStrings(leftStr, nodeStr, op) {
				return true
			}
		}
		return false
	}

	// For non-node-set values, follow XPath 1.0 type conversion rules
	switch op {
	case "==":
		// If both are booleans or one is a boolean, convert to boolean
		if left.Type() == XPathValueTypeBoolean || right.Type() == XPathValueTypeBoolean {
			return booleanValueOf(left) == booleanValueOf(right)
		}
		// If both are numbers or one is a number, convert to number
		if left.Type() == XPathValueTypeNumber || right.Type() == XPathValueTypeNumber {
			return numberValueOf(left) == numberValueOf(right)
		}
		// Otherwise compare as strings
		return stringValueOf(left) == stringValueOf(right)

	case "<", "<=", ">", ">=":
		// Relational operators always compare as numbers
		leftNum := numberValueOf(left)
		rightNum := numberValueOf(right)
		switch op {
		case "<":
			return leftNum < rightNum
		case "<=":
			return leftNum <= rightNum
		case ">":
			return leftNum > rightNum
		case ">=":
			return leftNum >= rightNum
		}
	}

	return false
}

// compareStrings compares two strings according to the given operator
func (n xpathBinaryOpNode) compareStrings(left, right, op string) bool {
	switch op {
	case "==":
		return left == right
	case "<":
		return left < right
	case "<=":
		return left <= right
	case ">":
		return left > right
	case ">=":
		return left >= right
	default:
		return false
	}
}

// xpathUnaryOpNode represents unary operations like -expr
type xpathUnaryOpNode struct {
	operator XPathOperator
	operand  XPathNode
}

func (n xpathUnaryOpNode) Type() XPathNodeType { return XPathNodeTypeUnary }

func (n xpathUnaryOpNode) Evaluate(ctx *XPathContext) (XPathValue, error) {
	operandValue, err := n.operand.Evaluate(ctx)
	if err != nil {
		return nil, err
	}

	switch n.operator {
	case XPathOperatorUnaryMinus:
		numVal := numberValueOf(operandValue)
		return NewXPathNumberValue(-numVal), nil
	default:
		return nil, NewXPathError(XPathErrorTypeType, "unsupported unary operator", 0)
	}
}

// xpathFunctionNode represents function calls
type xpathFunctionNode struct {
	name string
	args []XPathNode
}

func (n xpathFunctionNode) Type() XPathNodeType { return XPathNodeTypeFunction }

func (n xpathFunctionNode) Evaluate(ctx *XPathContext) (XPathValue, error) {
	if fn, exists := ctx.FunctionLibrary[n.name]; exists {
		// Evaluate arguments
		argValues := make([]XPathValue, len(n.args))
		for i, arg := range n.args {
			val, err := arg.Evaluate(ctx)
			if err != nil {
				return nil, err
			}
			argValues[i] = val
		}
		return fn.Call(ctx, argValues)
	}
	return nil, NewXPathError(XPathErrorTypeFunction, "Unknown function: "+n.name, 0)
}

// XPathError represents XPath-specific errors
type XPathError struct {
	Type     XPathErrorType
	Message  string
	Position int
}

func (e *XPathError) Error() string {
	return e.Message
}

// NewXPathError creates a new XPathError
func NewXPathError(errorType XPathErrorType, message string, position int) *XPathError {
	return &XPathError{
		Type:     errorType,
		Message:  message,
		Position: position,
	}
}

// NewXPathException creates a new XPath exception (alias for NewXPathError)
func NewXPathException(errorType string, message string) *XPathError {
	// Map string error types to XPathErrorType constants
	var errType XPathErrorType
	switch errorType {
	case XPathExceptionINVALID_EXPRESSION_ERR:
		errType = XPathErrorTypeSyntax
	case XPathExceptionTYPE_ERR:
		errType = XPathErrorTypeType
	case XPathExceptionWRONG_DOCUMENT_ERR:
		errType = XPathErrorTypeContext
	case XPathExceptionNAMESPACE_ERR:
		errType = XPathErrorTypeNamespace
	case XPathExceptionNOT_SUPPORTED_ERR:
		errType = XPathErrorTypeNotImplemented
	default:
		errType = XPathErrorTypeType
	}
	return NewXPathError(errType, message, 0)
}

// XPathErrorType represents different categories of XPath errors
// Following XPath 1.0 specification and DOM Level 3 XPath
type XPathErrorType uint8

const (
	XPathErrorTypeSyntax XPathErrorType = iota
	XPathErrorTypeType
	XPathErrorTypeFunction
	XPathErrorTypeAxis
	XPathErrorTypeContext
	XPathErrorTypeNamespace
	XPathErrorTypeNotImplemented
)

// XPath 1.0 standard exception codes
const (
	XPathExceptionINVALID_EXPRESSION_ERR = "INVALID_EXPRESSION_ERR" // Syntax error in expression
	XPathExceptionTYPE_ERR               = "TYPE_ERR"               // Type mismatch
	XPathExceptionWRONG_DOCUMENT_ERR     = "WRONG_DOCUMENT_ERR"     // Wrong document context
	XPathExceptionNAMESPACE_ERR          = "NAMESPACE_ERR"          // Namespace resolution error
	XPathExceptionNOT_SUPPORTED_ERR      = "NOT_SUPPORTED_ERR"      // Feature not supported
)

// Helper functions for type conversions following XPath 1.0 spec

func stringToNumber(s string) float64 {
	// XPath 1.0 string-to-number conversion rules:
	// 1. Skip leading and trailing whitespace
	// 2. If empty or contains non-numeric chars, return NaN
	// 3. Otherwise convert to number

	s = strings.TrimSpace(s)
	if s == "" {
		return math.NaN()
	}

	// Try to parse as float
	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return math.NaN()
	}

	return num
}

func numberToString(n float64) string {
	// XPath 1.0 number-to-string conversion rules:
	// 1. NaN -> "NaN"
	// 2. Positive infinity -> "Infinity"
	// 3. Negative infinity -> "-Infinity"
	// 4. Integer values without decimal point
	// 5. Decimal values with trailing zeros removed

	if math.IsNaN(n) {
		return "NaN"
	}
	if math.IsInf(n, 1) {
		return "Infinity"
	}
	if math.IsInf(n, -1) {
		return "-Infinity"
	}

	// Check if it's an integer
	if n == float64(int64(n)) {
		return strconv.FormatInt(int64(n), 10)
	}

	// Format as float, removing trailing zeros
	return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(n, 'f', -1, 64), "0"), ".")
}

func booleanToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func booleanToNumber(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func nodeSetToString(nodes []Node) string {
	// XPath 1.0: string value of node-set is string value of first node in document order
	if len(nodes) == 0 {
		return ""
	}
	node := nodes[0]

	// For attribute nodes, we need to get the attribute value, not text content
	if node.NodeType() == ATTRIBUTE_NODE {
		if attr, ok := node.(Attr); ok {
			return string(attr.Value())
		}
	}
	return string(node.TextContent())
}

func isNaN(f float64) bool {
	return f != f
}

// ===========================================================================
// XPath Implementation Structures
// ===========================================================================

// xpathExpression implements XPathExpression interface
type xpathExpression struct {
	expression       string
	resolver         XPathNSResolver
	ast              XPathNode
	document         *document
	variableBindings map[string]XPathValue
	mu               sync.RWMutex // Protect variable bindings
}

// SetVariableBindings sets variable bindings for the expression
func (xe *xpathExpression) SetVariableBindings(bindings map[string]XPathValue) {
	xe.mu.Lock()
	defer xe.mu.Unlock()
	// Make a copy to avoid external modifications
	xe.variableBindings = make(map[string]XPathValue)
	for k, v := range bindings {
		xe.variableBindings[k] = v
	}
}

func (xe *xpathExpression) Evaluate(contextNode Node, resultType uint16, result XPathResult) (XPathResult, error) {
	if contextNode == nil {
		return nil, NewXPathException("TYPE_ERR", "Context node cannot be null")
	}

	// Copy variable bindings under read lock for thread safety
	xe.mu.RLock()
	varBindings := make(map[string]XPathValue)
	for k, v := range xe.variableBindings {
		varBindings[k] = v
	}
	xe.mu.RUnlock()

	// Create evaluation context
	context := &XPathContext{
		ContextNode:       contextNode,
		ContextSize:       1,
		ContextPosition:   1,
		VariableBindings:  varBindings,
		FunctionLibrary:   getBuiltinFunctions(),
		NamespaceResolver: xe.resolver,
		Document:          xe.document,
	}

	// Evaluate AST
	value, err := xe.ast.Evaluate(context)
	if err != nil {
		return nil, NewXPathException("TYPE_ERR", err.Error())
	}

	// Convert to requested result type
	return xe.convertToResult(value, resultType, result)
}

func (xe *xpathExpression) convertToResult(value XPathValue, resultType uint16, result XPathResult) (XPathResult, error) {
	switch resultType {
	case XPATH_ANY_TYPE:
		// Return the most appropriate type
		switch value.(type) {
		case xpathNodeSetValue:
			return xe.convertToResult(value, XPATH_UNORDERED_NODE_ITERATOR_TYPE, result)
		case xpathStringValue:
			return xe.convertToResult(value, XPATH_STRING_TYPE, result)
		case xpathNumberValue:
			return xe.convertToResult(value, XPATH_NUMBER_TYPE, result)
		case xpathBooleanValue:
			return xe.convertToResult(value, XPATH_BOOLEAN_TYPE, result)
		}
	case XPATH_STRING_TYPE:
		strVal := stringValueOf(value)
		return &xpathStringResult{value: strVal}, nil
	case XPATH_NUMBER_TYPE:
		numVal := numberValueOf(value)
		return &xpathNumberResult{value: numVal}, nil
	case XPATH_BOOLEAN_TYPE:
		boolVal := booleanValueOf(value)
		return &xpathBooleanResult{value: boolVal}, nil
	case XPATH_UNORDERED_NODE_ITERATOR_TYPE, XPATH_ORDERED_NODE_ITERATOR_TYPE:
		if nodeSet, ok := value.(xpathNodeSetValue); ok {
			return &xpathNodeIteratorResult{
				nodes:        nodeSet.nodes,
				currentIndex: -1,
				resultType:   resultType,
			}, nil
		}
		return nil, NewXPathException("TYPE_ERR", "Cannot convert to node iterator")
	case XPATH_UNORDERED_NODE_SNAPSHOT_TYPE, XPATH_ORDERED_NODE_SNAPSHOT_TYPE:
		if nodeSet, ok := value.(xpathNodeSetValue); ok {
			return &xpathNodeSnapshotResult{
				nodes:      nodeSet.nodes,
				resultType: resultType,
			}, nil
		}
		return nil, NewXPathException("TYPE_ERR", "Cannot convert to node snapshot")
	case XPATH_ANY_UNORDERED_NODE_TYPE, XPATH_FIRST_ORDERED_NODE_TYPE:
		if nodeSet, ok := value.(xpathNodeSetValue); ok {
			if len(nodeSet.nodes) > 0 {
				return &xpathSingleNodeResult{
					node:       nodeSet.nodes[0],
					resultType: resultType,
				}, nil
			}
			return &xpathSingleNodeResult{
				node:       nil,
				resultType: resultType,
			}, nil
		}
		return nil, NewXPathException("TYPE_ERR", "Cannot convert to single node")
	}
	return nil, NewXPathException("TYPE_ERR", "Unsupported result type")
}

// Helper functions for value conversion
func stringValueOf(value XPathValue) string {
	switch v := value.(type) {
	case xpathStringValue:
		return v.value
	case xpathNumberValue:
		return numberToString(v.value)
	case xpathBooleanValue:
		return booleanToString(v.value)
	case xpathNodeSetValue:
		return nodeSetToString(v.nodes)
	}
	return ""
}

func numberValueOf(value XPathValue) float64 {
	switch v := value.(type) {
	case xpathStringValue:
		return stringToNumber(v.value)
	case xpathNumberValue:
		return v.value
	case xpathBooleanValue:
		return booleanToNumber(v.value)
	case xpathNodeSetValue:
		return stringToNumber(nodeSetToString(v.nodes))
	}
	return 0
}

func booleanValueOf(value XPathValue) bool {
	switch v := value.(type) {
	case xpathStringValue:
		return v.value != ""
	case xpathNumberValue:
		return v.value != 0 && !isNaN(v.value)
	case xpathBooleanValue:
		return v.value
	case xpathNodeSetValue:
		return len(v.nodes) > 0
	}
	return false
}

// ===========================================================================
// XPath Result Implementations
// ===========================================================================

// xpathStringResult implements XPathResult for string values
type xpathStringResult struct {
	value string
}

func (r *xpathStringResult) ResultType() uint16 {
	return XPATH_STRING_TYPE
}

func (r *xpathStringResult) NumberValue() (float64, error) {
	return 0, NewXPathException("TYPE_ERR", "Result is not a number")
}

func (r *xpathStringResult) StringValue() (string, error) {
	return r.value, nil
}

func (r *xpathStringResult) BooleanValue() (bool, error) {
	return false, NewXPathException("TYPE_ERR", "Result is not a boolean")
}

func (r *xpathStringResult) SingleNodeValue() (Node, error) {
	return nil, NewXPathException("TYPE_ERR", "Result is not a node")
}

func (r *xpathStringResult) InvalidIteratorState() bool {
	return false
}

func (r *xpathStringResult) SnapshotLength() (uint32, error) {
	return 0, NewXPathException("TYPE_ERR", "Result is not a snapshot")
}

func (r *xpathStringResult) IterateNext() (Node, error) {
	return nil, NewXPathException("TYPE_ERR", "Result is not an iterator")
}

func (r *xpathStringResult) SnapshotItem(index uint32) (Node, error) {
	return nil, NewXPathException("TYPE_ERR", "Result is not a snapshot")
}

// xpathNumberResult implements XPathResult for numeric values
type xpathNumberResult struct {
	value float64
}

func (r *xpathNumberResult) ResultType() uint16 {
	return XPATH_NUMBER_TYPE
}

func (r *xpathNumberResult) NumberValue() (float64, error) {
	return r.value, nil
}

func (r *xpathNumberResult) StringValue() (string, error) {
	return "", NewXPathException("TYPE_ERR", "Result is not a string")
}

func (r *xpathNumberResult) BooleanValue() (bool, error) {
	return false, NewXPathException("TYPE_ERR", "Result is not a boolean")
}

func (r *xpathNumberResult) SingleNodeValue() (Node, error) {
	return nil, NewXPathException("TYPE_ERR", "Result is not a node")
}

func (r *xpathNumberResult) InvalidIteratorState() bool {
	return false
}

func (r *xpathNumberResult) SnapshotLength() (uint32, error) {
	return 0, NewXPathException("TYPE_ERR", "Result is not a snapshot")
}

func (r *xpathNumberResult) IterateNext() (Node, error) {
	return nil, NewXPathException("TYPE_ERR", "Result is not an iterator")
}

func (r *xpathNumberResult) SnapshotItem(index uint32) (Node, error) {
	return nil, NewXPathException("TYPE_ERR", "Result is not a snapshot")
}

// xpathBooleanResult implements XPathResult for boolean values
type xpathBooleanResult struct {
	value bool
}

func (r *xpathBooleanResult) ResultType() uint16 {
	return XPATH_BOOLEAN_TYPE
}

func (r *xpathBooleanResult) NumberValue() (float64, error) {
	return 0, NewXPathException("TYPE_ERR", "Result is not a number")
}

func (r *xpathBooleanResult) StringValue() (string, error) {
	return "", NewXPathException("TYPE_ERR", "Result is not a string")
}

func (r *xpathBooleanResult) BooleanValue() (bool, error) {
	return r.value, nil
}

func (r *xpathBooleanResult) SingleNodeValue() (Node, error) {
	return nil, NewXPathException("TYPE_ERR", "Result is not a node")
}

func (r *xpathBooleanResult) InvalidIteratorState() bool {
	return false
}

func (r *xpathBooleanResult) SnapshotLength() (uint32, error) {
	return 0, NewXPathException("TYPE_ERR", "Result is not a snapshot")
}

func (r *xpathBooleanResult) IterateNext() (Node, error) {
	return nil, NewXPathException("TYPE_ERR", "Result is not an iterator")
}

func (r *xpathBooleanResult) SnapshotItem(index uint32) (Node, error) {
	return nil, NewXPathException("TYPE_ERR", "Result is not a snapshot")
}

// xpathSingleNodeResult implements XPathResult for single node values
type xpathSingleNodeResult struct {
	node       Node
	resultType uint16
}

func (r *xpathSingleNodeResult) ResultType() uint16 {
	return r.resultType
}

func (r *xpathSingleNodeResult) NumberValue() (float64, error) {
	return 0, NewXPathException("TYPE_ERR", "Result is not a number")
}

func (r *xpathSingleNodeResult) StringValue() (string, error) {
	return "", NewXPathException("TYPE_ERR", "Result is not a string")
}

func (r *xpathSingleNodeResult) BooleanValue() (bool, error) {
	return false, NewXPathException("TYPE_ERR", "Result is not a boolean")
}

func (r *xpathSingleNodeResult) SingleNodeValue() (Node, error) {
	return r.node, nil
}

func (r *xpathSingleNodeResult) InvalidIteratorState() bool {
	return false
}

func (r *xpathSingleNodeResult) SnapshotLength() (uint32, error) {
	return 0, NewXPathException("TYPE_ERR", "Result is not a snapshot")
}

func (r *xpathSingleNodeResult) IterateNext() (Node, error) {
	return nil, NewXPathException("TYPE_ERR", "Result is not an iterator")
}

func (r *xpathSingleNodeResult) SnapshotItem(index uint32) (Node, error) {
	return nil, NewXPathException("TYPE_ERR", "Result is not a snapshot")
}

// xpathNodeIteratorResult implements XPathResult for node iterator values
type xpathNodeIteratorResult struct {
	nodes        []Node
	currentIndex int
	resultType   uint16
}

func (r *xpathNodeIteratorResult) ResultType() uint16 {
	return r.resultType
}

func (r *xpathNodeIteratorResult) NumberValue() (float64, error) {
	return 0, NewXPathException("TYPE_ERR", "Result is not a number")
}

func (r *xpathNodeIteratorResult) StringValue() (string, error) {
	return "", NewXPathException("TYPE_ERR", "Result is not a string")
}

func (r *xpathNodeIteratorResult) BooleanValue() (bool, error) {
	return false, NewXPathException("TYPE_ERR", "Result is not a boolean")
}

func (r *xpathNodeIteratorResult) SingleNodeValue() (Node, error) {
	return nil, NewXPathException("TYPE_ERR", "Result is not a single node")
}

func (r *xpathNodeIteratorResult) InvalidIteratorState() bool {
	// TODO: Check for document mutations that would invalidate iterator
	return false
}

func (r *xpathNodeIteratorResult) SnapshotLength() (uint32, error) {
	return 0, NewXPathException("TYPE_ERR", "Result is not a snapshot")
}

func (r *xpathNodeIteratorResult) IterateNext() (Node, error) {
	r.currentIndex++
	if r.currentIndex >= len(r.nodes) {
		return nil, nil
	}
	return r.nodes[r.currentIndex], nil
}

func (r *xpathNodeIteratorResult) SnapshotItem(index uint32) (Node, error) {
	return nil, NewXPathException("TYPE_ERR", "Result is not a snapshot")
}

// xpathNodeSnapshotResult implements XPathResult for node snapshot values
type xpathNodeSnapshotResult struct {
	nodes      []Node
	resultType uint16
}

func (r *xpathNodeSnapshotResult) ResultType() uint16 {
	return r.resultType
}

func (r *xpathNodeSnapshotResult) NumberValue() (float64, error) {
	return 0, NewXPathException("TYPE_ERR", "Result is not a number")
}

func (r *xpathNodeSnapshotResult) StringValue() (string, error) {
	return "", NewXPathException("TYPE_ERR", "Result is not a string")
}

func (r *xpathNodeSnapshotResult) BooleanValue() (bool, error) {
	return false, NewXPathException("TYPE_ERR", "Result is not a boolean")
}

func (r *xpathNodeSnapshotResult) SingleNodeValue() (Node, error) {
	return nil, NewXPathException("TYPE_ERR", "Result is not a single node")
}

func (r *xpathNodeSnapshotResult) InvalidIteratorState() bool {
	return false
}

func (r *xpathNodeSnapshotResult) SnapshotLength() (uint32, error) {
	return uint32(len(r.nodes)), nil
}

func (r *xpathNodeSnapshotResult) IterateNext() (Node, error) {
	return nil, NewXPathException("TYPE_ERR", "Result is not an iterator")
}

func (r *xpathNodeSnapshotResult) SnapshotItem(index uint32) (Node, error) {
	if index >= uint32(len(r.nodes)) {
		return nil, nil
	}
	return r.nodes[index], nil
}

// ===========================================================================
// Built-in Functions
// ===========================================================================

// getBuiltinFunctions returns the standard XPath 1.0 built-in functions
func getBuiltinFunctions() map[string]XPathFunction {
	return map[string]XPathFunction{
		// Node set functions
		"last": &xpathBuiltinFunction{
			name:    "last",
			minArgs: 0,
			maxArgs: 0,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				return NewXPathNumberValue(float64(context.ContextSize)), nil
			},
		},
		"position": &xpathBuiltinFunction{
			name:    "position",
			minArgs: 0,
			maxArgs: 0,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				return NewXPathNumberValue(float64(context.ContextPosition)), nil
			},
		},
		"count": &xpathBuiltinFunction{
			name:    "count",
			minArgs: 1,
			maxArgs: 1,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				if nodeSet, ok := args[0].(xpathNodeSetValue); ok {
					return NewXPathNumberValue(float64(len(nodeSet.nodes))), nil
				}
				return nil, NewXPathException("TYPE_ERR", "count() argument must be a node-set")
			},
		},
		"id": &xpathBuiltinFunction{
			name:    "id",
			minArgs: 1,
			maxArgs: 1,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				// Get the ID string (may be space-separated list of IDs)
				idStr := stringValueOf(args[0])
				ids := strings.Fields(idStr) // Split on whitespace

				var nodes []Node
				for _, id := range ids {
					if element := context.Document.GetElementById(DOMString(id)); element != nil {
						nodes = append(nodes, element)
					}
				}

				return NewXPathNodeSetValue(nodes), nil
			},
		},
		"local-name": &xpathBuiltinFunction{
			name:    "local-name",
			minArgs: 0,
			maxArgs: 1,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				var node Node

				if len(args) == 0 {
					// Use context node
					node = context.ContextNode
				} else {
					// Use first node from node-set argument
					if nodeSet, ok := args[0].(xpathNodeSetValue); ok {
						if len(nodeSet.nodes) == 0 {
							return NewXPathStringValue(""), nil
						}
						node = nodeSet.nodes[0]
					} else {
						return nil, NewXPathException("TYPE_ERR", "local-name() argument must be a node-set")
					}
				}

				// Return local name of the node
				localName := node.LocalName()
				return NewXPathStringValue(string(localName)), nil
			},
		},
		"namespace-uri": &xpathBuiltinFunction{
			name:    "namespace-uri",
			minArgs: 0,
			maxArgs: 1,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				var node Node

				if len(args) == 0 {
					// Use context node
					node = context.ContextNode
				} else {
					// Use first node from node-set argument
					if nodeSet, ok := args[0].(xpathNodeSetValue); ok {
						if len(nodeSet.nodes) == 0 {
							return NewXPathStringValue(""), nil
						}
						node = nodeSet.nodes[0]
					} else {
						return nil, NewXPathException("TYPE_ERR", "namespace-uri() argument must be a node-set")
					}
				}

				// Return namespace URI of the node
				namespaceURI := node.NamespaceURI()
				return NewXPathStringValue(string(namespaceURI)), nil
			},
		},
		"name": &xpathBuiltinFunction{
			name:    "name",
			minArgs: 0,
			maxArgs: 1,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				var node Node

				if len(args) == 0 {
					// Use context node
					node = context.ContextNode
				} else {
					// Use first node from node-set argument
					if nodeSet, ok := args[0].(xpathNodeSetValue); ok {
						if len(nodeSet.nodes) == 0 {
							return NewXPathStringValue(""), nil
						}
						node = nodeSet.nodes[0]
					} else {
						return nil, NewXPathException("TYPE_ERR", "name() argument must be a node-set")
					}
				}

				// Return qualified name of the node
				nodeName := node.NodeName()
				return NewXPathStringValue(string(nodeName)), nil
			},
		},
		// String functions
		"string": &xpathBuiltinFunction{
			name:    "string",
			minArgs: 0,
			maxArgs: 1,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				if len(args) == 0 {
					// Convert context node to string
					strVal := context.ContextNode.TextContent()
					return NewXPathStringValue(string(strVal)), nil
				}
				strVal := stringValueOf(args[0])
				return NewXPathStringValue(strVal), nil
			},
		},
		// Number functions
		"number": &xpathBuiltinFunction{
			name:    "number",
			minArgs: 0,
			maxArgs: 1,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				if len(args) == 0 {
					// Convert context node to number
					strVal := context.ContextNode.TextContent()
					numVal := stringToNumber(string(strVal))
					return NewXPathNumberValue(numVal), nil
				}
				numVal := numberValueOf(args[0])
				return NewXPathNumberValue(numVal), nil
			},
		},
		"sum": &xpathBuiltinFunction{
			name:    "sum",
			minArgs: 1,
			maxArgs: 1,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				if nodeSet, ok := args[0].(xpathNodeSetValue); ok {
					var sum float64
					for _, node := range nodeSet.nodes {
						// Convert each node's string value to number and add to sum
						var nodeStr string
						if node.NodeType() == ATTRIBUTE_NODE {
							if attr, ok := node.(Attr); ok {
								nodeStr = string(attr.Value())
							}
						} else {
							nodeStr = string(node.TextContent())
						}
						nodeNum := stringToNumber(nodeStr)
						sum += nodeNum
					}
					return NewXPathNumberValue(sum), nil
				}
				return nil, NewXPathException("TYPE_ERR", "sum() argument must be a node-set")
			},
		},
		"floor": &xpathBuiltinFunction{
			name:    "floor",
			minArgs: 1,
			maxArgs: 1,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				numVal := numberValueOf(args[0])

				// Handle special values
				if math.IsNaN(numVal) {
					return NewXPathNumberValue(math.NaN()), nil
				}
				if math.IsInf(numVal, 0) {
					return NewXPathNumberValue(numVal), nil // Infinity remains infinity
				}

				return NewXPathNumberValue(math.Floor(numVal)), nil
			},
		},
		"ceiling": &xpathBuiltinFunction{
			name:    "ceiling",
			minArgs: 1,
			maxArgs: 1,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				numVal := numberValueOf(args[0])

				// Handle special values
				if math.IsNaN(numVal) {
					return NewXPathNumberValue(math.NaN()), nil
				}
				if math.IsInf(numVal, 0) {
					return NewXPathNumberValue(numVal), nil // Infinity remains infinity
				}

				return NewXPathNumberValue(math.Ceil(numVal)), nil
			},
		},
		"round": &xpathBuiltinFunction{
			name:    "round",
			minArgs: 1,
			maxArgs: 1,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				numVal := numberValueOf(args[0])

				// Handle special values
				if math.IsNaN(numVal) {
					return NewXPathNumberValue(math.NaN()), nil
				}
				if math.IsInf(numVal, 0) {
					return NewXPathNumberValue(numVal), nil // Infinity remains infinity
				}

				// XPath 1.0 rounding rules: round to nearest integer
				// For exactly .5 values, round away from zero (banker's rounding not used)
				if numVal >= 0 {
					return NewXPathNumberValue(math.Floor(numVal + 0.5)), nil
				} else {
					return NewXPathNumberValue(math.Ceil(numVal - 0.5)), nil
				}
			},
		},
		// Boolean functions
		"boolean": &xpathBuiltinFunction{
			name:    "boolean",
			minArgs: 1,
			maxArgs: 1,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				boolVal := booleanValueOf(args[0])
				return NewXPathBooleanValue(boolVal), nil
			},
		},
		"not": &xpathBuiltinFunction{
			name:    "not",
			minArgs: 1,
			maxArgs: 1,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				boolVal := booleanValueOf(args[0])
				return NewXPathBooleanValue(!boolVal), nil
			},
		},
		"true": &xpathBuiltinFunction{
			name:    "true",
			minArgs: 0,
			maxArgs: 0,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				return NewXPathBooleanValue(true), nil
			},
		},
		"false": &xpathBuiltinFunction{
			name:    "false",
			minArgs: 0,
			maxArgs: 0,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				return NewXPathBooleanValue(false), nil
			},
		},
		"lang": &xpathBuiltinFunction{
			name:    "lang",
			minArgs: 1,
			maxArgs: 1,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				testLang := strings.ToLower(stringValueOf(args[0]))

				// Walk up the tree looking for xml:lang attribute
				for node := context.ContextNode; node != nil; node = node.ParentNode() {
					if elem, ok := node.(Element); ok {
						// Check for xml:lang attribute
						if langAttr := elem.GetAttributeNS("http://www.w3.org/XML/1998/namespace", "lang"); string(langAttr) != "" {
							langValue := strings.ToLower(string(langAttr))

							// XPath 1.0 lang() function rules:
							// Returns true if the context node is in the language specified by the argument
							// Language matching is case-insensitive
							// Supports language subtag matching (e.g., "en" matches "en-US")

							// Exact match
							if langValue == testLang {
								return NewXPathBooleanValue(true), nil
							}

							// Check if testLang is a prefix followed by '-'
							// e.g., "en" should match "en-US", "en-GB", etc.
							if strings.HasPrefix(langValue, testLang+"-") {
								return NewXPathBooleanValue(true), nil
							}

							// No match found, return false
							return NewXPathBooleanValue(false), nil
						}
					}
				}

				// No xml:lang attribute found in the hierarchy
				return NewXPathBooleanValue(false), nil
			},
		},
		// Additional string functions
		"concat": &xpathBuiltinFunction{
			name:    "concat",
			minArgs: 2,
			maxArgs: -1, // unlimited args
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				var result strings.Builder
				for _, arg := range args {
					result.WriteString(stringValueOf(arg))
				}
				return NewXPathStringValue(result.String()), nil
			},
		},
		"contains": &xpathBuiltinFunction{
			name:    "contains",
			minArgs: 2,
			maxArgs: 2,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				str1 := stringValueOf(args[0])
				str2 := stringValueOf(args[1])
				return NewXPathBooleanValue(strings.Contains(str1, str2)), nil
			},
		},
		"starts-with": &xpathBuiltinFunction{
			name:    "starts-with",
			minArgs: 2,
			maxArgs: 2,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				str1 := stringValueOf(args[0])
				str2 := stringValueOf(args[1])
				return NewXPathBooleanValue(strings.HasPrefix(str1, str2)), nil
			},
		},
		"string-length": &xpathBuiltinFunction{
			name:    "string-length",
			minArgs: 0,
			maxArgs: 1,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				var str string
				if len(args) == 0 {
					// Use context node
					str = string(context.ContextNode.TextContent())
				} else {
					str = stringValueOf(args[0])
				}
				return NewXPathNumberValue(float64(len([]rune(str)))), nil // Count Unicode characters, not bytes
			},
		},
		"normalize-space": &xpathBuiltinFunction{
			name:    "normalize-space",
			minArgs: 0,
			maxArgs: 1,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				var str string
				if len(args) == 0 {
					// Use context node
					str = string(context.ContextNode.TextContent())
				} else {
					str = stringValueOf(args[0])
				}
				// Normalize whitespace: trim and collapse internal whitespace
				fields := strings.Fields(str)
				return NewXPathStringValue(strings.Join(fields, " ")), nil
			},
		},
		"substring": &xpathBuiltinFunction{
			name:    "substring",
			minArgs: 2,
			maxArgs: 3,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				str := stringValueOf(args[0])
				runes := []rune(str)
				start := int(numberValueOf(args[1])) - 1 // XPath uses 1-based indexing

				if start < 0 {
					start = 0
				}
				if start >= len(runes) {
					return NewXPathStringValue(""), nil
				}

				if len(args) == 3 {
					length := int(numberValueOf(args[2]))
					end := start + length
					if end > len(runes) {
						end = len(runes)
					}
					return NewXPathStringValue(string(runes[start:end])), nil
				} else {
					return NewXPathStringValue(string(runes[start:])), nil
				}
			},
		},
		"substring-before": &xpathBuiltinFunction{
			name:    "substring-before",
			minArgs: 2,
			maxArgs: 2,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				str := stringValueOf(args[0])
				substr := stringValueOf(args[1])

				// If substring is empty, return empty string
				if substr == "" {
					return NewXPathStringValue(""), nil
				}

				// Find first occurrence of substring
				index := strings.Index(str, substr)
				if index == -1 {
					// Substring not found, return empty string
					return NewXPathStringValue(""), nil
				}

				// Return everything before the substring
				return NewXPathStringValue(str[:index]), nil
			},
		},
		"substring-after": &xpathBuiltinFunction{
			name:    "substring-after",
			minArgs: 2,
			maxArgs: 2,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				str := stringValueOf(args[0])
				substr := stringValueOf(args[1])

				// If substring is empty, return the original string
				if substr == "" {
					return NewXPathStringValue(str), nil
				}

				// Find first occurrence of substring
				index := strings.Index(str, substr)
				if index == -1 {
					// Substring not found, return empty string
					return NewXPathStringValue(""), nil
				}

				// Return everything after the substring
				return NewXPathStringValue(str[index+len(substr):]), nil
			},
		},
		"translate": &xpathBuiltinFunction{
			name:    "translate",
			minArgs: 3,
			maxArgs: 3,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				str := stringValueOf(args[0])
				fromChars := stringValueOf(args[1])
				toChars := stringValueOf(args[2])

				// Convert from/to strings to runes for proper Unicode handling
				fromRunes := []rune(fromChars)
				toRunes := []rune(toChars)

				// Create translation map
				translationMap := make(map[rune]rune)
				for i, fromRune := range fromRunes {
					if i < len(toRunes) {
						// Map to corresponding character in toChars
						translationMap[fromRune] = toRunes[i]
					} else {
						// No corresponding character in toChars, mark for removal
						translationMap[fromRune] = 0 // Use 0 to indicate removal
					}
				}

				// Apply translation - range directly over string for efficiency
				var result []rune
				for _, r := range str { // Range over string directly, not []rune(str)
					if replacement, exists := translationMap[r]; exists {
						if replacement != 0 {
							// Replace with mapped character
							result = append(result, replacement)
						}
						// If replacement is 0, the character is removed (not appended)
					} else {
						// Character not in translation map, keep as-is
						result = append(result, r)
					}
				}

				return NewXPathStringValue(string(result)), nil
			},
		},
		// Additional XPath 1.0 functions
		"comment": &xpathBuiltinFunction{
			name:    "comment",
			minArgs: 0,
			maxArgs: 0,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				// XPath 1.0 comment() function - returns true if context node is a comment
				return NewXPathBooleanValue(context.ContextNode.NodeType() == COMMENT_NODE), nil
			},
		},
		"processing-instruction": &xpathBuiltinFunction{
			name:    "processing-instruction",
			minArgs: 0,
			maxArgs: 1,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				if len(args) == 0 {
					// Return true if context node is a processing instruction
					return NewXPathBooleanValue(context.ContextNode.NodeType() == PROCESSING_INSTRUCTION_NODE), nil
				}
				// Check if context node is a PI with the specified target
				target := stringValueOf(args[0])
				if context.ContextNode.NodeType() == PROCESSING_INSTRUCTION_NODE {
					pi, ok := context.ContextNode.(ProcessingInstruction)
					if ok && string(pi.Target()) == target {
						return NewXPathBooleanValue(true), nil
					}
				}
				return NewXPathBooleanValue(false), nil
			},
		},
		"text": &xpathBuiltinFunction{
			name:    "text",
			minArgs: 0,
			maxArgs: 0,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				// XPath 1.0 text() function - returns true if context node is a text node
				return NewXPathBooleanValue(context.ContextNode.NodeType() == TEXT_NODE), nil
			},
		},
		// Additional XPath 1.0 functions
		"node": &xpathBuiltinFunction{
			name:    "node",
			minArgs: 0,
			maxArgs: 0,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				// XPath 1.0 node() function - returns true for any node
				return NewXPathBooleanValue(true), nil
			},
		},
		"ancestor": &xpathBuiltinFunction{
			name:    "ancestor",
			minArgs: 0,
			maxArgs: 0,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				// XPath 1.0 ancestor axis - this is typically used as ancestor::
				return NewXPathBooleanValue(false), nil
			},
		},
		"ancestor-or-self": &xpathBuiltinFunction{
			name:    "ancestor-or-self",
			minArgs: 0,
			maxArgs: 0,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				// XPath 1.0 ancestor-or-self axis - this is typically used as ancestor-or-self::
				return NewXPathBooleanValue(true), nil
			},
		},
		"child": &xpathBuiltinFunction{
			name:    "child",
			minArgs: 0,
			maxArgs: 0,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				// XPath 1.0 child axis - this is typically used as child::
				return NewXPathBooleanValue(false), nil
			},
		},
		"descendant": &xpathBuiltinFunction{
			name:    "descendant",
			minArgs: 0,
			maxArgs: 0,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				// XPath 1.0 descendant axis - this is typically used as descendant::
				return NewXPathBooleanValue(false), nil
			},
		},
		"descendant-or-self": &xpathBuiltinFunction{
			name:    "descendant-or-self",
			minArgs: 0,
			maxArgs: 0,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				// XPath 1.0 descendant-or-self axis - this is typically used as descendant-or-self::
				return NewXPathBooleanValue(true), nil
			},
		},
		"following": &xpathBuiltinFunction{
			name:    "following",
			minArgs: 0,
			maxArgs: 0,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				// XPath 1.0 following axis - this is typically used as following::
				return NewXPathBooleanValue(false), nil
			},
		},
		"following-sibling": &xpathBuiltinFunction{
			name:    "following-sibling",
			minArgs: 0,
			maxArgs: 0,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				// XPath 1.0 following-sibling axis - this is typically used as following-sibling::
				return NewXPathBooleanValue(false), nil
			},
		},
		"parent": &xpathBuiltinFunction{
			name:    "parent",
			minArgs: 0,
			maxArgs: 0,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				// XPath 1.0 parent axis - this is typically used as parent::
				return NewXPathBooleanValue(false), nil
			},
		},
		"preceding": &xpathBuiltinFunction{
			name:    "preceding",
			minArgs: 0,
			maxArgs: 0,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				// XPath 1.0 preceding axis - this is typically used as preceding::
				return NewXPathBooleanValue(false), nil
			},
		},
		"preceding-sibling": &xpathBuiltinFunction{
			name:    "preceding-sibling",
			minArgs: 0,
			maxArgs: 0,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				// XPath 1.0 preceding-sibling axis - this is typically used as preceding-sibling::
				return NewXPathBooleanValue(false), nil
			},
		},
		"self": &xpathBuiltinFunction{
			name:    "self",
			minArgs: 0,
			maxArgs: 0,
			impl: func(context *XPathContext, args []XPathValue) (XPathValue, error) {
				// XPath 1.0 self axis - this is typically used as self::
				return NewXPathBooleanValue(true), nil
			},
		},
	}
}

// xpathBuiltinFunction implements XPathFunction for built-in functions
type xpathBuiltinFunction struct {
	name    string
	minArgs int
	maxArgs int
	impl    func(*XPathContext, []XPathValue) (XPathValue, error)
}

func (f *xpathBuiltinFunction) Call(context *XPathContext, args []XPathValue) (XPathValue, error) {
	if len(args) < f.minArgs || (f.maxArgs != -1 && len(args) > f.maxArgs) {
		return nil, NewXPathException("TYPE_ERR", fmt.Sprintf("%s() requires %d to %d arguments, got %d", f.name, f.minArgs, f.maxArgs, len(args)))
	}
	return f.impl(context, args)
}

func (f *xpathBuiltinFunction) MinArgs() int {
	return f.minArgs
}

func (f *xpathBuiltinFunction) MaxArgs() int {
	return f.maxArgs
}

// ===========================================================================
// XPath Parser Interface (Implementation in xpath_parser.go)
// ===========================================================================

// NewXPathParser creates a new XPath parser
func NewXPathParser() *XPathParser {
	return &XPathParser{}
}

// Parse function is implemented in xpath_parser.go
// This provides the main entry point for XPath expression parsing
