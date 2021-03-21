package record

// Returns the positions that are common to all the given arrays
// Expects each given list to be sorted from the smallest to the biggest position
// Returns a maximum of `limit` records, or everything if `limit` = 0
func JoinPositionLists(limit uint, lists ...PositionList) PositionList {
	if len(lists) == 0 {
		return make(PositionList, 0)
	}
	if len(lists) == 1 {
		if limit == 0 {
			return lists[0]
		} else {
			return lists[0][:limit]
		}
	}

	currentListIndexes := make([]int, len(lists))
	for i := range currentListIndexes {
		currentListIndexes[i] = 0
	}

	newList := make(PositionList, 0)
	for ; currentListIndexes[0] < len(lists[0]); currentListIndexes[0]++ {
		firstListPosition := lists[0][currentListIndexes[0]]

		foundInAllLists := true
		for listIndex := 1; listIndex < len(lists); listIndex++ {
			// Advancing the list up to the right position
			for lists[listIndex][currentListIndexes[listIndex]] < firstListPosition {
				currentListIndexes[listIndex]++
			}

			if lists[listIndex][currentListIndexes[listIndex]] != firstListPosition {
				foundInAllLists = false
				break
			}
		}

		if foundInAllLists {
			newList = append(newList, lists[0][currentListIndexes[0]])
			if limit > 0 && len(newList) >= int(limit) {
				break
			}
		}
	}

	return newList
}
