package Preprocessor

import (
	"fmt"
	"log"
)

//

// UTILITY FUNCTIONS AND STRUCTURES

type Status byte

// A Problem is a list of clauses & a number of vars.
type Problem struct {
	NbVars     int        // Total number of vars
	Clauses    []*Clause  // List of non-empty, non-unit clauses
	Status     Status     // Status of the problem. Can be trivially UNSAT (if empty clause was met or inferred by UP) or Indet.
	Units      []Lit      // List of unit literal found in the problem.
	Model      []decLevel // For each var, its inferred binding. 0 means unbound, 1 means bound to true, -1 means bound to false.
	minLits    []Lit      // For an optimisation problem, the list of lits whose sum must be minimized
	minWeights []int      // For an optimisation problem, the weight of each lit.
}

// CNF returns a DIMACS CNF representation of the problem.
func (pb *Problem) CNF() string {
	res := fmt.Sprintf("p cnf %d %d\n", pb.NbVars, len(pb.Clauses)+len(pb.Units))
	for _, unit := range pb.Units {
		res += fmt.Sprintf("%d 0\n", unit.Int())
	}
	for _, clause := range pb.Clauses {
		res += fmt.Sprintf("%s\n", clause.CNF())
	}
	return res
}

///// PROBLEM UTILITY FUNCTIONS FROM GOPHERSAT

func (pb *Problem) updateStatus(nbClauses int) {
	pb.Clauses = pb.Clauses[:nbClauses]
	if pb.Status == Undetermined && nbClauses == 0 {
		pb.Status = Sat
	}
}

func (pb *Problem) addUnit(lit Lit) {
	if lit.IsPositive() {
		if pb.Model[lit.Var()] == -1 {
			pb.Status = Unsat
			return
		}
		pb.Model[lit.Var()] = 1
	} else {
		if pb.Model[lit.Var()] == 1 {
			pb.Status = Unsat
			return
		}
		pb.Model[lit.Var()] = -1
	}
	pb.Units = append(pb.Units, lit)
}

// simplify simplifies the pure SAT problem, i.e runs unit propagation if possible.
func (pb *Problem) simplify2() {
	nbClauses := len(pb.Clauses)
	restart := true
	for restart {
		restart = false
		i := 0
		for i < nbClauses {
			c := pb.Clauses[i]
			nbLits := c.Len()
			clauseSat := false
			j := 0
			for j < nbLits {
				lit := c.Get(j)
				if pb.Model[lit.Var()] == 0 {
					j++
				} else if (pb.Model[lit.Var()] == 1) == lit.IsPositive() {
					clauseSat = true
					break
				} else {
					nbLits--
					c.Set(j, c.Get(nbLits))
				}
			}
			if clauseSat {
				nbClauses--
				pb.Clauses[i] = pb.Clauses[nbClauses]
			} else if nbLits == 0 {
				pb.Status = Unsat
				return
			} else if nbLits == 1 { // UP
				pb.addUnit(c.First())
				if pb.Status == Unsat {
					return
				}
				nbClauses--
				pb.Clauses[i] = pb.Clauses[nbClauses]
				restart = true // Must restart, since this lit might have made one more clause Unit or SAT.
			} else { // nb lits unbound > cardinality
				if c.Len() != nbLits {
					c.Shrink(nbLits)
				}
				i++
			}
		}
	}
	pb.updateStatus(nbClauses)
}

// Preprocess main function

func (pb *Problem) preprocess() {
	log.Printf("Preprocessing... %d clauses currently", len(pb.Clauses))
	occurs := make([][]int, pb.NbVars*2)
	for i, c := range pb.Clauses {
		for j := 0; j < c.Len(); j++ {
			occurs[c.Get(j)] = append(occurs[c.Get(j)], i)
		}
	}
	modified := true
	neverModified := true
	for modified {
		modified = false
		for i := 0; i < pb.NbVars; i++ {
			if pb.Model[i] != 0 {
				continue
			}
			v := Var(i)
			lit := v.Lit()
			nbLit := len(occurs[lit])
			nbLit2 := len(occurs[lit.Negation()])
			// slow method is only effective with less than 10 literals
			if (nbLit < 10 || nbLit2 < 10) && (nbLit != 0 || nbLit2 != 0) {
				modified = true
				neverModified = false
				// pb.deleted[v] = true
				log.Printf("%d can be removed: %d and %d", lit.Int(), len(occurs[lit]), len(occurs[lit.Negation()]))
				for _, idx1 := range occurs[lit] {
					for _, idx2 := range occurs[lit.Negation()] {
						c1 := pb.Clauses[idx1]
						c2 := pb.Clauses[idx2]
						// generate new clause with self-subsuming resolution
						newC := c1.Generate(c2, v)
						if !newC.Simplify() {
							switch newC.Len() {
							case 0:
								log.Printf("Inferred UNSAT")
								pb.Status = Unsat
								return
							case 1:
								log.Printf("Unit %d", newC.First().Int())
								lit2 := newC.First()
								if lit2.IsPositive() {
									if pb.Model[lit2.Var()] == -1 {
										pb.Status = Unsat
										return
									}
									pb.Model[lit2.Var()] = 1
								} else {
									if pb.Model[lit2.Var()] == 1 {
										pb.Status = Unsat
										return
									}
									pb.Model[lit2.Var()] = -1
								}
								pb.Units = append(pb.Units, lit2)
							default:
								pb.Clauses = append(pb.Clauses, newC)
							}
						}
					}
				}
				nbRemoved := 0
				for _, idx := range occurs[lit] {
					pb.Clauses[idx] = pb.Clauses[len(pb.Clauses)-nbRemoved-1]
					nbRemoved++
				}
				for _, idx := range occurs[lit.Negation()] {
					pb.Clauses[idx] = pb.Clauses[len(pb.Clauses)-nbRemoved-1]
					nbRemoved++
				}
				pb.Clauses = pb.Clauses[:len(pb.Clauses)-nbRemoved]
				log.Printf("clauses=%s", pb.CNF())
				// Redo occurs
				occurs = make([][]int, pb.NbVars*2)
				for i, c := range pb.Clauses {
					for j := 0; j < c.Len(); j++ {
						occurs[c.Get(j)] = append(occurs[c.Get(j)], i)
					}
				}
				continue
			}
		}
	}
	if !neverModified {
		log.Printf("There was no modifications to the boolean formula.")
	}
	log.Printf("Done. %d clauses now", len(pb.Clauses))
}
