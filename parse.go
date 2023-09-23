package bqt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

/* Describing Tests as structures */
type Mock struct {
	Filepath string            `json:"filepath"`
	Types    map[string]string `json:"types"`
}

type Output struct {
	Name string `json:"name"`
}

type Test struct {
	Name        string          `json:"name"`
	File        string          `json:"file"`
	Mocks       map[string]Mock `json:"mocks"`
	Output      Mock            `json:"output"`
	FileContent string
}

type Replacement struct {
	TableFullName  string
	ReplaceSql     string
	TableShortName string
}

/*
   returns a Test structure given a filepath
*/
func ParseTest(path string) (Test, error) {

	jsonFile, err := os.Open(path)
	if err != nil {
		return Test{}, err
	}
	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return Test{}, err
	}
	test := Test{}
	if err := json.Unmarshal(bytes, &test); err != nil {
		return Test{}, err
	}
	sqlQuery, err := ReadContents(test.File)
	test.FileContent = sqlQuery
	if err != nil {
		return Test{}, err
	}
	return test, nil
}

/*
   Given a folder returns a list of Test structs
*/
func ParseFolder(path string) ([]Test, error) {

	files, err := ioutil.ReadDir(path)
	tests := []Test{}
	if err != nil {
		return []Test{}, err
	}
	for _, f := range files {

		fullPath := filepath.Join(path, f.Name())
		fmt.Println(fullPath)
		test, err := ParseTest(fullPath)
		if err != nil {
			return []Test{}, err
		}
		tests = append(tests, test)

	}
	return tests, nil
}
