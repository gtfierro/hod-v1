package db

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRelationAdd1Value(t *testing.T) {
	assert := assert.New(t)

	// test 1 var
	rel1 := newRelation([]string{"var1"})
	assert.NotNil(rel1)
	var1Vals := generateKeyMap(5, 1)
	rel1.add1Value("var1", var1Vals)
	assert.Equal(0, rel1.vars["var1"])
	assert.Equal(len(rel1.rows), 5)
}

func TestRelationAdd2Value(t *testing.T) {
	assert := assert.New(t)

	rel2 := newRelation([]string{"var1", "var2"})
	rel2vals := generateValues(2, 10)
	rel2.add2Values("var1", "var2", rel2vals)
	assert.Equal(0, rel2.vars["var1"])
	assert.Equal(1, rel2.vars["var2"])
	assert.Equal(10, len(rel2.rows))
}

func TestRelationAdd3Value(t *testing.T) {
	assert := assert.New(t)

	rel3 := newRelation([]string{"var1", "var2", "var3"})
	rel3vals := generateValues(3, 10)
	rel3.add3Values("var1", "var2", "var3", rel3vals)
	assert.Equal(0, rel3.vars["var1"])
	assert.Equal(1, rel3.vars["var2"])
	assert.Equal(2, rel3.vars["var3"])
	assert.Equal(10, len(rel3.rows))
}

func TestRelationJoin1Value(t *testing.T) {
	assert := assert.New(t)

	// relation1 (var1)
	rel1 := newRelation([]string{"var1"})
	assert.NotNil(rel1)
	var1Vals := generateKeyMap(5, 1)
	rel1.add1Value("var1", var1Vals)
	assert.Equal(0, rel1.vars["var1"])
	assert.Equal(len(rel1.rows), 5)

	// relation2 (var1, var2)
	rel2 := newRelation([]string{"var1", "var2"})
	rel2vals := generateValues(2, 10)
	rel2.add2Values("var1", "var2", rel2vals)
	assert.Equal(1, rel2.vars["var2"])
	assert.Equal(10, len(rel2.rows))

	// inner join
	rel1.join(rel2, []string{"var1"}, nil)
	assert.Equal(3, len(rel1.rows))
}

func TestRelationJoin1ValueNoIntersection(t *testing.T) {
	assert := assert.New(t)

	// relation1 (var1)
	rel1 := newRelation([]string{"var1"})
	assert.NotNil(rel1)
	var1Vals := generateKeyMap(5, 100)
	rel1.add1Value("var1", var1Vals)
	assert.Equal(0, rel1.vars["var1"])
	assert.Equal(len(rel1.rows), 5)

	// relation2 (var1, var2)
	rel2 := newRelation([]string{"var1", "var2"})
	rel2vals := generateValues(2, 10)
	rel2.add2Values("var1", "var2", rel2vals)
	assert.Equal(1, rel2.vars["var2"])
	assert.Equal(10, len(rel2.rows))

	// inner join
	rel1.join(rel2, []string{"var1"}, nil)
	assert.Equal(0, len(rel1.rows))
}

func BenchmarkRelationAdd1Value(b *testing.B) {
	rel1 := newRelation([]string{"var1"})
	var1Vals := generateKeyMap(5, 1)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rel1.add1Value("var1", var1Vals)
	}
}

func BenchmarkRelationAdd2Value(b *testing.B) {
	rel2 := newRelation([]string{"var1", "var2"})
	rel2vals := generateValues(2, 10)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rel2.add2Values("var1", "var2", rel2vals)
	}
}

func BenchmarkRelationAdd3Value(b *testing.B) {
	rel3vals := generateValues(3, 10)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rel3 := newRelation([]string{"var1", "var2", "var3"})
		rel3.add3Values("var1", "var2", "var3", rel3vals)
	}
}
