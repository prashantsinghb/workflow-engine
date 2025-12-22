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
	Inputs   map[string]interface{}
	When     *Condition
	Retry    *RetryPolicy
}

type Condition struct {
	FromNode  string `json:"from_node" yaml:"from_node"`
	Key       string `json:"key" yaml:"key"`
	Equals    any    `json:"equals" yaml:"equals,omitempty"`
	NotEquals any    `json:"not_equals" yaml:"not_equals,omitempty"`
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
			Inputs:   n.Inputs,
		}
	}

	for id, n := range def.Nodes {
		node := g.Nodes[NodeID(id)]
		// Support both depends_on and depends
		deps := n.DependsOn
		if len(n.Depends) > 0 {
			deps = n.Depends
		}
		for _, dep := range deps {
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
