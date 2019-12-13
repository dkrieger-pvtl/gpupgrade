package checklist

import (
	"errors"
	"fmt"
)

type Action func(val int) error

type Node struct {
	name     string
	done     bool
	children []*Node
	execute  Action
}

func unset_func(val int) error {
	return errors.New("node action unset")
}

func NewNode() *Node {
	var n = Node{name: "---", done: false, children: nil, execute: unset_func}
	n.children = make([]*Node, 0)
	return &n
}
func (n *Node) String() string {
	return fmt.Sprintf("name: %s done: %v numChildren: %d",
		n.name, n.done, len(n.children))
}
func (n *Node) SetName(name string) {
	n.name = name
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

func (n *Node) WalkInOrder() {
	if len(n.children) == 0 {
		n.SetDone(true)
		fmt.Println(n)
		return
	}
	for _, c := range n.children {
		c.WalkInOrder()
	}
}
