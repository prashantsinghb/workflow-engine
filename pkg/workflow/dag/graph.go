package dag

import "github.com/prashantsinghb/workflow-engine/pkg/workflow"

type NodeID string

type Graph struct {
	Nodes map[NodeID][]NodeID
}

func Build(def *workflow.Definition) Graph {
	g := Graph{Nodes: map[NodeID][]NodeID{}}

	for id, node := range def.Nodes {
		deps := make([]NodeID, 0, len(node.DependsOn))
		for _, d := range node.DependsOn {
			deps = append(deps, NodeID(d))
		}
		g.Nodes[NodeID(id)] = deps
	}

	return g
}
