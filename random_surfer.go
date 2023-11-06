package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
)

type MoveType int

const (
	RandomInWeb = iota
	ViaEdge
)

type Node struct {
	Edges         []int
	ReversedEdges []int
	Branches      int
}

type Graph struct {
	Nodes []Node
	Size  int
}

func (g *Graph) AddEdge(from, to int) {
	g.Nodes[from].Edges = append(g.Nodes[from].Edges, to)
	g.Nodes[to].ReversedEdges = append(g.Nodes[to].ReversedEdges, from)
	g.Nodes[from].Branches++
}

func (g *Graph) Branches(node int) int {
	return g.Nodes[node].Branches
}

func (g *Graph) DanglingNodes() []float64 {
	nodes := make([]float64, g.Size)
	val := 1 / float64(g.Size) // val = 1 / n

	for i, node := range g.Nodes {
		if len(node.ReversedEdges) == 0 {
			nodes[i] = val
		} else {
			nodes[i] = 0
		}
	}
	return nodes
}

func NewGraph(n int) *Graph {
	nodes := make([]Node, n)
	return &Graph{Nodes: nodes, Size: n}
}

func ReadFromFile(path string) (*Graph, error) {
	var from, to int

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(file)
	n := 0
	firstLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	n, err = strconv.Atoi(strings.TrimSpace(firstLine))
	if err != nil {
		return nil, err
	}

	g := NewGraph(n)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// Will this handle EOF?
			break
		}

		numbers := strings.Fields(line)
		if len(numbers)%2 != 0 {
			return nil, fmt.Errorf("Invalid line: %s", line)
		}

		for i := 0; i < len(numbers); i += 2 {
			from, err = strconv.Atoi(numbers[i])
			if err != nil {
				return nil, fmt.Errorf("Invalid number: %s", line)
			}

			to, err = strconv.Atoi(numbers[i+1])
			if err != nil {
				return nil, fmt.Errorf("Invalid number: %s", line)
			}

			g.AddEdge(from, to)
		}
	}

	return g, nil
}

func (g *Graph) RandomSurfer() {
	type visitedNode struct {
		nodeId       int
		visitedTimes int
	}

	visitedMap := make(map[int]visitedNode)

	// Determine starting node
	nodeId := rand.Intn(g.Size)
	node := g.Nodes[nodeId]

	// Initiate now, determine real action in loop
	action := RandomInWeb

	for i := 0; i < randomSurferIterations; i++ {
		// TODO: move to own function
		if len(node.Edges) == 0 {
			// This is a dangling node, we can only move randomly in web
			action = RandomInWeb
		} else {
			// Determine if we'll surf the web though an edge or randomly
			if rand.Float64() < probMoveRandom {
				action = RandomInWeb
			} else {
				action = ViaEdge
			}
		}

		if _, ok := visitedMap[nodeId]; !ok {
			visitedMap[nodeId] = visitedNode{nodeId: nodeId, visitedTimes: 1}
		} else {
			n := visitedMap[nodeId]
			n.visitedTimes++
			visitedMap[nodeId] = n
		}

		switch action {
		case RandomInWeb:
			// Move randomly in web
			nodeId = rand.Intn(g.Size)
			node = g.Nodes[nodeId]
		case ViaEdge:
			// Move via edge
			nodeId = node.Edges[rand.Intn(len(node.Edges))]
			node = g.Nodes[nodeId]
		}
	}

	ranked := make([]visitedNode, len(visitedMap))

	for _, v := range visitedMap {
		ranked = append(ranked, v)
	}

	// Sort ranked by visited times
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].visitedTimes > ranked[j].visitedTimes
	})

	fmt.Printf("\nRandomSurfer top %d rankings after %d iterations\n", randomSurferPrintHowMany, randomSurferIterations)
	for i := 0; i < randomSurferPrintHowMany && i < g.Size; i++ {
		fmt.Printf("Rank: %d - Node: %d - Visited times: %d\n", i+1, ranked[i].nodeId, ranked[i].visitedTimes)
	}
}

func (g *Graph) CreateAMatrix(factor float64) [][]float64 {
	A := make([][]float64, g.Size)

	for i := 0; i < g.Size; i++ {
		A[i] = make([]float64, g.Size)
	}

	for i, node := range g.Nodes {
		for _, edge := range node.Edges {
			A[edge][i] = (1 / float64(g.Branches(i))) * factor
		}
	}

	return A
}
