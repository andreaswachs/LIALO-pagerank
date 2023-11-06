package main

import (
	"fmt"
	"os"
)

const (
	m                        = 0.15
	probMoveRandom           = m
	oneMinus_m               = 1 - m
	randomSurferIterations   = 10_000_000
	randomSurferPrintHowMany = 10
	pageRankPrintHowMany     = 10
	pageRankIterations       = 100
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage:", os.Args[0], "<graph_data_file> command")
		fmt.Println("Commands:")
		fmt.Println("\trandom-surfer")
		fmt.Println("\tpagerank")
		fmt.Println("\tboth")
		return
	}

	g, err := ReadFromFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	switch os.Args[2] {
	case "random-surfer":
		g.RandomSurfer()
	case "pagerank":
		PageRank(g)
	case "both":
		g.RandomSurfer()
		PageRank(g)
	default:
		fmt.Println("Unknown command")
	}
}
