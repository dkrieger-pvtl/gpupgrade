package checklist

import (
	"testing"
)

func TestNewTree(t *testing.T) {
	tree := NewTree()
	err := tree.Root.execute(1)
	if err == nil {
		t.Errorf("new tree function does not produce error")
	}
}

func TestNewNode(t *testing.T) {
	node := NewNode()
	if node.done != false {
		t.Errorf("new node alrady done")
	}
	if len(node.children) != 0 {
		t.Errorf("new node has children")
	}
}

func TestNode_SetDone(t *testing.T) {
	node := NewNode()
	node.SetDone(true)
	if node.done != true {
		t.Errorf("SetDone did not set to true")
	}
	node.SetDone(false)
	if node.done != false {
		t.Errorf("SetDone did not set to false")
	}
}
func TestNode_AddNode(t *testing.T) {
	node := NewNode()
	node.AddNode(NewNode())
	if len(node.children) != 1 {
		t.Errorf("AddNode did not add new node")
	}
}
func TestNode_SetAction(t *testing.T) {
	node := NewNode()
	node.SetAction(func(a int) error { return nil })
}
func TestNode_ExecuteAll(t *testing.T) {
	node := NewNode()
	node.SetAction(func(a int) error { return nil })
	c1 := NewNode()
	c1.SetAction(func(a int) error { return nil })
	c2 := NewNode()
	c2.SetAction(func(a int) error { return nil })
	err := node.ExecuteAll()
	if err != nil {
		t.Errorf("ExecuteAll failed")
	}
}
