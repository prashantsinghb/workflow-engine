package dag

func Ready(g Graph, completed map[NodeID]bool) []NodeID {
	ready := []NodeID{}

	for n, deps := range g.Nodes {
		if completed[n] {
			continue
		}

		ok := true
		for _, d := range deps {
			if !completed[d] {
				ok = false
				break
			}
		}

		if ok {
			ready = append(ready, n)
		}
	}

	return ready
}
