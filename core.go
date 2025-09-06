package xmldom

import (
	"fmt"
	"reflect"
	"strings"
	sync "sync"
)

// DOMString is a string type used in the DOM.
// The DOM specification defines DOMString as a sequence of 16-bit units,
// which corresponds to UTF-16. However, for pragmatic reasons, this implementation
// uses Go's native UTF-8 strings. This avoids costly conversions at every API
// boundary. This is a documented deviation from the specification.
type DOMString string

type NodeType = uint16

// DOM node type constants
const (
	ELEMENT_NODE                NodeType = 1
	ATTRIBUTE_NODE              NodeType = 2
	TEXT_NODE                   NodeType = 3
	CDATA_SECTION_NODE          NodeType = 4
	ENTITY_REFERENCE_NODE       NodeType = 5
	ENTITY_NODE                 NodeType = 6
	PROCESSING_INSTRUCTION_NODE NodeType = 7
	COMMENT_NODE                NodeType = 8
	DOCUMENT_NODE               NodeType = 9
	DOCUMENT_TYPE_NODE          NodeType = 10
	DOCUMENT_FRAGMENT_NODE      NodeType = 11
	NOTATION_NODE               NodeType = 12
)

type DocumentPositionType = uint16

// DocumentPosition constants
const (
	DOCUMENT_POSITION_DISCONNECTED            DocumentPositionType = 0x01
	DOCUMENT_POSITION_PRECEDING               DocumentPositionType = 0x02
	DOCUMENT_POSITION_FOLLOWING               DocumentPositionType = 0x04
	DOCUMENT_POSITION_CONTAINS                DocumentPositionType = 0x08
	DOCUMENT_POSITION_CONTAINED_BY            DocumentPositionType = 0x10
	DOCUMENT_POSITION_IMPLEMENTATION_SPECIFIC DocumentPositionType = 0x20
)

type NodeFilterType = uint16

// NodeFilter constants
const (
	FILTER_ACCEPT NodeFilterType = 1
	FILTER_REJECT NodeFilterType = 2
	FILTER_SKIP   NodeFilterType = 3
)

type ShowWhatType = uint32

// NodeIterator/TreeWalker whatToShow constants
const (
	SHOW_ALL                    ShowWhatType = 0xFFFFFFFF
	SHOW_ELEMENT                ShowWhatType = 0x00000001
	SHOW_ATTRIBUTE              ShowWhatType = 0x00000002
	SHOW_TEXT                   ShowWhatType = 0x00000004
	SHOW_CDATA_SECTION          ShowWhatType = 0x00000008
	SHOW_ENTITY_REFERENCE       ShowWhatType = 0x00000010
	SHOW_ENTITY                 ShowWhatType = 0x00000020
	SHOW_PROCESSING_INSTRUCTION ShowWhatType = 0x00000040
	SHOW_COMMENT                ShowWhatType = 0x00000080
	SHOW_DOCUMENT               ShowWhatType = 0x00000100
	SHOW_DOCUMENT_TYPE          ShowWhatType = 0x00000200
	SHOW_DOCUMENT_FRAGMENT      ShowWhatType = 0x00000400
	SHOW_NOTATION               ShowWhatType = 0x00000800
)

// NodeFilter interface
type NodeFilter interface {
	AcceptNode(node Node) uint16
}

// DOM exception codes

// ===========================================================================
// Interfaces
// ===========================================================================

// Node interface represents a node in the DOM tree
type Node interface {
	NodeType() NodeType
	NodeName() DOMString
	NodeValue() DOMString
	SetNodeValue(value DOMString) error
	ParentNode() Node
	ChildNodes() NodeList
	FirstChild() Node
	LastChild() Node
	PreviousSibling() Node
	NextSibling() Node
	Attributes() NamedNodeMap
	OwnerDocument() Document
	InsertBefore(newChild Node, refChild Node) (Node, error)
	ReplaceChild(newChild Node, oldChild Node) (Node, error)
	RemoveChild(oldChild Node) (Node, error)
	AppendChild(newChild Node) (Node, error)
	HasChildNodes() bool
	CloneNode(deep bool) Node
	Normalize()
	IsSupported(feature DOMString, version DOMString) bool
	NamespaceURI() DOMString
	Prefix() DOMString
	SetPrefix(prefix DOMString) error
	LocalName() DOMString
	HasAttributes() bool
	BaseURI() DOMString
	IsConnected() bool
	CompareDocumentPosition(otherNode Node) DocumentPositionType
	Contains(otherNode Node) bool
	GetRootNode() Node
	IsDefaultNamespace(namespaceURI DOMString) bool
	IsEqualNode(otherNode Node) bool
	IsSameNode(otherNode Node) bool
	LookupPrefix(namespaceURI DOMString) DOMString
	LookupNamespaceURI(prefix DOMString) DOMString
	TextContent() DOMString
	SetTextContent(value DOMString)
}

// Document interface represents a document node
type Document interface {
	Node
	Doctype() DocumentType
	Implementation() DOMImplementation
	DocumentElement() Element
	CreateElement(tagName DOMString) (Element, error)
	CreateDocumentFragment() DocumentFragment
	CreateTextNode(data DOMString) Text
	CreateComment(data DOMString) Comment
	CreateCDATASection(data DOMString) (CDATASection, error)
	CreateProcessingInstruction(target, data DOMString) (ProcessingInstruction, error)
	CreateAttribute(name DOMString) (Attr, error)
	CreateEntityReference(name DOMString) (EntityReference, error)
	GetElementsByTagName(tagname DOMString) NodeList
	ImportNode(importedNode Node, deep bool) (Node, error)
	CreateElementNS(namespaceURI, qualifiedName DOMString) (Element, error)
	CreateAttributeNS(namespaceURI, qualifiedName DOMString) (Attr, error)
	GetElementsByTagNameNS(namespaceURI, localName DOMString) NodeList
	GetElementById(elementId DOMString) Element
	AdoptNode(source Node) (Node, error)
	CreateNodeIterator(root Node, whatToShow ShowWhatType, filter NodeFilter) (NodeIterator, error)
	CreateTreeWalker(root Node, whatToShow ShowWhatType, filter NodeFilter) (TreeWalker, error)
	CreateRange() Range
	NormalizeDocument()
	RenameNode(node Node, namespaceURI, qualifiedName DOMString) (Node, error)

	// XPath evaluation methods following DOM Living Standard
	CreateExpression(expression string, resolver XPathNSResolver) (XPathExpression, error)
	CreateNSResolver(nodeResolver Node) Node
	Evaluate(expression string, contextNode Node, resolver XPathNSResolver,
		resultType uint16, result XPathResult) (XPathResult, error)

	// Document properties
	URL() DOMString
	DocumentURI() DOMString
	CharacterSet() DOMString
	Charset() DOMString
	InputEncoding() DOMString
	ContentType() DOMString
}

// Element interface represents an element node
type Element interface {
	Node
	TagName() DOMString
	GetAttribute(name DOMString) DOMString
	SetAttribute(name, value DOMString) error
	RemoveAttribute(name DOMString) error
	GetAttributeNode(name DOMString) Attr
	SetAttributeNode(newAttr Attr) (Attr, error)
	RemoveAttributeNode(oldAttr Attr) (Attr, error)
	GetElementsByTagName(name DOMString) NodeList
	GetAttributeNS(namespaceURI, localName DOMString) DOMString
	SetAttributeNS(namespaceURI, qualifiedName, value DOMString) error
	RemoveAttributeNS(namespaceURI, localName DOMString) error
	GetAttributeNodeNS(namespaceURI, localName DOMString) Attr
	SetAttributeNodeNS(newAttr Attr) (Attr, error)
	GetElementsByTagNameNS(namespaceURI, localName DOMString) NodeList
	HasAttribute(name DOMString) bool
	HasAttributeNS(namespaceURI, localName DOMString) bool

	// Element manipulation methods from Living Standard (applicable to XML)
	ToggleAttribute(name DOMString, force ...bool) bool
	Remove()
	ReplaceWith(nodes ...Node) error
	Before(nodes ...Node) error
	After(nodes ...Node) error
	Prepend(nodes ...Node) error
	Append(nodes ...Node) error

	// Element DOM properties from Living Standard
	Children() ElementList // Returns live collection of child elements
	FirstElementChild() Element
	LastElementChild() Element
	PreviousElementSibling() Element
	NextElementSibling() Element
	ChildElementCount() uint32
}

// Attr interface represents an attribute node
type Attr interface {
	Node
	Name() DOMString
	Value() DOMString
	SetValue(value DOMString)
	OwnerElement() Element
}

// CharacterData interface represents character data
type CharacterData interface {
	Node
	Data() DOMString
	SetData(data DOMString) error
	Length() uint
	SubstringData(offset, count uint) (DOMString, error)
	AppendData(arg DOMString) error
	InsertData(offset uint, arg DOMString) error
	DeleteData(offset, count uint) error
	ReplaceData(offset, count uint, arg DOMString) error

	// CharacterData manipulation methods from Living Standard
	Before(nodes ...Node) error
	After(nodes ...Node) error
	ReplaceWith(nodes ...Node) error
	Remove()
}

// Text interface represents a text node
type Text interface {
	CharacterData
	SplitText(offset uint) (Text, error)
}

// Comment interface represents a comment node
type Comment interface {
	CharacterData
}

// CDATASection interface represents a CDATA section
type CDATASection interface {
	Text
}

// DocumentType interface represents a document type node
type DocumentType interface {
	Node
	Name() DOMString
	Entities() NamedNodeMap
	Notations() NamedNodeMap
	PublicId() DOMString
	SystemId() DOMString
	InternalSubset() DOMString
}

// Notation interface represents a notation node
type Notation interface {
	Node
	PublicId() DOMString
	SystemId() DOMString
}

// Entity interface represents an entity node
type Entity interface {
	Node
	PublicId() DOMString
	SystemId() DOMString
	NotationName() DOMString
}

// EntityReference interface represents an entity reference node
type EntityReference interface {
	Node
}

// ProcessingInstruction interface represents a processing instruction node
type ProcessingInstruction interface {
	Node
	Target() DOMString
	Data() DOMString
	SetData(data DOMString) error
}

// DocumentFragment interface represents a document fragment node
type DocumentFragment interface {
	Node
}

// DOMImplementation interface provides methods for operations independent of any document instance
type DOMImplementation interface {
	HasFeature(feature, version DOMString) bool
	CreateDocumentType(qualifiedName, publicId, systemId DOMString) (DocumentType, error)
	CreateDocument(namespaceURI, qualifiedName DOMString, doctype DocumentType) (Document, error)
}

// NodeList interface represents an ordered collection of nodes
type NodeList interface {
	Item(index uint) Node
	Length() uint
}

// NamedNodeMap interface represents a collection of nodes accessible by name
type NamedNodeMap interface {
	GetNamedItem(name DOMString) Node
	SetNamedItem(arg Node) (Node, error)
	RemoveNamedItem(name DOMString) (Node, error)
	Item(index uint) Node
	Length() uint
	GetNamedItemNS(namespaceURI, localName DOMString) Node
	SetNamedItemNS(arg Node) (Node, error)
	RemoveNamedItemNS(namespaceURI, localName DOMString) (Node, error)
}

// NodeIterator interface
type NodeIterator interface {
	Root() Node
	ReferenceNode() Node
	PointerBeforeReferenceNode() bool
	WhatToShow() uint32
	Filter() NodeFilter
	NextNode() (Node, error)
	PreviousNode() (Node, error)
	Detach()
}

// TreeWalker interface
type TreeWalker interface {
	Root() Node
	WhatToShow() uint32
	Filter() NodeFilter
	CurrentNode() Node
	SetCurrentNode(node Node) error
	ParentNode() Node
	FirstChild() Node
	LastChild() Node
	PreviousSibling() Node
	NextSibling() Node
	PreviousNode() Node
	NextNode() Node
}

// Range interface
type Range interface {
	StartContainer() Node
	StartOffset() uint32
	EndContainer() Node
	EndOffset() uint32
	Collapsed() bool
	CommonAncestorContainer() Node

	SetStart(node Node, offset uint32) error
	SetEnd(node Node, offset uint32) error
	SetStartBefore(node Node) error
	SetStartAfter(node Node) error
	SetEndBefore(node Node) error
	SetEndAfter(node Node) error
	Collapse(toStart bool)
	SelectNode(node Node) error
	SelectNodeContents(node Node) error

	CompareBoundaryPoints(how uint16, sourceRange Range) (int16, error)
	DeleteContents() error
	ExtractContents() (DocumentFragment, error)
	CloneContents() (DocumentFragment, error)
	InsertNode(node Node) error
	SurroundContents(newParent Node) error

	CloneRange() Range
	Detach()
	IsPointInRange(node Node, offset uint32) (bool, error)
	ComparePoint(node Node, offset uint32) (int16, error)
	IntersectsNode(node Node) bool

	ToString() string
}

// Range comparison constants
const (
	START_TO_START uint16 = 0
	START_TO_END   uint16 = 1
	END_TO_END     uint16 = 2
	END_TO_START   uint16 = 3
)

// ElementList interface - collection of Element nodes
type ElementList interface {
	Length() uint
	Item(index uint) Element
}

// ===========================================================================
// Core Implementation Types
// ===========================================================================

// DOMException represents a DOM exception as defined in the Living Standard.
type DOMException struct {
	name    DOMString
	message string
}

func (e *DOMException) Error() string {
	return fmt.Sprintf("%s: %s", e.name, e.message)
}

// NewDOMException creates a new DOMException with the given name and message.
func NewDOMException(name DOMString, message string) *DOMException {
	return &DOMException{
		name:    name,
		message: message,
	}
}

// supportedFeatures defines the features and versions supported by this DOM implementation
var supportedFeatures = map[DOMString][]DOMString{
	"Core": {"2.0"},
	"XML":  {"1.0", "2.0"},
}

// liveList is a generic list implementation that can work with both Node and Element types
type liveList[T any] struct {
	items     []T                  // The cached list of items. For live lists, this is updated upon mutation.
	root      Node                 // The root node for the query that generated this list.
	filter    func(Node) bool      // The filter function to apply to the nodes.
	converter func(Node) (T, bool) // Function to convert Node to T type
	live      bool                 // True if the list is live and should be updated on mutations.
	doc       *document            // A pointer to the owner document, used to access activeNodeLists.
	update    func()               // The function to call to update the list of items.
}

// nodeList represents an ordered collection of nodes
type nodeList = liveList[Node]

// elementList represents an ordered collection of elements
type elementList = liveList[Element]

// Item returns the item at the given index in the list.
// If the index is out of bounds, it returns the zero value.
func (dl *liveList[T]) Item(index uint) T {
	var zero T
	if dl.doc != nil {
		dl.doc.mu.RLock()
		defer dl.doc.mu.RUnlock()
	}
	if index >= uint(len(dl.items)) {
		return zero
	}
	return dl.items[index]
}

// Length returns the number of items in the list.
func (dl *liveList[T]) Length() uint {
	if dl.doc != nil {
		dl.doc.mu.RLock()
		defer dl.doc.mu.RUnlock()
	}
	return uint(len(dl.items))
}

// namedNodeMap represents a collection of nodes accessible by name
type namedNodeMap struct {
	items map[DOMString]Node
	order []DOMString
}

func NewNamedNodeMap() *namedNodeMap {
	return &namedNodeMap{
		items: make(map[DOMString]Node),
		order: []DOMString{},
	}
}

func (nnm *namedNodeMap) GetNamedItem(name DOMString) Node {
	return nnm.items[name]
}

func (nnm *namedNodeMap) SetNamedItem(arg Node) (Node, error) {
	if arg.NodeType() != ATTRIBUTE_NODE {
		return nil, NewDOMException("HierarchyRequestError", "Node is not an attribute")
	}
	name := arg.NodeName()
	oldArg := nnm.items[name]
	if oldArg == nil {
		nnm.order = append(nnm.order, name)
	}
	nnm.items[name] = arg
	return oldArg, nil
}

func (nnm *namedNodeMap) RemoveNamedItem(name DOMString) (Node, error) {
	node := nnm.items[name]
	if node == nil {
		return nil, NewDOMException("NotFoundError", "Node not found")
	}
	delete(nnm.items, name)
	for i, n := range nnm.order {
		if n == name {
			nnm.order = append(nnm.order[:i], nnm.order[i+1:]...)
			break
		}
	}
	return node, nil
}

func (nnm *namedNodeMap) Item(index uint) Node {
	if index >= uint(len(nnm.order)) {
		return nil
	}
	return nnm.items[nnm.order[index]]
}

func (nnm *namedNodeMap) Length() uint {
	return uint(len(nnm.order))
}

func (nnm *namedNodeMap) GetNamedItemNS(namespaceURI, localName DOMString) Node {
	for _, node := range nnm.items {
		if node.NamespaceURI() == namespaceURI && node.LocalName() == localName {
			return node
		}
	}
	return nil
}

func (nnm *namedNodeMap) SetNamedItemNS(arg Node) (Node, error) {
	if arg.NodeType() != ATTRIBUTE_NODE {
		return nil, NewDOMException("HierarchyRequestError", "Node is not an attribute")
	}
	var oldArg Node
	var oldName DOMString
	for name, node := range nnm.items {
		if node.NamespaceURI() == arg.NamespaceURI() && node.LocalName() == arg.LocalName() {
			oldArg = node
			oldName = name
			break
		}
	}
	if oldArg != nil {
		delete(nnm.items, oldName)
		for i, name := range nnm.order {
			if name == oldName {
				nnm.order[i] = arg.NodeName()
				break
			}
		}
	} else {
		nnm.order = append(nnm.order, arg.NodeName())
	}
	nnm.items[arg.NodeName()] = arg
	return oldArg, nil
}

func (nnm *namedNodeMap) RemoveNamedItemNS(namespaceURI, localName DOMString) (Node, error) {
	for name, node := range nnm.items {
		if node.NamespaceURI() == namespaceURI && node.LocalName() == localName {
			delete(nnm.items, name)
			for i, n := range nnm.order {
				if n == name {
					nnm.order = append(nnm.order[:i], nnm.order[i+1:]...)
					break
				}
			}
			return node, nil
		}
	}
	return nil, NewDOMException("NotFoundError", "Node not found")
}

