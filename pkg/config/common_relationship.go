package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

type Relationship struct {
	Input         string                   `yaml:"input"`
	Index         string                   `yaml:"index"`
	IsArray       bool                     `yaml:"isArray"`
	Limit         uint                     `yaml:"limit"`
	Sort          []*Sort                  `yaml:"sort"`
	Match         []*RelationshipMatch     `yaml:"match"`
	Relationships map[string]*Relationship `yaml:"relationships"`
}

type RelationshipMatch struct {
	ParentColumn string `yaml:"parentColumn"`
	ChildColumn  string `yaml:"childColumn"`
}

func (config *Relationship) validate(
	rootConfig *Config,
	log *logrus.Entry,
	logPrefix string,
	parentIndexName string,
	parentIndex Index,
) error {
	if config.Limit == 0 && config.IsArray {
		log.Debug(logPrefix + "limit is not set. All relationships will be returned.")
	}

	if config.Index == "" {
		log.Debugf(logPrefix + "index is empty. Assuming 'default'.\n")
		config.Index = "default"
	}

	childIndex, childIndexExists := rootConfig.Indexes[config.Index]
	if !childIndexExists {
		return fmt.Errorf("index: Index '%v' not found in indexes list.", config.Index)
	}
	if !childIndex.DoesHandleInput(config.Input) {
		return fmt.Errorf("input: Index '%v' does not handle input '%v'.", config.Index, config.Input)
	}

	if config.Sort == nil {
		config.Sort = make([]*Sort, 0)
	}

	if len(config.Sort) > 0 && !config.IsArray {
		return fmt.Errorf("sort: You can only sort a relationship when isArray = 'true'.")
	}

	alreadyExistingSortColumns := make(map[string]bool)
	for sortIndex, sort := range config.Sort {
		err := sort.validate(rootConfig, log, "jsonObject.relationships[].sort.")
		if err != nil {
			return fmt.Errorf("sort.%v.%w", sortIndex, err)
		}

		if _, alreadyExists := alreadyExistingSortColumns[sort.Column]; alreadyExists {
			return fmt.Errorf("sort.%v.column: column %v is used twice for sorting", sortIndex, sort.Column)
		}
		alreadyExistingSortColumns[sort.Column] = true
	}

	alreadyExistingChildColumn := make(map[string]bool)
	for matchIndex, match := range config.Match {
		err := match.validate(
			rootConfig,
			log,
			config.Index,
			childIndex,
		)
		if err != nil {
			return fmt.Errorf("match.%v.%w", matchIndex, err)
		}

		if _, alreadyExists := alreadyExistingChildColumn[match.ChildColumn]; alreadyExists {
			return fmt.Errorf("match.%v.childColumn: Duplicate filter on childColumn %v", matchIndex, match.ChildColumn)
		}
		alreadyExistingChildColumn[match.ChildColumn] = true
	}

	for relationshipName, relationship := range config.Relationships {
		newPrefix := fmt.Sprintf("relationships.%v.", relationshipName)
		err := relationship.validate(rootConfig, log, logPrefix+newPrefix, config.Index, childIndex)
		if err != nil {
			return fmt.Errorf("%v%w", newPrefix, err)
		}
	}

	return nil
}

func (config *RelationshipMatch) validate(
	rootConfig *Config,
	log *logrus.Entry,
	childIndexName string,
	childIndex Index,
) error {
	// The parentColumn and childColumn will be validated at runtime (must be validated against the input and not the index)

	if !childIndex.DoesHandleColumn(config.ChildColumn) {
		return fmt.Errorf("childColumn: Index '%v' does not handle column '%v'.", childIndexName, config.ChildColumn)
	}

	return nil
}
