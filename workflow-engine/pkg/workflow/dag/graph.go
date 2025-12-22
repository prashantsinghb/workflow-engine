package dag

import (
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/api"
)

type NodeID string

type Node struct {
	ID       NodeID
	Depends  []NodeID
	Children []NodeID
	Executor string
	Uses     string
	With     map[string]interface{}
	Retry    *RetryPolicy
}

type RetryPolicy struct {
	MaxAttempts int
}

func (r *RetryPolicy) MaxAttemptsOrDefault() int {
	if r == nil || r.MaxAttempts <= 0 {
		return 1
	}
	return r.MaxAttempts
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
			Depends:  []NodeID{},
			Children: []NodeID{},
			Uses:     n.Uses,
			With:     n.With,
		}
	}

	for id, n := range def.Nodes {
		node := g.Nodes[NodeID(id)]
		for _, dep := range n.DependsOn {
			depID := NodeID(dep)
			node.Depends = append(node.Depends, depID)

			if parent, ok := g.Nodes[depID]; ok {
				parent.Children = append(parent.Children, NodeID(id))
			}
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
