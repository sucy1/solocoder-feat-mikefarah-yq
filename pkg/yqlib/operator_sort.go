package yqlib

import (
	"container/list"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type sortByPreferences struct {
	MissingFieldsLast bool
	Reverse           bool
}

type SortByPreferences = sortByPreferences

var ConfiguredSortByPreferences = sortByPreferences{}

func ApplySortByPreferences(node *ExpressionNode, prefs sortByPreferences) {
	if node == nil {
		return
	}
	if node.Operation != nil && node.Operation.OperationType == sortByOpType {
		node.Operation.Preferences = prefs
	}
	ApplySortByPreferences(node.LHS, prefs)
	ApplySortByPreferences(node.RHS, prefs)
}

func sortOperator(d *dataTreeNavigator, context Context, expressionNode *ExpressionNode) (Context, error) {
	selfExpression := &ExpressionNode{Operation: &Operation{OperationType: selfReferenceOpType}}
	expressionNode.RHS = selfExpression
	return sortByOperator(d, context, expressionNode)
}

func sortByOperator(d *dataTreeNavigator, context Context, expressionNode *ExpressionNode) (Context, error) {

	results := list.New()

	prefs := sortByPreferences{}
	if expressionNode.Operation.Preferences != nil {
		prefs = expressionNode.Operation.Preferences.(sortByPreferences)
	}
	if ConfiguredSortByPreferences.MissingFieldsLast {
		prefs.MissingFieldsLast = true
	}
	if ConfiguredSortByPreferences.Reverse {
		prefs.Reverse = true
	}

	for el := context.MatchingNodes.Front(); el != nil; el = el.Next() {
		candidate := el.Value.(*CandidateNode)

		var sortableArray sortableNodeArray

		if candidate.CanVisitValues() {
			sortableArray = make(sortableNodeArray, 0)
			visitor := func(valueNode *CandidateNode) error {
				compareContext, err := d.GetMatchingNodes(context.SingleReadonlyChildContext(valueNode), expressionNode.RHS)
				if err != nil {
					return err
				}
				sortableNode := sortableNode{Node: valueNode, CompareContext: compareContext, dateTimeLayout: context.GetDateTimeLayout(), prefs: prefs}
				sortableArray = append(sortableArray, sortableNode)
				return nil
			}
			if err := candidate.VisitValues(visitor); err != nil {
				return context, err
			}
		} else {
			return context, fmt.Errorf("node at path [%v] is not an array or map (it's a %v)", candidate.GetNicePath(), candidate.Tag)
		}

		sort.Stable(sortableArray)

		sortedList := candidate.CopyWithoutContent()
		switch candidate.Kind {
		case MappingNode:
			for _, sortedNode := range sortableArray {
				sortedList.AddKeyValueChild(sortedNode.Node.Key, sortedNode.Node)
			}
		case SequenceNode:
			for _, sortedNode := range sortableArray {
				sortedList.AddChild(sortedNode.Node)
			}
		}

		results.PushBack(sortedList)
	}
	return context.ChildContext(results), nil
}

type sortableNode struct {
	Node           *CandidateNode
	CompareContext Context
	dateTimeLayout string
	prefs          sortByPreferences
}

type sortableNodeArray []sortableNode

func (a sortableNodeArray) Len() int      { return len(a) }
func (a sortableNodeArray) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func (a sortableNodeArray) Less(i, j int) bool {
	lhsContext := a[i].CompareContext
	rhsContext := a[j].CompareContext

	lhsMissing := lhsContext.MatchingNodes.Len() == 0
	rhsMissing := rhsContext.MatchingNodes.Len() == 0

	if !lhsMissing && isAllNull(lhsContext) {
		lhsMissing = true
	}
	if !rhsMissing && isAllNull(rhsContext) {
		rhsMissing = true
	}

	if lhsMissing && rhsMissing {
		return false
	}
	if lhsMissing {
		return !a[i].prefs.MissingFieldsLast
	}
	if rhsMissing {
		return a[i].prefs.MissingFieldsLast
	}

	rhsEl := rhsContext.MatchingNodes.Front()

	var result int
	for lhsEl := lhsContext.MatchingNodes.Front(); lhsEl != nil && rhsEl != nil; lhsEl = lhsEl.Next() {
		lhs := lhsEl.Value.(*CandidateNode)
		rhs := rhsEl.Value.(*CandidateNode)

		result = a.compare(lhs, rhs, a[i].dateTimeLayout)

		if result < 0 {
			break
		} else if result > 0 {
			break
		}

		rhsEl = rhsEl.Next()
	}

	if result == 0 {
		result = lhsContext.MatchingNodes.Len() - rhsContext.MatchingNodes.Len()
	}

	if a[i].prefs.Reverse {
		return result > 0
	}
	return result < 0
}

func isAllNull(context Context) bool {
	for el := context.MatchingNodes.Front(); el != nil; el = el.Next() {
		node := el.Value.(*CandidateNode)
		tag := node.Tag
		if !strings.HasPrefix(tag, "!!") {
			tag = node.guessTagFromCustomType()
		}
		if tag != "!!null" {
			return false
		}
	}
	return context.MatchingNodes.Len() > 0
}

func (a sortableNodeArray) compare(lhs *CandidateNode, rhs *CandidateNode, dateTimeLayout string) int {
	lhsTag := lhs.Tag
	rhsTag := rhs.Tag

	if !strings.HasPrefix(lhsTag, "!!") {
		// custom tag - we have to have a guess
		lhsTag = lhs.guessTagFromCustomType()
	}

	if !strings.HasPrefix(rhsTag, "!!") {
		// custom tag - we have to have a guess
		rhsTag = rhs.guessTagFromCustomType()
	}

	isDateTime := lhsTag == "!!timestamp" && rhsTag == "!!timestamp"
	layout := dateTimeLayout
	// if the lhs is a string, it might be a timestamp in a custom format.
	if lhsTag == "!!str" && layout != time.RFC3339 {
		_, errLhs := parseDateTime(layout, lhs.Value)
		_, errRhs := parseDateTime(layout, rhs.Value)
		isDateTime = errLhs == nil && errRhs == nil
	}

	if lhsTag == "!!null" && rhsTag != "!!null" {
		return -1
	} else if lhsTag != "!!null" && rhsTag == "!!null" {
		return 1
	} else if lhsTag == "!!bool" && rhsTag != "!!bool" {
		return -1
	} else if lhsTag != "!!bool" && rhsTag == "!!bool" {
		return 1
	} else if lhsTag == "!!bool" && rhsTag == "!!bool" {
		lhsTruthy := isTruthyNode(lhs)

		rhsTruthy := isTruthyNode(rhs)
		if lhsTruthy == rhsTruthy {
			return 0
		} else if lhsTruthy {
			return 1
		}
		return -1
	} else if isDateTime {
		lhsTime, err := parseDateTime(layout, lhs.Value)
		if err != nil {
			log.Warningf("Could not parse time %v with layout %v for sort, sorting by string instead: %v", lhs.Value, layout, err)
			return strings.Compare(lhs.Value, rhs.Value)
		}
		rhsTime, err := parseDateTime(layout, rhs.Value)
		if err != nil {
			log.Warningf("Could not parse time %v with layout %v for sort, sorting by string instead: %v", rhs.Value, layout, err)
			return strings.Compare(lhs.Value, rhs.Value)
		}
		if lhsTime.Equal(rhsTime) {
			return 0
		} else if lhsTime.Before(rhsTime) {
			return -1
		}

		return 1
	} else if lhsTag == "!!int" && rhsTag == "!!int" {
		_, lhsNum, err := parseInt64(lhs.Value)
		if err != nil {
			return strings.Compare(lhs.Value, rhs.Value)
		}
		_, rhsNum, err := parseInt64(rhs.Value)
		if err != nil {
			return strings.Compare(lhs.Value, rhs.Value)
		}
		return int(lhsNum - rhsNum)
	} else if (lhsTag == "!!int" || lhsTag == "!!float") && (rhsTag == "!!int" || rhsTag == "!!float") {
		lhsNum, err := strconv.ParseFloat(lhs.Value, 64)
		if err != nil {
			return strings.Compare(lhs.Value, rhs.Value)
		}
		rhsNum, err := strconv.ParseFloat(rhs.Value, 64)
		if err != nil {
			return strings.Compare(lhs.Value, rhs.Value)
		}
		if lhsNum == rhsNum {
			return 0
		} else if lhsNum < rhsNum {
			return -1
		}

		return 1
	}

	return strings.Compare(lhs.Value, rhs.Value)
}
