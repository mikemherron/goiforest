package main

import (
	"encoding/csv"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"

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

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	dataSet, err := goitree.NewDataSetFromCSV(csv.NewReader(file),
		map[string]goitree.AttributeType{
			"V1":  goitree.AttributeTypeNumerical,
			"V2":  goitree.AttributeTypeNumerical,
			"V3":  goitree.AttributeTypeNumerical,
			"V4":  goitree.AttributeTypeNumerical,
			"V5":  goitree.AttributeTypeNumerical,
			"V6":  goitree.AttributeTypeNumerical,
			"V7":  goitree.AttributeTypeNumerical,
			"V8":  goitree.AttributeTypeNumerical,
			"V9":  goitree.AttributeTypeNumerical,
			"V10": goitree.AttributeTypeNumerical,
			"V11": goitree.AttributeTypeNumerical,
			"V12": goitree.AttributeTypeNumerical,
			"V13": goitree.AttributeTypeNumerical,
			"V14": goitree.AttributeTypeNumerical,
			"V15": goitree.AttributeTypeNumerical,
			"V16": goitree.AttributeTypeNumerical,
			"V17": goitree.AttributeTypeNumerical,
			"V18": goitree.AttributeTypeNumerical,
			"V19": goitree.AttributeTypeNumerical,
			"V20": goitree.AttributeTypeNumerical,
			"V21": goitree.AttributeTypeNumerical,
			"V22": goitree.AttributeTypeNumerical,
			"V23": goitree.AttributeTypeNumerical,
			"V24": goitree.AttributeTypeNumerical,
			"V25": goitree.AttributeTypeNumerical,
			"V26": goitree.AttributeTypeNumerical,
			"V27": goitree.AttributeTypeNumerical,
			"V28": goitree.AttributeTypeNumerical,
		},
	)

	if err != nil {
		panic(err)
	}

	// var treesDebug string
	forest := goitree.BuildForest(dataSet)
	// for i, t := range forest.Trees {
	// 	treesDebug += fmt.Sprintf("==============Tree %d==========\n", i)
	// 	treesDebug += t.String()
	// }

	// os.WriteFile("trees-debug.txt", []byte(treesDebug), 0644)

	file, err = os.Open(filePath)
	if err != nil {
		panic(err)
	}

	verifies := csv.NewReader(file)
	allRecords, err := verifies.ReadAll()
	if err != nil {
		panic(err)
	}

	headers := allRecords[0]
	for i, header := range headers {
		headers[i] = strings.TrimSpace(header)
	}

	allRecords = allRecords[1:]
	rand.Shuffle(len(allRecords), func(i, j int) {
		allRecords[i], allRecords[j] = allRecords[j], allRecords[i]
	})

	const reg = 80
	const anom = 20
	checkList := make([]map[string]string, 0)
	regular := 0
	anomalous := 0
	for _, record := range allRecords {
		label := record[len(record)-1]
		include := false
		if label == "0" && regular < reg {
			include = true
			regular++
		} else if label != "0" && anomalous < anom {
			include = true
			anomalous++
		}

		if include {
			checkRecord := make(map[string]string)
			for i, header := range headers {
				checkRecord[header] = record[i]
			}
			checkList = append(checkList, checkRecord)
		}

		if regular == reg && anomalous == anom {
			break
		}
	}

	labelAttr := goitree.Attribute{Name: "label", Type: goitree.AttributeTypeCategorical}
	scored := dataSet.CopyNoValues()
	scored.AddAttribute(labelAttr)
	for _, record := range checkList {
		result := forest.Score(record)
		attributes := result.Attributes
		attributes[labelAttr] = goitree.NewAttributeValue(labelAttr, record["Class"])
		scored.AddRow(result.Attributes)
		fmt.Printf("Label: %s, Score: %f Avg Path: %f\n", record["Class"], result.Score, result.AveragePathLength)
		// if record["label"] != "normal." {

		// 	tracesDebug := ""
		// 	for attr, val := range result.Attributes {
		// 		tracesDebug += fmt.Sprintf("%s: %s\n", attr.Name, attr.ValueToString(val))
		// 	}
		// 	tracesDebug += "\n"

		// 	for i, trace := range result.TreeTraces {
		// 		tracesDebug += fmt.Sprintf("================= Tree =========== %d\n", i)
		// 		tracesDebug += strings.Join(trace, "\n")
		// 		tracesDebug += "\n"
		// 	}
		// 	os.WriteFile(fmt.Sprintf("traces-%d.txt", i), []byte(tracesDebug), 0644)
		// }
	}

}
