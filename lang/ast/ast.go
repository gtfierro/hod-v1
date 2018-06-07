package ast

import (
	"fmt"
	"github.com/gtfierro/hod/lang/token"
	"github.com/gtfierro/hod/turtle"
	"github.com/kr/pretty"
	"strings"
)

type QueryType uint

const (
	SELECT_QUERY QueryType = 1 << iota
	INSERT_QUERY
	DELETE_QUERY
)

var debug = false

func SetDebug() {
	debug = true
}
func ClearDebug() {
	debug = false
}

type Query struct {
	Select    SelectClause
	From      FromClause
	Count     bool
	Insert    InsertClause
	Where     WhereClause
	Variables []string
	Type      QueryType
}

func (q Query) Dump() {
	for _, triple := range q.Where.Terms {
		fmt.Println(triple.String())
	}
	fmt.Println("----")
}

func (q Query) IsInsert() bool {
	return (q.Type & INSERT_QUERY) == INSERT_QUERY
}

func (q Query) IsSelect() bool {
	return (q.Type & SELECT_QUERY) == SELECT_QUERY
}

func (q Query) CopyWithNewTerms(terms []Triple) Query {
	newq := Query{
		Select:    q.Select,
		From:      q.From,
		Variables: q.Variables,
		Type:      q.Type,
	}
	newq.Where.Terms = make([]Triple, len(terms))
	copy(newq.Where.Terms, terms)

	newq.Insert.Terms = make([]Triple, len(terms))
	copy(newq.Insert.Terms, terms)
	return newq
}

func (q Query) Copy() *Query {
	return &Query{
		Select:    q.Select,
		From:      q.From,
		Variables: q.Variables,
		Where:     q.Where,
		Insert:    q.Insert,
		Count:     q.Count,
		Type:      q.Type,
	}
}

func NewQuery(selectclause, whereclause interface{}, count bool) (Query, error) {
	if debug {
		fmt.Printf("%# v", pretty.Formatter(whereclause.(WhereClause)))
	}
	q := Query{
		Where:  whereclause.(WhereClause),
		Select: selectclause.(SelectClause),
		Count:  count,
		Type:   SELECT_QUERY,
	}
	if q.From.Empty() {
		q.From.AllDBs = true
	}
	q.PopulateVars()
	if q.Select.AllVars {
		q.Select.Vars = q.Variables
	}
	return q, nil
}

func NewInsertQuery(insertclause, whereclause interface{}, count bool) (Query, error) {
	if debug {
		fmt.Printf("%# v", pretty.Formatter(whereclause.(WhereClause)))
		fmt.Printf("%# v", pretty.Formatter(insertclause.(InsertClause)))
	}
	q := Query{
		Where:  whereclause.(WhereClause),
		Select: SelectClause{AllVars: true},
		Insert: insertclause.(InsertClause),
		Count:  count,
		Type:   INSERT_QUERY,
	}
	if q.From.Empty() {
		q.From.AllDBs = true
	}
	q.PopulateVars()
	if q.Select.AllVars {
		q.Select.Vars = q.Variables
	}
	return q, nil
}

func NewQueryMulti(selectclause, fromclause, whereclause interface{}, count bool) (Query, error) {
	if debug {
		fmt.Printf("%# v", pretty.Formatter(whereclause.(WhereClause)))
	}
	q := Query{
		Where:  whereclause.(WhereClause),
		From:   fromclause.(FromClause),
		Select: selectclause.(SelectClause),
		Count:  count,
	}
	if q.From.Empty() {
		q.From.AllDBs = true
	}
	q.PopulateVars()
	if q.Select.AllVars {
		q.Select.Vars = q.Variables
	}
	return q, nil
}

func NewInsertQueryMulti(insertclause, fromclause, whereclause interface{}, count bool) (Query, error) {
	if debug {
		fmt.Printf("%# v", pretty.Formatter(whereclause.(WhereClause)))
		fmt.Printf("%# v", pretty.Formatter(insertclause.(InsertClause)))
	}
	q := Query{
		Where:  whereclause.(WhereClause),
		Select: SelectClause{AllVars: true},
		From:   fromclause.(FromClause),
		Insert: insertclause.(InsertClause),
		Count:  count,
		Type:   INSERT_QUERY,
	}
	if q.From.Empty() {
		q.From.AllDBs = true
	}
	q.PopulateVars()
	if q.Select.AllVars {
		q.Select.Vars = q.Variables
	}
	return q, nil
}

