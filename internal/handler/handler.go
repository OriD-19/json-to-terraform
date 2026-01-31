package handler

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/json-to-terraform/parser/internal/registry"
)

// refTraversal builds hcl.Traversal for a resource address and attribute (e.g. aws_vpc.node_3.id).
func refTraversal(addr, attr string) hcl.Traversal {
	var t hcl.Traversal
	idx := 0
	for i := 0; i < len(addr); i++ {
		if addr[i] == '.' {
			part := addr[idx:i]
			if len(t) == 0 {
				t = append(t, &hcl.TraverseRoot{Name: part})
			} else {
				t = append(t, &hcl.TraverseAttr{Name: part})
			}
			idx = i + 1
		}
	}
	if idx < len(addr) {
		part := addr[idx:]
		if len(t) == 0 {
			t = append(t, &hcl.TraverseRoot{Name: part})
		} else {
			t = append(t, &hcl.TraverseAttr{Name: part})
		}
	}
	if attr != "" {
		t = append(t, &hcl.TraverseAttr{Name: attr})
	}
	return t
}

// RefMap is an alias for registry.RefMap so handlers can use refs without importing registry in every signature.
// The actual type and interface live in registry to avoid import cycles.
type RefMap = registry.RefMap
