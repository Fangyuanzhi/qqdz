package jsonparse

import (
	"fmt"
	"github.com/buger/jsonparser"
	"io/ioutil"
)

var (
	JsonMap map[string]map[string]string
)

func ParseJsonFile(filePath string) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	cb := func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		fmt.Println(string(value))
	}
	jsonparser.ArrayEach(file, cb, "room", "robotName")
}
