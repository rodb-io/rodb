package sqlite

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rodb-io/rodb/pkg/input"
	"time"
)

// Current version of the indexing protocol
const CurrentVersion = uint16(1)

type Metadata struct {
	db                        *sql.DB
	indexName                 string
	version                   uint16
	inputFileModificationTime time.Time
	inputFileSize             int64
	completed                 bool
}

func NewMetadata(
	db *sql.DB,
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
	db *sql.DB,
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

	row := metadata.db.QueryRow(`
		SELECT
			"version",
			"inputFileModificationTime",
			"inputFileSize",
			"completed"
		FROM ` + tableIdentifier + `;
	`)
	if err := row.Err(); err != nil {
		return nil, err
	}

	var modificationTime int64
	if err = row.Scan(&metadata.version, &modificationTime, &metadata.inputFileSize, &metadata.completed); err != nil {
		return nil, err
	}
	metadata.inputFileModificationTime = time.Unix(modificationTime, 0)

	return metadata, nil
}

func HasMetadata(
	db *sql.DB,
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

	row := metadata.db.QueryRow(`
		SELECT COUNT(*)
		FROM sqlite_master
		WHERE type = 'table'
		AND name = ` + tableIdentifier + `;
	`)
	if err := row.Err(); err != nil {
		return false, err
	}

	var count int64
	if err = row.Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
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
		CREATE TABLE IF NOT EXISTS ` + tableIdentifier + ` (
			"version" INTEGER NOT NULL,
			"inputFileModificationTime" INTEGER NOT NULL,
			"inputFileSize" INTEGER NOT NULL,
			"completed" BOOLEAN NOT NULL
		);
	`)
	if err != nil {
		return err
	}

	_, err = metadata.db.Exec(`
		DELETE FROM ` + tableIdentifier + `;
	`)
	if err != nil {
		return err
	}

	_, err = metadata.db.Exec(
		`
			INSERT INTO `+tableIdentifier+` (
				"version",
				"inputFileModificationTime",
				"inputFileSize",
				"completed"
			) VALUES (?, ?, ?, ?);
		`,
		int64(metadata.version),
		metadata.inputFileModificationTime.Unix(),
		metadata.inputFileSize,
		metadata.completed,
	)
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
