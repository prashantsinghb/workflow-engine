package dag

import (
	"context"

	"github.com/prashantsinghb/workflow-engine/pkg/workflow/api"
)

type NodeID string

type Node struct {
	ID       NodeID
	Depends  []NodeID
	Executor string
}

type Executor interface {
	Execute(ctx context.Context, node *Node, inputs map[string]interface{}) (map[string]interface{}, error)
}

type Graph struct {
	Nodes map[NodeID]*Node
}

func Build(def *api.Definition) *Graph {
	g := &Graph{
		Nodes: make(map[NodeID]*Node),
	}

	for id, n := range def.Nodes {
		g.Nodes[NodeID(id)] = &Node{
			ID:       NodeID(id),
			Executor: n.Uses,
		}
	}

	for id, n := range def.Nodes {
		node := g.Nodes[NodeID(id)]
		for _, dep := range n.DependsOn {
			node.Depends = append(node.Depends, NodeID(dep))
		}
	}

	return g
}

func (g *Graph) NodeIDs() []NodeID {
	ids := make([]NodeID, 0, len(g.Nodes))
	for id := range g.Nodes {
		ids = append(ids, id)
	}
	return ids
}
