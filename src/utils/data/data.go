package data

import (
	"encoding/json"
	"fmt"
	"github.com/Jeffail/gabs"
	"github.com/pkg/errors"
	"math"
	"reflect"
	"regexp"
)

// func MakeHash[T any](s []T, index int) ([]T, error) {

// MakeHash removes an element specified by index for any type of slice
func MakeHash[T any](s []T) map[any]bool {
	m := make(map[any]bool)
	for _, v := range s {
		m[v] = true
	}

	return m
}

// RemoveIndexFromSlice removes an element specified by index for any type of slice
func RemoveIndexFromSlice[T any](s []T, index int) ([]T, error) {
	if index < 0 {
		return nil, errors.New("index must be a positive integer")
	}
	if len(s)-1 < index {
		return nil, errors.New("index out of bounds")
	}

	return append(s[:index], s[index+1:]...), nil
}

// FormatCSVFriendlyStrArr wraps the multi arr string with double quotes.
// Mostly used for CSV purposes
func FormatCSVFriendlyStrArr(arr [][]string) [][]string {
	for arrIdx, arrVal := range arr {
		arrInside := arrVal
		for arrInsideIdx, arrInsideVal := range arrInside {
			if arrInsideVal != "" {
				reComma := regexp.MustCompile(`"`)
				arrInsideVal = reComma.ReplaceAllString(arrInsideVal, `""`)
				arrInsideVal = `"` + arrInsideVal + `"`
			}
			arr[arrIdx][arrInsideIdx] = arrInsideVal
		}
	}
	return arr
}

func GetSliceValuesAsSliceOfStrings(slice interface{}) ([]string, error) {
	// Get data as raw value, and marshal into json
	rawValue := GetRawValue(slice)
	if reflect.TypeOf(rawValue).Kind() != reflect.Slice {
		return nil, errors.New("variable is not a slice")
	}
	marshalled, err := json.Marshal(rawValue)
	if err != nil {
		return nil, errors.New("error trying to marshal slice")
	}

	// Parse json contents and push into strArr
	var strArr []string
	jsonParsed, err := gabs.ParseJSON([]byte(marshalled))
	if err != nil {
		return nil, errors.Wrap(err, "error parsing the marshalled slice")
	}
	children, err := jsonParsed.Children()
	if err != nil {
		return nil, errors.Wrap(err, "error accessing the children of the parsed slice	")
	}
	for _, child := range children {
		strArr = append(strArr, child.String())
	}
	return strArr, nil
}

// GetRawValue returns the raw value of a variable if it's from a pointer
func GetRawValue(data interface{}) interface{} {
	if reflect.ValueOf(data).Kind() == reflect.Ptr {
		return reflect.Indirect(reflect.ValueOf(data)).Interface()
	}
	return data
}

// GetStructAsValue returns struct as value
func GetStructAsValue(_struct interface{}) (interface{}, error) {
	rawValue := GetRawValue(_struct)
	if reflect.TypeOf(rawValue).Kind() != reflect.Struct {
		actualType := reflect.TypeOf(_struct)
		return nil, fmt.Errorf("data provided is not struct: %v, but is of type: %v", _struct, actualType)
	}
	return rawValue, nil
}

// IntArrToInterfaceArr converts an array of ints to an array of interfaces
func IntArrToInterfaceArr(s []int) []interface{} {
	var i []interface{}
	for _, v := range s {
		i = append(i, v)
	}
	return i
}
func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}
