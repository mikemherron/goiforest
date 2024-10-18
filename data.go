package goitree

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"strings"
)

type FeatureType int

const FeatureTypeCategorical FeatureType = 0
const FeatureTypeNumerical FeatureType = 1

type Feature struct {
	Name string
	Type FeatureType
}

type FeatureValue struct {
	Str string
	Num float64
}

type DataSet struct {
	Features []Feature
	Values   map[Feature][]FeatureValue
	Size     int
}

func NewDataSetFromCSV(r *csv.Reader) *DataSet {
	ds := DataSet{
		Values: map[Feature][]FeatureValue{},
	}

	for {
		record, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			panic(err)
		}

		if ds.Size == 0 { // Header row
			ds.Features = make([]Feature, len(record))
			for i, name := range record {
				feature := Feature{
					Name: name,
					Type: FeatureTypeCategorical,
				}
				ds.Features[i] = feature
				ds.Values[feature] = []FeatureValue{}
			}
		} else {
			if len(record) != len(ds.Features) {
				panic(fmt.Sprintf("Row %d has %d columns, expected %d",
					ds.Size, len(record), len(ds.Features)))
			}

			for i, value := range record {
				ds.Values[ds.Features[i]] = append(ds.Values[ds.Features[i]],
					FeatureValue{Str: strings.TrimSpace(value)})
			}
		}
		ds.Size++
	}

	ds.Size-- //lol
	return &ds
}

func (d *DataSet) sample(size int) *DataSet {
	if size > d.Size {
		panic("Size greater than dataset size")
	}

	cp := d.copyNoValues()
	copied := map[int]bool{}
	for i := 0; i < size; i++ {
		// TODO: There must be a better way to do this..
		var randIdx int
		for {
			randIdx = rand.Intn(d.Size)
			if _, ok := copied[randIdx]; !ok {
				break
			}
		}
		copied[randIdx] = true
		cp.add(d.row(rand.Intn(d.Size)))
	}

	return cp
}

func (d *DataSet) copyNoValues() *DataSet {
	cp := DataSet{
		//Assumes Features are immutable, should maybe copy
		Features: d.Features,
		Values:   map[Feature][]FeatureValue{},
	}

	for feature := range d.Values {
		cp.Values[feature] = []FeatureValue{}
	}

	return &cp
}

func (d *DataSet) row(idx int) map[Feature]FeatureValue {
	row := map[Feature]FeatureValue{}
	for feature, values := range d.Values {
		row[feature] = values[idx]
	}
	return row
}

func (d *DataSet) add(row map[Feature]FeatureValue) {
	//TODO check has all features
	for feature, value := range row {
		if _, ok := d.Values[feature]; !ok {
			panic(fmt.Sprintf("Feature %v does not exist in dataset", feature))
		}
		d.Values[feature] = append(d.Values[feature], value)
	}
	d.Size++
}

type splitCondition struct {
	feature Feature
	//Always equal right now (goes to left, right now)
	value string
}

// TODO Support numeric ranges also
func (d *DataSet) Split() (*splitCondition, *DataSet, *DataSet) {
	slog.Info("Splitting Data Set",
		"feature_count", len(d.Features),
		"size", d.Size,
	)

	feature := d.Features[rand.Intn(len(d.Features))]
	//TODO: if numerical work out range from values
	//TODO: check if feature has only one value?
	condition := &splitCondition{
		feature: feature,
		value:   d.Values[feature][rand.Intn(len(d.Values[feature]))].Str,
	}

	matched, notMatches := d.splitOn(condition)

	return condition, matched, notMatches
}

func (d *DataSet) splitOn(condition *splitCondition) (*DataSet, *DataSet) {
	matched := d.copyNoValues()
	notMatches := d.copyNoValues()

	for i := 0; i < d.Size; i++ {
		row := d.row(i)
		if row[condition.feature].Str == condition.value {
			matched.add(row)
		} else {
			notMatches.add(row)
		}
	}

	return matched, notMatches
}