// ===========================================================================
// Base Node Implementation
// ===========================================================================

// node represents a node in the DOM tree
type node struct {
	nodeType        uint16
	nodeName        DOMString
	nodeValue       DOMString
	parentNode      Node
	childNodes      *nodeList
	firstChild      Node
	lastChild       Node
	previousSibling Node
	nextSibling     Node
	attributes      *namedNodeMap
	ownerDocument   Document
	namespaceURI    DOMString
	prefix          DOMString
	localName       DOMString
}

func (n *node) NodeType() uint16 {
	return n.nodeType
}

func (n *node) NodeName() DOMString {
	return n.nodeName
}

func (n *node) NodeValue() DOMString {
	return n.nodeValue
}

func (n *node) SetNodeValue(value DOMString) error {
	switch n.nodeType {
	case ATTRIBUTE_NODE:
		// This is handled by Attr.SetValue, but we need to allow it here for the interface
		n.nodeValue = value
		return nil
	case TEXT_NODE, COMMENT_NODE, CDATA_SECTION_NODE:
		// This is handled by CharacterData.SetData, but we need to allow it here for the interface
		n.nodeValue = value
		return nil
	case PROCESSING_INSTRUCTION_NODE:
		// This is handled by ProcessingInstruction.SetData, but we need to allow it here for the interface
		n.nodeValue = value
		return nil
	case ELEMENT_NODE, ENTITY_REFERENCE_NODE, ENTITY_NODE, DOCUMENT_NODE, DOCUMENT_TYPE_NODE, DOCUMENT_FRAGMENT_NODE, NOTATION_NODE:
		return NewDOMException("NoModificationAllowedError", "Node value is read-only for this node type")
	}
	return nil
}

func (n *node) ParentNode() Node {
	if n.parentNode == nil {
		return nil
	}
	return n.parentNode
}

func (n *node) ChildNodes() NodeList {
	if doc := n.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.RLock()
			defer d.mu.RUnlock()
		}
	}
	if n.childNodes == nil {
		doc, _ := n.ownerDocument.(*document)
		nl := &nodeList{
			live: true,
			doc:  doc,
		}
		nl.update = func() {
			nodes := []Node{}
			for child := n.firstChild; child != nil; child = child.NextSibling() {
				nodes = append(nodes, child)
			}
			nl.items = nodes
		}
		nl.update()

		if doc != nil {
			if doc.activeNodeLists == nil {
				doc.activeNodeLists = []*nodeList{}
			}
			doc.activeNodeLists = append(doc.activeNodeLists, nl)
		}
		n.childNodes = nl
	}
	return n.childNodes
}

func (n *node) FirstChild() Node {
	if n.firstChild == nil {
		return nil
	}
	return n.firstChild
}

func (n *node) LastChild() Node {
	if n.lastChild == nil {
		return nil
	}
	return n.lastChild
}

func (n *node) PreviousSibling() Node {
	if n.previousSibling == nil {
		return nil
	}
	return n.previousSibling
}

func (n *node) NextSibling() Node {
	if n.nextSibling == nil {
		return nil
	}
	return n.nextSibling
}

func (n *node) Attributes() NamedNodeMap {
	if doc := n.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.RLock()
			defer d.mu.RUnlock()
		}
	}
	if n.attributes != nil {
		return n.attributes
	}
	return nil
}

func (n *node) OwnerDocument() Document {
	if n.ownerDocument == nil {
		return nil
	}
	return n.ownerDocument
}

func (n *node) InsertBefore(newChild Node, refChild Node) (Node, error) {
	if newChild == nil {
		return nil, NewDOMException("HierarchyRequestError", "Invalid node")
	}

	// Handle self-insertion: inserting node before itself should be a no-op
	if newChild == refChild {
		return newChild, nil
	}

	// Check document ownership
	if newChild.OwnerDocument() != n.ownerDocument && n.ownerDocument != nil {
		return nil, NewDOMException("WrongDocumentError", "")
	}

	// HIERARCHY_REQUEST_ERR check - prevent cycles
	for ancestor := Node(n); ancestor != nil; ancestor = ancestor.ParentNode() {
		if ancestor == newChild {
			return nil, NewDOMException("HierarchyRequestError", "Cannot insert a node as a descendant of itself")
		}
	}

	// Handle DocumentFragment - insert its children instead of the fragment itself
	if newChild.NodeType() == DOCUMENT_FRAGMENT_NODE {
		// Collect all children of the fragment first
		var children []Node
		for child := newChild.FirstChild(); child != nil; child = child.NextSibling() {
			children = append(children, child)
		}

		// Insert each child using internal method to avoid infinite recursion
		for _, child := range children {
			// Remove from fragment first
			if df, ok := newChild.(*documentFragment); ok {
				df.removeChildInternal(child)
			} else {
				newChild.RemoveChild(child)
			}
			// Insert into target parent using internal method
			_, err := n.insertBeforeInternal(child, refChild)
			if err != nil {
				return nil, err
			}
		}

		// Return the fragment itself (which is now empty)
		return newChild, nil
	}

	return n.insertBeforeInternal(newChild, refChild)
}

// insertBeforeInternal handles the actual insertion without DocumentFragment expansion
func (n *node) insertBeforeInternal(newChild Node, refChild Node) (Node, error) {

	// Remove from current parent if exists - done internally to avoid deadlock
	if newChild.ParentNode() != nil {
		oldParent := newChild.ParentNode()
		oc := getInternalNode(newChild)

		// Update sibling links
		if oc.previousSibling != nil {
			if prevNode := getInternalNode(oc.previousSibling); prevNode != nil {
				prevNode.nextSibling = oc.nextSibling
			}
		}
		if oc.nextSibling != nil {
			if nextNode := getInternalNode(oc.nextSibling); nextNode != nil {
				nextNode.previousSibling = oc.previousSibling
			}
		}

		// Update parent's first/last child pointers
		if op := getInternalNode(oldParent); op != nil {
			if op.firstChild == newChild {
				op.firstChild = oc.nextSibling
			}
			if op.lastChild == newChild {
				op.lastChild = oc.previousSibling
			}
			// Update old parent's live NodeList if it exists
			if op.childNodes != nil && op.childNodes.update != nil {
				op.childNodes.update()
			}
		}

		// Clear the removed node's parent/sibling references
		oc.parentNode = nil
		oc.previousSibling = nil
		oc.nextSibling = nil
	}

	// Validate refChild
	if refChild != nil && refChild.ParentNode() != Node(n) {
		return nil, NewDOMException("NotFoundError", "refChild not found")
	}

	// Get internal nodes for manipulation
	nc := getInternalNode(newChild)
	rc := getInternalNode(refChild)

	// Append or insert operation
	if refChild == nil {
		nc.parentNode = Node(n)
		nc.previousSibling = n.lastChild
		nc.nextSibling = nil
		if n.lastChild != nil {
			if lastNode := getInternalNode(n.lastChild); lastNode != nil {
				lastNode.nextSibling = newChild
			}
		}
		n.lastChild = newChild
		if n.firstChild == nil {
			n.firstChild = newChild
		}
	} else { // Insert operation
		nc.parentNode = Node(n)
		nc.nextSibling = refChild
		nc.previousSibling = rc.previousSibling
		if rc.previousSibling != nil {
			if prevNode := getInternalNode(rc.previousSibling); prevNode != nil {
				prevNode.nextSibling = newChild
			}
		} else {
			n.firstChild = newChild
		}
		rc.previousSibling = newChild
	}

	// Update live NodeList if it exists
	if n.childNodes != nil && n.childNodes.update != nil {
		n.childNodes.update()
	}
	if doc := n.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.notifyMutation()
		}
	}
	return newChild, nil
}

func (n *node) ReplaceChild(newChild Node, oldChild Node) (Node, error) {
	if newChild == nil {
		return nil, NewDOMException("HierarchyRequestError", "Invalid node")
	}

	if oldChild.ParentNode() != Node(n) {
		return nil, NewDOMException("NotFoundError", "")
	}

	// Handle self-replacement: replacing node with itself should be a no-op
	if newChild == oldChild {
		return oldChild, nil
	}

	// Check document ownership
	if newChild.OwnerDocument() != n.ownerDocument && n.ownerDocument != nil {
		return nil, NewDOMException("WrongDocumentError", "")
	}

	// HIERARCHY_REQUEST_ERR check - prevent cycles
	for ancestor := Node(n); ancestor != nil; ancestor = ancestor.ParentNode() {
		if ancestor == newChild {
			return nil, NewDOMException("HierarchyRequestError", "Cannot insert a node as a descendant of itself")
		}
	}

	// Handle DocumentFragment - replace with its children
	if newChild.NodeType() == DOCUMENT_FRAGMENT_NODE {
		// Collect all children of the fragment first
		var children []Node
		for child := newChild.FirstChild(); child != nil; child = child.NextSibling() {
			children = append(children, child)
		}

		if len(children) == 0 {
			// Empty fragment, just remove the old child
			return n.RemoveChild(oldChild)
		}

		// Replace with first child, then insert remaining children after it
		firstChild := children[0]
		if df, ok := newChild.(*documentFragment); ok {
			df.removeChildInternal(firstChild)
		} else {
			newChild.RemoveChild(firstChild)
		}
		replaced, err := n.replaceChildInternal(firstChild, oldChild)
		if err != nil {
			return nil, err
		}

		// Insert remaining children after the first one
		refChild := firstChild.NextSibling()
		for _, child := range children[1:] {
			if df, ok := newChild.(*documentFragment); ok {
				df.removeChildInternal(child)
			} else {
				newChild.RemoveChild(child)
			}
			_, err := n.insertBeforeInternal(child, refChild)
			if err != nil {
				return nil, err
			}
		}

		return replaced, nil
	}

	return n.replaceChildInternal(newChild, oldChild)
}

// replaceChildInternal handles the actual replacement without DocumentFragment expansion
func (n *node) replaceChildInternal(newChild Node, oldChild Node) (Node, error) {
	// Remove from current parent if exists - done internally to avoid deadlock
	if newChild.ParentNode() != nil {
		oldParent := newChild.ParentNode()
		oc := getInternalNode(newChild)

		// Update sibling links
		if oc.previousSibling != nil {
			if prevNode := getInternalNode(oc.previousSibling); prevNode != nil {
				prevNode.nextSibling = oc.nextSibling
			}
		}
		if oc.nextSibling != nil {
			if nextNode := getInternalNode(oc.nextSibling); nextNode != nil {
				nextNode.previousSibling = oc.previousSibling
			}
		}

		// Update parent's first/last child pointers
		if op := getInternalNode(oldParent); op != nil {
			if op.firstChild == newChild {
				op.firstChild = oc.nextSibling
			}
			if op.lastChild == newChild {
				op.lastChild = oc.previousSibling
			}
			// Update old parent's live NodeList if it exists
			if op.childNodes != nil && op.childNodes.update != nil {
				op.childNodes.update()
			}
		}

		// Clear the removed node's parent/sibling references
		oc.parentNode = nil
		oc.previousSibling = nil
		oc.nextSibling = nil
	}

	nc := getInternalNode(newChild)
	oc := getInternalNode(oldChild)

	nc.nextSibling = oc.nextSibling
	nc.previousSibling = oc.previousSibling
	nc.parentNode = Node(n)

	if oc.previousSibling != nil {
		if prevNode := getInternalNode(oc.previousSibling); prevNode != nil {
			prevNode.nextSibling = newChild
		}
	} else {
		n.firstChild = newChild
	}

	if oc.nextSibling != nil {
		if nextNode := getInternalNode(oc.nextSibling); nextNode != nil {
			nextNode.previousSibling = newChild
		}
	} else {
		n.lastChild = newChild
	}

	oc.parentNode = nil
	oc.nextSibling = nil
	oc.previousSibling = nil

	// Update live NodeList if it exists
	if n.childNodes != nil && n.childNodes.update != nil {
		n.childNodes.update()
	}
	if doc := n.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.notifyMutation()
		}
	}
	return oldChild, nil
}

func (n *node) RemoveChild(oldChild Node) (Node, error) {
	if oldChild.ParentNode() != Node(n) {
		return nil, NewDOMException("NotFoundError", "")
	}

	oc := getInternalNode(oldChild)

	if oc.previousSibling != nil {
		if prevNode := getInternalNode(oc.previousSibling); prevNode != nil {
			prevNode.nextSibling = oc.nextSibling
		}
	} else {
		n.firstChild = oc.nextSibling
	}

	if oc.nextSibling != nil {
		if nextNode := getInternalNode(oc.nextSibling); nextNode != nil {
			nextNode.previousSibling = oc.previousSibling
		}
	} else {
		n.lastChild = oc.previousSibling
	}

	oc.parentNode = nil
	oc.nextSibling = nil
	oc.previousSibling = nil

	// Update live NodeList if it exists
	if n.childNodes != nil && n.childNodes.update != nil {
		n.childNodes.update()
	}
	if doc := n.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.notifyMutation()
		}
	}
	return oldChild, nil
}

func (n *node) AppendChild(newChild Node) (Node, error) {
	return n.InsertBefore(newChild, nil)
}

func (n *node) HasChildNodes() bool {
	return n.firstChild != nil
}

func (n *node) CloneNode(deep bool) Node {
	clone := &node{
		nodeType:      n.nodeType,
		nodeName:      n.nodeName,
		nodeValue:     n.nodeValue,
		ownerDocument: n.ownerDocument,
		namespaceURI:  n.namespaceURI,
		prefix:        n.prefix,
		localName:     n.localName,
	}

	if n.attributes != nil {
		clone.attributes = NewNamedNodeMap()
		for _, key := range n.attributes.order {
			attr := n.attributes.items[key]
			clonedAttr := attr.CloneNode(true)
			clone.attributes.SetNamedItem(clonedAttr)
		}
	}

	if deep {
		for child := n.firstChild; child != nil; child = child.NextSibling() {
			clone.AppendChild(child.CloneNode(true))
		}
	}
	return clone
}

func (n *node) Normalize() {
	// Collect children and build normalized list
	children := []Node{}
	var mergedText strings.Builder
	var hasTextNodes bool

	child := n.firstChild
	for child != nil {
		switch child.NodeType() {
		case TEXT_NODE:
			// Merge all text nodes, including empty ones
			data := string(child.NodeValue())
			mergedText.WriteString(data)
			hasTextNodes = true
			// Don't add text nodes to children - we're merging them
		default:
			// Non-text node: flush any pending merged text first
			if hasTextNodes {
				// Create text node with merged content (even if empty)
				mergedContent := mergedText.String()
				textNode := n.ownerDocument.CreateTextNode(DOMString(mergedContent))
				children = append(children, textNode)
				mergedText.Reset()
				hasTextNodes = false
			}
			// Recursively normalize element children
			if child.HasChildNodes() {
				child.Normalize()
			}
			children = append(children, child)
		}
		child = child.NextSibling()
	}

	// Flush any remaining text at the end
	if hasTextNodes {
		mergedContent := mergedText.String()
		textNode := n.ownerDocument.CreateTextNode(DOMString(mergedContent))
		children = append(children, textNode)
	}

	// Clear old child relationships first
	originalFirst := n.firstChild
	child = originalFirst
	for child != nil {
		next := child.NextSibling()
		if c := getInternalNode(child); c != nil {
			c.parentNode = nil
			c.previousSibling = nil
			c.nextSibling = nil
		}
		child = next
	}

	// Clear and rebuild children
	n.firstChild = nil
	n.lastChild = nil

	// Re-append normalized children
	for _, child := range children {
		n.AppendChild(child)
	}

	// Update live NodeList if it exists
	if n.childNodes != nil && n.childNodes.update != nil {
		n.childNodes.update()
	}
}

func (n *node) IsSupported(feature DOMString, version DOMString) bool {
	if versions, ok := supportedFeatures[feature]; ok {
		if version == "" {
			return true
		}
		for _, v := range versions {
			if v == version {
				return true
			}
		}
	}
	return false
}

func (n *node) NamespaceURI() DOMString {
	return n.namespaceURI
}

func (n *node) Prefix() DOMString {
	return n.prefix
}

func (n *node) SetPrefix(prefix DOMString) error {
	if prefix != "" && !IsValidName(prefix) {
		return NewDOMException("InvalidCharacterError", "Invalid character in prefix")
	}

	if n.namespaceURI == "" {
		return NewDOMException("NamespaceError", "Cannot set prefix for a node with no namespace URI")
	}

	if prefix == "xml" && n.namespaceURI != "http://www.w.org/XML/1998/namespace" {
		return NewDOMException("NamespaceError", "Invalid namespace URI for 'xml' prefix")
	}

	if prefix == "xmlns" && n.namespaceURI != "http://www.w3.org/2000/xmlns/" {
		return NewDOMException("NamespaceError", "Invalid namespace URI for 'xmlns' prefix")
	}

	if n.nodeType == ATTRIBUTE_NODE && n.NodeName() == "xmlns" {
		return NewDOMException("NamespaceError", "Cannot set a prefix on the 'xmlns' attribute")
	}

	// Read-only check
	// This is a simplification. A more accurate check would involve checking the read-only status of the node.
	switch n.nodeType {
	case ELEMENT_NODE, ATTRIBUTE_NODE:
		// All good
	default:
		return NewDOMException("NoModificationAllowedError", "Cannot set prefix on this node type")
	}

	n.prefix = prefix
	if n.localName != "" {
		if prefix != "" {
			n.nodeName = prefix + ":" + n.localName
		} else {
			n.nodeName = n.localName
		}
	}

	return nil
}

func (n *node) LocalName() DOMString {
	return n.localName
}

func (n *node) HasAttributes() bool {
	return n.attributes != nil && n.attributes.Length() > 0
}

