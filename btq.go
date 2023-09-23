package bqt

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"os"

	"cloud.google.com/go/bigquery"
	"github.com/fatih/color"
	"github.com/goccy/bigquery-emulator/server"
	"github.com/goccy/bigquery-emulator/types"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
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

/*
   Represents a Mock as SQL
*/
type SQLMock struct {
	Sql     string
	Columns []string
}

// converts a csv row's single column value into a SQL statement
func mockEntryToSql(columnName string, value string, columnType string) string {

	if value == "" {
		value = "null"
	} else {
		value = fmt.Sprintf("\"%s\"", value)
	}
	if columnType != "" {
		return fmt.Sprintf("CAST(%s AS %s) AS %s", value, columnType, columnName)
	}

	return fmt.Sprintf("%s AS %s", value, columnName)

}

/*
   Converts a Mock into a SQL statement that we can use in Replacements
*/
func mockToSql(m Mock) (SQLMock, error) {

	allColumns := []string{}
	file, err := os.Open(m.Filepath)
	if err != nil {
		return SQLMock{}, err

	}
	data := CSVToMap(file)
	var sqlStatements []string
	for _, row := range data {

		columnsValues := []string{}
		columns := make([]string, 0)
		// ordering columns so we can test
		for k, _ := range row {
			columns = append(columns, k)
		}
		sort.Strings(columns)
		if len(allColumns) == 0 {
			allColumns = columns
		}
		for _, column := range columns {
			value := row[column]
			columnType := m.Types[column]
			entry := mockEntryToSql(column, value, columnType)
			columnsValues = append(columnsValues, entry)

		}
		statement := fmt.Sprintf("\n SELECT %s", strings.Join(columnsValues, ", "))
		sqlStatements = append(sqlStatements, statement)
	}
	return SQLMock{Sql: strings.Join(sqlStatements, "\n UNION ALL \n"), Columns: allColumns}, nil

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

func RunTests(mode string, tests []Test) error {
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
		sqlQueries, err := GenerateTestSQL(t)

		if err != nil {
			return err
		}
		fmt.Println("Running: Query minus Expectation")
		err = RunQueryMinusExpectation(ctx, client, sqlQueries.QueryMinusExpected)
		if err != nil {
			lastErr = err
		}
		fmt.Println("Running: Expectation minus Query")
		err = RunExpectationMinusQuery(ctx, client, sqlQueries.ExpectedMinusQuery)
		if err != nil {
			lastErr = err
		}

	}
	return lastErr
}
