package index

import (
	"database/sql/driver"
	"fmt"
	gosqlite "github.com/mattn/go-sqlite3"
	"io"
	"reflect"
	sqlitePackage "rodb.io/pkg/index/sqlite"
	"rodb.io/pkg/input"
	"rodb.io/pkg/input/record"
	"rodb.io/pkg/util"
	"strings"
)

type Sqlite struct {
	config *SqliteConfig
	input  input.Input
	db     *gosqlite.SQLiteConn
}

func NewSqlite(
	config *SqliteConfig,
	inputs input.List,
) (*Sqlite, error) {
	input, inputExists := inputs[config.Input]
	if !inputExists {
		return nil, fmt.Errorf("Input '%v' not found in inputs list.", config.Input)
	}

	// We use the driver directly rather than the database/sql interfaces, because the generic
	// way tries to autodetect and normalize the return types. But we want to get the real types.
	db, err := sqlitePackage.Open(config.Dsn)
	if err != nil {
		return nil, fmt.Errorf("Error while opening the sqlite DSN: %w", err)
	}

	sqlite := &Sqlite{
		config: config,
		input:  input,
		db:     db,
	}

	metadataExists, err := sqlitePackage.HasMetadata(sqlite.db, sqlite.config.Name)
	if err != nil {
		return nil, fmt.Errorf("Error while checking metadata from the index: %w", err)
	}
	if !metadataExists {
		if err := sqlite.createIndex(); err != nil {
			return nil, fmt.Errorf("Error while creating the index: %w", err)
		}
	}

	return sqlite, nil
}

func (sqlite *Sqlite) Name() string {
	return sqlite.config.Name
}

func (sqlite *Sqlite) createIndex() error {
	metadata, err := sqlitePackage.NewMetadata(
		sqlite.db,
		sqlite.config.Name,
		sqlite.input,
	)
	if err != nil {
		return err
	}

	if err := metadata.Save(); err != nil {
		return err
	}

	updateProgress := util.TrackProgress(sqlite.input, sqlite.config.Logger)

	inputIterator, end, err := sqlite.input.IterateAll()
	if err != nil {
		return err
	}
	defer func() {
		if err := end(); err != nil {
			sqlite.config.Logger.Errorf("Error while closing the input iterator: %v", err)
		}
	}()

	columnDefinitions := make([]string, len(sqlite.config.Properties))
	columnIdentifiers := make([]string, len(sqlite.config.Properties))
	indexIdentifiers := make([]string, len(sqlite.config.Properties))
	insertPlaceholders := make([]string, len(sqlite.config.Properties))
	for propertyIndex, property := range sqlite.config.Properties {
		propertyIdentifier, err := sqlite.getPropertyIdentifier(property.Name)
		if err != nil {
			return err
		}
		indexIdentifier, err := sqlite.getIndexIdentifier(property.Name)
		if err != nil {
			return err
		}

		columnDefinitions[propertyIndex] = propertyIdentifier + " BLOB COLLATE " + property.Collate
		columnIdentifiers[propertyIndex] = propertyIdentifier
		indexIdentifiers[propertyIndex] = indexIdentifier
		insertPlaceholders[propertyIndex] = "?"
	}

	tableIdentifier, err := sqlite.getIndexTableIdentifier()
	if err != nil {
		return err
	}

	_, err = sqlite.db.Exec(`
		CREATE TABLE `+tableIdentifier+` (
			"offset" INTEGER NOT NULL,
			`+strings.Join(columnDefinitions, ", ")+`
		);
	`, []driver.Value{})
	if err != nil {
		return fmt.Errorf("Error while creating index table: %w", err)
	}

	for propertyIndex, indexIdentifier := range indexIdentifiers {
		_, err = sqlite.db.Exec(`
			CREATE INDEX `+indexIdentifier+` ON `+tableIdentifier+` (`+columnIdentifiers[propertyIndex]+`);
		`, []driver.Value{})
		if err != nil {
			return fmt.Errorf("Error while creating index table: %w", err)
		}
	}

	preparedInsert, err := sqlite.db.Prepare(`
		INSERT INTO ` + tableIdentifier + ` (
			"offset",
			` + strings.Join(columnIdentifiers, ", ") + `
		) VALUES (?, ` + strings.Join(insertPlaceholders, ", ") + `);
	`)
	if err != nil {
		return fmt.Errorf("Error while preparing index table insert query: %w", err)
	}

	valuesToInsert := make([]driver.Value, 1+len(sqlite.config.Properties))
	for {
		record, err := inputIterator()
		if err != nil {
			return err
		}
		if record == nil {
			break
		}

		updateProgress(record.Position())

		valuesToInsert[0] = record.Position()
		for propertyIndex, property := range sqlite.config.Properties {
			value, err := record.Get(property.Name)
			if err != nil {
				return err
			}

			if value != nil {
				value = reflect.ValueOf(value).Interface()
			}

			valuesToInsert[propertyIndex+1] = value
		}

		if _, err = preparedInsert.Exec(valuesToInsert); err != nil {
			return err
		}
	}

	metadata.SetCompleted(true)
	if err := metadata.Save(); err != nil {
		return err
	}

	rows, err := sqlite.db.Query(`SELECT COUNT(1) FROM `+tableIdentifier+`;`, []driver.Value{})
	if err != nil {
		return err
	}
	data := make([]driver.Value, 1)
	if err = rows.Next(data); err != nil {
		return err
	}
	defer rows.Close()

	sqlite.config.Logger.WithField("indexedRows", data[0].(int64)).Infof("Successfully finished indexing")

	return nil
}

