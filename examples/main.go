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

	//Read file
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	dataSet := goitree.NewDataSetFromCSV(csv.NewReader(file), map[string]goitree.FeatureType{
		"Flow Duration":               goitree.FeatureTypeNumerical,
		"Total Fwd Packets":           goitree.FeatureTypeNumerical,
		"Total Backward Packets":      goitree.FeatureTypeNumerical,
		"Total Length of Fwd Packets": goitree.FeatureTypeNumerical,
		"Total Length of Bwd Packets": goitree.FeatureTypeNumerical,
		"Fwd Packet Length Max":       goitree.FeatureTypeNumerical,
		"Fwd Packet Length Min":       goitree.FeatureTypeNumerical,
		"Fwd Packet Length Mean":      goitree.FeatureTypeNumerical,
		"Fwd Packet Length Std":       goitree.FeatureTypeNumerical,
		"Bwd Packet Length Max":       goitree.FeatureTypeNumerical,
		"Bwd Packet Length Min":       goitree.FeatureTypeNumerical,
		"Bwd Packet Length Mean":      goitree.FeatureTypeNumerical,
		"Bwd Packet Length Std":       goitree.FeatureTypeNumerical,
		"Flow Bytes/s":                goitree.FeatureTypeNumerical,
		"Flow Packets/s":              goitree.FeatureTypeNumerical,
		"Flow IAT Mean":               goitree.FeatureTypeNumerical,
		"Flow IAT Std":                goitree.FeatureTypeNumerical,
		"Flow IAT Max":                goitree.FeatureTypeNumerical,
		"Flow IAT Min":                goitree.FeatureTypeNumerical,
		"Fwd IAT Total":               goitree.FeatureTypeNumerical,
		"Fwd IAT Mean":                goitree.FeatureTypeNumerical,
		"Fwd IAT Std":                 goitree.FeatureTypeNumerical,
		"Fwd IAT Max":                 goitree.FeatureTypeNumerical,
		"Fwd IAT Min":                 goitree.FeatureTypeNumerical,
		"Bwd IAT Total":               goitree.FeatureTypeNumerical,
		"Bwd IAT Mean":                goitree.FeatureTypeNumerical,
		"Bwd IAT Std":                 goitree.FeatureTypeNumerical,
		"Bwd IAT Max":                 goitree.FeatureTypeNumerical,
		"Bwd IAT Min":                 goitree.FeatureTypeNumerical,
		"Fwd PSH Flags":               goitree.FeatureTypeNumerical,
		"Bwd PSH Flags":               goitree.FeatureTypeNumerical,
		"Fwd URG Flags":               goitree.FeatureTypeNumerical,
		"Bwd URG Flags":               goitree.FeatureTypeNumerical,
		"Fwd Header Length":           goitree.FeatureTypeNumerical,
		"Bwd Header Length":           goitree.FeatureTypeNumerical,
		"Fwd Packets/s":               goitree.FeatureTypeNumerical,
		"Bwd Packets/s":               goitree.FeatureTypeNumerical,
		"Min Packet Length":           goitree.FeatureTypeNumerical,
		"Max Packet Length":           goitree.FeatureTypeNumerical,
		"Packet Length Mean":          goitree.FeatureTypeNumerical,
		"Packet Length Std":           goitree.FeatureTypeNumerical,
		"Packet Length Variance":      goitree.FeatureTypeNumerical,
		"FIN Flag Count":              goitree.FeatureTypeNumerical,
		"SYN Flag Count":              goitree.FeatureTypeNumerical,
		"RST Flag Count":              goitree.FeatureTypeNumerical,
		"PSH Flag Count":              goitree.FeatureTypeNumerical,
		"ACK Flag Count":              goitree.FeatureTypeNumerical,
		"URG Flag Count":              goitree.FeatureTypeNumerical,
		"CWE Flag Count":              goitree.FeatureTypeNumerical,
		"ECE Flag Count":              goitree.FeatureTypeNumerical,
		"Down/Up Ratio":               goitree.FeatureTypeNumerical,
		"Average Packet Size":         goitree.FeatureTypeNumerical,
		"Avg Fwd Segment Size":        goitree.FeatureTypeNumerical,
		"Avg Bwd Segment Size":        goitree.FeatureTypeNumerical,
		"Fwd Avg Bytes/Bulk":          goitree.FeatureTypeNumerical,
		"Fwd Avg Packets/Bulk":        goitree.FeatureTypeNumerical,
		"Fwd Avg Bulk Rate":           goitree.FeatureTypeNumerical,
		"Bwd Avg Bytes/Bulk":          goitree.FeatureTypeNumerical,
		"Bwd Avg Packets/Bulk":        goitree.FeatureTypeNumerical,
		"Bwd Avg Bulk Rate":           goitree.FeatureTypeNumerical,
		"Subflow Fwd Packets":         goitree.FeatureTypeNumerical,
		"Subflow Fwd Bytes":           goitree.FeatureTypeNumerical,
		"Subflow Bwd Packets":         goitree.FeatureTypeNumerical,
		"Subflow Bwd Bytes":           goitree.FeatureTypeNumerical,
		"Init_Win_bytes_forward":      goitree.FeatureTypeNumerical,
		"Init_Win_bytes_backward":     goitree.FeatureTypeNumerical,
		"act_data_pkt_fwd":            goitree.FeatureTypeNumerical,
		"min_seg_size_forward":        goitree.FeatureTypeNumerical,
		"Active Mean":                 goitree.FeatureTypeNumerical,
		"Active Std":                  goitree.FeatureTypeNumerical,
		"Active Max":                  goitree.FeatureTypeNumerical,
		"Active Min":                  goitree.FeatureTypeNumerical,
		"Idle Mean":                   goitree.FeatureTypeNumerical,
		"Idle Std":                    goitree.FeatureTypeNumerical,
		"Idle Max":                    goitree.FeatureTypeNumerical,
		"Idle Min":                    goitree.FeatureTypeNumerical,
	}, map[string]bool{"Label": true})

	forest := goitree.BuildForest(dataSet)

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

	checkList := make([]map[string]string, 0)
	// Get 80 Regular records and 20 Anomalous records
	regular := 0
	anomalous := 0
	for _, record := range allRecords {
		label := record[len(record)-1]
		include := false
		if label == "BENIGN" && regular < 80 {
			include = true
			regular++
		} else if label != "BENIGN" && anomalous < 20 {
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

		if regular == 80 && anomalous == 20 {
			break
		}
	}

	for _, record := range checkList {
		fmt.Printf("Label: %s, Score: %f\n", record["Label"], forest.Score(record))
	}
}
