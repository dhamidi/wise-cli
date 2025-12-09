package queries

import (
	_ "embed"
)

//go:embed instructions.md
var InstructionsContent string

// GetInstructions returns the instructions for AI agents
func GetInstructions() string {
	return InstructionsContent
}
