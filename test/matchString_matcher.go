package test

import (
	"fmt"
	"strings"

	"github.com/onsi/gomega/types"
)

func diffStrings(a, b string) (string, bool) {
	if a == b {
		return "", true
	}

	// Split into lines for better comparison
	linesA := strings.Split(a, "\n")
	linesB := strings.Split(b, "\n")

	// Find first different line
	firstDiff := 0
	for firstDiff < len(linesA) && firstDiff < len(linesB) && linesA[firstDiff] == linesB[firstDiff] {
		firstDiff++
	}

	// Find last different line, going backwards
	lastDiffA := len(linesA) - 1
	lastDiffB := len(linesB) - 1
	for lastDiffA >= firstDiff && lastDiffB >= firstDiff && linesA[lastDiffA] == linesB[lastDiffB] {
		lastDiffA--
		lastDiffB--
	}

	// Build the diff output
	var diff strings.Builder
	diff.WriteString("\nDifference found starting at line " + fmt.Sprint(firstDiff+1) + ":\n\n")

	// Show context before the diff (up to 3 lines)
	contextStart := max(0, firstDiff-3)
	if contextStart < firstDiff {
		diff.WriteString("Context (unchanged lines):\n")
		for i := contextStart; i < firstDiff; i++ {
			diff.WriteString(fmt.Sprintf("  %d: %s\n", i+1, linesA[i]))
		}
		diff.WriteString("\n")
	}

	// Show the actual differences
	diff.WriteString("Actual output:\n")
	for i := firstDiff; i <= lastDiffA; i++ {
		diff.WriteString(fmt.Sprintf("- %d: %s\n", i+1, linesA[i]))
	}

	diff.WriteString("\nExpected output:\n")
	for i := firstDiff; i <= lastDiffB; i++ {
		diff.WriteString(fmt.Sprintf("+ %d: %s\n", i+1, linesB[i]))
	}

	// Show context after the diff (up to 3 lines)
	contextEnd := min(len(linesA), lastDiffA+4)
	if contextEnd > lastDiffA+1 {
		diff.WriteString("\nContext (unchanged lines):\n")
		for i := lastDiffA + 1; i < contextEnd; i++ {
			diff.WriteString(fmt.Sprintf("  %d: %s\n", i+1, linesA[i]))
		}
	}

	return diff.String(), false
}

type matchStringMatcher struct {
	expected string
}

func MatchString(expected string) types.GomegaMatcher {
	return &matchStringMatcher{expected: expected}
}

func (matcher *matchStringMatcher) Match(actual any) (bool, error) {
	actualStr, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("MatchString matcher expects a string")
	}

	// Normalize line endings
	actualStr = strings.ReplaceAll(actualStr, "\r\n", "\n")
	expected := strings.ReplaceAll(matcher.expected, "\r\n", "\n")

	diff, equal := diffStrings(actualStr, expected)
	if !equal {
		return false, fmt.Errorf("strings do not match:%s", diff)
	}

	return true, nil
}

func (matcher *matchStringMatcher) FailureMessage(actual any) string {
	return "Expected the output to match the expected content"
}

func (matcher *matchStringMatcher) NegatedFailureMessage(actual any) string {
	return "Expected the output not to match the expected content"
}