func (q *Query) PopulateVars() {
	vars := make(map[string]int)
	// get all variables
	for _, triple := range q.Where.Terms {
		AddIfVar(triple.Subject, vars)
		AddIfVar(triple.Object, vars)
		for _, path := range triple.Predicates {
			AddIfVar(path.Predicate, vars)
		}
	}
	for _, triple := range q.Insert.Terms {
		AddIfVar(triple.Subject, vars)
		AddIfVar(triple.Object, vars)
		for _, path := range triple.Predicates {
			AddIfVar(path.Predicate, vars)
		}
	}
	if q.Where.GraphGroup != nil {
		VarsFromGroup(*q.Where.GraphGroup, vars)
	}
	q.Variables = []string{} // clear
	for varname := range vars {
		q.Variables = append(q.Variables, varname)
	}
}

func (q Query) IterTriples(f func(t Triple) Triple) {
	for idx, triple := range q.Where.Terms {
		q.Where.Terms[idx] = f(triple)
	}
	for idx, triple := range q.Insert.Terms {
		q.Insert.Terms[idx] = f(triple)
	}
	if q.Where.GraphGroup != nil {
		q.Where.GraphGroup.IterTriples(f)
	}
}

func AddIfVar(uri turtle.URI, m map[string]int) {
	if uri.IsVariable() {
		m[uri.String()] = 1
	}
}

func VarsFromGroup(group GraphGroup, m map[string]int) {
	for _, triple := range group.Terms {
		AddIfVar(triple.Subject, m)
		AddIfVar(triple.Object, m)
		for _, path := range triple.Predicates {
			AddIfVar(path.Predicate, m)
		}
	}
	for _, union := range group.Unions {
		VarsFromGroup(union, m)
	}
}

func (grp GraphGroup) Expand() [][]Triple {
	var terms = make([]Triple, len(grp.Terms))
	copy(terms, grp.Terms)
	var groups [][]Triple

	if len(grp.Unions) > 0 {
		for _, union := range grp.Unions {
			for _, subgroup := range union.Expand() {
				groups = append(groups, append(terms, subgroup...))
			}
		}
	} else {
		groups = append(groups, terms)
	}
	return groups
}

func (grp GraphGroup) Iter(f func(t turtle.URI)) {
	for _, triple := range grp.Terms {
		f(triple.Subject)
		f(triple.Object)
		for _, path := range triple.Predicates {
			f(path.Predicate)
		}
	}
	for _, union := range grp.Unions {
		union.Iter(f)
	}
}

func (grp *GraphGroup) IterTriples(f func(t Triple) Triple) {
	for idx, triple := range grp.Terms {
		grp.Terms[idx] = f(triple)
	}
	for _, union := range grp.Unions {
		union.IterTriples(f)
	}
}

type SelectClause struct {
	Vars    []string
	AllVars bool
}

func NewAllSelectClause() (SelectClause, error) {
	return SelectClause{AllVars: true}, nil
}

func NewSelectClause(varlist interface{}) (SelectClause, error) {
	return SelectClause{Vars: varlist.([]string)}, nil
}

type InsertClause struct {
	Terms []Triple
}

func NewInsertClause(triples interface{}) (InsertClause, error) {
	return InsertClause{
		Terms: triples.([]Triple),
	}, nil
}

type FromClause struct {
	Databases []string
	AllDBs    bool
}

func (f FromClause) String() string {
	if f.AllDBs {
		return "*"
	}
	return strings.Join(f.Databases, " ")
}

func NewAllFromClause() (FromClause, error) {
	return FromClause{AllDBs: true}, nil
}

func NewFromClause(dblist interface{}) (FromClause, error) {
	return FromClause{Databases: dblist.([]string)}, nil
}

func (from FromClause) Empty() bool {
	return len(from.Databases) == 0 && !from.AllDBs
}

type WhereClause struct {
	Terms      []Triple
	GraphGroup *GraphGroup
}

func NewWhereClause(triples interface{}) (WhereClause, error) {
	return WhereClause{
		Terms: triples.([]Triple),
	}, nil
}

func NewWhereClauseWithGraphGroup(triples, group interface{}) (WhereClause, error) {
	g := group.(GraphGroup)
	return WhereClause{
		Terms:      triples.([]Triple),
		GraphGroup: &g,
	}, nil
}

func NewWhereClauseGraphGroup(group interface{}) (WhereClause, error) {
	g := group.(GraphGroup)
	return WhereClause{
		GraphGroup: &g,
	}, nil
}

type GraphGroup struct {
	Terms  []Triple
	Unions []GraphGroup
}

func GraphGroupFromTriples(triples interface{}) (GraphGroup, error) {
	return GraphGroup{
		Terms: triples.([]Triple),
	}, nil
}