func (n *node) BaseURI() DOMString {
	// For Document nodes, the base URI is the document's address.
	if n.nodeType == DOCUMENT_NODE {
		// Assuming Document struct has a URL field or similar.
		// For now, return an empty string or a placeholder.
		// This will need to be properly implemented when Document URL is added.
		return ""
	}

	// For other nodes, it's inherited from the parent.
	if n.parentNode != nil {
		return n.parentNode.BaseURI()
	}

	// If no parent and not a Document, then base URI is not available.
	return "" // Or a representation of null, like an empty string for DOMString
}

func (n *node) IsConnected() bool {
	// A node is connected if it has an owner document and is part of that document's tree.
	// We can check this by traversing up the parent chain.
	// If we reach the Document node and it's the ownerDocument, then it's connected.
	current := Node(n)
	for current != nil {
		if current.NodeType() == DOCUMENT_NODE {
			return current == n.ownerDocument // Check if it's the actual owner document
		}
		if parent := current.ParentNode(); parent != nil {
			current = parent
		} else {
			break
		}
	}
	return false
}

func (n *node) CompareDocumentPosition(otherNode Node) uint16 {
	// 1. Same Node
	if isSameNode(Node(n), otherNode) {
		return 0
	}

	// Ensure both nodes are concrete *node types for internal comparison
	thisNode := getInternalNode(n)
	thatNode := getInternalNode(otherNode)

	if thisNode == nil || thatNode == nil {
		// One or both nodes are invalid, treat as disconnected
		return DOCUMENT_POSITION_DISCONNECTED
	}

	// 2. Disconnected (different documents)
	// If one node has an owner document and the other doesn't, or they have different owner documents
	if thisNode.ownerDocument != thatNode.ownerDocument {
		return DOCUMENT_POSITION_DISCONNECTED
	}

	// If both are not connected to a document, they are disconnected from each other
	if thisNode.ownerDocument == nil && thatNode.ownerDocument == nil {
		return DOCUMENT_POSITION_DISCONNECTED
	}

	// Check if one node contains the other
	if n.Contains(otherNode) {
		return DOCUMENT_POSITION_CONTAINS | DOCUMENT_POSITION_FOLLOWING
	}
	if otherNode.Contains(n) {
		return DOCUMENT_POSITION_CONTAINED_BY | DOCUMENT_POSITION_PRECEDING
	}

	// Find common ancestor and check contains/contained by
	// Build paths from each node to the root
	pathThis := []Node{}
	curr := Node(n)
	for curr != nil {
		pathThis = append(pathThis, curr)
		if parent := curr.ParentNode(); parent != nil {
			curr = parent
		} else {
			break
		}
	}

	pathThat := []Node{}
	curr = otherNode
	for curr != nil {
		pathThat = append(pathThat, curr)
		if parent := curr.ParentNode(); parent != nil {
			curr = parent
		} else {
			break
		}
	}

	// Reverse paths to go from root to node
	reversePathThis := make([]Node, len(pathThis))
	for i, node := range pathThis {
		reversePathThis[len(pathThis)-1-i] = node
	}

	reversePathThat := make([]Node, len(pathThat))
	for i, node := range pathThat {
		reversePathThat[len(pathThat)-1-i] = node
	}

	var commonAncestor Node
	commonAncestorIndexThis := -1
	commonAncestorIndexThat := -1

	// Find the deepest common ancestor
	for i := 0; i < len(reversePathThis) && i < len(reversePathThat); i++ {
		if reversePathThis[i] == reversePathThat[i] {
			commonAncestor = reversePathThis[i]
			commonAncestorIndexThis = i
			commonAncestorIndexThat = i
		} else {
			break
		}
	}

	// If no common ancestor (and both are connected to a document, which means they are siblings of a document fragment or similar)
	// This case should ideally be covered by the ownerDocument check, but as a fallback.
	if commonAncestor == nil {
		return DOCUMENT_POSITION_DISCONNECTED
	}

	// Determine preceding/following based on common ancestor's children
	// Get the direct children of the common ancestor that are ancestors of n and otherNode
	var ancestorOfThis, ancestorOfThat Node
	if commonAncestorIndexThis+1 < len(reversePathThis) {
		ancestorOfThis = reversePathThis[commonAncestorIndexThis+1]
	}
	if commonAncestorIndexThat+1 < len(reversePathThat) {
		ancestorOfThat = reversePathThat[commonAncestorIndexThat+1]
	}

	// Compare siblings under the common ancestor
	// Iterate through children of commonAncestor to find order
	for child := commonAncestor.FirstChild(); child != nil; child = child.NextSibling() {
		if isSameNode(child, ancestorOfThis) {
			// We found the ancestor of 'this' first, so 'this' comes before 'other'
			// Therefore, from 'this' perspective, 'other' is FOLLOWING
			return DOCUMENT_POSITION_FOLLOWING
		}
		if isSameNode(child, ancestorOfThat) {
			// We found the ancestor of 'other' first, so 'other' comes before 'this'
			// Therefore, from 'this' perspective, 'other' is PRECEDING
			return DOCUMENT_POSITION_PRECEDING
		}
	}

	// Fallback for cases not explicitly handled (should ideally not be reached in a well-formed DOM)
	return DOCUMENT_POSITION_DISCONNECTED | DOCUMENT_POSITION_IMPLEMENTATION_SPECIFIC
}

func (n *node) Contains(otherNode Node) bool {
	// A node contains itself per DOM specification
	if isSameNode(Node(n), otherNode) {
		return true
	}

	// Traverse up the parent chain of otherNode
	current := otherNode.ParentNode()
	for current != nil {
		if isSameNode(Node(n), current) {
			return true // n is an ancestor of otherNode
		}
		current = current.ParentNode()
	}
	return false // n is not an ancestor of otherNode
}

func (n *node) GetRootNode() Node {
	current := Node(n)
	for parent := current.ParentNode(); parent != nil; parent = current.ParentNode() {
		current = parent
	}
	return current
}

func (n *node) IsDefaultNamespace(namespaceURI DOMString) bool {
	// If the node is an Attr node, it does not have a default namespace.
	if n.NodeType() == ATTRIBUTE_NODE {
		return false
	}

	// If the node has a namespace URI and it matches the given namespaceURI,
	// and it has no prefix, then it's the default namespace.
	if n.NamespaceURI() == namespaceURI && n.Prefix() == "" {
		return true
	}

	// For other nodes, check the parent.
	if n.ParentNode() != nil {
		return n.ParentNode().IsDefaultNamespace(namespaceURI)
	}

	// If no parent and no matching namespace, it's not the default.
	return false
}

func (n *node) IsEqualNode(otherNode Node) bool {
	// If both are nil, they are equal. If one is nil and other is not, they are not equal.
	if n == nil && otherNode == nil {
		return true
	}
	if n == nil || otherNode == nil {
		return false
	}

	// 1. Node type must be the same
	if n.NodeType() != otherNode.NodeType() {
		return false
	}

	// 2. Node name, local name, namespace URI, prefix, and node value must be the same
	if n.NodeName() != otherNode.NodeName() ||
		n.LocalName() != otherNode.LocalName() ||
		n.NamespaceURI() != otherNode.NamespaceURI() ||
		n.Prefix() != otherNode.Prefix() ||
		n.NodeValue() != otherNode.NodeValue() {
		return false
	}

	// 3. Attributes must be the same (for Element and Attr nodes)
	if n.NodeType() == ELEMENT_NODE || n.NodeType() == ATTRIBUTE_NODE {
		if n.Attributes() == nil && otherNode.Attributes() != nil ||
			n.Attributes() != nil && otherNode.Attributes() == nil ||
			(n.Attributes() != nil && otherNode.Attributes() != nil && n.Attributes().Length() != otherNode.Attributes().Length()) {
			return false
		}

		if n.Attributes() != nil {
			for i := uint(0); i < n.Attributes().Length(); i++ {
				attr1 := n.Attributes().Item(i).(Attr)
				attr2 := otherNode.Attributes().GetNamedItemNS(attr1.NamespaceURI(), attr1.LocalName()).(Attr)
				if attr2 == nil || !attr1.IsEqualNode(attr2) {
					return false
				}
			}
		}
	}

	// 4. Children must be the same (recursively)
	if n.HasChildNodes() != otherNode.HasChildNodes() {
		return false
	}

	if n.HasChildNodes() {
		child1 := n.FirstChild()
		child2 := otherNode.FirstChild()

		for child1 != nil && child2 != nil {
			if !child1.IsEqualNode(child2) {
				return false
			}
			child1 = child1.NextSibling()
			child2 = child2.NextSibling()
		}

		// If one has more children than the other
		if child1 != nil || child2 != nil {
			return false
		}
	}

	return true
}

func (n *node) IsSameNode(otherNode Node) bool {
	return isSameNode(Node(n), otherNode)
}

func (n *node) LookupPrefix(namespaceURI DOMString) DOMString {
	if namespaceURI == "" {
		return ""
	}

	// If this node is an element, check its attributes for namespace declarations
	if n.NodeType() == ELEMENT_NODE && n.attributes != nil {
		for i := uint(0); i < n.attributes.Length(); i++ {
			attr := n.attributes.Item(i).(Attr)
			if attr.Prefix() == "xmlns" && attr.NodeValue() == namespaceURI {
				return attr.LocalName()
			} else if attr.NodeName() == "xmlns" && attr.NodeValue() == namespaceURI {
				// Default namespace declaration
				return "" // Default namespace has no prefix
			}
		}
	}

	// If this node has a prefix and its namespace URI matches, return its prefix
	if n.Prefix() != "" && n.NamespaceURI() == namespaceURI {
		return n.Prefix()
	}

	// Recursively check parent
	if n.ParentNode() != nil {
		return n.ParentNode().LookupPrefix(namespaceURI)
	}

	return "" // No prefix found
}

func (n *node) LookupNamespaceURI(prefix DOMString) DOMString {
	// If this node is an element, check its attributes for namespace declarations
	if n.NodeType() == ELEMENT_NODE && n.attributes != nil {
		for i := uint(0); i < n.attributes.Length(); i++ {
			attr := n.attributes.Item(i).(Attr)
			if attr.Prefix() == "xmlns" && attr.LocalName() == prefix {
				return attr.NodeValue()
			} else if prefix == "" && attr.NodeName() == "xmlns" {
				// Default namespace declaration
				return attr.NodeValue()
			}
		}
	}

	// If this node has the prefix and a namespace URI, return its namespace URI
	if n.Prefix() == prefix && n.NamespaceURI() != "" {
		return n.NamespaceURI()
	}

	// Recursively check parent
	if n.ParentNode() != nil {
		return n.ParentNode().LookupNamespaceURI(prefix)
	}

	return "" // No namespace URI found
}

func (n *node) TextContent() DOMString {
	var buf strings.Builder
	visited := make(map[Node]bool)
	var traverse func(node Node)
	traverse = func(node Node) {
		if node == nil {
			return
		}
		if visited[node] {
			return // Prevent infinite loops
		}
		visited[node] = true

		switch node.NodeType() {
		case TEXT_NODE, CDATA_SECTION_NODE, COMMENT_NODE, PROCESSING_INSTRUCTION_NODE:
			buf.WriteString(string(node.NodeValue()))
		case ELEMENT_NODE, DOCUMENT_FRAGMENT_NODE:
			child := node.FirstChild()
			for child != nil {
				traverse(child)
				child = child.NextSibling()
			}
		}
	}
	traverse(n)
	return DOMString(buf.String())
}

func (n *node) SetTextContent(value DOMString) {
	// Direct removal of children to avoid RemoveChild complexity
	n.firstChild = nil
	n.lastChild = nil

	// Update child nodes live list if it exists
	if n.childNodes != nil && n.childNodes.update != nil {
		n.childNodes.update()
	}

	// If value is not empty, create a new text node and append it
	if value != "" {
		textNode := n.OwnerDocument().CreateTextNode(value)
		n.AppendChild(textNode)
	}
}

// ===========================================================================
// Document Implementation
// ===========================================================================

// document represents a document node
type document struct {
	node
	doctype         DocumentType
	implementation  DOMImplementation
	documentElement Element
	idMap           map[DOMString]Element
	activeNodeLists []*nodeList
	mu              sync.RWMutex // Mutex for protecting concurrent access to the DOM

	// Document properties
	url          DOMString
	documentURI  DOMString
	characterSet DOMString
	contentType  DOMString
}

func (d *document) InsertBefore(newChild Node, refChild Node) (Node, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	result, err := d.node.InsertBefore(newChild, refChild)
	if err != nil {
		return nil, err
	}
	// Fix the parent reference to be the document interface
	if nc := getInternalNode(newChild); nc != nil {
		nc.parentNode = d
	}

	// Set as document element if this is the first element child and no document element is set
	if newChild.NodeType() == ELEMENT_NODE && d.documentElement == nil {
		if elem, ok := newChild.(Element); ok {
			d.documentElement = elem
		}
	}

	return result, nil
}

func (d *document) AppendChild(newChild Node) (Node, error) {
	return d.InsertBefore(newChild, nil)
}

func (d *document) RemoveChild(oldChild Node) (Node, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	parent := oldChild.ParentNode()
	if parent == nil || parent != Node(d) {
		return nil, NewDOMException("NotFoundError", "")
	}

	oc := getInternalNode(oldChild)

	if oc.previousSibling != nil {
		if prevNode := getInternalNode(oc.previousSibling); prevNode != nil {
			prevNode.nextSibling = oc.nextSibling
		}
	} else {
		d.firstChild = oc.nextSibling
	}

	if oc.nextSibling != nil {
		if nextNode := getInternalNode(oc.nextSibling); nextNode != nil {
			nextNode.previousSibling = oc.previousSibling
		}
	} else {
		d.lastChild = oc.previousSibling
	}

	oc.parentNode = nil
	oc.nextSibling = nil
	oc.previousSibling = nil

	// Update live NodeList if it exists
	if d.childNodes != nil && d.childNodes.update != nil {
		d.childNodes.update()
	}
	return oldChild, nil
}

func (d *document) Doctype() DocumentType {
	if d.doctype != nil {
		return d.doctype
	}
	return nil
}

func (d *document) Implementation() DOMImplementation {
	return d.implementation
}

func (d *document) DocumentElement() Element {
	if d.documentElement != nil {
		return d.documentElement
	}
	return nil
}

func (d *document) CreateElement(tagName DOMString) (Element, error) {
	if !IsValidName(tagName) {
		return nil, NewDOMException("InvalidCharacterError", "Invalid character in element name")
	}
	elem := &element{
		node: node{
			nodeType:      ELEMENT_NODE,
			nodeName:      tagName,
			ownerDocument: d,
			attributes:    NewNamedNodeMap(),
		},
	}
	return elem, nil
}

func (d *document) CreateDocumentFragment() DocumentFragment {
	return &documentFragment{
		node: node{
			nodeType:      DOCUMENT_FRAGMENT_NODE,
			nodeName:      "#document-fragment",
			ownerDocument: d,
		},
	}
}

func (d *document) CreateTextNode(data DOMString) Text {
	return &text{
		characterData: characterData{
			node: node{
				nodeType:      TEXT_NODE,
				nodeName:      "#text",
				nodeValue:     data,
				ownerDocument: d,
			},
		},
	}
}

func (d *document) CreateComment(data DOMString) Comment {
	return &comment{
		characterData: characterData{
			node: node{
				nodeType:      COMMENT_NODE,
				nodeName:      "#comment",
				nodeValue:     data,
				ownerDocument: d,
			},
		},
	}
}

func (d *document) CreateCDATASection(data DOMString) (CDATASection, error) {
	return &cdataSection{
		text: text{
			characterData: characterData{
				node: node{
					nodeType:      CDATA_SECTION_NODE,
					nodeName:      "#cdata-section",
					nodeValue:     data,
					ownerDocument: d,
				},
			},
		},
	}, nil
}

func (d *document) CreateProcessingInstruction(target, data DOMString) (ProcessingInstruction, error) {
	if !IsValidName(target) || strings.EqualFold(string(target), "xml") {
		return nil, NewDOMException("InvalidCharacterError", "Invalid processing instruction target")
	}
	return &processingInstruction{
		node: node{
			nodeType:      PROCESSING_INSTRUCTION_NODE,
			nodeName:      target,
			nodeValue:     data,
			ownerDocument: d,
		},
		target: target,
		data:   data,
	}, nil
}

func (d *document) CreateAttribute(name DOMString) (Attr, error) {
	if !IsValidName(name) {
		return nil, NewDOMException("InvalidCharacterError", "Invalid character in attribute name")
	}
	return &attr{
		node: node{
			nodeType:      ATTRIBUTE_NODE,
			nodeName:      name,
			ownerDocument: d,
		},
	}, nil
}

func (d *document) CreateEntityReference(name DOMString) (EntityReference, error) {
	return &entityReference{
		node: node{
			nodeType:      ENTITY_REFERENCE_NODE,
			nodeName:      name,
			ownerDocument: d,
		},
	}, nil
}

func (d *document) GetElementsByTagName(tagname DOMString) NodeList {
	d.mu.RLock()
	defer d.mu.RUnlock()
	nl := &nodeList{
		root: d,
		filter: func(n Node) bool {
			return n.NodeType() == ELEMENT_NODE && (tagname == "*" || n.NodeName() == tagname)
		},
		live: true,
		doc:  d,
	}
	nl.update = func() {
		nodes := []Node{}
		var helper func(Node)
		helper = func(n Node) {
			if n == nil {
				return
			}
			if nl.filter(n) {
				nodes = append(nodes, n)
			}
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				helper(child)
			}
		}
		helper(nl.root)
		nl.items = nodes
	}
	nl.update() // initial population
	if d.activeNodeLists == nil {
		d.activeNodeLists = []*nodeList{}
	}
	d.activeNodeLists = append(d.activeNodeLists, nl)
	return nl
}

func (d *document) getElementsByTagNameHelper(n Node, tagname DOMString, result *[]Node) {
	internal := getInternalNode(n)
	if internal == nil {
		return
	}
	for child := internal.firstChild; child != nil; child = child.NextSibling() {
		if child.NodeType() == ELEMENT_NODE {
			if tagname == "*" || child.NodeName() == tagname {
				*result = append(*result, child)
			}
			d.getElementsByTagNameHelper(child, tagname, result)
		}
	}
}

