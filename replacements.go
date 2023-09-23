package bqt

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

/*
   Given a SQL query and a replacement Struct, it applies the recplament on the SQL and returns a new SQL query.   References to a Table are replaced
*/
func Replace(sql string, replacement Replacement) string {

	if (replacement == Replacement{}) {
		return sql
	}
	regexWithAlias, err := regexp.Compile(fmt.Sprintf(`(?i)%s\sas\s([A-z0-9_]+)\s`, replacement.TableFullName))
	if err != nil {
		panic(err)
	}
	useExistingAlias := fmt.Sprintf("(%s) AS $1 ", replacement.ReplaceSql)
	newSql := regexWithAlias.ReplaceAllString(sql, useExistingAlias)
	createAlias := fmt.Sprintf("(%s) AS %s", replacement.ReplaceSql, replacement.TableShortName)
	newSql = strings.ReplaceAll(newSql, fmt.Sprintf("%s", replacement.TableFullName), createAlias)
	return newSql
}

/*
   Given the SQL code of a model and an Expected Output mock,
   This function returns a SQL  query which asserts that the output table of SQL is equal to the data contained in the mock
*/
func queryMinusMock(sql string, m Mock) (string, error) {

	mockedSql, err := mockToSql(m)
	if err != nil {
		return "", err
	}
	columns := strings.Join(mockedSql.Columns, ",")
	return fmt.Sprintf("SELECT %s FROM( %s ) \n  EXCEPT DISTINCT \n SELECT %s FROM (%s)", columns, sql, columns, mockedSql.Sql), nil
}

func tableShortName(tableFullName string) string {

	parts := strings.Split(tableFullName, ".") // Split the string by dot
	lastItem := parts[len(parts)-1]            // Get the last item
	return lastItem
}

func sql(sqlToTest string, mocks map[string]Mock) (string, error) {
	for tablefullName, mock := range mocks {
		mockSql, err := mockToSql(mock)
		if err != nil {
			return "", err
		}
		tableShortName := tableShortName(tablefullName)
		r := Replacement{ReplaceSql: mockSql.Sql,
			TableFullName: tablefullName, TableShortName: tableShortName}
		sqlToTest = Replace(sqlToTest, r)
	}

	return sqlToTest, nil
}

func mockMinusQuery(sql string, output Mock) (string, error) {

	mockedSql, err := mockToSql(output)
	if err != nil {
		return "", err
	}
	columns := strings.Join(mockedSql.Columns, ",")
	return fmt.Sprintf("SELECT %s FROM( %s ) \n  EXCEPT DISTINCT \n SELECT %s FROM (%s)", columns, mockedSql.Sql, columns, sql), nil
}

type SQLTestQuery struct {
	ExpectedMinusQuery  string
	QueryMinusExpected  string
	QueryWithMockedData string
}

func ReadContents(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	c := string(content)
	return strings.TrimSpace(c), nil
}

/*
   Given a Test it generates the SQL code that mocks data, run the needed logic and asserts the output data
*/
func GenerateTestSQL(t Test) (SQLTestQuery, error) {
	queryWithMockedData, err := sql(t.FileContent, t.Mocks)
	if err != nil {
		return SQLTestQuery{}, err
	}
	sqlQueryMinusExpectation, err := queryMinusMock(queryWithMockedData, t.Output)
	if err != nil {
		return SQLTestQuery{}, err
	}
	sqlExpectationMinusQuery, err := mockMinusQuery(queryWithMockedData, t.Output)
	if err != nil {
		return SQLTestQuery{}, err
	}

	return SQLTestQuery{QueryMinusExpected: sqlQueryMinusExpectation, ExpectedMinusQuery: sqlExpectationMinusQuery, QueryWithMockedData: queryWithMockedData}, nil
}
