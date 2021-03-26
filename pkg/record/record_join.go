package record

// Returns the positions that are common to all the given iterators
// Expects each given iterator to be sorted from the smallest to the biggest position
func JoinPositionIterators(iterators ...PositionIterator) PositionIterator {
	if len(iterators) == 0 {
		return func() (*Position, error) {
			return nil, nil
		}
	}
	if len(iterators) == 1 {
		return iterators[0]
	}

	return func() (*Position, error) {
		currentIteratorsValues := make([]Position, len(iterators))
		for i := 1; i < len(currentIteratorsValues); i++ {
			position, err := iterators[i]()
			if err != nil {
				return nil, err
			}
			if position == nil {
				return nil, nil
			}
			currentIteratorsValues[i] = *position
		}

		for {
			firstListPosition, err := iterators[0]()
			if err != nil {
				return nil, err
			}
			if firstListPosition == nil {
				return nil, nil
			}

			foundInAllLists := true
			for listIndex := 1; listIndex < len(iterators); listIndex++ {
				// Advancing the list up to the right position
				for currentIteratorsValues[listIndex] < *firstListPosition {
					currentListValue, err := iterators[listIndex]()
					if err != nil {
						return nil, err
					}
					if currentListValue == nil {
						return nil, nil
					}
					currentIteratorsValues[listIndex] = *currentListValue
				}

				if currentIteratorsValues[listIndex] != *firstListPosition {
					foundInAllLists = false
					break
				}
			}

			if foundInAllLists {
				return firstListPosition, nil
			}
		}
	}
}
