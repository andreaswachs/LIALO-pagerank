package main

import (
	"fmt"
	"sort"
)

func PageRank(source *Graph) {
	// A is (1 - m)A
	A := source.CreateAMatrix(oneMinus_m)

	// Compute D as (1 - m)D
	D := source.DanglingNodes()
	for i := 0; i < len(D); i++ {
		D[i] *= oneMinus_m
	}

	// We can define mSx_k as a constant
	mSx_k := m / float64(source.Size)

	// Allocate memory for xk and xk_plus_1
	// This makes sure we only need to ever allocate memory for xk twice
	xk := make([]float64, source.Size)
	xk_plus_1 := make([]float64, source.Size)

	// Optimization: We can use two buffers to avoid allocating new memory
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
			addMatrixVector(A, xk, compBuf1)

			// Calculate component [2]
			mulVectorVector(D, xk, compBuf2)

			// Prematurely assign values to xk_plus_1
			addVectorVector(compBuf1, compBuf2, xk_plus_1)

			// Add component [3] after the fact
			addVectorScalar(xk_plus_1, mSx_k, xk_plus_1)
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

	fmt.Printf("\nPageRank top %d rankings after %d iterations\n", pageRankPrintHowMany, pageRankIterations)
	for i := 0; i < randomSurferPrintHowMany && i < source.Size; i++ {
		fmt.Printf("Rank: %d - Node: %d\n", i+1, ranked[i].nodeId)
	}

}

func addMatrixVector(A [][]float64, x []float64, res []float64) []float64 {
	for i := 0; i < len(x); i++ {
		for j := 0; j < len(x); j++ {
			res[i] += A[i][j] * x[j]
		}
	}

	return res
}

func addVectorVector(x []float64, y []float64, res []float64) []float64 {
	for i := 0; i < len(x); i++ {
		res[i] = x[i] + y[i]
	}

	return res
}

func addVectorScalar(x []float64, scalar float64, res []float64) []float64 {
	for i := 0; i < len(x); i++ {
		res[i] = x[i] + scalar
	}

	return res
}

func mulVectorVector(x []float64, y []float64, res []float64) []float64 {
	for i := 0; i < len(x); i++ {
		res[i] = x[i] * y[i]
	}

	return res
}
