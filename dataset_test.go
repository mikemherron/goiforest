package goiforest

import (
	"encoding/csv"
	"reflect"
	"strings"
	"testing"
)

func TestDataSetFromCSV(t *testing.T) {
	r := csv.NewReader(strings.NewReader(
		`Name,Color,Ignore,Cost
		apple,red,x,0.5
		banana,yellow,x,0.2
		pear,green,x,0.8`,
	))

	ds, err := NewDataSetFromCSV(r,
		map[string]AttributeType{
			"Name":  AttributeTypeCategorical,
			"Color": AttributeTypeCategorical,
			"Cost":  AttributeTypeNumerical,
		})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if ds.Size != 3 {
		t.Errorf("Expected size 3, got %d", ds.Size)
	}

	expected := map[Attribute][]AttributeValue{
		{Name: "Name", Type: AttributeTypeCategorical}: {
			{Str: "apple"},
			{Str: "banana"},
			{Str: "pear"},
		},
		{Name: "Color", Type: AttributeTypeCategorical}: {
			{Str: "red"},
			{Str: "yellow"},
			{Str: "green"},
		},
		{Name: "Cost", Type: AttributeTypeNumerical}: {
			{Num: 0.5},
			{Num: 0.2},
			{Num: 0.8},
		},
	}

	if !reflect.DeepEqual(ds.Values, expected) {
		t.Errorf("Expected %v, got %v", expected, ds.Values)
	}
}

func TestSplit(t *testing.T) {
	ds := DataSet{
		Attributes: []Attribute{
			{Name: "Name", Type: AttributeTypeCategorical},
			{Name: "Color", Type: AttributeTypeCategorical},
		},
		Values: map[Attribute][]AttributeValue{
			{Name: "Name", Type: AttributeTypeCategorical}: {
				{Str: "apple"},
				{Str: "raspberry"},
				{Str: "pear"},
			},
			{Name: "Color", Type: AttributeTypeCategorical}: {
				{Str: "red"},
				{Str: "red"},
				{Str: "green"},
			},
		},
		Size: 3,
	}

	actualLeft, actualRight := ds.splitOn(&splitCondition{
		attribute: Attribute{Name: "Color", Type: AttributeTypeCategorical},
		strVal:    "red",
	})

	expectedLeft := &DataSet{
		Attributes: []Attribute{
			{Name: "Name", Type: AttributeTypeCategorical},
			{Name: "Color", Type: AttributeTypeCategorical},
		},
		Values: map[Attribute][]AttributeValue{
			{Name: "Name", Type: AttributeTypeCategorical}: {
				{Str: "apple"},
				{Str: "raspberry"},
			},
			{Name: "Color", Type: AttributeTypeCategorical}: {
				{Str: "red"},
				{Str: "red"},
			},
		},
		Size: 2,
	}

	expectedRight := &DataSet{
		Attributes: []Attribute{
			{Name: "Name", Type: AttributeTypeCategorical},
			{Name: "Color", Type: AttributeTypeCategorical},
		},
		Values: map[Attribute][]AttributeValue{
			{Name: "Name", Type: AttributeTypeCategorical}: {
				{Str: "pear"},
			},
			{Name: "Color", Type: AttributeTypeCategorical}: {
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
