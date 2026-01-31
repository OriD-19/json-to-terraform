package dependency

import (
	"errors"

	"github.com/json-to-terraform/parser/internal/diagram"
)

// ErrCycle is returned when the dependency graph contains a cycle.
var ErrCycle = errors.New("dependency cycle detected")

// Resolve builds the dependency graph from edges and returns:
// - ordered: node IDs in topological order (dependencies first)
// - tiers: node IDs grouped by depth (tier 0 = no deps, tier 1 = depend only on tier 0, etc.)
func Resolve(d *diagram.Diagram) (ordered []string, tiers [][]string, err error) {
	if d == nil || len(d.Nodes) == 0 {
		return nil, nil, nil
	}

	nodeSet := make(map[string]bool)
	for i := range d.Nodes {
		nodeSet[d.Nodes[i].ID] = true
	}

	// target depends on source => inDegree[target] = number of edges into target
	inDegree := make(map[string]int)
	for id := range nodeSet {
		inDegree[id] = 0
	}
	for _, e := range d.Edges {
		if !nodeSet[e.Source] || !nodeSet[e.Target] || e.Source == e.Target {
			continue
		}
		inDegree[e.Target]++
	}

	var queue []string
	for id := range nodeSet {
		if inDegree[id] == 0 {
			queue = append(queue, id)
		}
	}

	ordered = make([]string, 0, len(d.Nodes))
	tiers = nil
	for len(queue) > 0 {
		tier := make([]string, len(queue))
		copy(tier, queue)
		tiers = append(tiers, tier)
		var nextQueue []string
		for _, u := range queue {
			ordered = append(ordered, u)
			for _, e := range d.Edges {
				if e.Source != u {
					continue
				}
				v := e.Target
				if !nodeSet[v] {
					continue
				}
				inDegree[v]--
				if inDegree[v] == 0 {
					nextQueue = append(nextQueue, v)
				}
			}
		}
		queue = nextQueue
	}

	if len(ordered) != len(nodeSet) {
		return nil, nil, ErrCycle
	}
	return ordered, tiers, nil
}
