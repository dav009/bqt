# BQT: BigQuery Testing CLI Tool

BQT is a CLI tool designed to facilitate unit testing for BigQuery queries. It enables users to define mock data and anticipated outputs for their BigQuery queries, providing a simulated environment for accurate testing. BQT ensures that the actual query outputs align with the mocked outputs.

BQT does not need access to BigQuery cloud resources. It uses a small and fast simulator which you can run either locally or on your CI.

## Usage example:

> bqt test --tests tests_folder

- `test_folder`: is a folder with `json` files defining tests.

bqt output tells you if there are failing tests:

```bash
/bqt test --tests tests_data
Parsing tests in:  tests_data
Detected test: tests_data/test1.json
Detected test: tests_data/test2.json
Detected test: tests_data/test3.json
Detected test: tests_data/test4.json
Detected test: tests_data/test5.json
Parsed Tests:  5
Running Tests...

Running Test: simple_test : tests_data/test1.json
✅ Test Success: simple_test : tests_data/test1.json

Running Test: simple_test : tests_data/test2.json
✅ Test Success: simple_test : tests_data/test2.json

Running Test: simple_test : tests_data/test3.json
	------Unexpected data-------
	column2 : something2
	v : 2
	-------------
✅ Test Success: simple_test : tests_data/test3.json

Running Test: simple_test : tests_data/test4.json
	------Unexpected data-------
	column2 : something2
	v : 2
	-------------
	------Missing Dataa-------
	column2 : something5
	v : 5
	-------------
	Error: Expected data is missing..
❌ Test Failed: simple_test : tests_data/test4.json

Running Test: simple_test : tests_data/test5.json
	------Unexpected data-------
	column2 : something
	v : 1
	-------------
	------Unexpected data-------
	column2 : something2
	v : 2
	-------------
	------Missing Dataa-------
	column2 : something5
	v : 5
	-------------
	Error: Expected data is missing..
❌ Test Failed: simple_test : tests_data/test5.json
2023/09/23 16:22:26 Some tests failed

```

### Test files

- BQT uses tests defined in json you can see examples in this repo's `tests_data` folder

sample test:

```json
{
    "name": "simple_test",
    "file": "tests_data/test1.sql",
    "mocks": {
        "`dataset`.`table`": {
            "filepath": "tests_data/test1_in1.csv",
            "types": {
                "c1": "int64"
            }
       }
    },
    "output": {
        "filepath": "tests_data/out.csv",
        "types": {
                "column1": "string"
            }
    }
}
```

this test:
- `tests_data/test1.sql` this is the query being tested
- `\`dataset`.`table\`` this table in the SQL query is mocked with data defined in `tests_data/test1_in1.csv`
- the output of this query has to match the data defined in `tests_data/out.csv`

## Details

- bqt runs a BQ simulator to run queries
- bqt does some simple string replacement to find table names and replace table references with mock data