func (sqlite *Sqlite) getPropertyIdentifier(propertyName string) (string, error) {
	return sqlitePackage.SanitizeIdentifier(sqlite.db, fmt.Sprintf("property_%v", propertyName))
}

func (sqlite *Sqlite) getIndexIdentifier(propertyName string) (string, error) {
	return sqlitePackage.SanitizeIdentifier(
		sqlite.db,
		fmt.Sprintf("index_%v_%v", sqlite.config.Name, propertyName),
	)
}

func (sqlite *Sqlite) getIndexTableIdentifier() (string, error) {
	return sqlitePackage.SanitizeIdentifier(sqlite.db, fmt.Sprintf("rodb_%v_index", sqlite.config.Name))
}

func (sqlite *Sqlite) GetRecordPositions(
	input input.Input,
	filters map[string]interface{},
) (record.PositionIterator, error) {
	if input != sqlite.input {
		return nil, fmt.Errorf("This index does not handle the input '%v'.", input.Name())
	}

	if len(filters) == 0 {
		return nil, fmt.Errorf("This index requires at least one filter.")
	}

	tableIdentifier, err := sqlite.getIndexTableIdentifier()
	if err != nil {
		return nil, err
	}

	clauses := make([]string, 0, len(filters))
	values := make([]driver.Value, 0, len(filters))
	for propertyName, filter := range filters {
		if !sqlite.config.DoesHandleProperty(propertyName) {
			return nil, fmt.Errorf("This index does not handle the property '%v'.", propertyName)
		}

		columnIdentifier, err := sqlite.getPropertyIdentifier(propertyName)
		if err != nil {
			return nil, err
		}

		clauses = append(clauses, columnIdentifier+" = ?")
		values = append(values, filter)
	}

	rows, err := sqlite.db.Query(`
		SELECT "offset"
		FROM `+tableIdentifier+`
		WHERE `+strings.Join(clauses, " AND ")+`
	`, values)
	if err != nil {
		return nil, err
	}

	rowData := make([]driver.Value, 1)
	return func() (*record.Position, error) {
		for {
			if err := rows.Next(rowData); err != nil {
				_ = rows.Close()
				if err == io.EOF {
					break
				} else {
					return nil, err
				}
			}

			position := record.Position(rowData[0].(int64))
			return &position, nil
		}

		return nil, nil
	}, nil
}

func (sqlite *Sqlite) Close() error {
	return sqlite.db.Close()
}
