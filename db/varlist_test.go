package db

import (
	"fmt"
	"testing"
)

func testConsistency(list *varlist, elements []string, t *testing.T) bool {
	if len(list.list) != len(elements) {
		t.Errorf("List wrong length. Should be %d but was %d (%s)", len(elements), len(list.list), list.dumplist())
		return false
	}
	if len(list.indexes) != len(elements) {
		t.Errorf("Indexes wrong length. Should be %d but was %d", len(elements), len(list.indexes))
		return false
	}
	if len(elements) > 1 {
		for i := 0; i < len(elements)-1; i++ {
			_cur := elements[i]
			_next := elements[i+1]
			cur := list.list[i]
			next := list.list[i+1]
			if cur.value != _cur {
				t.Errorf("wrong value for %v (%s) (%s)", cur, _cur, list.dumplist())
				return false
			}
			if next.value != _next {
				t.Errorf("wrong value for %v (%s)", next, _next)
				return false
			}
			if cur.next == nil || cur.next.value != _next {
				t.Errorf("Broken link at element %+v (next is %+v)", cur, cur.next)
				return false
			}
			if next.prev == nil || next.prev.value != _cur {
				t.Errorf("Broken link at element %+v (prev is %+v)", next, next.prev)
				return false
			}
			if list.indexes[_cur] != i {
				t.Errorf("Wrong index for %s. Should be %d but was %d (%s)", _cur, i, list.indexes[_cur], list.dumplist())
				return false
			}
			if list.indexes[_next] != i+1 {
				t.Errorf("Wrong index for %s. Should be %d but was %d (%s)", _next, i+1, list.indexes[_next], list.dumplist())
				return false
			}
		}
	}
	return true
}

func TestVarlistPushBack(t *testing.T) {
	for _, test := range []struct {
		elements []string
	}{
		{elements: []string{"a"}},
		{elements: []string{"a", "b"}},
		{elements: []string{"a", "b", "c"}},
	} {
		list := newvarlist()
		for _, elem := range test.elements {
			list.pushBack(elem)
		}
		if !testConsistency(list, test.elements, t) {
			fmt.Println(test.elements)
			return
		}
	}
}

func TestVarlistInsertAfter(t *testing.T) {
	for _, test := range []struct {
		elements    []string
		insertafter [][]string
		results     []string
	}{
		{
			elements:    []string{"a"},
			insertafter: [][]string{[]string{"b", "a"}},
			results:     []string{"a", "b"},
		},
		{
			elements:    []string{"a", "b"},
			insertafter: [][]string{[]string{"c", "a"}},
			results:     []string{"a", "c", "b"},
		},
		{
			elements:    []string{"a", "b"},
			insertafter: [][]string{[]string{"c", "b"}},
			results:     []string{"a", "b", "c"},
		},
		{
			elements:    []string{"a", "b", "c", "d"},
			insertafter: [][]string{[]string{"e", "d"}},
			results:     []string{"a", "b", "c", "d", "e"},
		},
		{
			elements:    []string{"a", "b", "c", "d"},
			insertafter: [][]string{[]string{"e", "c"}},
			results:     []string{"a", "b", "c", "e", "d"},
		},
		{
			elements:    []string{"a", "b", "c", "d"},
			insertafter: [][]string{[]string{"e", "a"}},
			results:     []string{"a", "e", "b", "c", "d"},
		},
	} {
		list := newvarlist()
		for _, elem := range test.elements {
			list.pushBack(elem)
		}
		for _, pair := range test.insertafter {
			list.insertAfter(pair[0], pair[1])
		}
		if !testConsistency(list, test.results, t) {
			fmt.Println(test)
			return
		}
	}
}

func TestVarlistMoveAfter(t *testing.T) {
	for _, test := range []struct {
		elements  []string
		moveafter [][]string
		results   []string
	}{
		{
			elements:  []string{"a", "b"},
			moveafter: [][]string{[]string{"b", "a"}},
			results:   []string{"a", "b"},
		},
		{
			elements:  []string{"a", "b", "c"},
			moveafter: [][]string{[]string{"c", "a"}},
			results:   []string{"a", "c", "b"},
		},
		{
			elements:  []string{"a", "b", "c"},
			moveafter: [][]string{[]string{"b", "a"}},
			results:   []string{"a", "b", "c"},
		},
		{
			elements:  []string{"a", "b", "c", "d"},
			moveafter: [][]string{[]string{"c", "a"}},
			results:   []string{"a", "c", "d", "b"},
		},
	} {
		list := newvarlist()
		for _, elem := range test.elements {
			list.pushBack(elem)
		}
		for _, pair := range test.moveafter {
			list.moveAfter(pair[0], pair[1])
		}
		if !testConsistency(list, test.results, t) {
			fmt.Println(test)
			return
		}
	}
}
