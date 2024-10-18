package goitree

import (
	"encoding/csv"
	"reflect"
	"strings"
	"testing"
)

func TestDataSetFromCSV(t *testing.T) {
	r := csv.NewReader(strings.NewReader(
		`Name,Color
		apple,red
		banana,yellow
		pear,green`,
	))

	ds := NewDataSetFromCSV(r)
	if ds.Size != 3 {
		t.Errorf("Expected size 3, got %d", ds.Size)
	}

	expected := map[Feature][]FeatureValue{
		{Name: "Name", Type: FeatureTypeCategorical}: {
			{Str: "apple"},
			{Str: "banana"},
			{Str: "pear"},
		},
		{Name: "Color", Type: FeatureTypeCategorical}: {
			{Str: "red"},
			{Str: "yellow"},
			{Str: "green"},
		},
	}

	if !reflect.DeepEqual(ds.Values, expected) {
		t.Errorf("Expected %v, got %v", expected, ds.Values)
	}
}

func TestSplit(t *testing.T) {
	ds := DataSet{
		Features: []Feature{
			{Name: "Name", Type: FeatureTypeCategorical},
			{Name: "Color", Type: FeatureTypeCategorical},
		},
		Values: map[Feature][]FeatureValue{
			{Name: "Name", Type: FeatureTypeCategorical}: {
				{Str: "apple"},
				{Str: "raspberry"},
				{Str: "pear"},
			},
			{Name: "Color", Type: FeatureTypeCategorical}: {
				{Str: "red"},
				{Str: "red"},
				{Str: "green"},
			},
		},
		Size: 3,
	}

	actualLeft, actualRight := ds.splitOn(&splitCondition{
		feature: Feature{Name: "Color", Type: FeatureTypeCategorical},
		value:   "red",
	})

	expectedLeft := &DataSet{
		Features: []Feature{
			{Name: "Name", Type: FeatureTypeCategorical},
			{Name: "Color", Type: FeatureTypeCategorical},
		},
		Values: map[Feature][]FeatureValue{
			{Name: "Name", Type: FeatureTypeCategorical}: {
				{Str: "apple"},
				{Str: "raspberry"},
			},
			{Name: "Color", Type: FeatureTypeCategorical}: {
				{Str: "red"},
				{Str: "red"},
			},
		},
		Size: 2,
	}

	expectedRight := &DataSet{
		Features: []Feature{
			{Name: "Name", Type: FeatureTypeCategorical},
			{Name: "Color", Type: FeatureTypeCategorical},
		},
		Values: map[Feature][]FeatureValue{
			{Name: "Name", Type: FeatureTypeCategorical}: {
				{Str: "pear"},
			},
			{Name: "Color", Type: FeatureTypeCategorical}: {
				{Str: "green"},
			},
		},
		Size: 1,
	}

	if !reflect.DeepEqual(actualLeft, expectedLeft) {
		t.Errorf("Expected %v, got %v", expectedLeft, actualLeft)
	}

	if !reflect.DeepEqual(actualRight, expectedRight) {
		t.Errorf("Expected %v, got %v", expectedRight, actualRight)
	}
}
