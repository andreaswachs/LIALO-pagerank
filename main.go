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
	m                        = 0.15
	probMoveRandom           = m
	probMoveViaEdges         = 1 - m
	randomSurferIterations   = 10_000_000
	randomSurferPrintHowMany = 10
	pageRankIterations       = 100
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

	fmt.Printf("RandomSurfer top %d rankings after %d iterations\n", randomSurferPrintHowMany, randomSurferIterations)
	for i := 0; i < randomSurferPrintHowMany; i++ {
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

func (g *Graph) CreateA_plus_DMatrix() [][]float64 {
	A := make([][]float64, g.Size)
	for i := 0; i < g.Size; i++ {
		A[i] = make([]float64, g.Size)
	}

	oneOverN := 1 / float64(g.Size)
	k := oneOverN

	for i, node := range g.Nodes {
		if g.Branches(i) != 0 {
			k = 1 / float64(g.Branches(i))
		} else {
			k = oneOverN
		}
		for _, edge := range node.Edges {
			A[edge][i] = k
		}
	}

	return A
}

func (g *Graph) CreateMMatrix() [][]float64 {
	A := g.CreateA_plus_DMatrix()
	M := make([][]float64, g.Size)
	mS := m / float64(g.Size)

	for i := 0; i < g.Size; i++ {
		M[i] = make([]float64, g.Size)
	}

	for i := 0; i < g.Size; i++ {
		for j := 0; j < g.Size; j++ {
			M[i][j] = probMoveViaEdges*A[i][j] + mS
		}
	}

	return M
}

func (g *Graph) CreateDAsVector() []float64 {
	return g.DanglingNodes()
}

func PageRank(source *Graph) {
	// A is (1 - m)A
	A := source.CreateAMatrix(probMoveViaEdges)

	// Compute D as (1 - m)D
	D := source.DanglingNodes()
	for i := 0; i < len(D); i++ {
		D[i] *= probMoveViaEdges
	}

	// We can define mSx_k as a constant
	mSx_k := m / float64(source.Size)

	xk := make([]float64, source.Size)
	xk_plus_1 := make([]float64, source.Size)

	// Buffer holds intermediate results when computing x_k+1
	compBuf1 := make([]float64, source.Size)

	compBuf2 := make([]float64, source.Size)

	// Use this helper when we move through iterations
	// as to use two vectors in total
	swapxk := func() {
		xk, xk_plus_1 = xk_plus_1, xk
	}

	// Initialize xk
	oneOverN := 1 / float64(source.Size)
	for i := 0; i < source.Size; i++ {
		xk[i] = oneOverN
	}

	// xk+1 = (1 − m)Axk + (1 − m)Dxk + mSxk
	//        [1]          [2]          [3]

	for i := 0; i < pageRankIterations; i++ {

		for j := 0; j < source.Size; j++ {
			// Calculate component [1]
			AddMatrixVector(A, xk, compBuf1)

			// Calculate component [2]
			MulVectorVector(D, xk, compBuf2)

			// Prematurely assign values to xk_plus_1
			AddVectorVector(compBuf1, compBuf2, xk_plus_1)

			// Add component [3]
			AddVectorScalar(xk_plus_1, mSx_k, xk_plus_1)
		}

		// We have completed this round's iteration, now we flip the ranking vectors
		swapxk()
	}

	// Move rankings into a map and sort it
	type rankedNode struct {
		nodeId int
		rank   float64
	}

	ranked := make([]rankedNode, source.Size)

	for i, rank := range xk {
		ranked[i] = rankedNode{nodeId: i, rank: rank}
	}

	// Sort ranked by visited 15:38:20
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].rank > ranked[j].rank
	})

	fmt.Printf("PageRank top %d rankings after %d iterations\n", randomSurferPrintHowMany, pageRankIterations)
	for i := 0; i < randomSurferPrintHowMany; i++ {
		fmt.Printf("Rank: %d - Node: %d\n", i+1, ranked[i].nodeId)
	}

}

func AddMatrixVector(A [][]float64, x []float64, res []float64) []float64 {
	for i := 0; i < len(x); i++ {
		for j := 0; j < len(x); j++ {
			res[i] += A[i][j] * x[j]
		}
	}

	return res
}

func AddMatrixPretendMatrix(A [][]float64, B []float64, res [][]float64) [][]float64 {
	// We pretend B is a matrix, when it really is a vector, but each row is identical
	for i := 0; i < len(A); i++ {
		for j := 0; j < len(A); j++ {
			res[i][j] = A[i][j] + B[j]
		}
	}

	return res
}

func AddVectorVector(x []float64, y []float64, res []float64) []float64 {
	for i := 0; i < len(x); i++ {
		res[i] = x[i] + y[i]
	}

	return res
}

func AddVectorScalar(x []float64, scalar float64, res []float64) []float64 {
	for i := 0; i < len(x); i++ {
		res[i] = x[i] + scalar
	}

	return res
}

func MulVectorVector(x []float64, y []float64, res []float64) []float64 {
	for i := 0; i < len(x); i++ {
		res[i] = x[i] * y[i]
	}

	return res
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

	g.RandomSurfer()

	PageRank(g)
}
