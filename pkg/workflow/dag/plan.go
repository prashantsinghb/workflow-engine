package dag

func TopologicalSort(g Graph) []NodeID {
	visited := map[NodeID]bool{}
	order := []NodeID{}

	var visit func(NodeID)
	visit = func(n NodeID) {
		if visited[n] {
			return
		}
		visited[n] = true
		for _, d := range g.Nodes[n] {
			visit(d)
		}
		order = append(order, n)
	}

	for n := range g.Nodes {
		visit(n)
	}

	return order
}
