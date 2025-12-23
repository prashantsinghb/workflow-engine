package dag

import (
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/api"
)

type NodeID string

type Node struct {
	ID         NodeID
	Depends    []NodeID
	Children   []NodeID
	Executor   string
	Uses       string
	With       map[string]interface{}
	Inputs     map[string]interface{}
	When       *Condition
	Retry      *RetryPolicy
	Compensate *Compensation
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

type Compensation struct {
	Uses string
	With map[string]interface{}
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
		node := &Node{
			ID:       NodeID(id),
			Depends:  []NodeID{},
			Children: []NodeID{},
			Uses:     n.Uses,
			With:     n.With,
			Inputs:   n.Inputs,
		}

		// Parse When condition if present
		if n.When != nil {
			when := &Condition{}
			if fromNode, ok := n.When["from_node"].(string); ok {
				when.FromNode = fromNode
			}
			if key, ok := n.When["key"].(string); ok {
				when.Key = key
			}
			if equals, ok := n.When["equals"]; ok {
				when.Equals = equals
			}
			if notEquals, ok := n.When["not_equals"]; ok {
				when.NotEquals = notEquals
			}
			if when.FromNode != "" && when.Key != "" {
				node.When = when
			}
		}

		// Parse Retry policy if present
		if n.Retry != nil {
			retry := &RetryPolicy{}
			if maxAttempts, ok := n.Retry["MaxAttempts"].(int); ok {
				retry.MaxAttempts = maxAttempts
			} else if maxAttempts, ok := n.Retry["max_attempts"].(int); ok {
				retry.MaxAttempts = maxAttempts
			} else if maxAttemptsFloat, ok := n.Retry["MaxAttempts"].(float64); ok {
				retry.MaxAttempts = int(maxAttemptsFloat)
			} else if maxAttemptsFloat, ok := n.Retry["max_attempts"].(float64); ok {
				retry.MaxAttempts = int(maxAttemptsFloat)
			}
			if retry.MaxAttempts > 0 {
				node.Retry = retry
			}
		}

		if n.Compensate != nil {
			comp := &Compensation{}
			if uses, ok := n.Compensate["uses"].(string); ok {
				comp.Uses = uses
			}
			if with, ok := n.Compensate["with"].(map[string]interface{}); ok {
				comp.With = with
			}
			if comp.Uses != "" {
				node.Compensate = comp
			}
		}

		g.Nodes[NodeID(id)] = node
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
