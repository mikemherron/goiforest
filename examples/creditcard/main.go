package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/mikemherron/goiforest"
)

// This example demonstrates how to use the Isolation Forest to score a dataset
// using the Credit Card Fraud dataset from Kaggle. The dataset is available at
// https://www.kaggle.com/mlg-ulb/creditcardfraud - to run the example download
// the dataset and provide the path to the CSV file as the first argument to the
// program.
func main() {
	if len(os.Args) < 2 {
		panic("Please provide a file path")
	}

	filePath := os.Args[1]

	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		panic(fmt.Sprintf("File %s does not exist", filePath))
	}

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	dataSet, err := goiforest.NewDataSetFromCSV(csv.NewReader(file),
		map[string]goiforest.AttributeType{
			"V1":    goiforest.AttributeTypeNumerical,
			"V2":    goiforest.AttributeTypeNumerical,
			"V3":    goiforest.AttributeTypeNumerical,
			"V4":    goiforest.AttributeTypeNumerical,
			"V5":    goiforest.AttributeTypeNumerical,
			"V6":    goiforest.AttributeTypeNumerical,
			"V7":    goiforest.AttributeTypeNumerical,
			"V8":    goiforest.AttributeTypeNumerical,
			"V9":    goiforest.AttributeTypeNumerical,
			"V10":   goiforest.AttributeTypeNumerical,
			"V11":   goiforest.AttributeTypeNumerical,
			"V12":   goiforest.AttributeTypeNumerical,
			"V13":   goiforest.AttributeTypeNumerical,
			"V14":   goiforest.AttributeTypeNumerical,
			"V15":   goiforest.AttributeTypeNumerical,
			"V16":   goiforest.AttributeTypeNumerical,
			"V17":   goiforest.AttributeTypeNumerical,
			"V18":   goiforest.AttributeTypeNumerical,
			"V19":   goiforest.AttributeTypeNumerical,
			"V20":   goiforest.AttributeTypeNumerical,
			"V21":   goiforest.AttributeTypeNumerical,
			"V22":   goiforest.AttributeTypeNumerical,
			"V23":   goiforest.AttributeTypeNumerical,
			"V24":   goiforest.AttributeTypeNumerical,
			"V25":   goiforest.AttributeTypeNumerical,
			"V26":   goiforest.AttributeTypeNumerical,
			"V27":   goiforest.AttributeTypeNumerical,
			"V28":   goiforest.AttributeTypeNumerical,
			"Class": goiforest.AttributeTypeCategorical,
		},
	)

	if err != nil {
		panic(err)
	}

	// Exclude the label attribute from the training data
	trainingDataSet, err := dataSet.Excluding("Class")
	if err != nil {
		panic(err)
	}

	// Build the forest
	forest := goiforest.BuildForest(trainingDataSet)

	// Grab 80 regular records and 20 anomalous records
	regularRecords := dataSet.Filter(
		func(record map[string]goiforest.AttributeValue) bool {
			return record["Class"].Str == "0"
		}).Sample(80)

	anomalousRecords := dataSet.Filter(
		func(record map[string]goiforest.AttributeValue) bool {
			return record["Class"].Str != "0"
		}).Sample(20)

	// Merge the two sets together
	testingSet, err := dataSet.CopyNoValues().Merge(anomalousRecords, regularRecords)
	if err != nil {
		panic(err)
	}

	// Score the testing set
	for i := 0; i < testingSet.Size; i++ {
		record := testingSet.GetRowPlain(i)
		result := forest.Score(record)
		fmt.Printf("Label: %s, Score: %f Avg Path: %f\n", record["Class"],
			result.Score, result.AveragePathLength)
	}

}
