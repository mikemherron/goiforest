package goitree

import "math/rand"

type IsolationForest struct {
	Trees []*IsolationTree
}

func (f *IsolationForest) Score(dataPoint DataPoint) float64 {
	score := 0.0
	pathLengths := []int{}
	for _, tree := range f.Trees {
		score += tree.traverse(dataPoint)
	}

	return score / NumTrees
}

type IsolationTree struct {
	Root *IsolationTreeNode
}

func (t *IsolationTree) traverse(dataPoint DataPoint) int {
	node := t.Root
	pathLength := 0

	for !node.IsLeaf {
		if dataPoint.Features[node.Feature] == node.Value {
			node = node.Left
		} else {
			node = node.Right
		}
		pathLength++
	}

	return pathLength
}

//types of comparison
// Equals, Not Equals
type IsolationTreeNode struct {
	Left   *IsolationTreeNode
	Right  *IsolationTreeNode
	IsLeaf bool
	// Feature to split on
	Feature FeatureName
	// Value to split on
	Value    FeatureValue
	IsEquals bool
}

type DataSet struct {
	DataPoints []DataPoint
}

func (d *DataSet) copy() *DataSet {
	dataPoints := make([]DataPoint, len(d.DataPoints))
	copy(dataPoints, d.DataPoints)
	return &DataSet{DataPoints: dataPoints}
}

func (d *DataSet) Split(feature FeatureName, value FeatureValue) (*DataSet, *DataSet) {
	left := DataSet{DataPoints: []DataPoint{}}
	right := DataSet{DataPoints: []DataPoint{}}

	for _, dataPoint := range d.DataPoints {
		if dataPoint.Features[feature] == value {
			left.DataPoints = append(left.DataPoints, dataPoint)
		} else {
			right.DataPoints = append(right.DataPoints, dataPoint)
		}
	}

	return &left, &right
}

// Should these be interfaces?
type DataPoint struct {
	// Each data point must have the same features?
	Features map[FeatureName]FeatureValue
}

type FeatureName = string

type FeatureValue struct {
	// Assume string features with categorical
	// selection only for now
	Str string
}

const NumTrees = 100

type FeatureSet struct {
	Names      []FeatureName
	Values     map[FeatureName]map[FeatureValue]bool
	ValuesList map[FeatureName][]FeatureValue
}

func (f *FeatureSet) AddFeature(name FeatureName) {
	if f.Has(name) {
		panic("Feature already exists")
	}
	f.Names = append(f.Names, name)
	f.Values[name] = map[FeatureValue]bool{}
}

func (f *FeatureSet) AddValue(name FeatureName, value FeatureValue) FeatureValue {
	if _, ok := f.Values[name]; !ok {
		panic("Feature does not exist")
	}
	f.Values[name][value] = true
	return value
}

func (f *FeatureSet) Has(name FeatureName) bool {
	_, ok := f.Values[name]
	return ok
}

func (f *FeatureSet) RandomFeatureSplit() (FeatureName, FeatureValue) {
	feature := f.Names[rand.Intn(len(f.Names))]
	value := f.ValuesList[feature][rand.Intn(len(f.ValuesList[feature]))]

	return feature, value
}

func newFeatureSet(dataSet *DataSet) *FeatureSet {
	featureSet := FeatureSet{
		Values:     map[FeatureName]map[FeatureValue]bool{},
		ValuesList: map[FeatureName][]FeatureValue{},
		Names:      []FeatureName{},
	}

	for feature := range dataSet.DataPoints[0].Features {
		featureSet.AddFeature(feature)
	}

	// Collect all features and values
	for _, dataPoint := range dataSet.DataPoints {
		for featureName, featureValue := range dataPoint.Features {
			if !featureSet.Has(featureName) {
				panic("Feature does not exist")
			}
			featureSet.AddValue(featureName, featureValue)
		}

		for featureName := range featureSet.Values {
			if _, ok := dataPoint.Features[featureName]; !ok {
				panic("Data point does not have all features")
			}
		}
	}

	// Build Values List
	for featureName, values := range featureSet.Values {
		for value := range values {
			featureSet.ValuesList[featureName] = append(featureSet.ValuesList[featureName], value)
		}
	}

	return &featureSet
}

func Build(dataSet *DataSet) *IsolationForest {
	featureSet := newFeatureSet(dataSet)
	forest := IsolationForest{Trees: []*IsolationTree{}}
	for i := 0; i < NumTrees; i++ {
		root := &IsolationTreeNode{}
		buildTree(root, dataSet.copy(), featureSet)

		forest.Trees = append(forest.Trees)
	}

	return &forest
}

func buildTree(parent *IsolationTreeNode, dataSet *DataSet, featureSet *FeatureSet) {
	if len(dataSet.DataPoints) <= 1 {
		parent.IsLeaf = true
		return
	}

	parent.Left = &IsolationTreeNode{}
	parent.Right = &IsolationTreeNode{}

	// TODO: Must be a much more efficient way to create new
	// feature set based on remaining data
	left, right := dataSet.Split(featureSet.RandomFeatureSplit())
	// NEXT: Capture the feature and value used to split so we can
	// use it later for scoring
	buildTree(parent.Left, left, newFeatureSet(left))
	buildTree(parent.Right, right, newFeatureSet(right))
}