func GraphGroupUnion(left, right interface{}) (GraphGroup, error) {
	return GraphGroup{
		Unions: []GraphGroup{left.(GraphGroup), right.(GraphGroup)},
	}, nil
}

func AddTriplesToGraphGroup(left, triples interface{}) (GraphGroup, error) {
	return GraphGroup{
		Terms:  append(left.(GraphGroup).Terms, triples.([]Triple)...),
		Unions: left.(GraphGroup).Unions,
	}, nil
}

func MergeGraphGroups(left, right interface{}) (GraphGroup, error) {
	return GraphGroup{
		Terms:  append(left.(GraphGroup).Terms, right.(GraphGroup).Terms...),
		Unions: append(left.(GraphGroup).Unions, right.(GraphGroup).Unions...),
	}, nil
}

type Triple struct {
	Subject    turtle.URI
	Predicates []PathPattern
	Object     turtle.URI
}

func (t Triple) String() string {
	s := "<" + t.Subject.String() + "|"
	for _, pp := range t.Predicates {
		s += " " + pp.String()
	}
	return s + " | " + t.Object.String() + ">"
}

func (t Triple) Copy() Triple {
	var p = make([]PathPattern, len(t.Predicates))
	copy(p, t.Predicates)
	return Triple{
		Subject:    t.Subject,
		Object:     t.Object,
		Predicates: p,
	}
}

func NewTriple(subject, predicates, object interface{}) (Triple, error) {
	return Triple{
		Subject:    subject.(turtle.URI),
		Predicates: predicates.([]PathPattern),
		Object:     object.(turtle.URI),
	}, nil
}

func NewTripleBlock(triple interface{}) ([]Triple, error) {
	return []Triple{triple.(Triple)}, nil
}

func AppendTripleBlock(block, triple interface{}) ([]Triple, error) {
	return append(block.([]Triple), triple.(Triple)), nil
}

func NewURI(value interface{}) (turtle.URI, error) {
	return turtle.ParseURI(value.(string)), nil
}

func NewVarList(_var interface{}) ([]string, error) {
	return []string{_var.(string)}, nil
}

func AppendVar(varlist, _var interface{}) ([]string, error) {
	return append(varlist.([]string), _var.(string)), nil
}

func NewStringList(_str interface{}) ([]string, error) {
	return []string{_str.(string)}, nil
}

func AppendString(strlist, _str interface{}) ([]string, error) {
	return append(strlist.([]string), _str.(string)), nil
}

func ParseString(_var interface{}) (string, error) {
	return string(_var.(*token.Token).Lit), nil
}

func ParseQuotedString(_var interface{}) (string, error) {
	return string(_var.(*token.Token).Lit), nil
}

func NewPathSequence(_pred interface{}) ([]PathPattern, error) {
	return []PathPattern{_pred.(PathPattern)}, nil
}

func AppendPathSequence(_seq, _pred interface{}) ([]PathPattern, error) {
	return append(_seq.([]PathPattern), _pred.(PathPattern)), nil
}

func NewPathPattern(_pred interface{}) (PathPattern, error) {
	pred, _ := ParseString(_pred)
	if pred == "a" {
		pred = "rdf:type"
	}
	return PathPattern{
		Predicate: turtle.ParseURI(pred),
		Pattern:   PATTERN_SINGLE,
	}, nil
}

func AddPathMod(_pred, _mod interface{}) (PathPattern, error) {
	return PathPattern{
		Predicate: _pred.(PathPattern).Predicate,
		Pattern:   _mod.(Pattern),
	}, nil
}

type PathPattern struct {
	Predicate turtle.URI
	Pattern   Pattern
}

func PathFromVar(_var interface{}) ([]PathPattern, error) {
	return []PathPattern{
		PathPattern{
			Predicate: turtle.ParseURI(_var.(string)),
			Pattern:   PATTERN_SINGLE,
		},
	}, nil
}

func (pp PathPattern) String() string {
	return pp.Predicate.String() + pp.Pattern.String()
}

type Pattern uint

const (
	PATTERN_SINGLE = iota + 1
	PATTERN_ZERO_ONE
	PATTERN_ONE_PLUS
	PATTERN_ZERO_PLUS
)

func (p Pattern) String() string {
	switch p {
	case PATTERN_SINGLE:
		return ""
	case PATTERN_ZERO_ONE:
		return "?"
	case PATTERN_ONE_PLUS:
		return "+"
	case PATTERN_ZERO_PLUS:
		return "*"
	}
	return "unknown"
}
