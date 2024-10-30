package goiforest

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

type AttributeType int

const AttributeTypeCategorical AttributeType = 0
const AttributeTypeNumerical AttributeType = 1

type Attribute struct {
	Name string
	Type AttributeType
}

type AttributeValue struct {
	Str string
	Num float64
}

func (a Attribute) ValueToString(v AttributeValue) string {
	if a.Type == AttributeTypeCategorical {
		return v.Str
	} else if a.Type == AttributeTypeNumerical {
		return fmt.Sprintf("%f", v.Num)
	}
	panic("Unknown feature type")
}

func NewAttributeValue(f Attribute, v string) AttributeValue {
	if f.Type == AttributeTypeCategorical {
		return AttributeValue{Str: v}
	} else if f.Type == AttributeTypeNumerical {
		num, err := strconv.ParseFloat(v, 64)
		if err != nil {
			panic(fmt.Sprintf("Error parsing value %v as float", v))
		}
		return AttributeValue{Num: num}
	}
	panic("Unknown feature type")
}

type DataSet struct {
	Attributes []Attribute
	Values     map[Attribute][]AttributeValue
	Size       int
}

func NewDataSet() *DataSet {
	return &DataSet{
		Attributes: []Attribute{},
		Values:     map[Attribute][]AttributeValue{},
	}
}

func NewDataSetFromCSV(r *csv.Reader, attributes map[string]AttributeType) (*DataSet, error) {
	header, err := r.Read()
	if errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("empty CSV file")
	} else if err != nil {
		return nil, fmt.Errorf("error reading CSV header: %w", err)
	}

	ds := DataSet{
		Values:     map[Attribute][]AttributeValue{},
		Attributes: make([]Attribute, 0, len(header)),
	}

	remainingAttributes := map[string]bool{}
	for name := range attributes {
		remainingAttributes[name] = true
	}

	attributeIdx := map[int]Attribute{}
	for i, name := range header {
		name = strings.TrimSpace(name)
		if _, ok := attributes[name]; !ok {
			continue
		}

		delete(remainingAttributes, name)

		attribute := Attribute{
			Name: name,
			Type: attributes[name],
		}
		ds.Attributes = append(ds.Attributes, attribute)
		ds.Values[attribute] = []AttributeValue{}

		attributeIdx[i] = attribute
	}

	if len(remainingAttributes) > 0 {
		return nil, fmt.Errorf("one more attributes not found in CSV file: %v", remainingAttributes)
	}

	for {
		record, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("error reading CSV row: %w", err)
		}

		for i, value := range record {
			attribute, ok := attributeIdx[i]
			if !ok {
				continue
			}

			attributeValue := AttributeValue{}
			if attribute.Type == AttributeTypeNumerical {
				attributeValue.Num, err = strconv.ParseFloat(value, 64)
				if err != nil {
					return nil, fmt.Errorf("error parsing value %v as float", value)
				}
			} else if attribute.Type == AttributeTypeCategorical {
				attributeValue.Str = strings.TrimSpace(value)
			}

			ds.Values[attribute] = append(ds.Values[attribute], attributeValue)
		}
		ds.Size++
	}

	return &ds, nil
}

type DataSetStats struct {
	Attributes []AttributeStats
}

func (dss DataSetStats) String() string {
	var sb strings.Builder
	sb.WriteString(
		fmt.Sprintf("%-20s %-20s %-20s %-20s %-20s %-20s %-20s\n",
			"Attribute",
			"Kurtosis",
			"Mean",
			"Variance",
			"Min",
			"Max",
			"Unique"))
	for _, attrStats := range dss.Attributes {
		sb.WriteString(fmt.Sprintf(
			"%-20s %-20.4f %-20.4f %-20.4f %-20.4f %-20.4f %-20d\n",
			attrStats.Attribute.Name,
			attrStats.Kurtosis,
			attrStats.Mean,
			attrStats.Variance,
			attrStats.Min,
			attrStats.Max,
			attrStats.Unique))
	}
	return sb.String()
}

type AttributeStats struct {
	Attribute Attribute
	Kurtosis  float64
	Mean      float64
	Variance  float64
	Unique    int
	Min       float64
	Max       float64
}

func (d *DataSet) Stats() DataSetStats {
	stats := DataSetStats{
		Attributes: make([]AttributeStats, len(d.Attributes)),
	}

	for _, attribute := range d.Attributes {
		attributeStats := AttributeStats{
			Attribute: attribute,
		}

		if attribute.Type != AttributeTypeNumerical {
			continue
		}

		values := make([]float64, d.Size)
		unique := make(map[float64]bool, 0)
		for _, v := range d.Values[attribute] {
			values = append(values, v.Num)
			unique[v.Num] = true
		}

		attributeStats.Kurtosis = kurtosis(values)
		attributeStats.Mean = mean(values)
		attributeStats.Variance = variance(values)
		attributeStats.Min = min(values)
		attributeStats.Max = max(values)
		attributeStats.Unique = len(unique)
		stats.Attributes = append(stats.Attributes, attributeStats)
	}
	return stats
}

