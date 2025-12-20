package dag

func Plan(g *Graph) []NodeID {
	visited := map[NodeID]bool{}
	order := []NodeID{}

	var visit func(NodeID)
	visit = func(id NodeID) {
		if visited[id] {
			return
		}
		visited[id] = true
		for _, dep := range g.Nodes[id].Depends {
			visit(dep)
		}
		order = append(order, id)
	}

	for id := range g.Nodes {
		visit(id)
	}

	return order
}
