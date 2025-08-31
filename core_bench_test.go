package xmldom_test

import (
	"fmt"
	"testing"

	"github.com/gogo-agent/xmldom"
)

// createTestDocument creates a document with a simple structure for testing.
func createTestDocument() xmldom.Document {
	impl := xmldom.NewDOMImplementation()
	doc, _ := impl.CreateDocument("", "root", nil)
	return doc
}

// createDeepDOM creates a DOM tree with a specified depth.
func createDeepDOM(b *testing.B, depth int) xmldom.Document {
	b.Helper()
	doc := createTestDocument()
	parent := doc.DocumentElement()
	for i := 0; i < depth; i++ {
		elem, _ := doc.CreateElement(xmldom.DOMString(fmt.Sprintf("elem%d", i)))
		parent.AppendChild(elem)
		parent = elem
	}
	return doc
}

// createWideDOM creates a DOM tree with a specified width (number of direct children).
func createWideDOM(b *testing.B, width int) xmldom.Document {
	b.Helper()
	doc := createTestDocument()
	root := doc.DocumentElement()
	for i := 0; i < width; i++ {
		elem, _ := doc.CreateElement("child")
		root.AppendChild(elem)
	}
	return doc
}

func BenchmarkGetElementsByTagName_Deep_Iteration(b *testing.B) {
	doc := createDeepDOM(b, 100) // A tree with 100 nested elements
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		nodeList := doc.GetElementsByTagName("elem99")
		if nodeList.Length() > 0 {
			_ = nodeList.Item(0)
		}
	}
}

func BenchmarkGetElementsByTagName_Wide_Iteration(b *testing.B) {
	doc := createWideDOM(b, 1000) // A root element with 1000 children
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		nodeList := doc.GetElementsByTagName("child")
		// Note: We're iterating over a potentially large NodeList
		for j := uint(0); j < nodeList.Length(); j++ {
			_ = nodeList.Item(j)
		}
	}
}

func BenchmarkCreateAndAppendElements(b *testing.B) {
	doc := createTestDocument()
	root := doc.DocumentElement()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		elem, _ := doc.CreateElement("child")
		root.AppendChild(elem)
	}
}

func BenchmarkSetAndRemoveAttribute(b *testing.B) {
	doc := createTestDocument()
	elem := doc.DocumentElement()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		elem.SetAttribute("name", "value")
		elem.RemoveAttribute("name")
	}
}

func BenchmarkChildNodes_Iteration(b *testing.B) {
	doc := createWideDOM(b, 1000) // A root element with 1000 children
	root := doc.DocumentElement()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		childNodes := root.ChildNodes()
		for j := uint(0); j < childNodes.Length(); j++ {
			_ = childNodes.Item(j)
		}
	}
}

func BenchmarkInsertBefore(b *testing.B) {
	doc := createTestDocument()
	root := doc.DocumentElement()
	refChild, _ := doc.CreateElement("ref_child")
	root.AppendChild(refChild)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		elem, _ := doc.CreateElement("new_element")
		_, _ = root.InsertBefore(elem, refChild)
	}
}

func BenchmarkRemoveChild(b *testing.B) {
	doc := createWideDOM(b, b.N)
	root := doc.DocumentElement()
	children := root.ChildNodes()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		child := children.Item(uint(i))
		if child != nil {
			_, _ = root.RemoveChild(child)
		}
	}
}

// ============================================================================
// Benchmarks for DOM Living Standard Features
// ============================================================================

// BenchmarkNodeIterator benchmarks NodeIterator traversal
func BenchmarkNodeIterator(b *testing.B) {
	doc := createWideDOM(b, 100)
	root := doc.DocumentElement()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		iter, _ := doc.CreateNodeIterator(root, 0xFFFFFFFF, nil)
		count := 0
		for {
			node, _ := iter.NextNode()
			if node == nil {
				break
			}
			count++
		}
	}
}

// BenchmarkTreeWalker benchmarks TreeWalker traversal
func BenchmarkTreeWalker(b *testing.B) {
	doc := createWideDOM(b, 100)
	root := doc.DocumentElement()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		walker, _ := doc.CreateTreeWalker(root, xmldom.SHOW_ELEMENT, nil)
		count := 0
		for walker.NextNode() != nil {
			count++
		}
	}
}

// BenchmarkRange benchmarks Range operations
func BenchmarkRange(b *testing.B) {
	doc := createTestDocument()
	root := doc.DocumentElement()
	text1 := doc.CreateTextNode("Hello World")
	text2 := doc.CreateTextNode("Another text")
	root.AppendChild(text1)
	root.AppendChild(text2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := doc.CreateRange()
		r.SetStart(text1, 0)
		r.SetEnd(text2, 5)
		r.SelectNode(text1)
		r.Collapse(true)
		_ = r.CloneRange()
	}
}

