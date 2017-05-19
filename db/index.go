//go:generate msgp
package db

type PredIndex map[string]*PredicateEntity
type RelshipIndex map[string]string
type NamespaceIndex map[string]string
