package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

type Relationship struct {
	Input         string                   `yaml:"input"`
	IsArray       bool                     `yaml:"isArray"`
	Limit         uint                     `yaml:"limit"`
	Sort          []*Sort                  `yaml:"sort"`
	Match         []*RelationshipMatch     `yaml:"match"`
	Relationships map[string]*Relationship `yaml:"relationships"`
}

type RelationshipMatch struct {
	ParentColumn string `yaml:"parentColumn"`
	ChildColumn  string `yaml:"childColumn"`
	ChildIndex   string `yaml:"childIndex"`
}

func (config *Relationship) validate(
	rootConfig *Config,
	log *logrus.Entry,
	logPrefix string,
) error {
	if config.Limit == 0 && config.IsArray {
		log.Debug(logPrefix + "limit is not set. All relationships will be returned.")
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
		logPrefix := fmt.Sprintf("match.%v.", matchIndex)
		err := match.validate(
			rootConfig,
			log,
			logPrefix,
			config.Input,
		)
		if err != nil {
			return fmt.Errorf("%v%w", logPrefix, err)
		}

		if _, alreadyExists := alreadyExistingChildColumn[match.ChildColumn]; alreadyExists {
			return fmt.Errorf("%vchildColumn: Duplicate filter on childColumn %v", logPrefix, match.ChildColumn)
		}
		alreadyExistingChildColumn[match.ChildColumn] = true
	}

	for relationshipName, relationship := range config.Relationships {
		newPrefix := fmt.Sprintf("relationships.%v.", relationshipName)
		err := relationship.validate(rootConfig, log, logPrefix+newPrefix)
		if err != nil {
			return fmt.Errorf("%v%w", newPrefix, err)
		}
	}

	return nil
}

func (config *RelationshipMatch) validate(
	rootConfig *Config,
	log *logrus.Entry,
	logPrefix string,
	parentInputName string,
) error {
	// The parentColumn and childColumn will be validated at runtime (must be validated against the input and not the index)

	if config.ChildIndex == "" {
		log.Debugf(logPrefix + "childIndex is empty. Assuming 'default'.\n")
		config.ChildIndex = "default"
	}

	childIndex, childIndexExists := rootConfig.Indexes[config.ChildIndex]
	if !childIndexExists {
		return fmt.Errorf("childIndex: Index '%v' not found in indexes list.", config.ChildIndex)
	}
	if !childIndex.DoesHandleInput(parentInputName) {
		return fmt.Errorf("childIndex: Index '%v' does not handle input '%v'.", config.ChildIndex, parentInputName)
	}
	if !childIndex.DoesHandleColumn(config.ChildColumn) {
		return fmt.Errorf("childColumn: Index '%v' does not handle column '%v'.", config.ChildIndex, config.ChildColumn)
	}

	return nil
}