func (d *DataSet) Filter(f func(map[string]AttributeValue) bool) *DataSet {
	cp := d.CopyNoValues()
	for i := 0; i < d.Size; i++ {
		row := d.GetRowWithNames(i)
		if f(row) {
			cp.AddRow(d.GetRow(i))
		}
	}
	return cp
}

func (d *DataSet) Limit(n int) *DataSet {
	cp := d.CopyNoValues()
	if n > d.Size {
		n = d.Size
	}
	for i := 0; i < n; i++ {
		cp.AddRow(d.GetRow(i))
	}
	return cp
}

func (d *DataSet) Shuffle() *DataSet {
	cp := d.Copy()
	rand.Shuffle(d.Size, func(i, j int) {
		for _, attr := range d.Attributes {
			cp.Values[attr][i], cp.Values[attr][j] = cp.Values[attr][j], cp.Values[attr][i]
		}
	})

	return cp
}

func (d *DataSet) Merge(dataSets ...*DataSet) (*DataSet, error) {
	cp := d.Copy()
	for i, toMerge := range dataSets {
		if err := d.attributesEqual(toMerge); err != nil {
			return nil, fmt.Errorf("data set %d attributes do not match: %v", i, err)
		}
		for i := 0; i < toMerge.Size; i++ {
			cp.AddRow(toMerge.GetRow(i))
		}
	}

	return cp, nil
}

func (d *DataSet) Sample(size int) *DataSet {
	if size > d.Size {
		size = d.Size
	}

	cp := d.CopyNoValues()
	copied := map[int]bool{}

	for i := 0; i < size; i++ {
		var randIdx int
		idxValid := false
		for !idxValid {
			randIdx = rand.Intn(d.Size)
			idxValid = !copied[randIdx]
		}
		copied[randIdx] = true
		//TODO: copy row here
		cp.AddRow(d.GetRow(randIdx))
	}

	return cp
}

func (d *DataSet) With(attributes []string) (*DataSet, error) {
	attributeSet := d.attributeSet()
	dsCopy := NewDataSet()
	for i, attrName := range attributes {
		if _, ok := attributeSet[attrName]; !ok {
			return nil, fmt.Errorf("attribute %v not found in dataset", attrName)
		}
		dsCopy.Attributes[i] = attributeSet[attrName]
	}

	for i := 0; i < d.Size; i++ {
		row := d.GetRow(i)
		rowCp := map[Attribute]AttributeValue{}
		for _, attr := range dsCopy.Attributes {
			rowCp[attr] = row[attr]
		}
		dsCopy.AddRow(rowCp)
	}

	return dsCopy, nil
}

func (d *DataSet) Excluding(attributes ...string) (*DataSet, error) {
	if err := d.containsAttributes(attributes); err != nil {
		return nil, err
	}

	if len(attributes) == len(d.Attributes) {
		return nil, fmt.Errorf("cannot exclude all attributes")
	}

	excludeSet := map[string]bool{}
	for _, attr := range attributes {
		excludeSet[attr] = true
	}

	cp := NewDataSet()
	for _, attr := range d.Attributes {
		if _, ok := excludeSet[attr.Name]; ok {
			continue
		}
		cp.Attributes = append(cp.Attributes, attr)
		cp.Values[attr] = []AttributeValue{}
	}

	for i := 0; i < d.Size; i++ {
		row := d.GetRow(i)
		newRow := map[Attribute]AttributeValue{}
		for _, feature := range d.Attributes {
			if _, ok := excludeSet[feature.Name]; ok {
				continue
			}
			newRow[feature] = row[feature]
		}
		cp.AddRow(newRow)
	}

	return cp, nil
}

func (d *DataSet) CopyNoValues() *DataSet {
	cp := &DataSet{
		Attributes: make([]Attribute, len(d.Attributes)),
		Values:     map[Attribute][]AttributeValue{},
	}
	for i, attribute := range d.Attributes {
		cp.Attributes[i] = attribute
		cp.Values[attribute] = []AttributeValue{}
	}

	return cp
}

func (d *DataSet) Copy() *DataSet {
	cp := d.CopyNoValues()
	for i := 0; i < d.Size; i++ {
		row := d.GetRow(i)
		cp.AddRow(row)
	}

	return cp
}

func (d *DataSet) containsAttributes(name []string) error {
	attributes := d.attributeSet()
	for _, n := range name {
		if _, ok := attributes[n]; !ok {
			return fmt.Errorf("attribute %v not found in dataset", n)
		}
	}
	return nil
}

