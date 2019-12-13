package checklist

import (
	"fmt"
	"strconv"
	"testing"
)

func TestNewNode(t *testing.T) {
	node := NewNode()
	if node.done != false {
		t.Errorf("new node alrady done")
	}
	if len(node.children) != 0 {
		t.Errorf("new node has children")
	}
}

func TestNode_SetName(t *testing.T) {
	node := NewNode()
	node.SetName("test")
	if node.name != "test" {
		t.Errorf("SetName did not set name")
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

func TestNode_WalkInOrder(t *testing.T) {
	node := NewNode()
	node.SetAction(func(a int) error { return nil })

	for i := 1; i <= 4; i++ {
		ni := NewNode()
		node.AddNode(ni)
		ni.SetName("-")
		for j := 1; j <= 3; j++ {
			nj := NewNode()
			ni.AddNode(nj)
			nj.SetName("--")
			for k := 1; k <= 2; k++ {
				nk := NewNode()
				nj.AddNode(nk)
				id := strconv.Itoa(i) + "-" + strconv.Itoa(j) + "-" + strconv.Itoa(k)
				nk.SetName(id)
				fmt.Println("INDEX:", id)
			}
		}
	}

	node.WalkInOrder()

}
