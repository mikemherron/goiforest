package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/mikemherron/goitree"
)

func main() {
	if len(os.Args) < 2 {
		panic("Please provide a file path")
	}

	filePath := os.Args[1]

	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		panic(fmt.Sprintf("File %s does not exist", filePath))
	}

	//Read file
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	csvReader := csv.NewReader(file)

	dataSet := goitree.NewDataSetFromCSV(csvReader)
	iForest := goitree.BuildForest(dataSet)

	fmt.Printf("Done:%v", iForest)

}