func (d *document) ImportNode(importedNode Node, deep bool) (Node, error) {
	if importedNode == nil {
		return nil, nil
	}

	// Clone the node
	newNode := importedNode.CloneNode(deep)

	// Set the owner document for the new node and its children
	var setOwner func(Node)
	setOwner = func(n Node) {
		if internalNode := getInternalNode(n); internalNode != nil {
			internalNode.ownerDocument = d
		}
		// Recursively set owner for children
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			setOwner(child)
		}
		// Recursively set owner for attributes
		if n.Attributes() != nil {
			for i := uint(0); i < n.Attributes().Length(); i++ {
				setOwner(n.Attributes().Item(i))
			}
		}
	}

	setOwner(newNode)

	return newNode, nil
}

func (d *document) CreateElementNS(namespaceURI, qualifiedName DOMString) (Element, error) {
	if !IsValidName(qualifiedName) {
		return nil, NewDOMException("InvalidCharacterError", "Invalid character in element qualified name")
	}

	// Reject reserved namespace URIs
	if namespaceURI == "http://www.w3.org/2000/xmlns/" || namespaceURI == "http://www.w3.org/XML/1998/namespace" {
		return nil, NewDOMException("NamespaceError", "Reserved namespace URI")
	}

	prefix, localName := parseQualifiedName(qualifiedName)
	return &element{
		node: node{
			nodeType:      ELEMENT_NODE,
			nodeName:      qualifiedName,
			ownerDocument: d,
			namespaceURI:  namespaceURI,
			prefix:        prefix,
			localName:     localName,
			attributes:    NewNamedNodeMap(),
		},
	}, nil
}

func (d *document) CreateAttributeNS(namespaceURI, qualifiedName DOMString) (Attr, error) {
	if !IsValidName(qualifiedName) {
		return nil, NewDOMException("InvalidCharacterError", "Invalid character in attribute qualified name")
	}
	prefix, localName := parseQualifiedName(qualifiedName)
	return &attr{
		node: node{
			nodeType:      ATTRIBUTE_NODE,
			nodeName:      qualifiedName,
			ownerDocument: d,
			namespaceURI:  namespaceURI,
			prefix:        prefix,
			localName:     localName,
		},
	}, nil
}

func (d *document) GetElementsByTagNameNS(namespaceURI, localName DOMString) NodeList {
	nl := &nodeList{
		root: d,
		filter: func(n Node) bool {
			return n.NodeType() == ELEMENT_NODE &&
				(namespaceURI == "*" || n.NamespaceURI() == namespaceURI) &&
				(localName == "*" || n.LocalName() == localName)
		},
		live: true,
		doc:  d,
	}
	nl.update = func() {
		nodes := []Node{}
		var helper func(Node)
		helper = func(n Node) {
			if n == nil {
				return
			}
			if nl.filter(n) {
				nodes = append(nodes, n)
			}
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				helper(child)
			}
		}
		helper(nl.root)
		nl.items = nodes
	}
	nl.update() // initial population
	if d.activeNodeLists == nil {
		d.activeNodeLists = []*nodeList{}
	}
	d.activeNodeLists = append(d.activeNodeLists, nl)
	return nl
}

func (d *document) getElementsByTagNameNSHelper(n Node, namespaceURI, localName DOMString, result *[]Node) {
	internal := getInternalNode(n)
	if internal == nil {
		return
	}
	for child := internal.firstChild; child != nil; child = child.NextSibling() {
		if child.NodeType() == ELEMENT_NODE {
			if (namespaceURI == "*" || child.NamespaceURI() == namespaceURI) &&
				(localName == "*" || child.LocalName() == localName) {
				*result = append(*result, child)
			}
			d.getElementsByTagNameNSHelper(child, namespaceURI, localName, result)
		}
	}
}

func (d *document) GetElementById(elementId DOMString) Element {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.idMap != nil {
		if elem := d.idMap[elementId]; elem != nil {
			return elem
		}
	}
	return nil
}

func (d *document) AdoptNode(source Node) (Node, error) {
	if source == nil {
		return nil, nil
	}

	// If the source node has a parent, remove it from its parent.
	if source.ParentNode() != nil {
		// If the source is the document element of its owner document, clear the documentElement field
		if sourceOwner := source.OwnerDocument(); sourceOwner != nil {
			if ownerDoc, ok := sourceOwner.(*document); ok {
				if ownerDoc.documentElement != nil && ownerDoc.documentElement == source {
					ownerDoc.documentElement = nil
				}
			}
		}
		source.ParentNode().RemoveChild(source)
	}

	// Set the owner document for the source node and its children (including attributes).
	var setOwner func(Node)
	setOwner = func(n Node) {
		if internalNode := getInternalNode(n); internalNode != nil {
			internalNode.ownerDocument = d
		}
		// Recursively set owner for children
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			setOwner(child)
		}
		// Recursively set owner for attributes
		if n.Attributes() != nil {
			for i := uint(0); i < n.Attributes().Length(); i++ {
				setOwner(n.Attributes().Item(i))
			}
		}
	}

	setOwner(source)

	return source, nil
}

// nodeIterator implements the NodeIterator interface.
type nodeIterator struct {
	root                       Node
	referenceNode              Node
	pointerBeforeReferenceNode bool
	whatToShow                 uint32
	filter                     NodeFilter
	active                     bool // Indicates if the iterator is still active
}

func (ni *nodeIterator) Root() Node {
	return ni.root
}

func (ni *nodeIterator) ReferenceNode() Node {
	return ni.referenceNode
}

func (ni *nodeIterator) PointerBeforeReferenceNode() bool {
	return ni.pointerBeforeReferenceNode
}

func (ni *nodeIterator) WhatToShow() uint32 {
	return ni.whatToShow
}

func (ni *nodeIterator) Filter() NodeFilter {
	return ni.filter
}

func (ni *nodeIterator) Detach() {
	ni.active = false
}

func (ni *nodeIterator) NextNode() (Node, error) {
	if !ni.active {
		return nil, NewDOMException("InvalidStateError", "NodeIterator is detached")
	}

	// Acquire read lock on the document
	if doc := ni.root.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.RLock()
			defer d.mu.RUnlock()
		}
	}

	var candidate Node
	if ni.pointerBeforeReferenceNode {
		candidate = ni.referenceNode
		ni.pointerBeforeReferenceNode = false
	} else {
		candidate = ni.getNextCandidate(ni.referenceNode)
	}

	for candidate != nil {
		// Check if candidate is within the root's subtree
		if !ni.isInSubtree(candidate) {
			candidate = ni.getNextCandidate(candidate)
			continue
		}

		// Apply whatToShow and filter
		if ni.acceptNode(candidate) {
			ni.referenceNode = candidate
			return candidate, nil
		}

		candidate = ni.getNextCandidate(candidate)
	}

	return nil, nil
}

func (ni *nodeIterator) PreviousNode() (Node, error) {
	if !ni.active {
		return nil, NewDOMException("InvalidStateError", "NodeIterator is detached")
	}

	// Acquire read lock on the document
	if doc := ni.root.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.RLock()
			defer d.mu.RUnlock()
		}
	}

	var candidate Node
	if !ni.pointerBeforeReferenceNode {
		// Return the current reference node first, then set pointer before it
		if ni.acceptNode(ni.referenceNode) {
			ni.pointerBeforeReferenceNode = true
			return ni.referenceNode, nil
		}
		candidate = ni.getPreviousCandidate(ni.referenceNode)
		ni.pointerBeforeReferenceNode = true
	} else {
		candidate = ni.getPreviousCandidate(ni.referenceNode)
	}

	for candidate != nil {
		// Check if candidate is within the root's subtree
		if !ni.isInSubtree(candidate) {
			candidate = ni.getPreviousCandidate(candidate)
			continue
		}

		// Apply whatToShow and filter
		if ni.acceptNode(candidate) {
			ni.referenceNode = candidate
			return candidate, nil
		}

		candidate = ni.getPreviousCandidate(candidate)
	}

	return nil, nil
}

func (ni *nodeIterator) getNextCandidate(current Node) Node {
	if current == nil {
		return nil
	}

	// Try first child
	if current.FirstChild() != nil {
		return current.FirstChild()
	}

	// Try next sibling
	if current.NextSibling() != nil {
		return current.NextSibling()
	}

	// Go up and try next sibling of ancestors
	parent := current.ParentNode()
	for parent != nil && parent != ni.root {
		if parent.NextSibling() != nil {
			return parent.NextSibling()
		}
		parent = parent.ParentNode()
	}

	return nil
}

func (ni *nodeIterator) getPreviousCandidate(current Node) Node {
	// Try previous sibling's last descendant
	if current.PreviousSibling() != nil {
		candidate := current.PreviousSibling()
		for candidate.LastChild() != nil {
			candidate = candidate.LastChild()
		}
		return candidate
	}

	// Try parent (including root)
	if current.ParentNode() != nil {
		return current.ParentNode()
	}

	return nil
}

func (ni *nodeIterator) isInSubtree(node Node) bool {
	if isSameNode(node, ni.root) {
		return true
	}
	return ni.root.Contains(node)
}

func (ni *nodeIterator) acceptNode(node Node) bool {
	// Check whatToShow
	if ni.whatToShow != 0xFFFFFFFF {
		// NodeType values are 1-indexed, so we need to shift by (NodeType - 1)
		mask := uint32(1 << (node.NodeType() - 1))
		if (ni.whatToShow & mask) == 0 {
			return false
		}
	}

	// Check filter
	if ni.filter != nil {
		return ni.filter.AcceptNode(node) == FILTER_ACCEPT
	}

	return true
}

func (d *document) CreateNodeIterator(root Node, whatToShow uint32, filter NodeFilter) (NodeIterator, error) {
	// Acquire read lock on the document
	d.mu.RLock()
	defer d.mu.RUnlock()

	// The root must be a Node within the document.
	if root == nil || (root.OwnerDocument() != d && root.NodeType() != DOCUMENT_NODE) {
		return nil, NewDOMException("NotFoundError", "The root is not a node in this document.")
	}

	ni := &nodeIterator{
		root:                       root,
		referenceNode:              root,
		pointerBeforeReferenceNode: true,
		whatToShow:                 whatToShow,
		filter:                     filter,
		active:                     true,
	}
	return ni, nil
}

// treeWalker implements the TreeWalker interface.
type treeWalker struct {
	root        Node
	whatToShow  uint32
	filter      NodeFilter
	currentNode Node
}

func (tw *treeWalker) Root() Node {
	return tw.root
}

func (tw *treeWalker) WhatToShow() uint32 {
	return tw.whatToShow
}

func (tw *treeWalker) Filter() NodeFilter {
	return tw.filter
}

func (tw *treeWalker) CurrentNode() Node {
	return tw.currentNode
}

func (tw *treeWalker) SetCurrentNode(node Node) error {
	if node == nil {
		return NewDOMException("NotSupportedError", "CurrentNode cannot be null")
	}
	tw.currentNode = node
	return nil
}

func (tw *treeWalker) acceptNode(node Node) uint16 {
	if node == nil {
		return FILTER_REJECT
	}

	// Check whatToShow
	if tw.whatToShow != 0xFFFFFFFF {
		// NodeType values are 1-indexed, so we need to shift by (NodeType - 1)
		mask := uint32(1 << (node.NodeType() - 1))
		if (tw.whatToShow & mask) == 0 {
			return FILTER_SKIP
		}
	}

	// Check filter
	if tw.filter != nil {
		return tw.filter.AcceptNode(node)
	}

	return FILTER_ACCEPT
}

func (tw *treeWalker) ParentNode() Node {
	node := tw.currentNode
	for node != nil && node != tw.root {
		node = node.ParentNode()
		if node != nil {
			result := tw.acceptNode(node)
			if result == FILTER_ACCEPT {
				tw.currentNode = node
				return node
			}
		}
	}
	return nil
}

func (tw *treeWalker) FirstChild() Node {
	return tw.traverseChildren(true)
}

func (tw *treeWalker) LastChild() Node {
	return tw.traverseChildren(false)
}

func (tw *treeWalker) traverseChildren(first bool) Node {
	node := tw.currentNode
	var child Node
	if first {
		child = node.FirstChild()
	} else {
		child = node.LastChild()
	}

	for child != nil {
		result := tw.acceptNode(child)
		if result == FILTER_ACCEPT {
			tw.currentNode = child
			return child
		}
		if result == FILTER_SKIP {
			var grandchild Node
			if first {
				grandchild = tw.traverseChildrenHelper(child, true)
			} else {
				grandchild = tw.traverseChildrenHelper(child, false)
			}
			if grandchild != nil {
				return grandchild
			}
		}
		if first {
			child = child.NextSibling()
		} else {
			child = child.PreviousSibling()
		}
	}
	return nil
}

func (tw *treeWalker) traverseChildrenHelper(node Node, first bool) Node {
	var child Node
	if first {
		child = node.FirstChild()
	} else {
		child = node.LastChild()
	}

	for child != nil {
		result := tw.acceptNode(child)
		if result == FILTER_ACCEPT {
			tw.currentNode = child
			return child
		}
		if result == FILTER_SKIP {
			var grandchild Node
			if first {
				grandchild = tw.traverseChildrenHelper(child, true)
			} else {
				grandchild = tw.traverseChildrenHelper(child, false)
			}
			if grandchild != nil {
				return grandchild
			}
		}
		if first {
			child = child.NextSibling()
		} else {
			child = child.PreviousSibling()
		}
	}
	return nil
}

func (tw *treeWalker) NextSibling() Node {
	return tw.traverseSiblings(true)
}

func (tw *treeWalker) PreviousSibling() Node {
	return tw.traverseSiblings(false)
}

func (tw *treeWalker) traverseSiblings(next bool) Node {
	node := tw.currentNode
	if node == tw.root {
		return nil
	}

	for {
		var sibling Node
		if next {
			sibling = node.NextSibling()
		} else {
			sibling = node.PreviousSibling()
		}

		for sibling != nil {
			result := tw.acceptNode(sibling)
			if result == FILTER_ACCEPT {
				tw.currentNode = sibling
				return sibling
			}
			if result == FILTER_SKIP {
				child := tw.traverseSiblingsHelper(sibling, next)
				if child != nil {
					return child
				}
			}
			if next {
				sibling = sibling.NextSibling()
			} else {
				sibling = sibling.PreviousSibling()
			}
		}

		node = node.ParentNode()
		if node == nil || node == tw.root {
			return nil
		}

		if tw.acceptNode(node) == FILTER_ACCEPT {
			continue
		}
	}
}

func (tw *treeWalker) traverseSiblingsHelper(node Node, next bool) Node {
	var child Node
	if next {
		child = node.FirstChild()
	} else {
		child = node.LastChild()
	}

	for child != nil {
		result := tw.acceptNode(child)
		if result == FILTER_ACCEPT {
			tw.currentNode = child
			return child
		}
		if result == FILTER_SKIP {
			grandchild := tw.traverseSiblingsHelper(child, next)
			if grandchild != nil {
				return grandchild
			}
		}
		if next {
			child = child.NextSibling()
		} else {
			child = child.PreviousSibling()
		}
	}
	return nil
}

func (tw *treeWalker) PreviousNode() Node {
	node := tw.currentNode
	for node != tw.root {
		var sibling Node = node.PreviousSibling()
		for sibling != nil {
			result := tw.acceptNode(sibling)
			node = sibling
			for result != FILTER_REJECT && node.LastChild() != nil {
				node = node.LastChild()
				result = tw.acceptNode(node)
			}
			if result == FILTER_ACCEPT {
				tw.currentNode = node
				return node
			}
			sibling = sibling.PreviousSibling()
		}

		if node == tw.root || node.ParentNode() == nil {
			return nil
		}
		node = node.ParentNode()

		if tw.acceptNode(node) == FILTER_ACCEPT {
			tw.currentNode = node
			return node
		}
	}
	return nil
}

func (tw *treeWalker) NextNode() Node {
	node := tw.currentNode
	result := FILTER_ACCEPT

	for {
		for result != FILTER_REJECT && node.FirstChild() != nil {
			node = node.FirstChild()
			result = tw.acceptNode(node)
			if result == FILTER_ACCEPT {
				tw.currentNode = node
				return node
			}
		}

		var sibling Node
		temp := node
		for temp != nil && temp != tw.root {
			sibling = temp.NextSibling()
			if sibling != nil {
				node = sibling
				break
			}
			temp = temp.ParentNode()
		}

		if sibling == nil {
			return nil
		}

		result = tw.acceptNode(node)
		if result == FILTER_ACCEPT {
			tw.currentNode = node
			return node
		}
	}
}

func (d *document) CreateTreeWalker(root Node, whatToShow uint32, filter NodeFilter) (TreeWalker, error) {
	// Acquire read lock on the document
	d.mu.RLock()
	defer d.mu.RUnlock()

	// The root must be a Node within the document.
	if root == nil {
		return nil, NewDOMException("NotSupportedError", "Root cannot be null")
	}

	tw := &treeWalker{
		root:        root,
		whatToShow:  whatToShow,
		filter:      filter,
		currentNode: root,
	}
	return tw, nil
}

// domRange implements the Range interface
type domRange struct {
	startContainer Node
	startOffset    uint32
	endContainer   Node
	endOffset      uint32
	doc            Document
	detached       bool
}

func (r *domRange) StartContainer() Node {
	return r.startContainer
}

func (r *domRange) StartOffset() uint32 {
	return r.startOffset
}

func (r *domRange) EndContainer() Node {
	return r.endContainer
}

func (r *domRange) EndOffset() uint32 {
	return r.endOffset
}

func (r *domRange) Collapsed() bool {
	return r.startContainer == r.endContainer && r.startOffset == r.endOffset
}

func (r *domRange) CommonAncestorContainer() Node {
	// Find the common ancestor of start and end containers
	if r.startContainer == r.endContainer {
		return r.startContainer
	}

	// Build ancestor chain for start container
	startAncestors := []Node{}
	for n := r.startContainer; n != nil; n = n.ParentNode() {
		startAncestors = append(startAncestors, n)
	}

	// Find first common ancestor with end container
	for n := r.endContainer; n != nil; n = n.ParentNode() {
		for _, ancestor := range startAncestors {
			if n == ancestor {
				return n
			}
		}
	}

	return nil
}

