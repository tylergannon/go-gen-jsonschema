package ptrfixture

import (
	"encoding/json"
	"testing"
)

// TestPointerRootRenderedSchemaExecutesProviders is recovered evidence (see
// ephemeral/issue-73/manager-validation/run/events/000001-validate.jsonl,
// event 87) that the validator ran and then discarded: it proves a
// pointer-root fluent chain (Declare((*Thing).Schema).Accessor(...).Method(...))
// actually executes its providers at runtime through the generated
// RenderedSchema(), not just that the scanner accepted and generated
// template-hole placeholders for them.
func TestPointerRootRenderedSchemaExecutesProviders(t *testing.T) {
	th := &Thing{Name: "widget", Count: 7}
	rendered, err := th.RenderedSchema()
	if err != nil {
		t.Fatalf("RenderedSchema() error = %v", err)
	}
	var v map[string]any
	if err := json.Unmarshal(rendered, &v); err != nil {
		t.Fatalf("rendered schema is not valid JSON: %v\n%s", err, rendered)
	}
	props := v["properties"].(map[string]any)
	name := props["name"].(map[string]any)
	count := props["count"].(map[string]any)
	if name["description"] != "pointer-root accessor provider ran" {
		t.Fatalf("Accessor provider did not run: %#v", name)
	}
	if count["description"] != "pointer-root method provider ran" {
		t.Fatalf("Method provider did not run: %#v", count)
	}
	t.Logf("rendered schema: %s", rendered)
}
