package goitree

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"strconv"
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

func newFeatureValue(f Feature, v string) FeatureValue {
	if f.Type == FeatureTypeCategorical {
		return FeatureValue{Str: v}
	} else if f.Type == FeatureTypeNumerical {
		num, err := strconv.ParseFloat(v, 64)
		if err != nil {
			panic(fmt.Sprintf("Error parsing value %v as float", v))
		}
		return FeatureValue{Num: num}
	}
	panic("Unknown feature type")
}

type DataSet struct {
	Features []Feature
	Values   map[Feature][]FeatureValue
	Size     int
}

func NewDataSetFromCSV(r *csv.Reader, types map[string]FeatureType, exclude map[string]bool) *DataSet {
	ds := DataSet{
		Values: map[Feature][]FeatureValue{},
	}

	header, err := r.Read()
	if errors.Is(err, io.EOF) {
		panic("Empty CSV")
	} else if err != nil {
		panic(err)
	}

	ds.Features = make([]Feature, 0, len(header))

	featureIdx := map[int]Feature{}
	for i, name := range header {
		name = strings.TrimSpace(name)
		if exclude[name] {
			continue
		}

		t := FeatureTypeCategorical
		if typ, ok := types[name]; ok {
			t = typ
		}

		feature := Feature{
			Name: name,
			Type: t,
		}
		ds.Features = append(ds.Features, feature)
		ds.Values[feature] = []FeatureValue{}

		featureIdx[i] = feature
	}

	for {
		record, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			panic(err)
		}

		for i, value := range record {
			feature, ok := featureIdx[i]
			if !ok { // Excluded features wont have a mapping
				continue
			}

			featureValue := FeatureValue{}
			if feature.Type == FeatureTypeNumerical {
				featureValue.Num, err = strconv.ParseFloat(value, 64)
				if err != nil {
					panic(fmt.Sprintf("Error parsing value %v as float", value))
				}
			} else if feature.Type == FeatureTypeCategorical {
				featureValue.Str = strings.TrimSpace(value)
			}

			ds.Values[feature] = append(ds.Values[feature], featureValue)
		}
		ds.Size++
	}

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
		//Assumes Feature slice is immutable, should maybe copy
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

type splitCheck func(val FeatureValue) bool

func splitCategory(splitVal string) splitCheck {
	return func(val FeatureValue) bool {
		return val.Str == splitVal
	}
}

func splitNumerical(splitVal float64) splitCheck {
	return func(val FeatureValue) bool {
		return val.Num >= splitVal
	}
}

type splitCondition struct {
	feature Feature
	check   splitCheck
}

// TODO Support numeric ranges also
func (d *DataSet) Split() (*splitCondition, *DataSet, *DataSet) {
	// slog.Info("Splitting Data Set",
	// 	"feature_count", len(d.Features),
	// 	"size", d.Size,
	// )

	var splitFunc splitCheck
	feature := d.Features[rand.Intn(len(d.Features))]
	if feature.Type == FeatureTypeCategorical {
		splitVal := d.Values[feature][rand.Intn(len(d.Values[feature]))]
		splitFunc = splitCategory(splitVal.Str)
	} else if feature.Type == FeatureTypeNumerical {
		min := math.Inf(1)
		max := math.Inf(-1)
		for _, value := range d.Values[feature] {
			if value.Num < min {
				min = value.Num
			}
			if value.Num > max {
				max = value.Num
			}
		}

		splitVal := min + (rand.Float64() * (max - min))
		splitFunc = splitNumerical(splitVal)
	}

	//TODO: check if feature has only one value left in dataset?
	// e.g if we only have a single useragent left in all the data, should we
	// still do a split based on useragent? That will mean we end up with a
	// empty branch
	condition := &splitCondition{
		feature: feature,
		check:   splitFunc,
	}

	matched, notMatched := d.splitOn(condition)

	return condition, matched, notMatched
}

func (d *DataSet) splitOn(condition *splitCondition) (*DataSet, *DataSet) {
	matched := d.copyNoValues()
	notMatches := d.copyNoValues()

	for i := 0; i < d.Size; i++ {
		row := d.row(i)
		if condition.check(row[condition.feature]) {
			matched.add(row)
		} else {
			notMatches.add(row)
		}
	}

	return matched, notMatches
}
