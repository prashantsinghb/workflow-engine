package dag

func Ready(g Graph, completed map[NodeID]bool) []NodeID {
	ready := []NodeID{}

	for id, node := range g.Nodes {
		if completed[id] {
			continue
		}

		ok := true
		for _, d := range node.Depends {
			if !completed[d] {
				ok = false
				break
			}
		}

		if ok {
			ready = append(ready, id)
		}
	}

	return ready
}
