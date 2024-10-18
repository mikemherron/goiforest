package goitree

import (
	"fmt"
	"math"
)

type IsolationForest struct {
	trees           []*IsolationTree
	trainingSize    int
	expectedAverage int
}

func (f *IsolationForest) Score(dataPoint map[string]string) float64 {
	score := 0.0

	var pathLengthTotal int
	for _, tree := range f.trees {
		pathLengthTotal += tree.traverse(dataPoint)
	}

	avgPathLength := float64(pathLengthTotal) / float64(len(f.trees))

	return score / NumTrees
}

type IsolationTree struct {
	Root *IsolationTreeNode
}

func (t *IsolationTree) traverse(dataPoint map[string]string) int {
	pathLength := 0
	node := t.Root
	for !node.isLeaf {
		if dataPoint.Features[node.Feature] == node.Value {
			node = node.Left
		} else {
			node = node.Right
		}
		pathLength++
	}

	return pathLength
}

// types of comparison
// Equals, Not Equals
type IsolationTreeNode struct {
	left   *IsolationTreeNode
	right  *IsolationTreeNode
	split  *splitCondition
	isLeaf bool
}

const NumTrees = 100
const SampleSize = 512

func BuildForest(dataSet *DataSet) *IsolationForest {
	forest := IsolationForest{
		trees:           []*IsolationTree{},
		trainingSize:    dataSet.Size,
		expectedAverage: 0,
	}

	maxDepth := int(math.Log(SampleSize) / math.Log(2))

	fmt.Print("Max Depth: ", maxDepth)
	for i := 0; i < NumTrees; i++ {
		forest.trees = append(forest.trees,
			&IsolationTree{Root: buildTree(dataSet.sample(SampleSize), 0, maxDepth)})
	}

	return &forest
}

func buildTree(dataSet *DataSet, depth int, maxDepth int) *IsolationTreeNode {
	node := &IsolationTreeNode{}
	if dataSet.Size <= 1 || depth >= maxDepth {
		node.isLeaf = true
	} else {
		split, left, right := dataSet.Split()
		node.split = split
		node.left = buildTree(left, depth+1, maxDepth)
		node.right = buildTree(right, depth+1, maxDepth)
	}

	return node
}

func harmonicNumber(n int) float64 {
	return float64(n) + 0.5772156649 // Euler-Mascheroni constant
}
