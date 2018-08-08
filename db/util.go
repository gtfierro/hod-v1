package db

import (
	sparql "github.com/gtfierro/hod/lang/ast"
	"github.com/zhangxinngang/murmur"
)

func reversePath(path []sparql.PathPattern) []sparql.PathPattern {
	newpath := make([]sparql.PathPattern, len(path))
	// for in-place, replace newpath with path
	if len(newpath) == 1 {
		return path
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		newpath[i], newpath[j] = path[j], path[i]
	}
	return newpath
}

func hashRowWithPos(row *relationRow, positions []int) uint32 {
	var b []byte
	for _, pos := range positions {
		b = append(b, row.content[pos*8:pos*8+8]...)
	}
	return murmur.Murmur3(b)
}

func compareResultMapList(rml1, rml2 []ResultMap) bool {
	var (
		found bool
	)

	if len(rml1) != len(rml2) {
		return false
	}

	for _, val1 := range rml1 {
		found = false
		for _, val2 := range rml2 {
			if compareResultMap(val1, val2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func compareResultMap(rm1, rm2 ResultMap) bool {
	if len(rm1) != len(rm2) {
		return false
	}
	for k, v := range rm1 {
		if v2, found := rm2[k]; !found {
			return false
		} else if v2 != v {
			return false
		}
	}
	return true
}
