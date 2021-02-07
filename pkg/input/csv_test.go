package input

import (
	"testing"
	"rods/pkg/config"
	"rods/pkg/source"
	"github.com/sirupsen/logrus"
)

func TestIterateAll(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		source := source.NewMock("test1,\"test2\"\n\"test\"\"3\",test 4")
		expectedRows := [][]string { {"test1", "test2"}, {"test\"3", "test 4"} }
		config := &config.CsvInputConfig{
			Path: "test",
			IgnoreFirstRow: false,
			Delimiter: ",",
		}
		csv, err := NewCsv(config, source, logrus.StandardLogger())
		if err != nil {
			t.Error(err)
		}

		rows, errors := csv.IterateAll()
		for i := 0; i < len(expectedRows); i++ {
			select {
				case row := <-rows:
					if len(row) != len(expectedRows[i]) {
						t.Errorf("Received row have %v columns, expected %v", len(row), len(expectedRows[i]))
					}
					for j := 0; j < len(row); j++ {
						if row[j] != expectedRows[i][j] {
							t.Errorf("Received '%v', expected '%v' for cell [%v][%v]", row[j], expectedRows[i][j], i, j)
						}
					}
				case err := <-errors:
					t.Error(err)
			}
		}
	})
}
