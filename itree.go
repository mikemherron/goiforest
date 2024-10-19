package goitree

import (
	"fmt"
	"math"
)

type IsolationForest struct {
	trees           []*IsolationTree
	features        map[string]Feature
	trainingSize    int
	expectedAverage float64
}

func (f *IsolationForest) Score(dataPoint map[string]string) float64 {
	score := 0.0

	dataPointFeatures := make(map[Feature]FeatureValue)
	for _, f := range f.features {
		val, exists := dataPoint[f.Name]
		if !exists {
			panic(fmt.Sprintf("Feature %s not found on %v", f.Name, dataPoint))
		}
		dataPointFeatures[f] = newFeatureValue(f, val)
	}

	var pathLengthTotal float64
	for _, tree := range f.trees {
		pathLengthTotal += tree.traverse(dataPointFeatures)
	}

	avgPathLength := float64(pathLengthTotal) / float64(len(f.trees))
	score = math.Pow(2, -(avgPathLength / f.expectedAverage))
	return score
}

type IsolationTree struct {
	Root *IsolationTreeNode
}

func (t *IsolationTree) traverse(dataPoint map[Feature]FeatureValue) float64 {
	pathLength := 0
	node := t.Root
	for !node.isLeaf {
		if node.split.check(dataPoint[node.split.feature]) {
			node = node.left
		} else {
			node = node.right
		}
		pathLength++
	}

	return float64(pathLength) + avgPathLen(node.remainingSize)
}

// types of comparison
// Equals, Not Equals
type IsolationTreeNode struct {
	left          *IsolationTreeNode
	right         *IsolationTreeNode
	split         *splitCondition
	remainingSize int
	isLeaf        bool
}

const NumTrees = 100
const SampleSize = 256

func BuildForest(dataSet *DataSet) *IsolationForest {
	forest := IsolationForest{
		trees:           []*IsolationTree{},
		trainingSize:    dataSet.Size,
		features:        make(map[string]Feature),
		expectedAverage: 0,
	}

	maxDepth := uint(math.Ceil(math.Log2(float64(SampleSize))))

	for i := 0; i < NumTrees; i++ {
		forest.trees = append(forest.trees,
			&IsolationTree{Root: buildTree(dataSet.sample(SampleSize), 0, maxDepth)})
	}

	for _, feature := range dataSet.Features {
		forest.features[feature.Name] = feature
	}

	forest.expectedAverage = avgPathLen(SampleSize)

	fmt.Printf("For dataset of size %d, expected average path length is %f (max depth:%d)\n",
		SampleSize, forest.expectedAverage, maxDepth)

	return &forest
}

func buildTree(dataSet *DataSet, depth uint, maxDepth uint) *IsolationTreeNode {
	node := &IsolationTreeNode{}
	if dataSet.Size <= 1 || depth >= maxDepth {
		fmt.Printf("Leaf node at depth %d with remaining size %d\n", depth, dataSet.Size)
		node.isLeaf = true
		node.remainingSize = dataSet.Size
	} else {
		split, left, right := dataSet.Split()

		var zero *DataSet
		if left.Size == 0 {
			zero = left
		} else if right.Size == 0 {
			zero = right
		}

		if zero != nil {
			fmt.Printf("Zero split on feature %s with func %v\n", split.feature.Name, split.check)
		}

		if right.Size == 0 && left.Size == 0 {
			panic(fmt.Sprintf("Split failed: left size %d, right size %d, zero size %d",
				left.Size, right.Size, zero.Size))
		}

		node.isLeaf = true
		node.remainingSize = dataSet.Size

		node.split = split
		node.left = buildTree(left, depth+1, maxDepth)
		node.right = buildTree(right, depth+1, maxDepth)
	}

	return node
}

func harmonicNumber(n int) float64 {
	return math.Log(float64(n)) + 0.5772156649 // Euler-Mascheroni constant
}

func avgPathLen(size int) float64 {
	if size <= 1 {
		return 0
	}
	return 2*harmonicNumber(size-1) - ((2 * float64(size-1)) / float64(size))
}
