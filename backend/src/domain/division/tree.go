package division

import "activist-base/src/domain/shared"

// ValidateCreateRoot returns an error if a root node already exists.
func ValidateCreateRoot(existingRootID *shared.DivisionID) error {
	if existingRootID != nil {
		return ErrRootExists
	}
	return nil
}

// ValidateCreateChild ensures a new child node can be placed under parent.
func ValidateCreateChild(childID shared.DivisionID, parent Division) error {
	if childID == parent.ID {
		return ErrSelfParent
	}
	return parent.CanAcceptChild()
}

// ValidateMove ensures that moving node to newParent does not violate tree invariants.
// ancestorsOfNewParent must include all ancestor IDs of newParent (not including newParent itself).
func ValidateMove(nodeID shared.DivisionID, newParent Division, ancestorsOfNewParent []shared.DivisionID) error {
	if nodeID == newParent.ID {
		return ErrSelfParent
	}
	for _, id := range ancestorsOfNewParent {
		if id == nodeID {
			return ErrCycleOnMove
		}
	}
	return newParent.CanAcceptChild()
}