// BenchmarkNormalizeDocument benchmarks document normalization
func BenchmarkNormalizeDocument(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		doc := createTestDocument()
		root := doc.DocumentElement()
		// Create many adjacent text nodes
		for j := 0; j < 100; j++ {
			text := doc.CreateTextNode(xmldom.DOMString(fmt.Sprintf("text%d", j)))
			root.AppendChild(text)
		}
		b.StartTimer()

		doc.NormalizeDocument()
	}
}

// Benchmarkxmldom.ElementToggleAttribute benchmarks ToggleAttribute
func BenchmarkElementToggleAttribute(b *testing.B) {
	doc := createTestDocument()
	elem := doc.DocumentElement()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		elem.ToggleAttribute("disabled")
	}
}

// Benchmarkxmldom.ElementRemove benchmarks the Remove method
func BenchmarkElementRemove(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		doc := createTestDocument()
		root := doc.DocumentElement()
		elem, _ := doc.CreateElement("child")
		root.AppendChild(elem)
		b.StartTimer()

		elem.Remove()
	}
}

// Benchmarkxmldom.ElementReplaceWith benchmarks ReplaceWith
func BenchmarkElementReplaceWith(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		doc := createTestDocument()
		root := doc.DocumentElement()
		elem1, _ := doc.CreateElement("elem1")
		root.AppendChild(elem1)
		b.StartTimer()

		elem2, _ := doc.CreateElement("elem2")
		elem1.ReplaceWith(elem2)
	}
}

// Benchmarkxmldom.ElementBefore benchmarks Before method
func BenchmarkElementBefore(b *testing.B) {
	doc := createTestDocument()
	root := doc.DocumentElement()
	anchor, _ := doc.CreateElement("anchor")
	root.AppendChild(anchor)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		elem, _ := doc.CreateElement("before")
		anchor.Before(elem)
	}
}

// Benchmarkxmldom.ElementAfter benchmarks After method
func BenchmarkElementAfter(b *testing.B) {
	doc := createTestDocument()
	root := doc.DocumentElement()
	anchor, _ := doc.CreateElement("anchor")
	root.AppendChild(anchor)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		elem, _ := doc.CreateElement("after")
		anchor.After(elem)
	}
}

// Benchmarkxmldom.ElementPrepend benchmarks Prepend method
func BenchmarkElementPrepend(b *testing.B) {
	doc := createTestDocument()
	root := doc.DocumentElement()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		elem, _ := doc.CreateElement("prepend")
		root.Prepend(elem)
	}
}

// Benchmarkxmldom.ElementAppend benchmarks Append method
func BenchmarkElementAppend(b *testing.B) {
	doc := createTestDocument()
	root := doc.DocumentElement()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		elem, _ := doc.CreateElement("append")
		root.Append(elem)
	}
}

// Benchmarkxmldom.ElementChildren benchmarks Children property
func BenchmarkElementChildren(b *testing.B) {
	doc := createWideDOM(b, 100)
	root := doc.DocumentElement()
	// Add some text nodes to make it mixed content
	for i := 0; i < 50; i++ {
		text := doc.CreateTextNode("text")
		root.AppendChild(text)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		children := root.Children()
		for j := uint(0); j < children.Length(); j++ {
			_ = children.Item(j)
		}
	}
}

// Benchmarkxmldom.ElementFirstxmldom.ElementChild benchmarks Firstxmldom.ElementChild
func BenchmarkElementFirstElementChild(b *testing.B) {
	doc := createTestDocument()
	root := doc.DocumentElement()
	// Add text nodes before elements
	for i := 0; i < 10; i++ {
		text := doc.CreateTextNode("text")
		root.AppendChild(text)
	}
	elem, _ := doc.CreateElement("first")
	root.AppendChild(elem)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = root.FirstElementChild()
	}
}

// Benchmarkxmldom.ElementLastxmldom.ElementChild benchmarks Lastxmldom.ElementChild
func BenchmarkElementLastElementChild(b *testing.B) {
	doc := createTestDocument()
	root := doc.DocumentElement()
	elem, _ := doc.CreateElement("last")
	root.AppendChild(elem)
	// Add text nodes after elements
	for i := 0; i < 10; i++ {
		text := doc.CreateTextNode("text")
		root.AppendChild(text)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = root.LastElementChild()
	}
}

// Benchmarkxmldom.ElementChildxmldom.ElementCount benchmarks Childxmldom.ElementCount
func BenchmarkElementChildElementCount(b *testing.B) {
	doc := createWideDOM(b, 100)
	root := doc.DocumentElement()
	// Add text nodes to make it mixed
	for i := 0; i < 100; i++ {
		text := doc.CreateTextNode("text")
		root.AppendChild(text)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = root.ChildElementCount()
	}
}

