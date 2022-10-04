package main

import (
	"fmt"
	"testing"
)

func Test_ShouldContinueBasedOnResults(t *testing.T) {
	tests := map[string]struct {
		input          []int
		syncAll        bool
		expectedOutput bool
	}{
		"empty input": {
			input:          []int{},
			expectedOutput: true,
		},
		"one input element": {
			input:          []int{20},
			expectedOutput: true,
		},
		"multiple input": {
			input:          []int{20, 20, 10, 9, 1, 0, 10},
			expectedOutput: true,
		},
		"last one zero": {
			input:          []int{20, 10, 10, 0},
			expectedOutput: true,
		},
		"last two zero": {
			input:          []int{20, 20, 10, 0, 0},
			expectedOutput: false,
		},
		"syncAll overrides everything else": {
			input:          []int{20, 20, 10, 0, 0},
			syncAll:        true,
			expectedOutput: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := ShouldContinueBasedOnResults(test.input, test.syncAll)
			if result != test.expectedOutput {
				fmt.Println(fmt.Sprintf("%s: %v does not equal %v", name, result, test.expectedOutput))
				t.Fail()
			}
		})
	}
}
