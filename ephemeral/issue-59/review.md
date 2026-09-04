# Adversarial review: issue #59 helper fix

Target: `f1f4e21f831775c297e462fd382f3c1cea68ceab` in the `v1-helper-fixes` worktree.

Evidence inspected:

- GitHub issue #59 and its acceptance criteria.
- Parent contract draft `../v1-initial/docs/spec/v1.md`, especially the manual-helper decisions at lines 84-88.
- The complete commit diff for `json_schema.go`, `json_schema_helpers_test.go`, and the worklog.
- Existing helper tests covering strict and non-strict `ObjectSchema`/`JSONSchema` behavior.
- Focused command: `go test . -run 'Issue59|ObjectSchema|JSONSchemaStrict' -count=1 -v` (passed).

Findings: none. Named `~int`/`~string` values classify by underlying kind; property keys are encoded through `json.Marshal`; strict `JSONSchema` and `ObjectSchema` force all declared properties and `additionalProperties:false` while non-strict paths preserve caller settings; strict map-derived required order is sorted; and empty `EnumSchema` returns a `SchemaNode` whose marshal reports an error without construction panic. Arbitrary custom JSON hook transformations are explicitly conditional/outside the v1 guarantee in the parent contract, so their absence from this helper fix is not a material defect.

Outcome: no findings
