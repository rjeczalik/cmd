package memfs

// Compare returns true when the structure of the lhs and rhs is the same.
// It does not compare the value of the Files between the trees. If both trees
// are empty it returns true.
func Compare(lhs, rhs *FS) bool {
	type node struct{ lhs, rhs Directory }
	var (
		glob = []node{{lhs: lhs.Tree, rhs: rhs.Tree}}
		nod  node
	)
	for len(glob) > 0 {
		nod, glob = glob[len(glob)-1], glob[:len(glob)-1]
		if len(nod.lhs) != len(nod.rhs) {
			return false
		}
		for k, lv := range nod.lhs {
			rv, ok := nod.rhs[k]
			if !ok {
				return false
			}
			switch l := lv.(type) {
			case File:
				if _, ok := rv.(File); !ok {
					return false
				}
			case Directory:
				r, ok := rv.(Directory)
				if !ok {
					return false
				}
				glob = append(glob, node{lhs: l, rhs: r})
			default:
				return false
			}
		}
	}
	return true
}
