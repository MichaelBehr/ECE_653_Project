package Preprocessor

import (
	"fmt"
	"sort"
)

type decLevel int
type Lit int32
type Var int32
const (
	// Indet means the problem is not proven sat or unsat yet.
	Undetermined = Status(iota)
	// Sat means the problem or clause is satisfied.
	Sat
	// Unsat means the problem or clause is unsatisfied.
	Unsat
)

// Utility functions for pre-processor inpsired by implementations/pseudocode from http://fmv.jku.at/papers/EenBiere-SAT05.pdf,
// MaxSatPreprocessor and GopherSat

// data used in Pseudo Boolean constraints.
type pbData struct {
	weights []int  // weight of each literal. If nil, weights are all 1.
	watched []bool // indices of watched literals.
}

// clause structure
type Clause struct {
	lits []Lit
	pbData   *pbData
}

// First returns the first literal from the clause.
func (c *Clause) First() Lit {
	return c.lits[0]
}

// Get returns the ith literal from the clause.
func (c *Clause) Get(i int) Lit {
	return c.lits[i]
}

// returns number of literals in the clause
func (c *Clause) Len() int {
	return len(c.lits)
}

// sorts the literals in the clause
func (c *Clause) Sort(){
	sort.Slice(c.lits, func(i, j int) bool {
		return c.lits[i] < c.lits[j]
	})
}

// Set sets the ith literal of the clause.
func (c *Clause) Set(i int, l Lit) {
	c.lits[i] = l
}

// Shrink reduces the length of the clauses, by removing all lits
// starting from position newLen.
func (c *Clause) Shrink(newLen int) {
	c.lits = c.lits[:newLen]
	if c.pbData != nil {
		c.pbData.weights = c.pbData.weights[:newLen]
		c.pbData.watched = c.pbData.watched[:newLen]
	}
}

// NewClause returns a clause whose lits are given as an argument.
func NewClause(lits []Lit) *Clause {
	return &Clause{lits: lits}
}

// IntToLit converts a CNF literal to a Lit.
func IntToLit(i int32) Lit {
	if i < 0 {
		return Lit(2*(-i-1) + 1)
	}
	return Lit(2 * (i - 1))
}

// IsPositive is true iff l is > 0
func (l Lit) IsPositive() bool {
	return l%2 == 0
}

// Int returns the equivalent CNF literal.
func (l Lit) Int() int32 {
	sign := l&1 == 1
	res := int32((l / 2) + 1)
	if sign {
		return -res
	}
	return res
}

func (l Lit) Negation() Lit {
	// bitwise XOR on the literal. Remember we have encoded it as a 32bit integer
	return l ^ 1
}

// Lit returns the positive Lit associated to v.
func (v Var) Lit() Lit {
	return Lit(v * 2)
}


func (l Lit) Var() Var {
	return Var(l / 2)
}

// CNF returns a DIMACS CNF representation of the clause.
func (c *Clause) CNF() string {
	res := ""
	for _, lit := range c.lits {
		res += fmt.Sprintf("%d ", lit.Int())
	}
	return fmt.Sprintf("%s0", res)
}

//////

// Function Subsumes will return true iff c subsumes c2
// also assumes we have sorted the literals in the clause
func (c *Clause) Subsumes(c2 *Clause) bool {
	// size of c must be less than c2
	if c.Len() > c2.Len() {
		return false
	}
	for _, lit := range c.lits {
		match := false
		for _, lit2 := range c2.lits {
			if lit == lit2 {
				match = true
				break
			}
			// we will not find a matching literal anymore so return false
			if lit2 > lit {
				return false
			}
		}
		// if for any literal in clause c, there is no match in clause c2 return false
		if !match {
			return false
		}
	}
	return true
}

// SelfSubsumes returns true iff c self-subsumes c2.
func (c *Clause) SelfSubsumes(c2 *Clause) bool {
	oneNeg := false
	for _, lit := range c.lits {
		found := false
		for _, lit2 := range c2.lits {
			if lit == lit2 {
				found = true
				break
			}
			if lit == lit2.Negation() {
				// we only want one matching negative literal
				if oneNeg {
					return false
				}
				oneNeg = true
				found = true
				break
			}
			// We won't find it anymore
			if lit2 > lit {
				return false
			}
		}
		if !found {
			return false
		}
	}
	return oneNeg
}

// Simplify simplifies the given clause by removing redundant lits.
// If the clause is trivially satisfied (i.e contains both a lit and its negation),
// true is returned. Otherwise, false is returned.
func (c *Clause) Simplify() (isSat bool) {
	c.Sort()
	lits := make([]Lit, 0, len(c.lits))
	i := 0
	for i < len(c.lits) {
		if i < len(c.lits)-1 && c.lits[i] == c.lits[i+1].Negation() {
			return true
		}
		lit := c.lits[i]
		lits = append(lits, lit)
		i++
		for i < len(c.lits) && c.lits[i] == lit {
			i++
		}
	}
	if len(lits) < len(c.lits) {
		c.lits = lits
	}
	return false
}

// Generate returns a subsumed clause from c and c2, by removing v.
func (c *Clause) Generate(c2 *Clause, v Var) *Clause {
	c3 := &Clause{lits: make([]Lit, 0, len(c.lits)+len(c2.lits)-2)}
	for _, lit := range c.lits {
		if lit.Var() != v {
			c3.lits = append(c3.lits, lit)
		}
	}
	for _, lit2 := range c2.lits {
		if lit2.Var() != v {
			c3.lits = append(c3.lits, lit2)
		}
	}
	return c3
}
