## Outcome

Publish v1.0.0 once the remaining legal, documentation, and practical release checks are complete.

## 1.0 acceptance

- Record the maintainer's intended license and add the license plus any required notices. The repository currently has no tracked LICENSE/COPYING file.
- Ensure concise compatibility and migration guidance covers Go 1.27+, Optional/Nullable semantics, generated union/enum codecs, TypeScript output, custom-method ownership, and YAML support.
- Complete the bounded consumer check in #62 and the advertised-example fixes in #77. Resolve the silent-registration bug in #80.
- Package and install the bounded Codex plugin from #64.
- Tag the exact reviewed commit as v1.0.0, publish useful release notes, and verify a fresh consumer can install that tag without a local `replace`, generate, run its roundtrip/validation path, and pass `--no-changes`.

## Evidence already established

- #36 proved a cold module-cache test run.
- #71/#78 proved the central Go → TypeScript → Go transport path.
- v1.0.0-rc.5 was installed from the public module in a fresh consumer and passed generation, TypeScript compilation, roundtrip execution, and unchanged regeneration; #79 retains the smoke evidence.

Do not require a full agent/model evaluation program, a new structured CLI protocol, or an exhaustive platform matrix for the stable tag. Those are follow-on product investments.

Final dependencies: #77, #80, #73, #64, and the bounded #62 check.
