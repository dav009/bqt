package bqt

import (
	"context"
	"errors"
	"fmt"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"cloud.google.com/go/bigquery"
	"github.com/fatih/color"
	"github.com/goccy/bigquery-emulator/server"
	"github.com/goccy/bigquery-emulator/types"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// returns a Manifest structure out of a .json file
func ParseManifest(path string) Manifest {

	jsonFile, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	bytes, _ := ioutil.ReadAll(jsonFile)
	manifest := Manifest{}
	if err := json.Unmarshal(bytes, &manifest); err != nil {
		panic(err)
	}
	return manifest
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

func SaveSQL(path string, sql string) error {
	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return err
	}
	data := []byte(sql)
	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func RunQueryMinusExpectation(ctx context.Context, client *bigquery.Client, query string) error {
	fmt.Println("qyerying...")
	q := client.Query((query))
	fmt.Println("reading....")
	it, err := q.Read(ctx)
	if err != nil {
		return err
	}
	for {

		var row []bigquery.Value
		if err := it.Next(&row); err != nil {
			if err == iterator.Done {
				break
			}
			return err
		}

		color.Green("-------------")
		for i, field := range it.Schema {
			record := fmt.Sprintf("%s : %v", field.Name, row[i])
			color.Green(record)

		}
		color.Green("-------------")
		//color.Green(strings.Join(columns, "\t"))
		//color.Green(fmt.Sprintf("%v", row))
		err = errors.New("Query returned extra data compared to expectation..")
	}

	return err
}

func RunExpectationMinusQuery(ctx context.Context, client *bigquery.Client, query string) error {
	it, err := client.Query((query)).Read(ctx)
	if err != nil {
		return err
	}
	for {
		var row []bigquery.Value
		if err := it.Next(&row); err != nil {
			if err == iterator.Done {
				break
			}
			return err
		}
		color.Red("-------------")
		for i, field := range it.Schema {
			record := fmt.Sprintf("%s : %v", field.Name, row[i])
			color.Red(record)

		}
		color.Red("-------------")
		color.Red(fmt.Sprintf("%v", row))
		err = errors.New("Expected data was not fully completed..")

	}
	return err
}

func RunTests(mode string, tests []Test, m Manifest) error {
	ctx := context.Background()
	const (
		projectID = "fq-stage-bigquery"
		datasetID = "dataset1"
		routineID = "routine1"
	)
	bqServer, err := server.New(server.TempStorage)
	if err != nil {
		return err
	}
	if err := bqServer.Load(
		server.StructSource(
			types.NewProject(
				projectID,
				types.NewDataset(
					datasetID,
				),
			),
		),
	); err != nil {
		return err
	}
	if err := bqServer.SetProject(projectID); err != nil {
		return err
	}
	testServer := bqServer.TestServer()
	defer testServer.Close()

	var client *bigquery.Client
	if mode == "local" {
		client, err = bigquery.NewClient(
			ctx,
			projectID,
			option.WithEndpoint(testServer.URL),
			option.WithoutAuthentication(),
		)
	} else {
		client, err = bigquery.NewClient(
			ctx,
			projectID,
		)
	}

	if err != nil {
		return err
	}
	defer client.Close()

	var lastErr error = nil

	for _, t := range tests {
		sqlQueries, err := GenerateTestSQL(t, m)

		if err != nil {
			return err
		}
		fmt.Println("Running: Query minus Expectation")
		fmt.Println(sqlQueries.QueryMinusExpected)
		fmt.Println("end of query..")
		err = RunQueryMinusExpectation(ctx, client, sqlQueries.QueryMinusExpected)
		if err != nil {
			lastErr = err
		}
		fmt.Println("Running:Expectation minus Query")
		err = RunExpectationMinusQuery(ctx, client, sqlQueries.ExpectedMinusQuery)
		if err != nil {
			lastErr = err
		}

	}
	return lastErr
}
