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
	ParentProperty string `yaml:"parentProperty"`
	ChildProperty  string `yaml:"childProperty"`
	ChildIndex     string `yaml:"childIndex"`
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

	input, inputExists := rootConfig.Inputs[config.Input]
	if !inputExists {
		return fmt.Errorf("memoryMap.input: Input '%v' not found in inputs list.", config.Input)
	}

	alreadyExistingSortProperties := make(map[string]bool)
	for sortIndex, sort := range config.Sort {
		err := sort.validate(rootConfig, input, log, "jsonObject.relationships[].sort.")
		if err != nil {
			return fmt.Errorf("sort.%v.%w", sortIndex, err)
		}

		if _, alreadyExists := alreadyExistingSortProperties[sort.Property]; alreadyExists {
			return fmt.Errorf("sort.%v.property: property %v is used twice for sorting", sortIndex, sort.Property)
		}
		alreadyExistingSortProperties[sort.Property] = true
	}

	alreadyExistingChildProperty := make(map[string]bool)
	for matchIndex, match := range config.Match {
		logPrefix := fmt.Sprintf("match.%v.", matchIndex)
		err := match.validate(
			rootConfig,
			log,
			logPrefix,
			input,
		)
		if err != nil {
			return fmt.Errorf("%v%w", logPrefix, err)
		}

		if _, alreadyExists := alreadyExistingChildProperty[match.ChildProperty]; alreadyExists {
			return fmt.Errorf("%vchildProperty: Duplicate filter on childProperty %v", logPrefix, match.ChildProperty)
		}
		alreadyExistingChildProperty[match.ChildProperty] = true
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
	input Input,
) error {
	// The parentProperty and childProperty will be validated at runtime (must be validated against the input and not the index)

	if config.ChildIndex == "" {
		log.Debugf(logPrefix + "childIndex is empty. Assuming 'default'.\n")
		config.ChildIndex = "default"
	}

	childIndex, childIndexExists := rootConfig.Indexes[config.ChildIndex]
	if !childIndexExists {
		return fmt.Errorf("childIndex: Index '%v' not found in indexes list.", config.ChildIndex)
	}
	if !childIndex.DoesHandleInput(input) {
		return fmt.Errorf("childIndex: Index '%v' does not handle input '%v'.", config.ChildIndex, input.Name())
	}
	if !childIndex.DoesHandleProperty(config.ChildProperty) {
		return fmt.Errorf("childProperty: Index '%v' does not handle property '%v'.", config.ChildIndex, config.ChildProperty)
	}

	return nil
}
