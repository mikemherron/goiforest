package goiforest

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

type IsolationForest struct {
	Trees           []*IsolationTree
	attributes      map[string]Attribute
	expectedAverage float64
}

type ScoreResult struct {
	Score             float64
	Attributes        map[Attribute]AttributeValue
	AveragePathLength float64
	TreeTraces        [][]string
}

func (f *IsolationForest) Score(dataPoint map[string]string) ScoreResult {
	score := 0.0

	dataPointAttributes := make(map[Attribute]AttributeValue)
	for _, f := range f.attributes {
		val, exists := dataPoint[f.Name]
		if !exists {
			panic(fmt.Sprintf("Attribute %s not found on %v", f.Name, dataPoint))
		}
		dataPointAttributes[f] = NewAttributeValue(f, val)
	}

	traces := make([][]string, len(f.Trees))
	var pathLengthTotal float64
	for i, tree := range f.Trees {
		n, trace := tree.traverse(dataPointAttributes)
		traces[i] = trace
		pathLengthTotal += n
	}

	avgPathLength := float64(pathLengthTotal) / float64(len(f.Trees))
	score = math.Pow(2, (-avgPathLength / f.expectedAverage))
	return ScoreResult{
		Score:             score,
		Attributes:        dataPointAttributes,
		AveragePathLength: avgPathLength,
		TreeTraces:        traces,
	}
}

type IsolationTree struct {
	Root *IsolationTreeNode
}

func (t *IsolationTree) String() string {
	return t.Root.String(0)
}

func (t *IsolationTree) traverse(dataPoint map[Attribute]AttributeValue) (float64, []string) {
	var pathLength float64 = 0.0
	var traces []string
	node := t.Root
	for !node.isLeaf {
		att := node.split.attribute
		val := dataPoint[att]
		if node.split.check(val) {
			traces = append(traces, fmt.Sprintf("%s (%s)", node.split.String(false), att.ValueToString(val)))
			node = node.left
		} else {
			traces = append(traces, fmt.Sprintf("%s (%s)", node.split.String(true), att.ValueToString(val)))
			node = node.right
		}
		pathLength++
	}

	traces = append(traces, fmt.Sprintf("Hit root, path length %f, remaining size: %d\n", pathLength, node.remainingSize))

	return pathLength + avgPathLen(node.remainingSize), traces
}

type IsolationTreeNode struct {
	left          *IsolationTreeNode
	right         *IsolationTreeNode
	split         *splitCondition
	remainingSize int
	isLeaf        bool
}

func (n *IsolationTreeNode) String(depth int) string {
	prefix := fmt.Sprintf("%"+strconv.Itoa(depth)+"s", "")
	var s string
	if n.isLeaf {
		s += prefix + fmt.Sprintf("Leaf [%d]\n", n.remainingSize)
	} else {
		s += prefix + fmt.Sprintf("Node [%d]\n", n.remainingSize)
		s += prefix + fmt.Sprintf("%s\n", n.split.String(false))
		s += prefix + n.left.String(depth+1)
		s += prefix + fmt.Sprintf("%s\n", n.split.String(true))
		s += prefix + n.right.String(depth+1)
	}
	return s
}

const NumTrees = 100
const SampleSize = 256

func BuildForest(dataSet *DataSet) *IsolationForest {
	forest := IsolationForest{
		Trees:           []*IsolationTree{},
		attributes:      make(map[string]Attribute),
		expectedAverage: avgPathLen(SampleSize),
	}

	maxDepth := uint(math.Ceil(math.Log2(float64(SampleSize))))

	for i := 0; i < NumTrees; i++ {
		forest.Trees = append(forest.Trees,
			&IsolationTree{Root: buildTree(dataSet.Sample(SampleSize), 0, maxDepth, make(map[Attribute]bool))})
	}

	for _, feature := range dataSet.Attributes {
		forest.attributes[feature.Name] = feature
	}

	return &forest
}

func buildTree(dataSet *DataSet, depth uint, maxDepth uint, exclude map[Attribute]bool) *IsolationTreeNode {
	node := &IsolationTreeNode{}
	node.remainingSize = dataSet.Size
	if dataSet.Size <= 1 || depth >= maxDepth {
		node.isLeaf = true
	} else {
		split, left, right, err := dataSet.Split(exclude)
		if errors.Is(err, ErrNotSplittable) {
			node.isLeaf = true
		} else {
			node.split = split
			// If a split has resulted in a dataset with no elements on one side,
			// don't use that split again in this tree. This can happens when all
			// elements have the same value for an attribute.
			if left.Size == 0 || right.Size == 0 {
				exclude = addExclusion(exclude, split.attribute)
			}
			node.left = buildTree(left, depth+1, maxDepth, exclude)
			node.right = buildTree(right, depth+1, maxDepth, exclude)
		}
	}

	return node
}

func addExclusion(exclude map[Attribute]bool, attr Attribute) map[Attribute]bool {
	newMap := make(map[Attribute]bool)
	for k, v := range exclude {
		newMap[k] = v
	}
	newMap[attr] = true
	return newMap
}

func harmonicNumber(n int) float64 {
	return math.Log(float64(n)) + 0.5772156649
}

func avgPathLen(size int) float64 {
	if size <= 1 {
		return 0
	}
	return 2*harmonicNumber(size-1) - ((2 * float64(size-1)) / float64(size))
}
