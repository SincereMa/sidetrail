package seed

import (
	"context"
	"fmt"
	"os"
	"strings"
)

const maxDocumentSize = 5000

// GeneratePrompt reads the specified files and generates a structured prompt
// for the host agent to extract decisions, constraints, and signals.
func GeneratePrompt(_ context.Context, files []string) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("no files specified")
	}

	var builder strings.Builder
	builder.WriteString("# SideTrail Seed Prompt\n\n")
	builder.WriteString("Extract decisions, constraints, and signals from the following project documents.\n")
	builder.WriteString("For each item found, generate a JSON record matching the SideTrail schema.\n\n")
	builder.WriteString("## Schema Reference\n\n")
	builder.WriteString("Required fields: id, kind, scope, subject, reason, source_type, author, created_at, last_verified_at, status\n")
	builder.WriteString("Kinds: decision, constraint, signal, experiment, incident\n")
	builder.WriteString("Source types: human, agent-suggested, scrape, derived\n\n")
	builder.WriteString("## Documents\n\n")

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("read %q: %w", file, err)
		}
		builder.WriteString(fmt.Sprintf("### %s\n\n", file))
		content := string(data)
		if len(content) > maxDocumentSize {
			content = content[:maxDocumentSize] + "\n... (truncated)"
		}
		builder.WriteString(content)
		builder.WriteString("\n\n")
	}

	builder.WriteString("## Output Format\n\n")
	builder.WriteString("Return a JSON array of records. Each record should have:\n")
	builder.WriteString("- kind: decision, constraint, or signal (based on content)\n")
	builder.WriteString("- scope: file or directory path this applies to\n")
	builder.WriteString("- subject: brief title of the item\n")
	builder.WriteString("- reason: why this decision/constraint exists\n")
	builder.WriteString("- status: active\n")
	builder.WriteString("- source_type: derived (since extracted from docs)\n")
	builder.WriteString("- author: agent\n")

	return builder.String(), nil
}