// BenchmarkNodeComparison benchmarks node comparison methods
func BenchmarkNodeComparison(b *testing.B) {
	doc := createDeepDOM(b, 10)
	root := doc.DocumentElement()
	deepChild := root
	for deepChild.FirstChild() != nil {
		deepChild = deepChild.FirstChild().(xmldom.Element)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = root.Contains(deepChild)
		_ = root.CompareDocumentPosition(deepChild)
		_ = deepChild.IsConnected()
	}
}

// BenchmarkTextContent benchmarks TextContent getter
func BenchmarkTextContentGet(b *testing.B) {
	doc := createTestDocument()
	root := doc.DocumentElement()
	// Create deep structure with text
	parent := root
	for i := 0; i < 10; i++ {
		elem, _ := doc.CreateElement("elem")
		text := doc.CreateTextNode(xmldom.DOMString(fmt.Sprintf("text%d", i)))
		elem.AppendChild(text)
		parent.AppendChild(elem)
		parent = elem
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = root.TextContent()
	}
}

// BenchmarkTextContent benchmarks TextContent setter
func BenchmarkTextContentSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		doc := createWideDOM(b, 100)
		root := doc.DocumentElement()
		b.StartTimer()

		root.SetTextContent("New content")
	}
}

// BenchmarkAdoptNode benchmarks AdoptNode
func BenchmarkAdoptNode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		doc1 := createTestDocument()
		doc2 := createTestDocument()
		elem, _ := doc1.CreateElement("adopted")
		doc1.DocumentElement().AppendChild(elem)
		b.StartTimer()

		doc2.AdoptNode(elem)
	}
}

// BenchmarkRenameNode benchmarks RenameNode
func BenchmarkRenameNode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		doc := createTestDocument()
		elem, _ := doc.CreateElement("oldName")
		doc.DocumentElement().AppendChild(elem)
		b.StartTimer()

		doc.RenameNode(elem, "http://example.com", "newName")
	}
}

// BenchmarkCharacterDataManipulation benchmarks CharacterData manipulation methods
func BenchmarkCharacterDataRemove(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		doc := createTestDocument()
		root := doc.DocumentElement()
		text := doc.CreateTextNode("text")
		root.AppendChild(text)
		b.StartTimer()

		text.Remove()
	}
}

func BenchmarkCharacterDataReplaceWith(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		doc := createTestDocument()
		root := doc.DocumentElement()
		text := doc.CreateTextNode("text")
		root.AppendChild(text)
		b.StartTimer()

		elem, _ := doc.CreateElement("replacement")
		text.ReplaceWith(elem)
	}
}

// BenchmarkDomListGeneric benchmarks the generic domList implementation
func BenchmarkNodeListIteration(b *testing.B) {
	doc := createWideDOM(b, 1000)
	nodeList := doc.DocumentElement().ChildNodes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := uint(0); j < nodeList.Length(); j++ {
			_ = nodeList.Item(j)
		}
	}
}

func BenchmarkElementListIteration(b *testing.B) {
	doc := createWideDOM(b, 1000)
	// Add text nodes to make it mixed
	root := doc.DocumentElement()
	for i := 0; i < 500; i++ {
		text := doc.CreateTextNode("text")
		root.AppendChild(text)
	}

	elementList := root.Children()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := uint(0); j < elementList.Length(); j++ {
			_ = elementList.Item(j)
		}
	}
}

// BenchmarkIsEqualNode benchmarks IsEqualNode comparison
func BenchmarkIsEqualNode(b *testing.B) {
	doc := createTestDocument()
	elem1, _ := doc.CreateElement("elem")
	elem1.SetAttribute("attr1", "value1")
	elem1.SetAttribute("attr2", "value2")
	text1 := doc.CreateTextNode("content")
	elem1.AppendChild(text1)

	elem2, _ := doc.CreateElement("elem")
	elem2.SetAttribute("attr1", "value1")
	elem2.SetAttribute("attr2", "value2")
	text2 := doc.CreateTextNode("content")
	elem2.AppendChild(text2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = elem1.IsEqualNode(elem2)
	}
}

// BenchmarkGetRootNode benchmarks GetRootNode
func BenchmarkGetRootNode(b *testing.B) {
	doc := createDeepDOM(b, 100)
	// Get the deepest node
	deepest := doc.DocumentElement()
	for deepest.FirstChild() != nil {
		deepest = deepest.FirstChild().(xmldom.Element)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = deepest.GetRootNode()
	}
}
