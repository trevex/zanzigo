package zanzigo

import (
	"cmp"
	"slices"
	"testing"
)

func TestSortCommands(t *testing.T) {
	commands := []Command{
		&CheckIndirectCommand{},
		&CheckDirectUsersetCommand{},
		&CheckDirectCommand{},
	}
	commands = sortCommands(commands)
	expected := []Command{
		&CheckDirectCommand{},
		&CheckDirectUsersetCommand{},
		&CheckIndirectCommand{},
	}
	if slices.CompareFunc(commands, expected, func(a Command, b Command) int {
		return cmp.Compare(a.Kind(), b.Kind())
	}) != 0 {
		t.Fatalf("Expected order %v but got %v", expected, commands)
	}
}