func (r *domRange) SetStart(node Node, offset uint32) error {
	if node == nil {
		return NewDOMException("InvalidNodeTypeError", "Node cannot be null")
	}

	// Validate offset
	if err := r.validateOffset(node, offset); err != nil {
		return err
	}

	r.startContainer = node
	r.startOffset = offset

	// Ensure start is not after end
	if r.comparePositions(r.startContainer, r.startOffset, r.endContainer, r.endOffset) > 0 {
		r.endContainer = node
		r.endOffset = offset
	}

	return nil
}

func (r *domRange) SetEnd(node Node, offset uint32) error {
	if node == nil {
		return NewDOMException("InvalidNodeTypeError", "Node cannot be null")
	}

	// Validate offset
	if err := r.validateOffset(node, offset); err != nil {
		return err
	}

	r.endContainer = node
	r.endOffset = offset

	// Ensure end is not before start
	if r.comparePositions(r.startContainer, r.startOffset, r.endContainer, r.endOffset) > 0 {
		r.startContainer = node
		r.startOffset = offset
	}

	return nil
}

func (r *domRange) SetStartBefore(node Node) error {
	if node == nil || node.ParentNode() == nil {
		return NewDOMException("InvalidNodeTypeError", "Node must have a parent")
	}

	offset := r.getNodeIndex(node)
	return r.SetStart(node.ParentNode(), offset)
}

func (r *domRange) SetStartAfter(node Node) error {
	if node == nil || node.ParentNode() == nil {
		return NewDOMException("InvalidNodeTypeError", "Node must have a parent")
	}

	offset := r.getNodeIndex(node) + 1
	return r.SetStart(node.ParentNode(), offset)
}

func (r *domRange) SetEndBefore(node Node) error {
	if node == nil || node.ParentNode() == nil {
		return NewDOMException("InvalidNodeTypeError", "Node must have a parent")
	}

	offset := r.getNodeIndex(node)
	return r.SetEnd(node.ParentNode(), offset)
}

func (r *domRange) SetEndAfter(node Node) error {
	if node == nil || node.ParentNode() == nil {
		return NewDOMException("InvalidNodeTypeError", "Node must have a parent")
	}

	offset := r.getNodeIndex(node) + 1
	return r.SetEnd(node.ParentNode(), offset)
}

func (r *domRange) Collapse(toStart bool) {
	if toStart {
		r.endContainer = r.startContainer
		r.endOffset = r.startOffset
	} else {
		r.startContainer = r.endContainer
		r.startOffset = r.endOffset
	}
}

func (r *domRange) SelectNode(node Node) error {
	if node == nil || node.ParentNode() == nil {
		return NewDOMException("InvalidNodeTypeError", "Node must have a parent")
	}

	parent := node.ParentNode()
	offset := r.getNodeIndex(node)

	r.startContainer = parent
	r.startOffset = offset
	r.endContainer = parent
	r.endOffset = offset + 1

	return nil
}

func (r *domRange) SelectNodeContents(node Node) error {
	if node == nil {
		return NewDOMException("InvalidNodeTypeError", "Node cannot be null")
	}

	r.startContainer = node
	r.startOffset = 0
	r.endContainer = node
	r.endOffset = r.getNodeLength(node)

	return nil
}

func (r *domRange) CompareBoundaryPoints(how uint16, sourceRange Range) (int16, error) {
	if sourceRange == nil {
		return 0, NewDOMException("InvalidStateError", "sourceRange cannot be null")
	}

	var thisNode, otherNode Node
	var thisOffset, otherOffset uint32

	switch how {
	case START_TO_START:
		thisNode = r.startContainer
		thisOffset = r.startOffset
		otherNode = sourceRange.StartContainer()
		otherOffset = sourceRange.StartOffset()
	case START_TO_END:
		thisNode = r.endContainer
		thisOffset = r.endOffset
		otherNode = sourceRange.StartContainer()
		otherOffset = sourceRange.StartOffset()
	case END_TO_END:
		thisNode = r.endContainer
		thisOffset = r.endOffset
		otherNode = sourceRange.EndContainer()
		otherOffset = sourceRange.EndOffset()
	case END_TO_START:
		thisNode = r.startContainer
		thisOffset = r.startOffset
		otherNode = sourceRange.EndContainer()
		otherOffset = sourceRange.EndOffset()
	default:
		return 0, NewDOMException("NotSupportedError", "Invalid comparison type")
	}

	return int16(r.comparePositions(thisNode, thisOffset, otherNode, otherOffset)), nil
}

func (r *domRange) DeleteContents() error {
	if r.Collapsed() {
		return nil
	}

	// This is a simplified implementation
	// A full implementation would handle partial text node deletion
	// and maintain proper DOM structure

	return nil
}

func (r *domRange) ExtractContents() (DocumentFragment, error) {
	// This is a simplified implementation
	// A full implementation would extract and return the contents
	frag := r.doc.CreateDocumentFragment()
	return frag, nil
}

func (r *domRange) CloneContents() (DocumentFragment, error) {
	// This is a simplified implementation
	// A full implementation would clone and return the contents
	frag := r.doc.CreateDocumentFragment()
	return frag, nil
}

func (r *domRange) InsertNode(node Node) error {
	if node == nil {
		return NewDOMException("InvalidNodeTypeError", "Node cannot be null")
	}

	// This is a simplified implementation
	// A full implementation would properly insert the node at the start position

	return nil
}

func (r *domRange) SurroundContents(newParent Node) error {
	if newParent == nil {
		return NewDOMException("InvalidNodeTypeError", "Node cannot be null")
	}

	// This is a simplified implementation
	// A full implementation would extract contents and insert them into newParent

	return nil
}

func (r *domRange) CloneRange() Range {
	return &domRange{
		startContainer: r.startContainer,
		startOffset:    r.startOffset,
		endContainer:   r.endContainer,
		endOffset:      r.endOffset,
		doc:            r.doc,
	}
}

func (r *domRange) Detach() {
	// Mark the range as detached
	r.detached = true
}

func (r *domRange) IsPointInRange(node Node, offset uint32) (bool, error) {
	if r.detached {
		return false, NewDOMException("InvalidStateError", "Range is detached")
	}

	if node == nil {
		return false, NewDOMException("InvalidNodeTypeError", "Node cannot be null")
	}

	if err := r.validateOffset(node, offset); err != nil {
		return false, err
	}

	cmp := r.comparePositions(node, offset, r.startContainer, r.startOffset)
	if cmp < 0 {
		return false, nil
	}

	cmp = r.comparePositions(node, offset, r.endContainer, r.endOffset)
	if cmp > 0 {
		return false, nil
	}

	return true, nil
}

func (r *domRange) ComparePoint(node Node, offset uint32) (int16, error) {
	if node == nil {
		return 0, NewDOMException("InvalidNodeTypeError", "Node cannot be null")
	}

	if err := r.validateOffset(node, offset); err != nil {
		return 0, err
	}

	if r.comparePositions(node, offset, r.startContainer, r.startOffset) < 0 {
		return -1, nil
	}

	if r.comparePositions(node, offset, r.endContainer, r.endOffset) > 0 {
		return 1, nil
	}

	return 0, nil
}

func (r *domRange) IntersectsNode(node Node) bool {
	if node == nil || node.ParentNode() == nil {
		return false
	}

	parent := node.ParentNode()
	offset := r.getNodeIndex(node)

	cmp := r.comparePositions(parent, offset, r.endContainer, r.endOffset)
	if cmp > 0 {
		return false
	}

	cmp = r.comparePositions(parent, offset+1, r.startContainer, r.startOffset)
	return cmp >= 0
}

func (r *domRange) ToString() string {
	// This is a simplified implementation
	// A full implementation would extract text content from the range
	return ""
}

// Helper methods
func (r *domRange) validateOffset(node Node, offset uint32) error {
	length := r.getNodeLength(node)
	if offset > length {
		return NewDOMException("IndexSizeError", "Offset out of bounds")
	}
	return nil
}

func (r *domRange) getNodeLength(node Node) uint32 {
	switch node.NodeType() {
	case TEXT_NODE, COMMENT_NODE, PROCESSING_INSTRUCTION_NODE:
		if cd, ok := node.(CharacterData); ok {
			return uint32(cd.Length())
		}
	default:
		count := uint32(0)
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			count++
		}
		return count
	}
	return 0
}

func (r *domRange) getNodeIndex(node Node) uint32 {
	index := uint32(0)
	for sibling := node.ParentNode().FirstChild(); sibling != nil && sibling != node; sibling = sibling.NextSibling() {
		index++
	}
	return index
}

func (r *domRange) comparePositions(node1 Node, offset1 uint32, node2 Node, offset2 uint32) int {
	if node1 == node2 {
		if offset1 < offset2 {
			return -1
		} else if offset1 > offset2 {
			return 1
		}
		return 0
	}

	// Handle container relationships according to DOM Range spec
	position := node1.CompareDocumentPosition(node2)

	// If node1 contains node2
	if position&DOCUMENT_POSITION_CONTAINS != 0 {
		// node1 contains node2, so we need to find which child of node1 contains node2
		// and compare offset1 with the index of that child
		childIndex := uint32(0)
		for child := node1.FirstChild(); child != nil; child = child.NextSibling() {
			if child == node2 || child.Contains(node2) {
				// Found the child that contains node2
				// Special case: if offset1 equals childIndex and child == node2 and offset2 == 0,
				// then these positions are equivalent (before child vs start of child)
				if offset1 == childIndex && child == node2 && offset2 == 0 {
					return 0 // equivalent positions
				}
				if offset1 <= childIndex {
					return -1 // position is before this child
				} else {
					return 1 // position is after this child
				}
			}
			childIndex++
		}
		return 1 // shouldn't reach here in a well-formed DOM
	}

	// If node2 contains node1
	if position&DOCUMENT_POSITION_CONTAINED_BY != 0 {
		// node2 contains node1, so we need to find which child of node2 contains node1
		// and compare offset2 with the index of that child
		childIndex := uint32(0)
		for child := node2.FirstChild(); child != nil; child = child.NextSibling() {
			if child == node1 || child.Contains(node1) {
				// Found the child that contains node1
				// Special case: if offset2 equals childIndex and child == node1 and offset1 == 0,
				// then these positions are equivalent (before child vs start of child)
				if offset2 == childIndex && child == node1 && offset1 == 0 {
					return 0 // equivalent positions
				}
				if offset2 <= childIndex {
					return 1 // position is after this child
				} else {
					return -1 // position is before this child
				}
			}
			childIndex++
		}
		return -1 // shouldn't reach here in a well-formed DOM
	}

	// Neither contains the other - use document order
	if position&DOCUMENT_POSITION_FOLLOWING != 0 {
		return -1 // node1 is before node2
	} else if position&DOCUMENT_POSITION_PRECEDING != 0 {
		return 1 // node1 is after node2
	}

	return 0
}

func (d *document) CreateRange() Range {
	d.mu.RLock()
	defer d.mu.RUnlock()

	r := &domRange{
		startContainer: d,
		startOffset:    0,
		endContainer:   d,
		endOffset:      0,
		doc:            d,
	}
	return r
}

func (d *document) NormalizeDocument() {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Normalize the document element if it exists
	if d.documentElement != nil {
		d.documentElement.Normalize()
	} else {
		// If no document element, normalize the document itself
		d.Normalize()
	}
}

func normalizeNode(node Node) {
	// First recursively normalize all element children
	child := node.FirstChild()
	for child != nil {
		next := child.NextSibling()
		if child.NodeType() == ELEMENT_NODE {
			normalizeNode(child)
		}
		child = next
	}

	// Now normalize this node's direct children - collect and merge text nodes
	children := []Node{}
	var mergedText strings.Builder
	var hasPendingText bool

	child = node.FirstChild()
	for child != nil {
		switch child.NodeType() {
		case TEXT_NODE:
			if data := string(child.NodeValue()); data != "" {
				mergedText.WriteString(data)
				hasPendingText = true
			}
			// Don't add the original text node to children - we're merging
		default:
			// Non-text node: flush any pending merged text
			if hasPendingText {
				textNode := node.OwnerDocument().CreateTextNode(DOMString(mergedText.String()))
				children = append(children, textNode)
				mergedText.Reset()
				hasPendingText = false
			}
			children = append(children, child)
		}
		child = child.NextSibling()
	}

	// Flush any remaining text at the end
	if hasPendingText {
		textNode := node.OwnerDocument().CreateTextNode(DOMString(mergedText.String()))
		children = append(children, textNode)
	}

	// Clear and rebuild children if we have changes to make
	if n := getInternalNode(node); n != nil {
		// Clear all child relationships
		oldChild := n.firstChild
		for oldChild != nil {
			next := oldChild.NextSibling()
			if oc := getInternalNode(oldChild); oc != nil {
				oc.parentNode = nil
				oc.previousSibling = nil
				oc.nextSibling = nil
			}
			oldChild = next
		}

		n.firstChild = nil
		n.lastChild = nil

		// Re-append normalized children
		for _, child := range children {
			node.AppendChild(child)
		}

		// Update live NodeList if it exists
		if n.childNodes != nil && n.childNodes.update != nil {
			n.childNodes.update()
		}
	}
}

func (d *document) RenameNode(node Node, namespaceURI, qualifiedName DOMString) (Node, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if node == nil {
		return nil, NewDOMException("InvalidNodeTypeError", "Node cannot be null")
	}

	// Only element and attribute nodes can be renamed
	switch node.NodeType() {
	case ELEMENT_NODE:
		if elem, ok := node.(*element); ok {
			// Validate the qualified name
			if qualifiedName == "" {
				return nil, NewDOMException("InvalidCharacterError", "Qualified name cannot be empty")
			}

			// Update the element's name and namespace directly
			elem.nodeName = qualifiedName
			elem.namespaceURI = namespaceURI
			elem.localName = qualifiedName // Should be parsed from qualifiedName

			return elem, nil
		}

	case ATTRIBUTE_NODE:
		if attr, ok := node.(*attr); ok {
			// Validate the qualified name
			if qualifiedName == "" {
				return nil, NewDOMException("InvalidCharacterError", "Qualified name cannot be empty")
			}

			// Update the attribute's name and namespace
			attr.nodeName = qualifiedName
			attr.namespaceURI = namespaceURI
			attr.localName = qualifiedName // Should be parsed from qualifiedName

			return attr, nil
		}

	default:
		return nil, NewDOMException("NotSupportedError", "Only element and attribute nodes can be renamed")
	}

	return nil, NewDOMException("InvalidNodeTypeError", "Invalid node type")
}

// Document property methods
func (d *document) URL() DOMString {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.url
}

func (d *document) DocumentURI() DOMString {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.documentURI
}

func (d *document) CharacterSet() DOMString {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.characterSet == "" {
		return "UTF-8" // Default to UTF-8
	}
	return d.characterSet
}

func (d *document) Charset() DOMString {
	// Alias for CharacterSet
	return d.CharacterSet()
}

func (d *document) InputEncoding() DOMString {
	// Alias for CharacterSet
	return d.CharacterSet()
}

func (d *document) ContentType() DOMString {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.contentType == "" {
		return "application/xml" // Default for DOM documents
	}
	return d.contentType
}

// XPath methods implementation

// CreateExpression compiles an XPath expression for reuse
func (d *document) CreateExpression(expression string, resolver XPathNSResolver) (XPathExpression, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if expression == "" {
		return nil, NewXPathException("INVALID_EXPRESSION_ERR", "Expression cannot be empty")
	}

	// Check cache first
	var ast XPathNode
	var err error

	if cachedAst, found := getCachedExpression(expression); found {
		ast = cachedAst
	} else {
		// Parse the XPath expression into AST
		parser := NewXPathParser()
		ast, err = parser.Parse(expression)
		if err != nil {
			return nil, err
		}

		// Store in cache for future use
		setCachedExpression(expression, ast)
	}

	// Create compiled expression
	return &xpathExpression{
		expression: expression,
		resolver:   resolver,
		ast:        ast,
		document:   d,
	}, nil
}

// CreateNSResolver creates a namespace resolver from a node
func (d *document) CreateNSResolver(nodeResolver Node) Node {
	// Return the node as-is - it already implements LookupNamespaceURI
	// which is what we need for XPath namespace resolution
	return nodeResolver
}

// Evaluate evaluates an XPath expression on a context node
func (d *document) Evaluate(expression string, contextNode Node, resolver XPathNSResolver,
	resultType uint16, result XPathResult) (XPathResult, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if expression == "" {
		return nil, NewXPathException("INVALID_EXPRESSION_ERR", "Expression cannot be empty")
	}

	if contextNode == nil {
		// Use document as default context node
		contextNode = d
	}

	// Validate that context node belongs to this document
	if contextNode.OwnerDocument() != d && contextNode.NodeType() != DOCUMENT_NODE {
		return nil, NewXPathException("WRONG_DOCUMENT_ERR", "Context node must belong to this document")
	}

	// Create and compile expression
	expr, err := d.CreateExpression(expression, resolver)
	if err != nil {
		return nil, err
	}

	// Evaluate expression
	result, err = expr.Evaluate(contextNode, resultType, result)
	return result, err
}

func (d *document) removeIdMapping(idValue DOMString) {
	if d.idMap != nil && idValue != "" {
		delete(d.idMap, idValue)
	}
}

func (d *document) updateIdMappingForElement(element Element, attributeName DOMString, oldValue, newValue DOMString) {
	if attributeName != "id" {
		return
	}

	// Remove old mapping if it exists
	if oldValue != "" {
		d.removeIdMapping(oldValue)
	}

	// Add new mapping if new value is not empty
	if newValue != "" {
		if d.idMap == nil {
			d.idMap = make(map[DOMString]Element)
		}
		d.idMap[newValue] = element
	}
}

