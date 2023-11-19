package zanzigo

import (
	"cmp"
	"slices"
	"testing"
)

func TestSortCommands(t *testing.T) {
	commands := []Check{
		&CheckIndirect{},
		&CheckDirectUserset{},
		&CheckDirect{},
	}
	commands = sortChecks(commands)
	expected := []Check{
		&CheckDirect{},
		&CheckDirectUserset{},
		&CheckIndirect{},
	}
	if slices.CompareFunc(commands, expected, func(a Check, b Check) int {
		return cmp.Compare(a.Kind(), b.Kind())
	}) != 0 {
		t.Fatalf("Expected order %v but got %v", expected, commands)
	}
}
