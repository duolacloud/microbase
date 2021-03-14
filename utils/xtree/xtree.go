package xtree

import (
	"context"
)

type XTree struct {
}

func NewXTree() *XTree {
	return &XTree{}
}

type NodeVisitor interface {
	PreVisit(c context.Context, current *Node) error
	PostVisit(c context.Context, current *Node) error
}

type Node struct {
	Visitor  NodeVisitor
	Children []*Node
	Parent   *Node
}

func (n *Node) PreVisit(c context.Context) error {
	if n.Visitor != nil {
		err := n.Visitor.PreVisit(c, n)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) PostVisit(c context.Context) error {
	if n.Visitor != nil {
		err := n.Visitor.PostVisit(c, n)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *XTree) Travel(c context.Context, node *Node) error {
	return p.dfs(c, node)
}

func (p *XTree) dfs(c context.Context, node *Node) error {
	err := node.PreVisit(c)
	if err != nil {
		return err
	}

	for _, child := range node.Children {
		if err := p.dfs(c, child); err != nil {
			return err
		}
	}

	return node.PostVisit(c)
}