// notifyMutation is called whenever the DOM tree is mutated. It iterates
// over all active live NodeLists and calls their update function to keep
// them in sync with the DOM.
func (d *document) notifyMutation() {
	for _, nl := range d.activeNodeLists {
		if nl.update != nil {
			nl.update()
		}
	}
}

// ===========================================================================
// Element Implementation
// ===========================================================================

// element represents an element node
type element struct {
	node
}

func (e *element) InsertBefore(newChild Node, refChild Node) (Node, error) {
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}

	if newChild == nil {
		return nil, NewDOMException("HierarchyRequestError", "Invalid node")
	}

	// Handle self-insertion: inserting node before itself should be a no-op
	if newChild == refChild {
		return newChild, nil
	}

	// Check document ownership
	if newChild.OwnerDocument() != e.ownerDocument && e.ownerDocument != nil {
		return nil, NewDOMException("WrongDocumentError", "")
	}

	// HIERARCHY_REQUEST_ERR check - prevent cycles
	for ancestor := Node(e); ancestor != nil; ancestor = ancestor.ParentNode() {
		if ancestor == newChild {
			return nil, NewDOMException("HierarchyRequestError", "Cannot insert a node as a descendant of itself")
		}
	}

	// Handle DocumentFragment - insert its children instead of the fragment itself
	if newChild.NodeType() == DOCUMENT_FRAGMENT_NODE {
		// Collect all children of the fragment first
		var children []Node
		for child := newChild.FirstChild(); child != nil; child = child.NextSibling() {
			children = append(children, child)
		}

		// Insert each child using internal method to avoid infinite recursion
		for _, child := range children {
			// Remove from fragment first
			if df, ok := newChild.(*documentFragment); ok {
				df.removeChildInternal(child)
			} else {
				newChild.RemoveChild(child)
			}
			// Insert into target parent using internal method
			_, err := e.insertBeforeInternal(child, refChild)
			if err != nil {
				return nil, err
			}
		}

		// Return the fragment itself (which is now empty)
		return newChild, nil
	}

	return e.insertBeforeInternal(newChild, refChild)
}

// insertBeforeInternal handles the actual insertion without DocumentFragment expansion for elements
func (e *element) insertBeforeInternal(newChild Node, refChild Node) (Node, error) {

	// Remove from current parent if exists - done internally to avoid deadlock
	if newChild.ParentNode() != nil {
		oldParent := newChild.ParentNode()
		oc := getInternalNode(newChild)

		// Update sibling links
		if oc.previousSibling != nil {
			if prevNode := getInternalNode(oc.previousSibling); prevNode != nil {
				prevNode.nextSibling = oc.nextSibling
			}
		}
		if oc.nextSibling != nil {
			if nextNode := getInternalNode(oc.nextSibling); nextNode != nil {
				nextNode.previousSibling = oc.previousSibling
			}
		}

		// Update parent's first/last child pointers
		if op := getInternalNode(oldParent); op != nil {
			if op.firstChild == newChild {
				op.firstChild = oc.nextSibling
			}
			if op.lastChild == newChild {
				op.lastChild = oc.previousSibling
			}
			// Update old parent's live NodeList if it exists
			if op.childNodes != nil && op.childNodes.update != nil {
				op.childNodes.update()
			}
		}

		// Clear the removed node's parent/sibling references
		oc.parentNode = nil
		oc.previousSibling = nil
		oc.nextSibling = nil
	}

	// Validate refChild
	if refChild != nil && refChild.ParentNode() != Node(e) {
		return nil, NewDOMException("NotFoundError", "refChild not found")
	}

	// Get internal nodes for manipulation
	nc := getInternalNode(newChild)
	rc := getInternalNode(refChild)

	// Append or insert operation
	if refChild == nil {
		nc.parentNode = Node(e) // Store as element interface
		nc.previousSibling = e.lastChild
		nc.nextSibling = nil
		if e.lastChild != nil {
			if lastNode := getInternalNode(e.lastChild); lastNode != nil {
				lastNode.nextSibling = newChild
			}
		}
		e.lastChild = newChild
		if e.firstChild == nil {
			e.firstChild = newChild
		}
	} else { // Insert operation
		nc.parentNode = Node(e) // Store as element interface
		nc.nextSibling = refChild
		nc.previousSibling = rc.previousSibling
		if rc.previousSibling != nil {
			if prevNode := getInternalNode(rc.previousSibling); prevNode != nil {
				prevNode.nextSibling = newChild
			}
		} else {
			e.firstChild = newChild
		}
		rc.previousSibling = newChild
	}

	// Update live NodeList if it exists
	if e.childNodes != nil && e.childNodes.update != nil {
		e.childNodes.update()
	}
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.notifyMutation()
		}
	}
	return newChild, nil
}

func (e *element) AppendChild(newChild Node) (Node, error) {
	return e.InsertBefore(newChild, nil)
}

func (e *element) ReplaceChild(newChild Node, oldChild Node) (Node, error) {
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}

	if newChild == nil {
		return nil, NewDOMException("HierarchyRequestError", "Invalid node")
	}

	if oldChild.ParentNode() != Node(e) {
		return nil, NewDOMException("NotFoundError", "")
	}

	// Handle self-replacement: replacing node with itself should be a no-op
	if newChild == oldChild {
		return oldChild, nil
	}

	// Check document ownership
	if newChild.OwnerDocument() != e.ownerDocument && e.ownerDocument != nil {
		return nil, NewDOMException("WrongDocumentError", "")
	}

	// HIERARCHY_REQUEST_ERR check - prevent cycles
	for ancestor := Node(e); ancestor != nil; ancestor = ancestor.ParentNode() {
		if ancestor == newChild {
			return nil, NewDOMException("HierarchyRequestError", "Cannot insert a node as a descendant of itself")
		}
	}

	// Handle DocumentFragment - replace with its children
	if newChild.NodeType() == DOCUMENT_FRAGMENT_NODE {
		// Collect all children of the fragment first
		var children []Node
		for child := newChild.FirstChild(); child != nil; child = child.NextSibling() {
			children = append(children, child)
		}

		if len(children) == 0 {
			// Empty fragment, just remove the old child
			return e.RemoveChild(oldChild)
		}

		// Replace with first child, then insert remaining children after it
		firstChild := children[0]
		if df, ok := newChild.(*documentFragment); ok {
			df.removeChildInternal(firstChild)
		} else {
			newChild.RemoveChild(firstChild)
		}
		replaced, err := e.replaceChildInternal(firstChild, oldChild)
		if err != nil {
			return nil, err
		}

		// Insert remaining children after the first one
		refChild := firstChild.NextSibling()
		for _, child := range children[1:] {
			if df, ok := newChild.(*documentFragment); ok {
				df.removeChildInternal(child)
			} else {
				newChild.RemoveChild(child)
			}
			_, err := e.insertBeforeInternal(child, refChild)
			if err != nil {
				return nil, err
			}
		}

		return replaced, nil
	}

	return e.replaceChildInternal(newChild, oldChild)
}

// replaceChildInternal handles the actual replacement without DocumentFragment expansion for elements
func (e *element) replaceChildInternal(newChild Node, oldChild Node) (Node, error) {

	// Remove from current parent if exists - done internally to avoid deadlock
	if newChild.ParentNode() != nil {
		oldParent := newChild.ParentNode()
		oc := getInternalNode(newChild)

		// Update sibling links
		if oc.previousSibling != nil {
			if prevNode := getInternalNode(oc.previousSibling); prevNode != nil {
				prevNode.nextSibling = oc.nextSibling
			}
		}
		if oc.nextSibling != nil {
			if nextNode := getInternalNode(oc.nextSibling); nextNode != nil {
				nextNode.previousSibling = oc.previousSibling
			}
		}

		// Update parent's first/last child pointers
		if op := getInternalNode(oldParent); op != nil {
			if op.firstChild == newChild {
				op.firstChild = oc.nextSibling
			}
			if op.lastChild == newChild {
				op.lastChild = oc.previousSibling
			}
			// Update old parent's live NodeList if it exists
			if op.childNodes != nil && op.childNodes.update != nil {
				op.childNodes.update()
			}
		}

		// Clear the removed node's parent/sibling references
		oc.parentNode = nil
		oc.previousSibling = nil
		oc.nextSibling = nil
	}

	nc := getInternalNode(newChild)
	oc := getInternalNode(oldChild)

	nc.nextSibling = oc.nextSibling
	nc.previousSibling = oc.previousSibling
	nc.parentNode = Node(e)

	if oc.previousSibling != nil {
		if prevNode := getInternalNode(oc.previousSibling); prevNode != nil {
			prevNode.nextSibling = newChild
		}
	} else {
		e.firstChild = newChild
	}

	if oc.nextSibling != nil {
		if nextNode := getInternalNode(oc.nextSibling); nextNode != nil {
			nextNode.previousSibling = newChild
		}
	} else {
		e.lastChild = newChild
	}

	oc.parentNode = nil
	oc.nextSibling = nil
	oc.previousSibling = nil

	// Update live NodeList if it exists
	if e.childNodes != nil && e.childNodes.update != nil {
		e.childNodes.update()
	}
	return oldChild, nil
}

func (e *element) RemoveChild(oldChild Node) (Node, error) {
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}
	// Check parent - need to compare interface values properly
	parent := oldChild.ParentNode()
	if parent == nil || parent != Node(e) {
		return nil, NewDOMException("NotFoundError", "")
	}

	oc := getInternalNode(oldChild)

	if oc.previousSibling != nil {
		if prevNode := getInternalNode(oc.previousSibling); prevNode != nil {
			prevNode.nextSibling = oc.nextSibling
		}
	} else {
		e.firstChild = oc.nextSibling
	}

	if oc.nextSibling != nil {
		if nextNode := getInternalNode(oc.nextSibling); nextNode != nil {
			nextNode.previousSibling = oc.previousSibling
		}
	} else {
		e.lastChild = oc.previousSibling
	}

	oc.parentNode = nil
	oc.nextSibling = nil
	oc.previousSibling = nil

	// Update live NodeList if it exists
	if e.childNodes != nil && e.childNodes.update != nil {
		e.childNodes.update()
	}
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.notifyMutation()
		}
	}
	return oldChild, nil
}

func (e *element) CloneNode(deep bool) Node {
	clone := &element{
		node: node{
			nodeType:      e.nodeType,
			nodeName:      e.nodeName,
			nodeValue:     e.nodeValue,
			ownerDocument: e.ownerDocument,
			namespaceURI:  e.namespaceURI,
			prefix:        e.prefix,
			localName:     e.localName,
		},
	}

	if e.attributes != nil {
		clone.attributes = NewNamedNodeMap()
		for _, key := range e.attributes.order {
			attr := e.attributes.items[key]
			clonedAttr := attr.CloneNode(true)
			clone.attributes.SetNamedItem(clonedAttr)
		}
	}

	if deep {
		for child := e.firstChild; child != nil; child = child.NextSibling() {
			clone.AppendChild(child.CloneNode(true))
		}
	}
	return clone
}

// removeChildInternal removes a child without acquiring document lock
// This is used internally to avoid deadlocks when the lock is already held
func (e *element) removeChildInternal(oldChild Node) error {
	// Check parent - need to compare interface values properly
	parent := oldChild.ParentNode()
	if parent == nil || parent != Node(e) {
		return NewDOMException("NotFoundError", "")
	}

	oc := getInternalNode(oldChild)

	if oc.previousSibling != nil {
		if prevNode := getInternalNode(oc.previousSibling); prevNode != nil {
			prevNode.nextSibling = oc.nextSibling
		}
	} else {
		e.firstChild = oc.nextSibling
	}

	if oc.nextSibling != nil {
		if nextNode := getInternalNode(oc.nextSibling); nextNode != nil {
			nextNode.previousSibling = oc.previousSibling
		}
	} else {
		e.lastChild = oc.previousSibling
	}

	oc.parentNode = nil
	oc.nextSibling = nil
	oc.previousSibling = nil

	// Update live NodeList if it exists
	if e.childNodes != nil && e.childNodes.update != nil {
		e.childNodes.update()
	}
	return nil
}

func (e *element) TagName() DOMString {
	return e.nodeName
}

func (e *element) GetAttribute(name DOMString) DOMString {
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.RLock()
			defer d.mu.RUnlock()
		}
	}
	if e.attributes != nil {
		if attr := e.attributes.GetNamedItem(name); attr != nil {
			return attr.NodeValue()
		}
	}
	return ""
}

func (e *element) SetAttribute(name, value DOMString) error {
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}
	if e.attributes == nil {
		e.attributes = NewNamedNodeMap()
	}

	var oldValue DOMString
	if existingAttr := e.attributes.GetNamedItem(name); existingAttr != nil {
		oldValue = existingAttr.NodeValue()
		existingAttr.SetNodeValue(value)
	} else {
		oldValue = ""
		newAttr, err := e.ownerDocument.CreateAttribute(name)
		if err != nil {
			return err
		}
		if newAttr == nil {
			return NewDOMException("InvalidCharacterError", "Invalid attribute name")
		}
		a := newAttr.(*attr)
		a.ownerElement = e
		newAttr.SetValue(value)
		e.attributes.SetNamedItem(newAttr)
	}

	// Update ID mapping
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.updateIdMappingForElement(e, name, oldValue, value)
			d.notifyMutation()
		}
	}
	return nil
}

func (e *element) RemoveAttribute(name DOMString) error {
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}

	var oldValue DOMString
	if e.attributes != nil {
		if existingAttr := e.attributes.GetNamedItem(name); existingAttr != nil {
			oldValue = existingAttr.NodeValue()
		}
		e.attributes.RemoveNamedItem(name)
	}

	// Update ID mapping
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.updateIdMappingForElement(e, name, oldValue, "")
			d.notifyMutation()
		}
	}
	return nil
}

func (e *element) GetAttributeNode(name DOMString) Attr {
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.RLock()
			defer d.mu.RUnlock()
		}
	}
	if e.attributes != nil {
		if n := e.attributes.GetNamedItem(name); n != nil {
			return n.(Attr)
		}
	}
	return nil
}

func (e *element) SetAttributeNode(newAttr Attr) (Attr, error) {
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}
	a := newAttr.(*attr)
	if a.ownerElement != nil {
		return nil, NewDOMException("InUseAttributeError", "Attribute already in use")
	}
	if e.attributes == nil {
		e.attributes = NewNamedNodeMap()
	}

	var oldValue DOMString
	oldNode, _ := e.attributes.SetNamedItem(newAttr)
	if oldNode != nil {
		oldValue = oldNode.NodeValue()
		// Clear ownerElement of the replaced attribute
		if oldAttr, ok := oldNode.(*attr); ok {
			oldAttr.ownerElement = nil
		}
	}

	a.ownerElement = e

	// Update ID mapping
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.updateIdMappingForElement(e, newAttr.NodeName(), oldValue, newAttr.NodeValue())
			d.notifyMutation()
		}
	}
	if oldNode != nil {
		return oldNode.(Attr), nil
	}
	return nil, nil
}

func (e *element) RemoveAttributeNode(oldAttr Attr) (Attr, error) {
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}
	if e.attributes == nil {
		return nil, NewDOMException("NotFoundError", "Attribute not found")
	}
	removedNode, err := e.attributes.RemoveNamedItem(oldAttr.Name())
	if err != nil {
		return nil, err
	}

	// Nullify the ownerElement of the removed attribute
	if removedAttr, ok := removedNode.(*attr); ok {
		removedAttr.ownerElement = nil
	}

	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.notifyMutation()
		}
	}
	return removedNode.(Attr), nil
}

func (e *element) GetElementsByTagName(name DOMString) NodeList {
	doc, ok := e.ownerDocument.(*document)
	if !ok {
		// Should not happen in a well-formed document
		return &nodeList{items: []Node{}}
	}
	if doc != nil {
		doc.mu.RLock()
		defer doc.mu.RUnlock()
	}
	nl := &nodeList{
		root: e,
		filter: func(n Node) bool {
			return n.NodeType() == ELEMENT_NODE && (name == "*" || n.NodeName() == name)
		},
		live: true,
		doc:  doc,
	}
	nl.update = func() {
		nodes := []Node{}
		var helper func(Node)
		helper = func(n Node) {
			if n == nil {
				return
			}
			if nl.filter(n) {
				nodes = append(nodes, n)
			}
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				helper(child)
			}
		}
		// Start with children of root, not root itself (DOM spec: descendants only)
		for child := nl.root.FirstChild(); child != nil; child = child.NextSibling() {
			helper(child)
		}
		nl.items = nodes
	}
	nl.update() // initial population
	if doc.activeNodeLists == nil {
		doc.activeNodeLists = []*nodeList{}
	}
	doc.activeNodeLists = append(doc.activeNodeLists, nl)
	return nl
}

func (e *element) GetAttributeNS(namespaceURI, localName DOMString) DOMString {
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.RLock()
			defer d.mu.RUnlock()
		}
	}
	if e.attributes != nil {
		if attr := e.attributes.GetNamedItemNS(namespaceURI, localName); attr != nil {
			return attr.NodeValue()
		}
	}
	return ""
}

func (e *element) SetAttributeNS(namespaceURI, qualifiedName, value DOMString) error {
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}
	if e.attributes == nil {
		e.attributes = NewNamedNodeMap()
	}
	_, localName := parseQualifiedName(qualifiedName)
	var oldValue DOMString
	if existingAttr := e.attributes.GetNamedItemNS(namespaceURI, localName); existingAttr != nil {
		oldValue = existingAttr.NodeValue()
		existingAttr.SetNodeValue(value)
	} else {
		oldValue = ""
		newAttr, _ := e.ownerDocument.CreateAttributeNS(namespaceURI, qualifiedName)
		a := newAttr.(*attr)
		a.ownerElement = e
		newAttr.SetValue(value)
		e.attributes.SetNamedItemNS(newAttr)
	}

	// Update ID mapping
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.updateIdMappingForElement(e, localName, oldValue, value)
			d.notifyMutation()
		}
	}
	return nil
}