func (d *DataSet) attributesEqual(other *DataSet) error {
	attributes := d.attributeSet()
	otherAttributes := other.attributeSet()

	missing := make([]string, 0)
	// Check all attributes in this data set are in the other data set
	for name := range attributes {
		if _, ok := otherAttributes[name]; !ok {
			missing = append(missing, name)
		}
	}

	// Check all attributes in the other data set are in this data set
	for name := range otherAttributes {
		if _, ok := attributes[name]; !ok {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("attributes not equal: %v", missing)
	}

	return nil
}

func (d *DataSet) attributeSet() map[string]Attribute {
	attributes := make(map[string]Attribute)
	for _, attr := range d.Attributes {
		attributes[attr.Name] = attr
	}
	return attributes
}

// func (d *DataSet) AddAttribute(a Attribute) {
// 	if d.Size > 0 {
// 		panic("Cannot add attribute to non-empty dataset")
// 	}

// 	d.Attributes = append(d.Attributes, a)
// 	d.Values[a] = []AttributeValue{}
// }

func (d *DataSet) GetRow(idx int) map[Attribute]AttributeValue {
	row := map[Attribute]AttributeValue{}
	for attr, values := range d.Values {
		row[attr] = values[idx]
	}
	return row
}

func (d *DataSet) GetRowWithNames(idx int) map[string]AttributeValue {
	row := map[string]AttributeValue{}
	for attr, values := range d.Values {
		row[attr.Name] = values[idx]
	}
	return row
}

func (d *DataSet) GetRowPlain(idx int) map[string]string {
	row := map[string]string{}
	for attr, values := range d.Values {
		row[attr.Name] = attr.ValueToString(values[idx])
	}
	return row
}

func (d *DataSet) AddRow(row map[Attribute]AttributeValue) {
	for _, attr := range d.Attributes {
		if _, ok := row[attr]; !ok {
			panic(fmt.Sprintf("Feature %v missing in row", attr))
		}
	}

	for feature, value := range row {
		if _, ok := d.Values[feature]; !ok {
			panic(fmt.Sprintf("Feature %v does not exist in dataset", feature))
		}
		d.Values[feature] = append(d.Values[feature], value)
	}
	d.Size++
}

func (d *DataSet) ToCSV(w io.Writer) error {
	writer := csv.NewWriter(w)

	header := make([]string, len(d.Attributes))
	for i, attribute := range d.Attributes {
		header[i] = attribute.Name
	}
	err := writer.Write(header)
	if err != nil {
		return err
	}

	for i := 0; i < d.Size; i++ {
		record := make([]string, len(d.Attributes))
		for j, feature := range d.Attributes {
			record[j] = feature.ValueToString(d.Values[feature][i])
		}

		err = writer.Write(record)
		if err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

type splitCondition struct {
	attribute Attribute
	strVal    string
	numVal    float64
}

func (s *splitCondition) check(val AttributeValue) bool {
	if s.attribute.Type == AttributeTypeCategorical {
		return val.Str == s.strVal
	} else if s.attribute.Type == AttributeTypeNumerical {
		return val.Num >= s.numVal
	}
	panic("Unknown feature type")
}

func (s *splitCondition) String(inverse bool) string {
	if s.attribute.Type == AttributeTypeCategorical {
		if inverse {
			return fmt.Sprintf("%s != %s", s.attribute.Name, s.strVal)
		} else {
			return fmt.Sprintf("%s == %s", s.attribute.Name, s.strVal)
		}
	} else {
		if inverse {
			return fmt.Sprintf("%s < %f", s.attribute.Name, s.numVal)
		} else {
			return fmt.Sprintf("%s >= %f", s.attribute.Name, s.numVal)
		}
	}
}

var ErrNotSplittable = errors.New("dataset not splittable")

func (d *DataSet) Split(exclude map[Attribute]bool) (*splitCondition, *DataSet, *DataSet, error) {

	splittable := make([]Attribute, 0)
	for _, attr := range d.Attributes {
		if _, ok := exclude[attr]; !ok {
			splittable = append(splittable, attr)
		}
	}

	if len(splittable) == 0 {
		return nil, nil, nil, ErrNotSplittable
	}

	splitAttr := splittable[rand.Intn(len(splittable))]
	condition := &splitCondition{attribute: splitAttr}
	if splitAttr.Type == AttributeTypeCategorical {
		condition.strVal = d.Values[splitAttr][rand.Intn(len(d.Values[splitAttr]))].Str
	} else if splitAttr.Type == AttributeTypeNumerical {
		min := math.Inf(1)
		max := math.Inf(-1)
		for _, value := range d.Values[splitAttr] {
			if value.Num < min {
				min = value.Num
			}
			if value.Num > max {
				max = value.Num
			}
		}
		condition.numVal = min + (rand.Float64() * (max - min))
	}

	matched, notMatched := d.splitOn(condition)

	return condition, matched, notMatched, nil
}

func (d *DataSet) splitOn(condition *splitCondition) (*DataSet, *DataSet) {
	matched := d.CopyNoValues()
	notMatched := d.CopyNoValues()

	for i := 0; i < d.Size; i++ {
		row := d.GetRow(i)
		if condition.check(row[condition.attribute]) {
			matched.AddRow(row)
		} else {
			notMatched.AddRow(row)
		}
	}

	return matched, notMatched
}
