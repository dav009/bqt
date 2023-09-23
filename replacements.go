package bqt

import (
	"fmt"
	"regexp"
	"strings"
)

type Replacement struct {
	TableFullName  string
	ReplaceSql     string
	TableShortName string
}

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
func assertSQLCode(sql string, m Mock) (string, error) {

	mockedSql, err := mockToSql(m)
	if err != nil {
		return "", err
	}
	columns := strings.Join(mockedSql.Columns, ",")
	return fmt.Sprintf("SELECT %s FROM( %s ) \n  EXCEPT DISTINCT \n SELECT %s FROM (%s)", columns, sql, columns, mockedSql.Sql), nil
}

func sql(queryToTest string, mocks map[string]Mock) (Replacement, error) {
	sqlCode := queryToTest // load sql file here
	replacement := Replacement{ReplaceSql: sqlCode, TableFullName: fullname, TableShortName: shortname}
	sqlCode = Replace(sqlCode, replacement)
	return replacement, nil
}

func sql(nodeKey string, mocks map[string]Mock) (Replacement, error) {
	sqlCode = Replace(sqlCode, replacement)
	replacement := Replacement{ReplaceSql: sqlCode, TableFullName: fullname, TableShortName: shortname}
	return replacement, nil
}

/*
   Given a Test it generates the SQL code that mocks data, run the needed logic and asserts the output data
*/

func GenerateTestSQL(t Test) (string, error) {

	replacement, err := sql(t.Model, t.Mocks)
	if err != nil {
		return "", err
	}
	sql, err := assertSQLCode(replacement.ReplaceSql, t.Output)
	if err != nil {
		return "", err
	}
	return sql, nil
}
