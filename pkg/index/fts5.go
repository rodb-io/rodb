package index

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"reflect"
	sqlitePackage "rodb.io/pkg/index/sqlite"
	"rodb.io/pkg/input"
	"rodb.io/pkg/input/record"
	"rodb.io/pkg/util"
	"strconv"
	"strings"
)

type Fts5 struct {
	config *Fts5Config
	input  input.Input
	db     *sql.DB
}

func NewFts5(
	config *Fts5Config,
	inputs input.List,
) (*Fts5, error) {
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

	sqlite := &Fts5{
		config: config,
		input:  input,
		db:     db,
	}

	metadataExists, err := sqlitePackage.HasMetadata(sqlite.db, sqlite.config.Name)
	if err != nil {
		return nil, fmt.Errorf("Error while checking metadata from the index: %w", err)
	}
	if metadataExists {
		metadata, err := sqlitePackage.LoadMetadata(sqlite.db, sqlite.config.Name)
		if err != nil {
			return nil, err
		}

		if err := metadata.AssertValid(sqlite.input); err != nil {
			return nil, err
		}
	} else {
		if err := sqlite.createIndex(); err != nil {
			return nil, fmt.Errorf("Error while creating the index: %w", err)
		}
	}

	return sqlite, nil
}

func (sqlite *Fts5) Name() string {
	return sqlite.config.Name
}

func (sqlite *Fts5) createIndex() error {
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

	columnIdentifiers := make([]string, len(sqlite.config.Properties))
	insertPlaceholders := make([]string, len(sqlite.config.Properties))
	for propertyIndex, property := range sqlite.config.Properties {
		if property == "__offset" {
			return errors.New("__offset is a reserved property name for the fts5 index.")
		}

		propertyIdentifier, err := sqlitePackage.SanitizeIdentifier(sqlite.db, property)
		if err != nil {
			return err
		}

		columnIdentifiers[propertyIndex] = propertyIdentifier
		insertPlaceholders[propertyIndex] = "?"
	}

	tableIdentifier, err := sqlite.getIndexTableIdentifier()
	if err != nil {
		return err
	}

	prefixString := ""
	for _, prefixValue := range sqlite.config.Prefix {
		if prefixString != "" {
			prefixString += " "
		}
		prefixString += strconv.Itoa(prefixValue)
	}

	tokenizeString, err := sqlitePackage.SanitizeIdentifier(sqlite.db, sqlite.config.Tokenize)
	if err != nil {
		return err
	}

	_, err = sqlite.db.Exec(`
		CREATE VIRTUAL TABLE ` + tableIdentifier + ` USING fts5(
			"__offset" UNINDEXED,
			` + strings.Join(columnIdentifiers, ", ") + `,
			prefix = '` + prefixString + `',
			tokenize = ` + tokenizeString + `
		);
	`)
	if err != nil {
		return fmt.Errorf("Error while creating index table: %w", err)
	}

	preparedInsert, err := sqlite.db.Prepare(`
		INSERT INTO ` + tableIdentifier + ` (
			"__offset",
			` + strings.Join(columnIdentifiers, ", ") + `
		) VALUES (?, ` + strings.Join(insertPlaceholders, ", ") + `);
	`)
	if err != nil {
		return fmt.Errorf("Error while preparing index table insert query: %w", err)
	}

	valuesToInsert := make([]interface{}, 1+len(sqlite.config.Properties))
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
		for propertyIndex, propertyName := range sqlite.config.Properties {
			value, err := record.Get(propertyName)
			if err != nil {
				return err
			}

			if value != nil {
				value = reflect.ValueOf(value).Interface()
			}

			valuesToInsert[propertyIndex+1] = value
		}

		if _, err = preparedInsert.Exec(valuesToInsert...); err != nil {
			return err
		}
	}

	metadata.SetCompleted(true)
	if err := metadata.Save(); err != nil {
		return err
	}

	row := sqlite.db.QueryRow(`SELECT COUNT(1) FROM ` + tableIdentifier + `;`)
	if err := row.Err(); err != nil {
		return err
	}

	var indexedRows int64
	if err = row.Scan(&indexedRows); err != nil {
		return err
	}

	sqlite.config.Logger.WithField("indexedRows", indexedRows).Infof("Successfully finished indexing")

	return nil
}

func (sqlite *Fts5) getIndexIdentifier(propertyName string) (string, error) {
	return sqlitePackage.SanitizeIdentifier(
		sqlite.db,
		fmt.Sprintf("index_%v_%v", sqlite.config.Name, propertyName),
	)
}

func (sqlite *Fts5) getIndexTableIdentifier() (string, error) {
	return sqlitePackage.SanitizeIdentifier(sqlite.db, fmt.Sprintf("rodb_%v_index", sqlite.config.Name))
}

func (sqlite *Fts5) GetRecordPositions(
	input input.Input,
	filters map[string]interface{},
) (record.PositionIterator, error) {
	if input != sqlite.input {
		return nil, fmt.Errorf("This index does not handle the input '%v'.", input.Name())
	}

	// We want to expose the index MATCH operator, so we don't really need a parameter name.
	// However, for the sake of simplicity, we keep the sqlite's logic and require
	// the index name as parameter.
	if len(filters) != 1 {
		return nil, fmt.Errorf("This index can only have one filter at a time.")
	}
	if _, filterExists := filters["match"]; !filterExists {
		return nil, fmt.Errorf("This index must receive a single filter named 'match'.")
	}

	tableIdentifier, err := sqlite.getIndexTableIdentifier()
	if err != nil {
		return nil, err
	}

	rows, err := sqlite.db.Query(`
		SELECT "__offset"
		FROM `+tableIdentifier+`
		WHERE `+tableIdentifier+` MATCH ?
	`, filters["match"])
	if err != nil {
		return nil, err
	}

	return func() (*record.Position, error) {
		for rows.Next() {
			var positionValue int64
			if err := rows.Scan(&positionValue); err != nil {
				_ = rows.Close()
				return nil, err
			}

			position := record.Position(positionValue)
			return &position, nil
		}

		return nil, nil
	}, nil
}

func (sqlite *Fts5) Close() error {
	return sqlite.db.Close()
}
