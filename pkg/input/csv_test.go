package input

import (
	"testing"
	"rods/pkg/config"
	"rods/pkg/source"
	"github.com/sirupsen/logrus"
)

func TestIterateAll(t *testing.T) {
	testCases := []struct{
		name string
		file string
		expectedRows [][]string
	} {
		{
			name: "normal",
			file: "test1,\"test2\"\n\"test\"\"3\",test 4",
			expectedRows: [][]string { {"test1", "test2"}, {"test\"3", "test 4"} },
		}, {
			name: "empty",
			file: "",
			expectedRows: [][]string {},
		}, {
			name: "empty row",
			file: "test1,test2\n\ntest3,test4",
			expectedRows: [][]string { {"test1", "test2"}, {"test3", "test4"} },
		}, {
			name: "end after one row",
			file: "test1,test2",
			expectedRows: [][]string { {"test1", "test2"} },
		}, {
			name: "too many columns",
			file: "test1,test2\ntest3,test4,test5",
			expectedRows: [][]string { {"test1", "test2"}, {"test3", "test4", "test5"} },
		}, {
			name: "not enough columns",
			file: "test1,test2\ntest3",
			expectedRows: [][]string { {"test1", "test2"}, {"test3"} },
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			source := source.NewMock(testCase.file)
			config := &config.CsvInputConfig{
				Path: "test",
				IgnoreFirstRow: false,
				Delimiter: ",",
				Columns: map[string]config.CsvInputColumnConfig{
					"a": { Type: "string" },
					"b": { Type: "string" },
				},
			}
			csv, err := NewCsv(config, source, logrus.StandardLogger())
			if err != nil {
				t.Error(err)
			}

			rows, errors := csv.IterateAll()
			for i := 0; i < len(testCase.expectedRows); i++ {
				select {
					case row := <-rows:
						if len(row) != len(testCase.expectedRows[i]) {
							t.Errorf("Received row have %v columns, expected %v", len(row), len(testCase.expectedRows[i]))
						}
						for j := 0; j < len(row) && j < len(testCase.expectedRows); j++ {
							if row[j] != testCase.expectedRows[i][j] {
								t.Errorf("Received '%v', expected '%v' for cell [%v][%v]", row[j], testCase.expectedRows[i][j], i, j)
							}
						}
					case err := <-errors:
						t.Error(err)
				}
			}
		})
	}
}