func (e *element) RemoveAttributeNS(namespaceURI, localName DOMString) error {
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}
	if e.attributes != nil {
		e.attributes.RemoveNamedItemNS(namespaceURI, localName)
	}
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.notifyMutation()
		}
	}
	return nil
}

func (e *element) GetAttributeNodeNS(namespaceURI, localName DOMString) Attr {
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.RLock()
			defer d.mu.RUnlock()
		}
	}
	if e.attributes != nil {
		if n := e.attributes.GetNamedItemNS(namespaceURI, localName); n != nil {
			return n.(Attr)
		}
	}
	return nil
}

func (e *element) SetAttributeNodeNS(newAttr Attr) (Attr, error) {
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}
	a := newAttr.(*attr)
	if a.ownerElement != nil {
		return nil, NewDOMException("InUseAttributeError", "Attribute already in use")
	}
	if e.attributes == nil {
		e.attributes = NewNamedNodeMap()
	}
	oldNode, _ := e.attributes.SetNamedItemNS(newAttr)
	a.ownerElement = e
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.notifyMutation()
		}
	}
	if oldNode != nil {
		return oldNode.(Attr), nil
	}
	return nil, nil
}

func (e *element) GetElementsByTagNameNS(namespaceURI, localName DOMString) NodeList {
	doc, ok := e.ownerDocument.(*document)
	if !ok {
		// Should not happen in a well-formed document
		return &nodeList{items: []Node{}}
	}
	if doc != nil {
		doc.mu.RLock()
		defer doc.mu.RUnlock()
	}
	nl := &nodeList{
		root: e,
		filter: func(n Node) bool {
			return n.NodeType() == ELEMENT_NODE &&
				(namespaceURI == "*" || n.NamespaceURI() == namespaceURI) &&
				(localName == "*" || n.LocalName() == localName)
		},
		live: true,
		doc:  doc,
	}
	nl.update = func() {
		nodes := []Node{}
		var helper func(Node)
		helper = func(n Node) {
			if n == nil {
				return
			}
			if nl.filter(n) {
				nodes = append(nodes, n)
			}
			for child := n.FirstChild(); child != nil; child = child.NextSibling() {
				helper(child)
			}
		}
		// Start with children of root, not root itself (DOM spec: descendants only)
		for child := nl.root.FirstChild(); child != nil; child = child.NextSibling() {
			helper(child)
		}
		nl.items = nodes
	}
	nl.update() // initial population
	if doc.activeNodeLists == nil {
		doc.activeNodeLists = []*nodeList{}
	}
	doc.activeNodeLists = append(doc.activeNodeLists, nl)
	return nl
}

// hasAttributeInternal checks if an attribute exists without acquiring locks
// This is used internally when locks are already held to avoid deadlocks
func (e *element) hasAttributeInternal(name DOMString) bool {
	return e.attributes != nil && e.attributes.GetNamedItem(name) != nil
}

func (e *element) HasAttribute(name DOMString) bool {
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.RLock()
			defer d.mu.RUnlock()
		}
	}
	return e.hasAttributeInternal(name)
}

// hasAttributeNSInternal checks if a namespaced attribute exists without acquiring locks
func (e *element) hasAttributeNSInternal(namespaceURI, localName DOMString) bool {
	return e.attributes != nil && e.attributes.GetNamedItemNS(namespaceURI, localName) != nil
}

func (e *element) HasAttributeNS(namespaceURI, localName DOMString) bool {
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.RLock()
			defer d.mu.RUnlock()
		}
	}
	return e.hasAttributeNSInternal(namespaceURI, localName)
}

// Element manipulation methods from Living Standard

func (e *element) ToggleAttribute(name DOMString, force ...bool) bool {
	if doc := e.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}

	// Use internal helper to avoid nested lock acquisition
	hasAttr := e.hasAttributeInternal(name)

	// Determine action based on force parameter
	var shouldAdd bool
	if len(force) > 0 {
		shouldAdd = force[0]
	} else {
		shouldAdd = !hasAttr
	}

	if shouldAdd && !hasAttr {
		// Create attribute internally without additional locking
		if e.attributes == nil {
			e.attributes = NewNamedNodeMap()
		}

		// Create a new attribute
		newAttr, _ := e.ownerDocument.CreateAttribute(name)
		a := newAttr.(*attr)
		a.ownerElement = e
		newAttr.SetValue("")
		e.attributes.SetNamedItem(newAttr)

		// Update ID index if this is an ID attribute
		if doc := e.OwnerDocument(); doc != nil && name == "id" {
			if d, ok := doc.(*document); ok {
				d.updateIdMappingForElement(e, name, "", "")
			}
		}
		return true
	} else if !shouldAdd && hasAttr {
		// Remove attribute internally without additional locking
		if e.attributes != nil {
			if attr := e.attributes.GetNamedItem(name); attr != nil {
				// Update ID index if this is an ID attribute
				if doc := e.OwnerDocument(); doc != nil && name == "id" {
					if d, ok := doc.(*document); ok {
						d.updateIdMappingForElement(e, name, attr.NodeValue(), "")
					}
				}
				e.attributes.RemoveNamedItem(name)
			}
		}
		return false
	}

	return shouldAdd
}

func (e *element) Remove() {
	if parent := e.ParentNode(); parent != nil {
		parent.RemoveChild(e)
	}
}

func (e *element) ReplaceWith(nodes ...Node) error {
	parent := e.ParentNode()
	if parent == nil {
		return nil // No parent, nothing to do
	}

	if len(nodes) == 0 {
		parent.RemoveChild(e)
		return nil
	}

	// Create a document fragment to hold the nodes
	doc := e.OwnerDocument()
	if doc == nil {
		return NewDOMException("InvalidStateError", "Element has no owner document")
	}

	frag := doc.CreateDocumentFragment()
	for _, node := range nodes {
		frag.AppendChild(node)
	}

	// Replace this element with the fragment
	_, err := parent.ReplaceChild(frag, e)
	return err
}

func (e *element) Before(nodes ...Node) error {
	parent := e.ParentNode()
	if parent == nil {
		return nil // No parent, nothing to do
	}

	if len(nodes) == 0 {
		return nil
	}

	// Create a document fragment to hold the nodes
	doc := e.OwnerDocument()
	if doc == nil {
		return NewDOMException("InvalidStateError", "Element has no owner document")
	}

	frag := doc.CreateDocumentFragment()
	for _, node := range nodes {
		frag.AppendChild(node)
	}

	// Insert the fragment before this element
	_, err := parent.InsertBefore(frag, e)
	return err
}

func (e *element) After(nodes ...Node) error {
	parent := e.ParentNode()
	if parent == nil {
		return nil // No parent, nothing to do
	}

	if len(nodes) == 0 {
		return nil
	}

	// Create a document fragment to hold the nodes
	doc := e.OwnerDocument()
	if doc == nil {
		return NewDOMException("InvalidStateError", "Element has no owner document")
	}

	frag := doc.CreateDocumentFragment()
	for _, node := range nodes {
		frag.AppendChild(node)
	}

	// Insert the fragment after this element
	nextSibling := e.NextSibling()
	if nextSibling != nil {
		_, err := parent.InsertBefore(frag, nextSibling)
		return err
	} else {
		_, err := parent.AppendChild(frag)
		return err
	}
}

func (e *element) Prepend(nodes ...Node) error {
	if len(nodes) == 0 {
		return nil
	}

	// Create a document fragment to hold the nodes
	doc := e.OwnerDocument()
	if doc == nil {
		return NewDOMException("InvalidStateError", "Element has no owner document")
	}

	frag := doc.CreateDocumentFragment()
	for _, node := range nodes {
		frag.AppendChild(node)
	}

	// Insert at the beginning
	firstChild := e.FirstChild()
	if firstChild != nil {
		_, err := e.InsertBefore(frag, firstChild)
		return err
	} else {
		_, err := e.AppendChild(frag)
		return err
	}
}

func (e *element) Append(nodes ...Node) error {
	if len(nodes) == 0 {
		return nil
	}

	// Create a document fragment to hold the nodes
	doc := e.OwnerDocument()
	if doc == nil {
		return NewDOMException("InvalidStateError", "Element has no owner document")
	}

	frag := doc.CreateDocumentFragment()
	for _, node := range nodes {
		frag.AppendChild(node)
	}

	// Append to the end
	_, err := e.AppendChild(frag)
	return err
}

// Element DOM properties from Living Standard

func (e *element) Children() ElementList {
	// Return a live collection of child elements only
	doc, ok := e.ownerDocument.(*document)
	if !ok {
		// Should not happen in a well-formed document
		return &elementList{items: []Element{}}
	}
	if doc != nil {
		doc.mu.RLock()
		defer doc.mu.RUnlock()
	}

	el := &elementList{
		root: e,
		filter: func(n Node) bool {
			// Only include element nodes that are direct children
			if n.NodeType() != ELEMENT_NODE {
				return false
			}
			return n.ParentNode() == e
		},
		converter: func(n Node) (Element, bool) {
			elem, ok := n.(Element)
			return elem, ok
		},
		live: true,
		doc:  doc,
	}

	// Initial population - collect child elements
	child := e.FirstChild()
	for child != nil {
		if el.filter != nil && el.filter(child) {
			if elem, ok := el.converter(child); ok {
				el.items = append(el.items, elem)
			}
		}
		child = child.NextSibling()
	}

	return el
}

func (e *element) FirstElementChild() Element {
	child := e.FirstChild()
	for child != nil {
		if child.NodeType() == ELEMENT_NODE {
			if elem, ok := child.(Element); ok {
				return elem
			}
		}
		child = child.NextSibling()
	}
	return nil
}

func (e *element) LastElementChild() Element {
	child := e.LastChild()
	for child != nil {
		if child.NodeType() == ELEMENT_NODE {
			if elem, ok := child.(Element); ok {
				return elem
			}
		}
		child = child.PreviousSibling()
	}
	return nil
}

func (e *element) PreviousElementSibling() Element {
	sibling := e.PreviousSibling()
	for sibling != nil {
		if sibling.NodeType() == ELEMENT_NODE {
			if elem, ok := sibling.(Element); ok {
				return elem
			}
		}
		sibling = sibling.PreviousSibling()
	}
	return nil
}

func (e *element) NextElementSibling() Element {
	sibling := e.NextSibling()
	for sibling != nil {
		if sibling.NodeType() == ELEMENT_NODE {
			if elem, ok := sibling.(Element); ok {
				return elem
			}
		}
		sibling = sibling.NextSibling()
	}
	return nil
}

func (e *element) ChildElementCount() uint32 {
	count := uint32(0)
	child := e.FirstChild()
	for child != nil {
		if child.NodeType() == ELEMENT_NODE {
			count++
		}
		child = child.NextSibling()
	}
	return count
}

// ===========================================================================
// Attribute Implementation
// ===========================================================================

// attr represents an attribute node
type attr struct {
	node
	ownerElement Element
}

func (a *attr) Name() DOMString {
	return a.nodeName
}

func (a *attr) Value() DOMString {
	return a.nodeValue
}

func (a *attr) SetValue(value DOMString) {
	oldValue := a.nodeValue
	a.nodeValue = value
	if a.ownerElement != nil && a.ownerElement.OwnerDocument() != nil {
		if doc, ok := a.ownerElement.OwnerDocument().(*document); ok {
			doc.updateIdMappingForElement(a.ownerElement, a.nodeName, oldValue, value)
		}
	}
}

func (a *attr) OwnerElement() Element {
	if a.ownerElement != nil {
		return a.ownerElement
	}
	return nil
}

// ===========================================================================
// CharacterData Implementation
// ===========================================================================

// characterData represents character data
type characterData struct {
	node
}

func (cd *characterData) Data() DOMString {
	return cd.nodeValue
}

func (cd *characterData) SetData(data DOMString) error {
	if doc := cd.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}
	cd.nodeValue = data
	return nil
}

func (cd *characterData) Length() uint {
	return uint(len(cd.nodeValue))
}

func (cd *characterData) SubstringData(offset, count uint) (DOMString, error) {
	if doc := cd.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.RLock()
			defer d.mu.RUnlock()
		}
	}
	length := cd.Length()
	if offset > length {
		return "", NewDOMException("IndexSizeError", "Offset out of bounds")
	}
	end := offset + count
	if end > length {
		end = length
	}
	return cd.nodeValue[offset:end], nil
}

func (cd *characterData) AppendData(arg DOMString) error {
	if doc := cd.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}
	cd.nodeValue = cd.nodeValue + arg
	return nil
}

func (cd *characterData) InsertData(offset uint, arg DOMString) error {
	if doc := cd.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}
	length := cd.Length()
	if offset > length {
		return NewDOMException("IndexSizeError", "Offset out of bounds")
	}
	cd.nodeValue = cd.nodeValue[:offset] + arg + cd.nodeValue[offset:]
	return nil
}

func (cd *characterData) DeleteData(offset, count uint) error {
	if doc := cd.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}
	length := cd.Length()
	if offset > length {
		return NewDOMException("IndexSizeError", "Offset out of bounds")
	}
	end := offset + count
	if end > length {
		end = length
	}
	cd.nodeValue = cd.nodeValue[:offset] + cd.nodeValue[end:]
	return nil
}

func (cd *characterData) ReplaceData(offset, count uint, arg DOMString) error {
	if err := cd.DeleteData(offset, count); err != nil {
		return err
	}
	return cd.InsertData(offset, arg)
}

// CharacterData manipulation methods from Living Standard

func (cd *characterData) Before(nodes ...Node) error {
	parent := cd.ParentNode()
	if parent == nil {
		return nil // No parent, nothing to do
	}

	if len(nodes) == 0 {
		return nil
	}

	// Insert each node directly before this node
	for _, node := range nodes {
		_, err := parent.InsertBefore(node, cd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cd *characterData) After(nodes ...Node) error {
	parent := cd.ParentNode()
	if parent == nil {
		return nil // No parent, nothing to do
	}

	if len(nodes) == 0 {
		return nil
	}

	// Create a document fragment to hold the nodes
	doc := cd.OwnerDocument()
	if doc == nil {
		return NewDOMException("InvalidStateError", "CharacterData has no owner document")
	}

	frag := doc.CreateDocumentFragment()
	for _, node := range nodes {
		frag.AppendChild(node)
	}

	// Insert the fragment after this node
	nextSibling := cd.NextSibling()
	if nextSibling != nil {
		_, err := parent.InsertBefore(frag, nextSibling)
		return err
	} else {
		_, err := parent.AppendChild(frag)
		return err
	}
}

func (cd *characterData) ReplaceWith(nodes ...Node) error {
	parent := cd.ParentNode()
	if parent == nil {
		return nil // No parent, nothing to do
	}

	if len(nodes) == 0 {
		parent.RemoveChild(cd)
		return nil
	}

	// Create a document fragment to hold the nodes
	doc := cd.OwnerDocument()
	if doc == nil {
		return NewDOMException("InvalidStateError", "CharacterData has no owner document")
	}

	frag := doc.CreateDocumentFragment()
	for _, node := range nodes {
		frag.AppendChild(node)
	}

	// Replace this node with the fragment
	_, err := parent.ReplaceChild(frag, cd)
	return err
}

func (cd *characterData) Remove() {
	if parent := cd.ParentNode(); parent != nil {
		parent.RemoveChild(cd)
	}
}

// ===========================================================================
// Text Node Implementation
// ===========================================================================

// text represents a text node
type text struct {
	characterData
}

func (t *text) SplitText(offset uint) (Text, error) {
	if doc := t.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}
	length := t.Length()
	if offset > length {
		return nil, NewDOMException("IndexSizeError", "Offset out of bounds")
	}

	newData := t.nodeValue[offset:]
	t.nodeValue = t.nodeValue[:offset]

	newText := t.ownerDocument.CreateTextNode(newData).(*text)

	// Manually insert the new text node after the current one
	if t.parentNode != nil {
		newText.parentNode = t.parentNode

		// Update sibling pointers
		newText.previousSibling = t
		newText.nextSibling = t.nextSibling

		if t.nextSibling != nil {
			if ns := getInternalNode(t.nextSibling); ns != nil {
				ns.previousSibling = newText
			}
		}
		t.nextSibling = newText

		// Update parent's lastChild if needed
		if p := getInternalNode(t.parentNode); p != nil {
			if p.lastChild == t {
				p.lastChild = newText
			}
			// Update parent's live NodeList if it exists
			if p.childNodes != nil && p.childNodes.update != nil {
				p.childNodes.update()
			}
		}
	}

	return newText, nil
}

// ===========================================================================
// Other Node Types
// ===========================================================================

// comment represents a comment node
type comment struct {
	characterData
}

// cdataSection represents a CDATA section
type cdataSection struct {
	text
}

// documentType represents a document type node
type documentType struct {
	node
	name           DOMString
	entities       *namedNodeMap
	notations      *namedNodeMap
	publicId       DOMString
	systemId       DOMString
	internalSubset DOMString
}

func (dt *documentType) Name() DOMString {
	return dt.name
}

func (dt *documentType) Entities() NamedNodeMap {
	return dt.entities
}

func (dt *documentType) Notations() NamedNodeMap {
	return dt.notations
}

func (dt *documentType) PublicId() DOMString {
	return dt.publicId
}

func (dt *documentType) SystemId() DOMString {
	return dt.systemId
}

func (dt *documentType) InternalSubset() DOMString {
	return dt.internalSubset
}

// notation represents a notation node
type notation struct {
	node
	publicId DOMString
	systemId DOMString
}

func (n *notation) PublicId() DOMString {
	return n.publicId
}

func (n *notation) SystemId() DOMString {
	return n.systemId
}

// entity represents an entity node
type entity struct {
	node
	publicId     DOMString
	systemId     DOMString
	notationName DOMString
}

func (e *entity) PublicId() DOMString {
	return e.publicId
}

func (e *entity) SystemId() DOMString {
	return e.systemId
}

func (e *entity) NotationName() DOMString {
	return e.notationName
}

// entityReference represents an entity reference node
type entityReference struct {
	node
}

// processingInstruction represents a processing instruction node
type processingInstruction struct {
	node
	target DOMString
	data   DOMString
}

func (pi *processingInstruction) Target() DOMString {
	return pi.target
}

func (pi *processingInstruction) Data() DOMString {
	return pi.data
}

func (pi *processingInstruction) SetData(data DOMString) error {
	if doc := pi.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}
	pi.data = data
	pi.nodeValue = data
	return nil
}

