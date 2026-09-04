package inspection

import "runtime/debug"

func BuildIdentity() Tool {
	tool := Tool{
		Name: ToolName, Version: "unknown", VersionState: "unknown",
		Revision: "unknown", RevisionState: "unknown",
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return tool
	}
	if info.Main.Version != "" {
		tool.Version = info.Main.Version
		if info.Main.Version == "(devel)" {
			tool.VersionState = "development"
		} else {
			tool.VersionState = "release"
		}
	}
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			if setting.Value != "" {
				tool.Revision = setting.Value
				tool.RevisionState = "known"
			}
		case "vcs.modified":
			tool.Modified = setting.Value == "true"
		}
	}
	return tool
}

func Version() Result {
	result := NewResult("version")
	result.Capabilities = InstalledCapabilities()
	return result
}

func InstalledCapabilities() []Capability {
	return []Capability{
		{Name: "inspection", Status: StatusSupported},
		{Name: "schema.generation", Status: StatusSupported, Detail: "documented v1 shapes"},
		{Name: "json.standard_codec", Status: StatusSupported, Detail: "documented ordinary shapes"},
		{Name: "json.validation", Status: StatusSupported},
		{Name: "yaml.input", Status: StatusSupported},
		{Name: "provider.runtime_schema", Status: StatusSupported},
		{Name: "provider.generated_codec", Status: StatusUnsupported},
		{Name: "provider.static_validation", Status: StatusUnsupported},
		{Name: "union.schema", Status: StatusSupported},
		{Name: "union.json_encode", Status: StatusUnsupported, Detail: "pending generated codec support"},
		{Name: "union.json_decode", Status: StatusSupported},
		{Name: "enum.string_mode.schema", Status: StatusSupported},
		{Name: "enum.string_mode.json_encode", Status: StatusUnsupported, Detail: "pending generated codec support"},
		{Name: "enum.string_mode.json_decode", Status: StatusUnsupported, Detail: "pending generated codec support"},
		{Name: "preview", Status: StatusUnsupported},
	}
}
