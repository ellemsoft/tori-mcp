package main

import "testing"

func TestToriMCPToolsDeclareOutputSchemas(t *testing.T) {
	srv := newToriMCPServer()
	tools := srv.ListTools()

	tests := []struct {
		name        string
		schemaField string
	}{
		{name: "search", schemaField: "docs"},
		{name: "show", schemaField: "id"},
		{name: "filters", schemaField: "filters"},
		{name: "categories", schemaField: "categories"},
		{name: "locations", schemaField: "locations"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := tools[tt.name]
			if tool == nil {
				t.Fatalf("tool %q not registered", tt.name)
			}

			outputSchema := tool.Tool.OutputSchema
			if outputSchema.Type != "object" {
				t.Fatalf("output schema type = %q, want object", outputSchema.Type)
			}
			if _, ok := outputSchema.Properties[tt.schemaField]; !ok {
				t.Fatalf("output schema missing property %q", tt.schemaField)
			}

			annotations := tool.Tool.Annotations
			if annotations.ReadOnlyHint == nil || !*annotations.ReadOnlyHint {
				t.Fatalf("readOnlyHint = %v, want true", annotations.ReadOnlyHint)
			}
			if annotations.DestructiveHint == nil || *annotations.DestructiveHint {
				t.Fatalf("destructiveHint = %v, want false", annotations.DestructiveHint)
			}
			if annotations.IdempotentHint == nil || !*annotations.IdempotentHint {
				t.Fatalf("idempotentHint = %v, want true", annotations.IdempotentHint)
			}
			if annotations.OpenWorldHint == nil || !*annotations.OpenWorldHint {
				t.Fatalf("openWorldHint = %v, want true", annotations.OpenWorldHint)
			}
		})
	}
}
