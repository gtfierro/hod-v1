package db

import (
	"container/list"
	"fmt"
	"github.com/gtfierro/hod/query"
	"reflect"
	"strings"
)

// struct to hold the graph of the query plan
type dependencyGraph struct {
	selectVars []string
	roots      []*queryTerm
	// map of variable name -> resolved?
	variables map[string]bool
}

// initializes the query plan struct
func makeDependencyGraph(q query.Query) *dependencyGraph {
	dg := &dependencyGraph{
		selectVars: []string{},
		roots:      []*queryTerm{},
		variables:  make(map[string]bool),
	}
	for _, v := range q.Select.Variables {
		dg.selectVars = append(dg.selectVars, v.Var.String())
	}
	return dg
}

func (dg *dependencyGraph) iter() chan *queryTerm {
	iter := make(chan *queryTerm)

	go func() {
		stack := list.New()
		//for _, r := range dg.roots {
		//	stack.PushFront(r)
		//}
		stack.PushFront(dg.roots[0])
		for stack.Len() > 0 {
			node := stack.Remove(stack.Front()).(*queryTerm)
			iter <- node
			log.Debug("send", node)
			for _, c := range node.children {
				stack.PushFront(c)
			}
		}
		close(iter)
	}()

	return iter
}

// returns true of the query plan or any of its children
// already includes the given query term
func (dg *dependencyGraph) hasChild(qt *queryTerm) bool {
	for _, r := range dg.roots {
		if r.equals(qt) {
			return true
		}
		if r.hasChild(qt) {
			return true
		}
	}
	return false
}

// adds the query term to the root set if it is
// not already there
func (dg *dependencyGraph) addRootTerm(qt *queryTerm) {
	if !dg.hasChild(qt) {
		// loop through and append to a node if we share a variable with it
		for _, root := range dg.roots {
			if root.bubbleDownDepends(qt) {
				return
			}
		}
		// otherwise, add it to the roots
		dg.roots = append(dg.roots, qt)
	}
}

func (dg *dependencyGraph) dump() {
	for _, r := range dg.roots {
		r.dump(0)
	}
}

// Firstly, if qt is already in the plan, we return
// iterate through in a breadth first search for any node
// [qt] shares a variable with. We attach qt as a child of that
// term
// Returns true if the node was added
func (dg *dependencyGraph) addChild(qt *queryTerm) bool {
	if dg.hasChild(qt) {
		fmt.Println("dg already has", qt.String())
		return false
	}
	stack := list.New()
	// push the roots onto the stack
	//for _, r := range dg.roots {
	//	stack.PushFront(r)
	//}
	stack.PushFront(dg.roots[0])
addchildloop:
	for stack.Len() > 0 {
		node := stack.Remove(stack.Front()).(*queryTerm)
		// if depends on, attach and return
		if qt.dependsOn(node) {
			//fmt.Println("node", qt.String(), "depends on", node.String())
			for _, child := range node.children {
				if qt.dependsOn(child) {
					stack.PushFront(child)
					continue addchildloop
				}
			}
			node.children = append(node.children, qt)
			return true
		}
		// add node children to back of stack
		for _, c := range node.children {
			stack.PushFront(c)
		}
	}
	dg.roots[0].children = append(dg.roots[0].children, qt)
	return true
}

// stores the state/variables for a particular triple
// from a SPARQL query
type queryTerm struct {
	query.Filter
	children  []*queryTerm
	variables []string
}

// initializes a queryTerm from a given Filter
func (dg *dependencyGraph) makeQueryTerm(f query.Filter) *queryTerm {
	qt := &queryTerm{
		f,
		[]*queryTerm{},
		[]string{},
	}
	// TODO: handle the predicates
	if qt.Subject.IsVariable() {
		dg.variables[qt.Subject.String()] = false
		qt.variables = append(qt.variables, qt.Subject.String())
	}
	if qt.Object.IsVariable() {
		dg.variables[qt.Object.String()] = false
		qt.variables = append(qt.variables, qt.Object.String())
	}
	return qt
}

// returns true if the term or any of its children has
// the given child
func (qt *queryTerm) hasChild(child *queryTerm) bool {
	for _, c := range qt.children {
		if c.equals(child) {
			return true
		}
		if c.hasChild(child) {
			return true
		}
	}
	return false
}

// returns true if two query terms are equal
func (qt *queryTerm) equals(qt2 *queryTerm) bool {
	return qt.Subject == qt2.Subject &&
		qt.Object == qt2.Object &&
		reflect.DeepEqual(qt.Path, qt2.Path)
}

func (qt *queryTerm) String() string {
	return fmt.Sprintf("<%s %s %s>", qt.Subject, qt.Path, qt.Object)
}

func (qt *queryTerm) dump(indent int) {
	fmt.Println(strings.Repeat("  ", indent), qt.String())
	for _, c := range qt.children {
		c.dump(indent + 1)
	}
}

func (qt *queryTerm) dependsOn(other *queryTerm) bool {
	for _, v := range qt.variables {
		for _, vv := range other.variables {
			if vv == v {
				return true
			}
		}
	}
	return false
}

// adds "other" as a child of the furthest node down the children tree
// that the node depends on. Returns true if "other" was added to the tree,
// and false otherwise
func (qt *queryTerm) bubbleDownDepends(other *queryTerm) bool {
	if !other.dependsOn(qt) {
		return false
	}
	for _, child := range qt.children {
		if other.bubbleDownDepends(child) {
			return true
		}
	}
	qt.children = append(qt.children, other)
	return true
}

// removes all terms in the removeList from removeFrom and returns
// the result
func filterTermList(removeFrom, removeList []*queryTerm) []*queryTerm {
	var ret = []*queryTerm{}
	for _, a := range removeFrom {
		keep := true
		for _, b := range removeList {
			if a.equals(b) {
				keep = false
				break
			}
		}
		if keep {
			ret = append(ret, a)
		}
	}
	return ret
}
