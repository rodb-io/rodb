package sqlite

import (
	"database/sql/driver"
	"fmt"
	gosqlite "github.com/mattn/go-sqlite3"
	"rodb.io/pkg/input"
	"time"
)

// Current version of the indexing protocol
const CurrentVersion = uint16(1)

type Metadata struct {
	db                        *gosqlite.SQLiteConn
	indexName                 string
	version                   uint16
	inputFileModificationTime time.Time
	inputFileSize             int64
	completed                 bool
}

func NewMetadata(
	db *gosqlite.SQLiteConn,
	indexName string,
	input input.Input,
) (*Metadata, error) {
	size, err := input.Size()
	if err != nil {
		return nil, err
	}

	modTime, err := input.ModTime()
	if err != nil {
		return nil, err
	}

	metadata := &Metadata{
		db:                        db,
		indexName:                 indexName,
		version:                   CurrentVersion,
		inputFileModificationTime: modTime,
		inputFileSize:             size,
		completed:                 false,
	}

	return metadata, nil
}

func LoadMetadata(
	db *gosqlite.SQLiteConn,
	indexName string,
) (*Metadata, error) {
	metadata := &Metadata{
		db:        db,
		indexName: indexName,
	}

	tableIdentifier, err := metadata.GetTableIdentifier()
	if err != nil {
		return nil, err
	}

	rows, err := metadata.db.Query(`
		SELECT
			"version",
			"inputFileModificationTime",
			"inputFileSize",
			"completed"
		FROM `+tableIdentifier+`;
	`, []driver.Value{})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data := make([]driver.Value, 4)
	if err = rows.Next(data); err != nil {
		return nil, err
	}

	metadata.version = uint16(data[0].(int64))
	metadata.inputFileModificationTime = time.Unix(data[1].(int64), 0)
	metadata.inputFileSize = data[2].(int64)
	metadata.completed = data[3].(bool)

	return metadata, nil
}

func HasMetadata(
	db *gosqlite.SQLiteConn,
	indexName string,
) (bool, error) {
	metadata := &Metadata{
		db:        db,
		indexName: indexName,
	}

	tableIdentifier, err := metadata.GetTableIdentifier()
	if err != nil {
		return false, err
	}

	rows, err := metadata.db.Query(`
		SELECT COUNT(*)
		FROM sqlite_master
		WHERE type = 'table'
		AND name = `+tableIdentifier+`;
	`, []driver.Value{})
	if err != nil {
		return false, err
	}
	defer rows.Close()

	data := make([]driver.Value, 1)
	if err = rows.Next(data); err != nil {
		return false, err
	}

	return (data[0].(int64) > 0), nil
}

func (metadata *Metadata) GetTableIdentifier() (string, error) {
	return SanitizeIdentifier(metadata.db, fmt.Sprintf("rodb_%v_metadata", metadata.indexName))
}

// Sets the completed flag, which records wether or not the index
// generation has been finished
func (metadata *Metadata) SetCompleted(completed bool) {
	metadata.completed = completed
}

func (metadata *Metadata) Save() error {
	tableIdentifier, err := metadata.GetTableIdentifier()
	if err != nil {
		return err
	}

	_, err = metadata.db.Exec(`
		CREATE TABLE IF NOT EXISTS `+tableIdentifier+` (
			"version" INTEGER NOT NULL,
			"inputFileModificationTime" INTEGER NOT NULL,
			"inputFileSize" INTEGER NOT NULL,
			"completed" BOOLEAN NOT NULL
		);
	`, []driver.Value{})
	if err != nil {
		return err
	}

	_, err = metadata.db.Exec(`
		DELETE FROM `+tableIdentifier+`;
	`, []driver.Value{})
	if err != nil {
		return err
	}

	_, err = metadata.db.Exec(`
		INSERT INTO `+tableIdentifier+` (
			"version",
			"inputFileModificationTime",
			"inputFileSize",
			"completed"
		) VALUES (?, ?, ?, ?);
	`, []driver.Value{
		int64(metadata.version),
		metadata.inputFileModificationTime.Unix(),
		metadata.inputFileSize,
		metadata.completed,
	})
	if err != nil {
		return err
	}

	return nil
}

// Validates that the metadata of the file is a valid RODB index
// and matches the given configuration as well as the current version
func (metadata *Metadata) AssertValid(input input.Input) error {
	if metadata.version != CurrentVersion {
		return fmt.Errorf("The index file is not compatible with the current version of this software.")
	}

	modTime, err := input.ModTime()
	if err != nil {
		return err
	}
	if metadata.inputFileModificationTime.Unix() != modTime.Unix() {
		return fmt.Errorf("The input file has been modified since the index generation.")
	}

	size, err := input.Size()
	if err != nil {
		return err
	}
	if metadata.inputFileSize != size {
		return fmt.Errorf("The input file size has changed since the index generation.")
	}

	if !metadata.completed {
		return fmt.Errorf("The previous indexing process has not ended properly. Please remove the corrupted file and try again.")
	}

	return nil
}
