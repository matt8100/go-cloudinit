package cloudinit

import "fmt"

// MountEntry represents one mounts item.
type MountEntry struct {
	Spec    string
	File    string
	VFSType string
	Options string
	Freq    string
	PassNo  string
}

// MarshalYAML implements yaml.Marshaler.
func (m MountEntry) MarshalYAML() (any, error) {
	values := []string{m.Spec, m.File, m.VFSType, m.Options, m.Freq, m.PassNo}
	if values[0] == "" {
		return nil, fmt.Errorf("mount entry requires at least Spec")
	}

	last := len(values) - 1
	for last > 0 && values[last] == "" {
		last--
	}

	return values[:last+1], nil
}

// MountDefaults represents mount_default_fields.
type MountDefaults [6]*string

// MarshalYAML implements yaml.Marshaler.
func (m MountDefaults) MarshalYAML() (any, error) {
	values := make([]any, len(m))
	for i, v := range m {
		if v == nil {
			values[i] = nil
			continue
		}
		values[i] = *v
	}
	return values, nil
}

// SwapDefinition configures the swap module.
type SwapDefinition struct {
	Filename string       `yaml:"filename,omitempty"`
	Size     *SizeOrValue `yaml:"size,omitempty"`
	MaxSize  *SizeOrValue `yaml:"maxsize,omitempty"`
}

// SizeOrValue represents swap sizes that may be auto, integer bytes, or strings.
type SizeOrValue struct {
	Auto    bool
	Integer *int64
	String  string
}

// AutoSize returns a SizeOrValue that renders as auto.
func AutoSize() *SizeOrValue { return &SizeOrValue{Auto: true} }

// BytesSize returns a SizeOrValue backed by a byte count.
func BytesSize(v int64) *SizeOrValue { return &SizeOrValue{Integer: &v} }

// HumanSize returns a SizeOrValue backed by a human-readable size string.
func HumanSize(v string) *SizeOrValue { return &SizeOrValue{String: v} }

// MarshalYAML implements yaml.Marshaler.
func (s SizeOrValue) MarshalYAML() (any, error) {
	switch {
	case s.Auto:
		return "auto", nil
	case s.Integer != nil:
		return *s.Integer, nil
	case s.String != "":
		return s.String, nil
	default:
		return nil, nil
	}
}

// PartitionLayout represents one disk layout partition entry.
type PartitionLayout struct {
	Size int
	Type string
}

// MarshalYAML implements yaml.Marshaler.
func (p PartitionLayout) MarshalYAML() (any, error) {
	if p.Size <= 0 {
		return nil, fmt.Errorf("partition size must be > 0")
	}
	if p.Type == "" {
		return p.Size, nil
	}
	return []any{p.Size, p.Type}, nil
}

// DiskLayout represents the layout form accepted by disk_setup.
type DiskLayout struct {
	Remove  bool
	Enabled *bool
	Parts   []PartitionLayout
}

// RemoveLayout returns a disk layout that removes partition data.
func RemoveLayout() *DiskLayout {
	return &DiskLayout{Remove: true}
}

// SimpleLayout returns a bool-backed disk layout.
func SimpleLayout(enabled bool) *DiskLayout {
	return &DiskLayout{Enabled: &enabled}
}

// MarshalYAML implements yaml.Marshaler.
func (l DiskLayout) MarshalYAML() (any, error) {
	switch {
	case l.Remove && l.Enabled == nil && len(l.Parts) == 0:
		return "remove", nil
	case !l.Remove && l.Enabled != nil && len(l.Parts) == 0:
		return *l.Enabled, nil
	case !l.Remove && l.Enabled == nil && len(l.Parts) > 0:
		return l.Parts, nil
	default:
		return nil, fmt.Errorf("layout must define exactly one of Remove, Enabled, or Parts")
	}
}

// DiskSetupConfig represents one disk_setup entry.
type DiskSetupConfig struct {
	TableType string      `yaml:"table_type,omitempty"`
	Layout    *DiskLayout `yaml:"layout,omitempty"`
	Overwrite *bool       `yaml:"overwrite,omitempty"`
}

// DiskSetupMap maps device names to disk setup configuration.
type DiskSetupMap map[string]DiskSetupConfig

// FilesystemSetup represents one fs_setup entry.
type FilesystemSetup struct {
	Label      string               `yaml:"label,omitempty"`
	Filesystem string               `yaml:"filesystem,omitempty"`
	Device     string               `yaml:"device,omitempty"`
	Partition  *StringOrInt         `yaml:"partition,omitempty"`
	Overwrite  *bool                `yaml:"overwrite,omitempty"`
	ReplaceFS  string               `yaml:"replace_fs,omitempty"`
	ExtraOpts  *StringOrStringSlice `yaml:"extra_opts,omitempty"`
	Cmd        *StringOrStringSlice `yaml:"cmd,omitempty"`
}
