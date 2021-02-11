package input

import (
	"io"
	"testing"
	"rods/pkg/config"
	"rods/pkg/source"
	"rods/pkg/utils"
	"github.com/sirupsen/logrus"
)

func TestIterateAll(t *testing.T) {
	testCases := []struct{
		name string
		file string
		expectedRows [][]*string
	} {
		{
			name: "normal",
			file: "test1,\"test2\"\n\"test\"\"3\",test 4",
			expectedRows: [][]*string { {utils.PString("test1"), utils.PString("test2")}, {utils.PString("test\"3"), utils.PString("test 4")} },
		}, {
			name: "empty",
			file: "",
			expectedRows: [][]*string {},
		}, {
			name: "empty row",
			file: "test1,test2\n\ntest3,test4",
			expectedRows: [][]*string { {utils.PString("test1"), utils.PString("test2")}, {utils.PString("test3"), utils.PString("test4")} },
		}, {
			name: "end after one row",
			file: "test1,test2",
			expectedRows: [][]*string { {utils.PString("test1"), utils.PString("test2")} },
		}, {
			name: "too many columns",
			file: "test1,test2\ntest3,test4,test5",
			expectedRows: [][]*string { {utils.PString("test1"), utils.PString("test2")}, {utils.PString("test3"), utils.PString("test4"), utils.PString("test5")} },
		}, {
			name: "not enough columns",
			file: "test1,test2\ntest3",
			expectedRows: [][]*string { {utils.PString("test1"), utils.PString("test2")}, {utils.PString("test3"), nil} },
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			source := source.NewMock(testCase.file)
			config := &config.CsvInputConfig{
				Path: "test",
				IgnoreFirstRow: false,
				Delimiter: ",",
				Columns: []config.CsvInputColumnConfig{
					{ Name: "a", Type: "string" },
					{ Name: "b", Type: "string" },
				},
				ColumnIndexByName: map[string]int{
					"a": 0,
					"b": 1,
				},
			}
			csv, err := NewCsv(config, source, logrus.StandardLogger())
			if err != nil {
				t.Error(err)
			}

			records, errors := csv.IterateAll()
			for i := 0; i < len(testCase.expectedRows); i++ {
				select {
					case record := <-records:
						for j, k := range []string { "a", "b" } {
							result, err := record.GetString(k)
							if err != nil {
								t.Errorf("Got error '%v', expected '%v' for cell [%v][%v]", err, testCase.expectedRows[i][j], i, j)
								continue
							}

							expected := testCase.expectedRows[i][j]
							if !(result == nil && expected == nil) && *result != *expected {
								t.Errorf("Received '%v', expected '%v' for cell [%v][%v]", result, testCase.expectedRows[i][j], i, j)
							}
						}
					case err := <-errors:
						t.Error(err)
				}

				// Asserts that IterateAll does not fail with concurrent accesses
				csv.sourceReader.Seek(0, io.SeekStart)
			}
		})
	}
}
