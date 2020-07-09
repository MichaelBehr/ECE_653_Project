package Preprocessor

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// COPIES OF GOPHERSAT'S PARSE FUNCTIONS FOR TESTING


func parseHeader(r *bufio.Reader) (nbVars, nbClauses int, err error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return 0, 0, fmt.Errorf("cannot read header: %v", err)
	}
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return 0, 0, fmt.Errorf("invalid syntax %q in header", line)
	}
	nbVars, err = strconv.Atoi(fields[1])
	if err != nil {
		return 0, 0, fmt.Errorf("nbvars not an int : %q", fields[1])
	}
	nbClauses, err = strconv.Atoi(fields[2])
	if err != nil {
		return 0, 0, fmt.Errorf("nbClauses not an int : '%s'", fields[2])
	}
	return nbVars, nbClauses, nil
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// readInt reads an int from r.
// 'b' is the last read byte. It can be a space, a '-' or a digit.
// The int can be negated.
// All spaces before the int value are ignored.
// Can return EOF.
func readInt(b *byte, r *bufio.Reader) (res int, err error) {
	for err == nil && isSpace(*b) {
		*b, err = r.ReadByte()
	}
	if err == io.EOF {
		return res, io.EOF
	}
	if err != nil {
		return res, fmt.Errorf("could not read digit: %v", err)
	}
	neg := 1
	if *b == '-' {
		neg = -1
		*b, err = r.ReadByte()
		if err != nil {
			return 0, fmt.Errorf("cannot read int: %v", err)
		}
	}
	for err == nil {
		if *b < '0' || *b > '9' {
			return 0, fmt.Errorf("cannot read int: %q is not a digit", *b)
		}
		res = 10*res + int(*b-'0')
		*b, err = r.ReadByte()
		if isSpace(*b) {
			break
		}
	}
	res *= neg
	return res, err
}

// ParseCNF parses a CNF file and returns the corresponding Problem.
func ParseCNF(f io.Reader) (*Problem, error) {
	r := bufio.NewReader(f)
	var (
		nbClauses int
		pb        Problem
	)
	b, err := r.ReadByte()
	for err == nil {
		if b == 'c' { // Ignore comment
			b, err = r.ReadByte()
			for err == nil && b != '\n' {
				b, err = r.ReadByte()
			}
		} else if b == 'p' { // Parse header
			pb.NbVars, nbClauses, err = parseHeader(r)
			if err != nil {
				return nil, fmt.Errorf("cannot parse CNF header: %v", err)
			}
			pb.Model = make([]decLevel, pb.NbVars)
			pb.Clauses = make([]*Clause, 0, nbClauses)
		} else {
			lits := make([]Lit, 0, 3) // Make room for some lits to improve performance
			for {
				val, err := readInt(&b, r)
				fmt.Printf("Value: " + string(val))
				if err == io.EOF {
					if len(lits) != 0 { // This is not a trailing space at the end...
						return nil, fmt.Errorf("unfinished clause while EOF found")
					}
					break // When there are only several useless spaces at the end of the file, that is ok
				}
				if err != nil {
					return nil, fmt.Errorf("cannot parse clause: %v", err)
				}
				if val == 0 {
					pb.Clauses = append(pb.Clauses, NewClause(lits))
					break
				} else {
					if val > pb.NbVars || -val > pb.NbVars {
						return nil, fmt.Errorf("invalid literal %d for problem with %d vars only", val, pb.NbVars)
					}
					lits = append(lits, IntToLit(int32(val)))
				}
			}
		}
		b, err = r.ReadByte()
	}
	if err != io.EOF {
		return nil, err
	}
	pb.simplify2()
	return &pb, nil
}