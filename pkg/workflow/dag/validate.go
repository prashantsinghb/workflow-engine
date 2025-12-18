package dag

import "fmt"

func Validate(g Graph) error {
	visited := map[NodeID]bool{}
	stack := map[NodeID]bool{}

	var visit func(NodeID) error
	visit = func(n NodeID) error {
		if stack[n] {
			return fmt.Errorf("cycle detected: %s", n)
		}
		if visited[n] {
			return nil
		}
		visited[n] = true
		stack[n] = true
		for _, dep := range g.Nodes[n] {
			if _, ok := g.Nodes[dep]; !ok {
				return fmt.Errorf("node %s depends on unknown node %s", n, dep)
			}
			if err := visit(dep); err != nil {
				return err
			}
		}
		stack[n] = false
		return nil
	}

	for n := range g.Nodes {
		if err := visit(n); err != nil {
			return err
		}
	}
	return nil
}
