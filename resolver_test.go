package zanzigo

import (
	"cmp"
	"slices"
	"testing"
)

func TestSortCommands(t *testing.T) {
	commands := []CheckCommand{
		&CheckIndirectCommand{},
		&CheckDirectUsersetCommand{},
		&CheckDirectCommand{},
	}
	commands = sortCommands(commands)
	expected := []CheckCommand{
		&CheckDirectCommand{},
		&CheckDirectUsersetCommand{},
		&CheckIndirectCommand{},
	}
	if slices.CompareFunc(commands, expected, func(a CheckCommand, b CheckCommand) int {
		return cmp.Compare(a.Kind(), b.Kind())
	}) != 0 {
		t.Fatalf("Expected order %v but got %v", expected, commands)
	}
}
