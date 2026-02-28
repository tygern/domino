package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/tygern/domino/internal/coxeter"
	"github.com/tygern/domino/internal/tableau"
	"github.com/tygern/domino/internal/tikz"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "info":
		runInfo(os.Args[2:])
	case "tableau":
		runTableau(os.Args[2:])
	case "heap":
		runHeap(os.Args[2:])
	case "bad":
		runBad(os.Args[2:])
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: domino <command> [flags]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  info     -perm 1,-4,3,-2       Element info (length, descents, bad, reduced)")
	fmt.Fprintln(os.Stderr, "  info     -expr 1,2,3,4,3 -rank 4")
	fmt.Fprintln(os.Stderr, "  tableau  -perm 1,-4,3,-2       Right and left tableaux as TikZ")
	fmt.Fprintln(os.Stderr, "  heap     -perm 1,-4,3,-2       Heap as TikZ")
	fmt.Fprintln(os.Stderr, "  bad      -rank 4               List all bad elements of D_n")
}

func runInfo(args []string) {
	elem := parseElement(args)

	re := elem.ReducedExpression()
	fmt.Printf("Permutation:   %s\n", elem)
	fmt.Printf("Length:        %d\n", elem.Length())
	fmt.Printf("Right descent: %s\n", formatSet(elem.RightDescentSet()))
	fmt.Printf("Left descent:  %s\n", formatSet(elem.LeftDescentSet()))
	fmt.Printf("Bad:           %t\n", elem.IsBad())
	fmt.Printf("Reduced:       %s\n", re)
}

func runTableau(args []string) {
	elem := parseElement(args)

	fmt.Println("Right tableau:")
	fmt.Println(tikz.RenderTableau(tableau.RightTableau(elem)))
	fmt.Println("Left tableau:")
	fmt.Println(tikz.RenderTableau(tableau.LeftTableau(elem)))
}

func runHeap(args []string) {
	elem := parseElement(args)
	fmt.Println(tikz.RenderHeap(tableau.NewHeap(elem)))
}

func runBad(args []string) {
	rank := 0
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "-rank" {
			var err error
			rank, err = strconv.Atoi(args[i+1])
			if err != nil || rank < 1 {
				fmt.Fprintln(os.Stderr, "Error: invalid rank")
				os.Exit(1)
			}
		}
	}
	if rank == 0 {
		fmt.Fprintln(os.Stderr, "Error: -rank is required")
		os.Exit(1)
	}

	bad := coxeter.BadElements(rank)
	for _, elem := range bad {
		fmt.Println(elem)
	}
	fmt.Fprintf(os.Stderr, "%d bad elements in D_%d\n", len(bad), rank)
}

func parseElement(args []string) coxeter.Element {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "-perm" {
			return parsePerm(args[i+1])
		}
		if args[i] == "-expr" {
			rank := 0
			for j := i + 2; j < len(args)-1; j++ {
				if args[j] == "-rank" {
					var err error
					rank, err = strconv.Atoi(args[j+1])
					if err != nil {
						fmt.Fprintln(os.Stderr, "Error: invalid rank")
						os.Exit(1)
					}
				}
			}
			if rank == 0 {
				fmt.Fprintln(os.Stderr, "Error: -rank is required with -expr")
				os.Exit(1)
			}
			return parseExpr(args[i+1], rank)
		}
	}
	fmt.Fprintln(os.Stderr, "Error: -perm or -expr is required")
	os.Exit(1)
	return coxeter.Element{}
}

func parsePerm(s string) coxeter.Element {
	parts := strings.Split(s, ",")
	perm := make([]int, len(parts))
	for i, p := range parts {
		val, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid integer %q\n", p)
			os.Exit(1)
		}
		perm[i] = val
	}
	elem, err := coxeter.NewElement(perm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	return elem
}

func parseExpr(s string, rank int) coxeter.Element {
	parts := strings.Split(s, ",")
	gens := make([]int, len(parts))
	for i, p := range parts {
		val, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid integer %q\n", p)
			os.Exit(1)
		}
		gens[i] = val
	}
	expr, err := coxeter.NewExpression(gens, rank)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	return expr.ToElement()
}

func formatSet(items []int) string {
	if len(items) == 0 {
		return "{}"
	}
	parts := make([]string, len(items))
	for i, v := range items {
		parts[i] = strconv.Itoa(v)
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
