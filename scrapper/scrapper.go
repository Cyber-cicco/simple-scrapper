package scrapper

import (
	"bytes"
	"errors"

	"github.com/Cyber-cicco/tree-sitter-query-builder/querier"
	sitter "github.com/smacker/go-tree-sitter"
)

type SelectorType string

const (
	ST_ID    = "ST_ID"
	ST_CLASS = "ST_CLASS"
	ST_BASE  = "ST_BASE"
)

type DOMElement struct {
	Node     *sitter.Node
	document *DOMStructure
}

type DOMStructure struct {
	RootNode *sitter.Node
	content  []byte
}

type Selectable interface {
	QuerySelector(selector string) (*DOMElement, bool)
	QuerySelectorAll(selector string) ([]*DOMElement, bool)
}

func ToDOM(n *sitter.Node, content []byte) (*DOMStructure, error) {
	if n.Type() == "document" || n.Type() == "element" {
		return &DOMStructure{
			RootNode: n,
			content:  content,
		}, nil
	}
	return nil, errors.New("node is not an HTML element")
}

func (s *DOMStructure) QuerySelector(query string) (*DOMElement, bool) {
	return querySelector(query, s.RootNode, s.content, s)
}

func (s *DOMElement) QuerySelector(query string) (*DOMElement, bool) {
	return querySelector(query, s.Node, s.document.content, s.document)
}

func (s *DOMStructure) QuerySelectorAll(query string) ([]*DOMElement, bool) {
	return querySelectorAll(query, s.RootNode, s.content, s)
}

func (s *DOMElement) QuerySelectorAll(query string) ([]*DOMElement, bool) {
	return querySelectorAll(query, s.Node, s.document.content, s.document)
}

func querySelectorAll(query string, rootNode *sitter.Node, content []byte, s *DOMStructure) ([]*DOMElement, bool) {

    nodes := []*sitter.Node{}

	if len(query) == 0 {
		return nil, false
	}

	selector, err := parseSelector(query)

	if err != nil {
		return nil, false
	}

	switch selector.sType {

	case ST_BASE:
		nodes = querier.GetChildrenMatching(rootNode, func(n *sitter.Node) bool {
			isEl := n.Type() == "element"

			if !isEl {
				return false
			}

			return getTagName(n, content) == selector.matched
		}, nodes)

	case ST_ID:
		nodes = querier.GetChildrenMatching(rootNode, func(n *sitter.Node) bool {
			return elementWithAttributeEquals(n, "id", selector.matched, content)
		}, nodes)

	case ST_CLASS:
		nodes = querier.GetChildrenMatching(rootNode, func(n *sitter.Node) bool {
			return elementWithAttributeEquals(n, "class", selector.matched, content)
		}, nodes)
	}

    elements := make([]*DOMElement, len(nodes)) 
    
    for i, node := range nodes {

        if selector.sType == ST_CLASS || selector.sType == ST_ID {
            elements[i] = &DOMElement{
            	Node:     node.Parent(),
            	document: s,
            }
        } else {
            elements[i] = &DOMElement{
            	Node:     node,
            	document: s,
            }
        }
    }

    return elements, true
}

func querySelector(query string, rootNode *sitter.Node, content []byte, s *DOMStructure) (*DOMElement, bool) {

	var element *sitter.Node

	if len(query) == 0 {
		return nil, false
	}

	selector, err := parseSelector(query)

	if err != nil {
		return nil, false
	}

	switch selector.sType {

	case ST_BASE:
		element = querier.GetFirstMatch(rootNode, func(n *sitter.Node) bool {
			isEl := n.Type() == "element"

			if !isEl {
				return false
			}

			return getTagName(n, content) == selector.matched
		})

	case ST_ID:
		element = querier.GetFirstMatch(rootNode, func(n *sitter.Node) bool {
			return elementWithAttributeEquals(n, "id", selector.matched, content)
		})
        if element != nil && element.Parent() != nil {
            element = element.Parent()
        }

	case ST_CLASS:
		element = querier.GetFirstMatch(rootNode, func(n *sitter.Node) bool {
			return elementWithAttributeEquals(n, "class", selector.matched, content)
		})
        if element != nil && element.Parent() != nil {
            element = element.Parent()
        }
	}

	return &DOMElement{
		Node:     element,
		document: s,
	}, element != nil

}

func elementWithAttributeEquals(n *sitter.Node, attributeName, matched string, content []byte) bool {

	isEl := n.Type() == "start_tag"

	if !isEl {
		return false
	}

	el := querier.GetFirstMatch(n, func(n *sitter.Node) bool {
		return attributeEquals(n, attributeName, matched, content)
	})

	return el != nil
}

func attributeEquals(n *sitter.Node, attributeName, matched string, content []byte) bool {

	isSeachedAttribute := n.Type() == "attribute" && n.Child(0) != nil && n.Child(0).Content(content) == attributeName

	if !isSeachedAttribute {
		return false
	}

	if n.Child(2) == nil || n.Child(2).Child(1) == nil {
		return false
	}

	return n.Child(2).Child(1).Content(content) == matched
}

type selector struct {
	matched string
	sType   SelectorType
}

func parseSelector(query string) (selector, error) {
	switch query[0] {
	case '.':

		if len(query) < 2 {
			return selector{}, errors.New("Erreur dans la syntaxe du sélecteur")
		}

		return selector{
			matched: query[1:],
			sType:   ST_CLASS,
		}, nil

	case '#':

		if len(query) < 2 {
			return selector{}, errors.New("Erreur dans la syntaxe du sélecteur")
		}

		return selector{
			matched: query[1:],
			sType:   ST_ID,
		}, nil

	default:

		return selector{
			matched: query,
			sType:   ST_BASE,
		}, nil
	}
}

func (s *DOMElement) InnerText() []byte {

	var buffer bytes.Buffer
	nodes := []*sitter.Node{}
	nodes = querier.GetChildrenMatching(s.Node, func(n *sitter.Node) bool {
		return n.Type() == "text" || n.Type() == "entity"
	}, nodes)

	for _, match := range nodes {

		if match.Type() == "text" {
            toWrite := []byte(match.Content(s.document.content))
            toWrite = trimLargeWhitespacesAndDeleteCR(toWrite)
            if match.StartByte() > 0 && s.document.content[match.StartByte()-1] == ' ' {
                toWrite = append([]byte{' '}, toWrite...)
            }
            if match.EndByte() < uint32(len(s.document.content)) && s.document.content[match.EndByte()] == ' ' {
                toWrite = append(toWrite, ' ')
            }
			buffer.Write(toWrite)
		}

		if match.Type() == "entity" {
			char, ok := specialChars[match.Content(s.document.content)]
			if !ok {
				continue
			}
			buffer.Write([]byte{char})
		}
	}

    return buffer.Bytes()
}

func trimLargeWhitespacesAndDeleteCR(in []byte) []byte {
    out := []byte{}
    lastWasWS := false
    for i := 0; i < len(in); i++ {
        if (lastWasWS && in[i] == ' ') || in[i] == '\n' || in[i] == '\r' {
            continue
        }
        lastWasWS = in[i] == ' '
        out = append(out,in[i])
    }
    return out
}

func (s *DOMElement) ToString() string {
	return s.Node.Content(s.document.content)
}

func (s *DOMElement) TagName() string {
	return getTagName(s.Node, s.document.content)
}

func getTagName(n *sitter.Node, content []byte) string {
	return n.Child(0).Child(1).Content(content)
}
