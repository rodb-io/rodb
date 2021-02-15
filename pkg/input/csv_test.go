package input

import (
	"github.com/sirupsen/logrus"
	"io"
	"rods/pkg/config"
	"rods/pkg/source"
	"rods/pkg/util"
	"testing"
)

func TestIterateAll(t *testing.T) {
	testCases := []struct {
		name         string
		file         string
		expectedRows [][]*string
	}{
		{
			name:         "normal",
			file:         "test1,\"test2\"\n\"test\"\"3\",test 4",
			expectedRows: [][]*string{{util.PString("test1"), util.PString("test2")}, {util.PString("test\"3"), util.PString("test 4")}},
		}, {
			name:         "empty",
			file:         "",
			expectedRows: [][]*string{},
		}, {
			name:         "empty row",
			file:         "test1,test2\n\ntest3,test4",
			expectedRows: [][]*string{{util.PString("test1"), util.PString("test2")}, {util.PString("test3"), util.PString("test4")}},
		}, {
			name:         "end after one row",
			file:         "test1,test2",
			expectedRows: [][]*string{{util.PString("test1"), util.PString("test2")}},
		}, {
			name:         "too many columns",
			file:         "test1,test2\ntest3,test4,test5",
			expectedRows: [][]*string{{util.PString("test1"), util.PString("test2")}, {util.PString("test3"), util.PString("test4"), util.PString("test5")}},
		}, {
			name:         "not enough columns",
			file:         "test1,test2\ntest3",
			expectedRows: [][]*string{{util.PString("test1"), util.PString("test2")}, {util.PString("test3"), nil}},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			source := source.NewMock(testCase.file)
			config := &config.CsvInput{
				Path:           "test",
				IgnoreFirstRow: false,
				Delimiter:      ",",
				Columns: []config.CsvInputColumn{
					{Name: "a", Type: "string"},
					{Name: "b", Type: "string"},
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

			resultsChannel := csv.IterateAll()
			for i := 0; i < len(testCase.expectedRows); i++ {
				select {
				case result := <-resultsChannel:
					if result.Error != nil {
						t.Error(err)
					} else {
						for j, k := range []string{"a", "b"} {
							result, err := result.Record.GetString(k)
							if err != nil {
								t.Errorf("Got error '%v', expected '%v' for cell [%v][%v]", err, testCase.expectedRows[i][j], i, j)
								continue
							}

							expected := testCase.expectedRows[i][j]
							if !(result == nil && expected == nil) && *result != *expected {
								t.Errorf("Received '%v', expected '%v' for cell [%v][%v]", result, testCase.expectedRows[i][j], i, j)
							}
						}
					}
				}

				// Asserts that IterateAll does not fail with concurrent accesses
				csv.sourceReader.Seek(0, io.SeekStart)
			}
		})
	}
}
