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

const (
	m                      = 0.15
	probMoveRandom         = m
	probMoveViaEdges       = 1 - m
	randomSurferIterations = 10_000_000
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

func (g *Graph) DanglingNodes() []int {
	nodes := make([]int, 0)
	for i, node := range g.Nodes {
		if len(node.ReversedEdges) == 0 {
			nodes = append(nodes, i)
		}
	}
	return nodes
}

func NewGraph(n int) *Graph {
	nodes := make([]Node, n)
	return &Graph{Nodes: nodes, Size: n}
}

func ReadFromFile(path string) (*Graph, error) {
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

	var from, to int
	g := NewGraph(n)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// Will this handle EOF?
			break
		}

		numbers := strings.Fields(line)
		if len(numbers) != 2 {
			return nil, fmt.Errorf("Invalid line: %s", line)
		}

		from, err = strconv.Atoi(numbers[0])
		if err != nil {
			return nil, fmt.Errorf("Invalid number: %s", line)
		}

		to, err = strconv.Atoi(numbers[1])
		if err != nil {
			return nil, fmt.Errorf("Invalid number: %s", line)
		}

		g.AddEdge(from, to)
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

	fmt.Println("RandomSurfer ranking:")
	for i := 0; i < 10; i++ {
		fmt.Println("Rank", i+1, "-", ranked[i].nodeId)
	}

}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a file path to graph data file")
		return
	}

	g, err := ReadFromFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	// Print graph size
	fmt.Println(g.Size)

	// Print RandomSurfer ranking
	g.RandomSurfer()
}
