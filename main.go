package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"Preprocessor"
)

func main() {
	var (
		help    bool
	)
	flag.BoolVar(&help, "help", false, "displays help")
	flag.Parse()
	if !help && len(flag.Args()) != 1 {
		fmt.Printf("This is GoPreProcessor. Functions taken from Gophersat. Modifications/additions by Michael Behr.\n")
		fmt.Fprintf(os.Stderr, "Syntax : %s [options] (file.cnf|file.wcnf|file.bf|file.opb)\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	if help {
		fmt.Printf("This is GoPreProcessor version 1.0, a SAT pre-processor by Michael Behr and Jared Lenos.\n")
		fmt.Printf("Syntax : %s [options] (file.cnf|file.wcnf|file.bf|file.opb)\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}
	path := flag.Args()[0]
	fmt.Printf("c solving %s\n", path)
	if strings.HasSuffix(path, ".cnf") {
		if pb, err := parse(flag.Args()[0]); err != nil {
			fmt.Fprintf(os.Stderr, "could not parse problem: %v\n", err)
			os.Exit(1)
		} else {
			fmt.Printf("\nCNF FORMULA:\n\n",pb.CNF())
			// run pre-processing
			pb.Preprocess()
			fmt.Printf("\nSIMPLIFIED FORMULA:\n\n",pb.CNF())
		}
	} else{
		fmt.Fprintf(os.Stderr, "Could not parse problem. Make sure it is in CNF form.")
	}

}
func parse(path string) (pb *Preprocessor.Problem, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %q: %v", path, err)
	}
	defer f.Close()
	if strings.HasSuffix(path, ".cnf") {
		pb, err := Preprocessor.ParseCNF(f)
		if err != nil {
			return nil, fmt.Errorf("could not parse DIMACS file %q: %v", path, err)
		}
		return pb,nil
	}
	return nil, fmt.Errorf("invalid file format for %q", path)
}