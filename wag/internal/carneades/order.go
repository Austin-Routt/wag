package carneades

import (
	"sort"
	"strconv"
	"strings"
)

// SortArgumentIDs sorts argument IDs in their natural order: "a2" before
// "a10", numeric IDs of the form aN first, any other ID after them in
// lexical order. Go randomises map iteration, so every list of IDs taken from
// Graph.Arguments must be sorted before its order can matter (defect #3 of the
// 2026-10-06 review: seeded games were not reproducible).
func SortArgumentIDs(ids []string) {
	sort.SliceStable(ids, func(i, j int) bool { return argIDLess(ids[i], ids[j]) })
}

func argIDLess(x, y string) bool {
	nx, okx := argNumber(x)
	ny, oky := argNumber(y)
	switch {
	case okx && oky:
		if nx != ny {
			return nx < ny
		}
		return x < y
	case okx != oky:
		return okx
	default:
		return x < y
	}
}

func argNumber(id string) (int, bool) {
	if !strings.HasPrefix(id, "a") || len(id) < 2 {
		return 0, false
	}
	n, err := strconv.Atoi(id[1:])
	return n, err == nil
}

// SortedKeys returns the keys of a string-keyed map in lexical order.
func SortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