func (pi *processingInstruction) InsertBefore(newChild Node, refChild Node) (Node, error) {
	if doc := pi.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}
	result, err := pi.node.InsertBefore(newChild, refChild)
	if err != nil {
		return nil, err
	}
	// Fix the parent reference to be the processingInstruction interface
	if nc := getInternalNode(newChild); nc != nil {
		nc.parentNode = pi
	}
	return result, nil
}

func (pi *processingInstruction) AppendChild(newChild Node) (Node, error) {
	return pi.InsertBefore(newChild, nil)
}

func (pi *processingInstruction) ReplaceChild(newChild Node, oldChild Node) (Node, error) {
	if doc := pi.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}

	if newChild == nil {
		return nil, NewDOMException("HierarchyRequestError", "Invalid node")
	}

	if oldChild.ParentNode() != pi {
		return nil, NewDOMException("NotFoundError", "")
	}

	// Handle self-replacement: replacing node with itself should be a no-op
	if newChild == oldChild {
		return oldChild, nil
	}

	// Check document ownership
	if newChild.OwnerDocument() != pi.ownerDocument && pi.ownerDocument != nil {
		return nil, NewDOMException("WrongDocumentError", "")
	}

	// HIERARCHY_REQUEST_ERR check - prevent cycles
	for ancestor := Node(pi); ancestor != nil; ancestor = ancestor.ParentNode() {
		if ancestor == newChild {
			return nil, NewDOMException("HierarchyRequestError", "Cannot insert a node as a descendant of itself")
		}
	}

	// Handle DocumentFragment - replace with its children
	if newChild.NodeType() == DOCUMENT_FRAGMENT_NODE {
		// Collect all children of the fragment first
		var children []Node
		for child := newChild.FirstChild(); child != nil; child = child.NextSibling() {
			children = append(children, child)
		}

		if len(children) == 0 {
			// Empty fragment, just remove the old child
			return pi.RemoveChild(oldChild)
		}

		// Replace with first child, then insert remaining children after it
		firstChild := children[0]
		if srcFrag, ok := newChild.(*documentFragment); ok {
			srcFrag.removeChildInternal(firstChild)
		} else {
			newChild.RemoveChild(firstChild)
		}

		// Use replaceChildInternal to avoid infinite recursion
		replaced, err := pi.replaceChildInternal(firstChild, oldChild)
		if err != nil {
			return nil, err
		}

		// Insert remaining children after the first one
		refChild := firstChild.NextSibling()
		for _, child := range children[1:] {
			if srcFrag, ok := newChild.(*documentFragment); ok {
				srcFrag.removeChildInternal(child)
			} else {
				newChild.RemoveChild(child)
			}
			_, err := pi.node.insertBeforeInternal(child, refChild)
			if err != nil {
				return nil, err
			}
			// Fix the parent reference to be the processingInstruction interface
			if nc := getInternalNode(child); nc != nil {
				nc.parentNode = pi
			}
		}

		return replaced, nil
	}

	return pi.replaceChildInternal(newChild, oldChild)
}

// replaceChildInternal handles the actual replacement without DocumentFragment expansion
func (pi *processingInstruction) replaceChildInternal(newChild Node, oldChild Node) (Node, error) {
	// Remove from current parent if exists - done internally to avoid deadlock
	if newChild.ParentNode() != nil {
		oldParent := newChild.ParentNode()
		oc := getInternalNode(newChild)

		// Update sibling links
		if oc.previousSibling != nil {
			if prevNode := getInternalNode(oc.previousSibling); prevNode != nil {
				prevNode.nextSibling = oc.nextSibling
			}
		}
		if oc.nextSibling != nil {
			if nextNode := getInternalNode(oc.nextSibling); nextNode != nil {
				nextNode.previousSibling = oc.previousSibling
			}
		}

		// Update parent's first/last child pointers
		if op := getInternalNode(oldParent); op != nil {
			if op.firstChild == newChild {
				op.firstChild = oc.nextSibling
			}
			if op.lastChild == newChild {
				op.lastChild = oc.previousSibling
			}
			// Update old parent's live NodeList if it exists
			if op.childNodes != nil && op.childNodes.update != nil {
				op.childNodes.update()
			}
		}

		// Clear the removed node's parent/sibling references
		oc.parentNode = nil
		oc.previousSibling = nil
		oc.nextSibling = nil
	}

	nc := getInternalNode(newChild)
	oc := getInternalNode(oldChild)

	nc.nextSibling = oc.nextSibling
	nc.previousSibling = oc.previousSibling
	nc.parentNode = pi // Store as processingInstruction interface

	if oc.previousSibling != nil {
		if prevNode := getInternalNode(oc.previousSibling); prevNode != nil {
			prevNode.nextSibling = newChild
		}
	} else {
		pi.firstChild = newChild
	}

	if oc.nextSibling != nil {
		if nextNode := getInternalNode(oc.nextSibling); nextNode != nil {
			nextNode.previousSibling = newChild
		}
	} else {
		pi.lastChild = newChild
	}

	oc.parentNode = nil
	oc.nextSibling = nil
	oc.previousSibling = nil

	// Update live NodeList if it exists
	if pi.childNodes != nil && pi.childNodes.update != nil {
		pi.childNodes.update()
	}
	if doc := pi.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.notifyMutation()
		}
	}
	return oldChild, nil
}

func (pi *processingInstruction) RemoveChild(oldChild Node) (Node, error) {
	if doc := pi.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}
	parent := oldChild.ParentNode()
	if parent == nil || parent != Node(pi) {
		return nil, NewDOMException("NotFoundError", "")
	}

	oc := getInternalNode(oldChild)

	if oc.previousSibling != nil {
		if prevNode := getInternalNode(oc.previousSibling); prevNode != nil {
			prevNode.nextSibling = oc.nextSibling
		}
	} else {
		pi.firstChild = oc.nextSibling
	}

	if oc.nextSibling != nil {
		if nextNode := getInternalNode(oc.nextSibling); nextNode != nil {
			nextNode.previousSibling = oc.previousSibling
		}
	} else {
		pi.lastChild = oc.previousSibling
	}

	oc.parentNode = nil
	oc.nextSibling = nil
	oc.previousSibling = nil

	// Update live NodeList if it exists
	if pi.childNodes != nil && pi.childNodes.update != nil {
		pi.childNodes.update()
	}
	if doc := pi.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.notifyMutation()
		}
	}
	return oldChild, nil
}

// documentFragment represents a document fragment node
type documentFragment struct {
	node
}

func (df *documentFragment) InsertBefore(newChild Node, refChild Node) (Node, error) {
	if doc := df.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}
	result, err := df.node.InsertBefore(newChild, refChild)
	if err != nil {
		return nil, err
	}
	// Fix the parent reference to be the documentFragment interface
	if nc := getInternalNode(newChild); nc != nil {
		nc.parentNode = df
	}
	return result, nil
}

func (df *documentFragment) AppendChild(newChild Node) (Node, error) {
	return df.InsertBefore(newChild, nil)
}

func (df *documentFragment) RemoveChild(oldChild Node) (Node, error) {
	if doc := df.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}
	return df.removeChildInternal(oldChild)
}

func (df *documentFragment) ReplaceChild(newChild Node, oldChild Node) (Node, error) {
	if doc := df.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.mu.Lock()
			defer d.mu.Unlock()
		}
	}

	if newChild == nil {
		return nil, NewDOMException("HierarchyRequestError", "Invalid node")
	}

	// Compare with df (the interface) instead of Node(df.node)
	if oldChild.ParentNode() != df {
		return nil, NewDOMException("NotFoundError", "")
	}

	// Handle self-replacement: replacing node with itself should be a no-op
	if newChild == oldChild {
		return oldChild, nil
	}

	// Check document ownership
	if newChild.OwnerDocument() != df.ownerDocument && df.ownerDocument != nil {
		return nil, NewDOMException("WrongDocumentError", "")
	}

	// HIERARCHY_REQUEST_ERR check - prevent cycles
	for ancestor := Node(df); ancestor != nil; ancestor = ancestor.ParentNode() {
		if ancestor == newChild {
			return nil, NewDOMException("HierarchyRequestError", "Cannot insert a node as a descendant of itself")
		}
	}

	// Handle DocumentFragment - replace with its children
	if newChild.NodeType() == DOCUMENT_FRAGMENT_NODE {
		// Collect all children of the fragment first
		var children []Node
		for child := newChild.FirstChild(); child != nil; child = child.NextSibling() {
			children = append(children, child)
		}

		if len(children) == 0 {
			// Empty fragment, just remove the old child
			return df.RemoveChild(oldChild)
		}

		// Replace with first child, then insert remaining children after it
		firstChild := children[0]
		if srcFrag, ok := newChild.(*documentFragment); ok {
			srcFrag.removeChildInternal(firstChild)
		} else {
			newChild.RemoveChild(firstChild)
		}

		// Use replaceChildInternal to avoid infinite recursion
		replaced, err := df.replaceChildInternal(firstChild, oldChild)
		if err != nil {
			return nil, err
		}

		// Insert remaining children after the first one
		refChild := firstChild.NextSibling()
		for _, child := range children[1:] {
			if srcFrag, ok := newChild.(*documentFragment); ok {
				srcFrag.removeChildInternal(child)
			} else {
				newChild.RemoveChild(child)
			}
			_, err := df.node.insertBeforeInternal(child, refChild)
			if err != nil {
				return nil, err
			}
			// Fix the parent reference to be the documentFragment interface
			if nc := getInternalNode(child); nc != nil {
				nc.parentNode = df
			}
		}

		return replaced, nil
	}

	return df.replaceChildInternal(newChild, oldChild)
}

// replaceChildInternal handles the actual replacement without DocumentFragment expansion
func (df *documentFragment) replaceChildInternal(newChild Node, oldChild Node) (Node, error) {
	// Remove from current parent if exists - done internally to avoid deadlock
	if newChild.ParentNode() != nil {
		oldParent := newChild.ParentNode()
		oc := getInternalNode(newChild)

		// Update sibling links
		if oc.previousSibling != nil {
			if prevNode := getInternalNode(oc.previousSibling); prevNode != nil {
				prevNode.nextSibling = oc.nextSibling
			}
		}
		if oc.nextSibling != nil {
			if nextNode := getInternalNode(oc.nextSibling); nextNode != nil {
				nextNode.previousSibling = oc.previousSibling
			}
		}

		// Update parent's first/last child pointers
		if op := getInternalNode(oldParent); op != nil {
			if op.firstChild == newChild {
				op.firstChild = oc.nextSibling
			}
			if op.lastChild == newChild {
				op.lastChild = oc.previousSibling
			}
			// Update old parent's live NodeList if it exists
			if op.childNodes != nil && op.childNodes.update != nil {
				op.childNodes.update()
			}
		}

		// Clear the removed node's parent/sibling references
		oc.parentNode = nil
		oc.previousSibling = nil
		oc.nextSibling = nil
	}

	nc := getInternalNode(newChild)
	oc := getInternalNode(oldChild)

	nc.nextSibling = oc.nextSibling
	nc.previousSibling = oc.previousSibling
	nc.parentNode = df // Store as documentFragment interface

	if oc.previousSibling != nil {
		if prevNode := getInternalNode(oc.previousSibling); prevNode != nil {
			prevNode.nextSibling = newChild
		}
	} else {
		df.firstChild = newChild
	}

	if oc.nextSibling != nil {
		if nextNode := getInternalNode(oc.nextSibling); nextNode != nil {
			nextNode.previousSibling = newChild
		}
	} else {
		df.lastChild = newChild
	}

	oc.parentNode = nil
	oc.nextSibling = nil
	oc.previousSibling = nil

	// Update live NodeList if it exists
	if df.childNodes != nil && df.childNodes.update != nil {
		df.childNodes.update()
	}
	if doc := df.OwnerDocument(); doc != nil {
		if d, ok := doc.(*document); ok {
			d.notifyMutation()
		}
	}
	return oldChild, nil
}

// removeChildInternal handles the actual removal without acquiring locks
func (df *documentFragment) removeChildInternal(oldChild Node) (Node, error) {
	parent := oldChild.ParentNode()
	if parent == nil || parent != Node(df) {
		return nil, NewDOMException("NotFoundError", "")
	}

	oc := getInternalNode(oldChild)

	if oc.previousSibling != nil {
		if prevNode := getInternalNode(oc.previousSibling); prevNode != nil {
			prevNode.nextSibling = oc.nextSibling
		}
	} else {
		df.firstChild = oc.nextSibling
	}

	if oc.nextSibling != nil {
		if nextNode := getInternalNode(oc.nextSibling); nextNode != nil {
			nextNode.previousSibling = oc.previousSibling
		}
	} else {
		df.lastChild = oc.previousSibling
	}

	oc.parentNode = nil
	oc.nextSibling = nil
	oc.previousSibling = nil

	// Update live NodeList if it exists
	if df.childNodes != nil && df.childNodes.update != nil {
		df.childNodes.update()
	}
	return oldChild, nil
}

// ===========================================================================
// DOMImplementation
// ===========================================================================

// domImplementation provides methods for operations independent of any document instance
type domImplementation struct{}

func NewDOMImplementation() DOMImplementation {
	return &domImplementation{}
}

func (di *domImplementation) HasFeature(feature, version DOMString) bool {
	if versions, ok := supportedFeatures[feature]; ok {
		if version == "" {
			return true
		}
		for _, v := range versions {
			if v == version {
				return true
			}
		}
	}
	return false
}

func (di *domImplementation) CreateDocumentType(qualifiedName, publicId, systemId DOMString) (DocumentType, error) {
	return &documentType{
		node: node{
			nodeType: DOCUMENT_TYPE_NODE,
			nodeName: qualifiedName,
		},
		name:      qualifiedName,
		publicId:  publicId,
		systemId:  systemId,
		entities:  NewNamedNodeMap(),
		notations: NewNamedNodeMap(),
	}, nil
}

func (di *domImplementation) CreateDocument(namespaceURI, qualifiedName DOMString, doctype DocumentType) (Document, error) {
	doc := &document{
		node: node{
			nodeType:   DOCUMENT_NODE,
			nodeName:   "#document",
			attributes: NewNamedNodeMap(),
		},
		implementation: di,
	}
	doc.node.ownerDocument = doc

	if doctype != nil {
		doc.doctype = doctype
	}

	if qualifiedName != "" {
		elem, err := doc.CreateElementNS(namespaceURI, qualifiedName)
		if err != nil {
			return nil, err
		}
		doc.AppendChild(elem)
		doc.documentElement = elem
	}

	return doc, nil
}

// ===========================================================================
// Helper Functions
// ===========================================================================

// getInternalNode extracts the internal *node from any Node interface implementation
func getInternalNode(n Node) *node {
	if n == nil {
		return nil
	}

	// Handle all concrete types that embed node
	switch v := n.(type) {
	case *node:
		return v
	case *document:
		return &v.node
	case *element:
		return &v.node
	case *attr:
		return &v.node
	case *text:
		return &v.characterData.node
	case *comment:
		return &v.characterData.node
	case *cdataSection:
		return &v.text.characterData.node
	case *documentFragment:
		return &v.node
	case *documentType:
		return &v.node
	case *processingInstruction:
		return &v.node
	case *entity:
		return &v.node
	case *entityReference:
		return &v.node
	case *notation:
		return &v.node
	case *characterData:
		return &v.node
	default:
		// This should not happen if all types are handled
		return nil
	}
}

// isSameNode compares two nodes for identity (same object)
func isSameNode(a, b Node) bool {
	if a == nil || b == nil {
		return a == b
	}
	// Compare using reflection to get the actual pointer values
	// This handles the case where *element and *node point to same memory
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)
	if aVal.Kind() == reflect.Ptr && bVal.Kind() == reflect.Ptr {
		return aVal.Pointer() == bVal.Pointer()
	}
	return a == b
}

// IsValidName checks if a string is a valid XML Name.
// See https://www.w3.org/TR/xml/#NT-Name
func IsValidName(name DOMString) bool {
	if name == "" {
		return false
	}
	for i, r := range string(name) {
		if i == 0 {
			if !isNameStartChar(r) {
				return false
			}
		} else {
			if !isNameChar(r) {
				return false
			}
		}
	}
	return true
}

func isNameStartChar(r rune) bool {
	return r == ':' ||
		(r >= 'A' && r <= 'Z') ||
		r == '_' ||
		(r >= 'a' && r <= 'z') ||
		(r >= 0xC0 && r <= 0xD6) ||
		(r >= 0xD8 && r <= 0xF6) ||
		(r >= 0xF8 && r <= 0x2FF) ||
		(r >= 0x370 && r <= 0x37D) ||
		(r >= 0x37F && r <= 0x1FFF) ||
		(r >= 0x200C && r <= 0x200D) ||
		(r >= 0x2070 && r <= 0x218F) ||
		(r >= 0x2C00 && r <= 0x2FEF) ||
		(r >= 0x3001 && r <= 0xD7FF) ||
		(r >= 0xF900 && r <= 0xFDCF) ||
		(r >= 0xFDF0 && r <= 0xFFFD) ||
		(r >= 0x10000 && r <= 0xEFFFF)
}

func isNameChar(r rune) bool {
	return isNameStartChar(r) ||
		r == '-' ||
		r == '.' ||
		(r >= '0' && r <= '9') ||
		r == 0xB7 ||
		(r >= 0x0300 && r <= 0x036F) ||
		(r >= 0x203F && r <= 0x2040)
}

// parseQualifiedName parses a qualified name into prefix and local name
func parseQualifiedName(qualifiedName DOMString) (prefix, localName DOMString) {
	parts := strings.SplitN(string(qualifiedName), ":", 2)
	if len(parts) == 2 {
		return DOMString(parts[0]), DOMString(parts[1])
	}
	return "", qualifiedName
}
