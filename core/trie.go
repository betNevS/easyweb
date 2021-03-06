package core

import (
	"errors"
	"strings"
)

type Tree struct {
	root *node
}

type node struct {
	isLast   bool
	segment  string
	handlers []ControllerHandler
	childs   []*node
	parent   *node
}

func (n *node) filterChildNodes(segment string) []*node {
	if n.childs == nil {
		return nil
	}

	if isWildSegment(segment) {
		return n.childs
	}

	nodes := make([]*node, 0, len(n.childs))
	for _, cnode := range n.childs {
		if isWildSegment(cnode.segment) || cnode.segment == segment {
			nodes = append(nodes, cnode)
		}
	}

	return nodes
}

func (n *node) matchNode(uri string) *node {
	segments := strings.SplitN(uri, "/", 2)

	segment := segments[0]
	if !isWildSegment(segment) {
		segment = strings.ToUpper(segment)
	}

	cnodes := n.filterChildNodes(segment)
	if cnodes == nil {
		return nil
	}

	if len(segments) == 1 {
		for _, tn := range cnodes {
			if tn.isLast {
				return tn
			}
		}
		return nil
	}

	for _, tn := range cnodes {
		tnMatch := tn.matchNode(segments[1])
		if tnMatch != nil {
			return tnMatch
		}
	}
	return nil
}

func (n *node) findChildNode(segment string) *node {
	for _, node := range n.childs {
		if node.segment == segment {
			return node
		}
	}
	return nil
}

func (n *node) parseParamsFromEndNode(uri string) map[string]string {
	ret := make(map[string]string)
	segments := strings.Split(uri, "/")

	cur := n
	for i := len(segments) - 1; i >= 0; i-- {
		if cur.segment == "" {
			break
		}
		if isWildSegment(cur.segment) {
			ret[cur.segment[1:]] = segments[i]
		}
		cur = cur.parent
	}
	return ret
}

func NewTree() *Tree {
	return &Tree{
		root: &node{},
	}
}

func (t *Tree) AddRouter(uri string, handlers []ControllerHandler) error {
	n := t.root

	if n.matchNode(uri) != nil {
		return errors.New("route conflict: " + uri)
	}

	segments := strings.Split(uri, "/")

	for i, segment := range segments {
		if !isWildSegment(segment) {
			segment = strings.ToUpper(segment)
		}

		isLast := i == len(segments)-1

		cnode := n.findChildNode(segment)
		if cnode != nil {
			n = cnode
		} else {
			newNode := &node{segment: segment}
			newNode.parent = n
			if isLast {
				newNode.isLast = isLast
				newNode.handlers = handlers
			}
			n.childs = append(n.childs, newNode)
			n = newNode
		}
	}
	return nil
}

func (t *Tree) FindHandler(uri string) []ControllerHandler {
	matchNode := t.root.matchNode(uri)
	if matchNode == nil {
		return nil
	}
	return matchNode.handlers
}

func (t *Tree) FindNode(uri string) *node {
	matchNode := t.root.matchNode(uri)
	if matchNode == nil {
		return nil
	}

	return matchNode
}

func isWildSegment(segment string) bool {
	return strings.HasPrefix(segment, ":")
}
