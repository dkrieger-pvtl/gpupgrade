package checklist

import (
	"errors"
)

type Tree struct {
	Root *Node
}

type Action func(val int) error

type Node struct {
	done     bool
	children []*Node
	execute  Action
}

func unset_func(val int) error {
	return errors.New("node action unset")
}

func NewTree() Tree {
	var n = NewNode()
	var t = Tree{n}
	return t
}

func NewNode() *Node {
	var n = Node{done: false, children: nil, execute: unset_func}
	n.children = make([]*Node, 0)
	return &n
}
func (n *Node) SetDone(val bool) {
	n.done = val
}
func (n *Node) AddNode(newN *Node) {
	n.children = append(n.children, newN)
}
func (n *Node) SetAction(a Action) {
	n.execute = a
}
func (n *Node) ExecuteAll() error {
	for _, n := range n.children {
		err := n.execute(0)
		if err != nil {
			return err
		}
	}
	return nil
}
