package cloudinit

import (
	"fmt"
	"net/url"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Bool returns a pointer to v.
func Bool(v bool) *bool { return &v }

// Int returns a pointer to v.
func Int(v int) *int { return &v }

// Int64 returns a pointer to v.
func Int64(v int64) *int64 { return &v }

// StringPtr returns a pointer to v.
func StringPtr(v string) *string { return &v }

// SupportedModuleSet returns SupportedModules as a set for fast membership checks.
func SupportedModuleSet() map[string]struct{} {
	modules := make(map[string]struct{}, len(SupportedModules))
	for _, name := range SupportedModules {
		modules[name] = struct{}{}
	}
	return modules
}

func validateURI(v string) bool {
	parsed, err := url.Parse(v)
	if err != nil || parsed.Scheme == "" {
		return false
	}
	if parsed.Scheme == "file" {
		return strings.HasPrefix(parsed.Path, "/")
	}
	return parsed.Host != "" || parsed.Path != "" || parsed.Opaque != ""
}

func valueTrue(v *bool) bool {
	return v != nil && *v
}

func marshalNode(value any) (*yaml.Node, error) {
	var node yaml.Node
	if err := node.Encode(value); err != nil {
		return nil, err
	}
	return &node, nil
}

// BoolOrStringValue represents a YAML value that may be either a bool or string.
type BoolOrStringValue struct {
	Bool   *bool
	String string
}

// BoolValue constructs a BoolOrStringValue from a bool.
func BoolValue(v bool) *BoolOrStringValue {
	return &BoolOrStringValue{Bool: &v}
}

// StringValue constructs a BoolOrStringValue from a string.
func StringValue(v string) *BoolOrStringValue {
	return &BoolOrStringValue{String: v}
}

// MarshalYAML implements yaml.Marshaler.
func (v BoolOrStringValue) MarshalYAML() (any, error) {
	if v.Bool != nil && v.String != "" {
		return nil, fmt.Errorf("cannot define both Bool and String")
	}
	if v.Bool != nil {
		return *v.Bool, nil
	}
	if v.String != "" {
		return v.String, nil
	}
	return nil, nil
}

// IntBoolStringValue represents a YAML value that may be an int, bool, or string.
type IntBoolStringValue struct {
	Int    *int
	Bool   *bool
	String string
}

// IntValue constructs an IntBoolStringValue from an int.
func IntValue(v int) *IntBoolStringValue {
	return &IntBoolStringValue{Int: &v}
}

// IntBoolValue constructs an IntBoolStringValue from a bool.
func IntBoolValue(v bool) *IntBoolStringValue {
	return &IntBoolStringValue{Bool: &v}
}

// IntStringValue constructs an IntBoolStringValue from a string.
func IntStringValue(v string) *IntBoolStringValue {
	return &IntBoolStringValue{String: v}
}

// MarshalYAML implements yaml.Marshaler.
func (v IntBoolStringValue) MarshalYAML() (any, error) {
	count := 0
	if v.Int != nil {
		count++
	}
	if v.Bool != nil {
		count++
	}
	if v.String != "" {
		count++
	}
	if count > 1 {
		return nil, fmt.Errorf("must define exactly one of Int, Bool, or String")
	}
	switch {
	case v.Int != nil:
		return *v.Int, nil
	case v.Bool != nil:
		return *v.Bool, nil
	case v.String != "":
		return v.String, nil
	default:
		return nil, nil
	}
}

// StringOrStringSlice represents a YAML value that may be a string or []string.
type StringOrStringSlice struct {
	String string
	List   []string
}

// MarshalYAML implements yaml.Marshaler.
func (v StringOrStringSlice) MarshalYAML() (any, error) {
	switch {
	case v.String != "" && len(v.List) == 0:
		return v.String, nil
	case v.String == "" && len(v.List) > 0:
		return v.List, nil
	case v.String == "" && len(v.List) == 0:
		return nil, fmt.Errorf("must define either String or List")
	default:
		return nil, fmt.Errorf("must define either String or List")
	}
}

// StringOrInt represents a YAML value that may be a string or int.
type StringOrInt struct {
	String string
	Int    *int
}

// MarshalYAML implements yaml.Marshaler.
func (v StringOrInt) MarshalYAML() (any, error) {
	switch {
	case v.String != "" && v.Int == nil:
		return v.String, nil
	case v.String == "" && v.Int != nil:
		return *v.Int, nil
	case v.String == "" && v.Int == nil:
		return nil, fmt.Errorf("must define either String or Int")
	default:
		return nil, fmt.Errorf("must define either String or Int")
	}
}

// NullableString represents either a string value or an explicit YAML null.
type NullableString struct {
	String *string
	Null   bool
}

// StringSetting constructs a NullableString from a string.
func StringSetting(value string) *NullableString {
	return &NullableString{String: &value}
}

// NullSetting constructs a NullableString that renders as YAML null.
func NullSetting() *NullableString {
	return &NullableString{Null: true}
}

// MarshalYAML implements yaml.Marshaler.
func (v NullableString) MarshalYAML() (any, error) {
	switch {
	case v.String != nil && !v.Null:
		return *v.String, nil
	case v.String == nil && v.Null:
		return nil, nil
	case v.String == nil && !v.Null:
		return nil, nil
	default:
		return nil, fmt.Errorf("must define either String or Null")
	}
}

// RawObject represents an untyped YAML object subtree.
type RawObject map[string]any

func mapHasKeys(m map[string]any, keys ...string) bool {
	for _, key := range keys {
		if _, ok := m[key]; ok {
			return true
		}
	}
	return false
}

func nestedMap(m map[string]any, key string) map[string]any {
	value, ok := m[key]
	if !ok {
		return nil
	}
	child, _ := value.(map[string]any)
	return child
}

// CommandMap maps names to command definitions.
type CommandMap map[string]Command

// CommandCollection represents either a list or map of commands.
type CommandCollection struct {
	List []Command
	Map  CommandMap
}

// MarshalYAML implements yaml.Marshaler.
func (c CommandCollection) MarshalYAML() (any, error) {
	switch {
	case len(c.List) > 0 && len(c.Map) == 0:
		return c.List, nil
	case len(c.Map) > 0 && len(c.List) == 0:
		return c.Map, nil
	case len(c.List) == 0 && len(c.Map) == 0:
		return nil, nil
	default:
		return nil, fmt.Errorf("must define either List or Map")
	}
}

// StringCollection represents either a list or map of strings.
type StringCollection struct {
	List []string
	Map  map[string]string
}

// MarshalYAML implements yaml.Marshaler.
func (c StringCollection) MarshalYAML() (any, error) {
	switch {
	case len(c.List) > 0 && len(c.Map) == 0:
		return c.List, nil
	case len(c.Map) > 0 && len(c.List) == 0:
		return c.Map, nil
	case len(c.List) == 0 && len(c.Map) == 0:
		return nil, nil
	default:
		return nil, fmt.Errorf("must define either List or Map")
	}
}

// ManageEtcHostsMode is the mode accepted by manage_etc_hosts.
type ManageEtcHostsMode string

// ManageEtcHostsMode values supported by cloud-init.
const (
	ManageEtcHostsTrue      ManageEtcHostsMode = "true"
	ManageEtcHostsFalse     ManageEtcHostsMode = "false"
	ManageEtcHostsLocalhost ManageEtcHostsMode = "localhost"
	ManageEtcHostsTemplate  ManageEtcHostsMode = "template"
)

// ManageEtcHostsValue wraps the manage_etc_hosts setting.
type ManageEtcHostsValue struct {
	Mode ManageEtcHostsMode
}

// ManageEtcHostsEnabled enables manage_etc_hosts.
func ManageEtcHostsEnabled() *ManageEtcHostsValue {
	return &ManageEtcHostsValue{Mode: ManageEtcHostsTrue}
}

// ManageEtcHostsDisabled disables manage_etc_hosts.
func ManageEtcHostsDisabled() *ManageEtcHostsValue {
	return &ManageEtcHostsValue{Mode: ManageEtcHostsFalse}
}

// ManageEtcHostsLocalhostValue sets manage_etc_hosts to localhost mode.
func ManageEtcHostsLocalhostValue() *ManageEtcHostsValue {
	return &ManageEtcHostsValue{Mode: ManageEtcHostsLocalhost}
}

// MarshalYAML implements yaml.Marshaler.
func (m ManageEtcHostsValue) MarshalYAML() (any, error) {
	switch m.Mode {
	case ManageEtcHostsTrue:
		return true, nil
	case ManageEtcHostsFalse:
		return false, nil
	case ManageEtcHostsLocalhost, ManageEtcHostsTemplate:
		return string(m.Mode), nil
	default:
		return nil, fmt.Errorf("unsupported manage_etc_hosts mode %q", m.Mode)
	}
}

// LocaleValue represents either a locale name or an enabled/disabled toggle.
type LocaleValue struct {
	Name    string
	Enabled *bool
}

// Locale constructs a LocaleValue from a locale name.
func Locale(name string) *LocaleValue {
	return &LocaleValue{Name: name}
}

// LocaleEnabled constructs a LocaleValue from a boolean.
func LocaleEnabled(enabled bool) *LocaleValue {
	return &LocaleValue{Enabled: &enabled}
}

// MarshalYAML implements yaml.Marshaler.
func (l LocaleValue) MarshalYAML() (any, error) {
	if l.Enabled != nil && l.Name != "" {
		return nil, fmt.Errorf("cannot define both Name and Enabled")
	}
	if l.Enabled != nil {
		return *l.Enabled, nil
	}
	if l.Name != "" {
		return l.Name, nil
	}
	return nil, nil
}

// ResizeRootFSMode is the string mode accepted by resize_rootfs.
type ResizeRootFSMode string

// ResizeRootFSNoBlock is the non-blocking resize_rootfs mode.
const ResizeRootFSNoBlock ResizeRootFSMode = "noblock"

// ResizeRootFSValue represents either a bool or string resize_rootfs value.
type ResizeRootFSValue struct {
	Enabled *bool
	Mode    ResizeRootFSMode
}

// ResizeRootFSEnabled constructs a bool resize_rootfs value.
func ResizeRootFSEnabled(enabled bool) *ResizeRootFSValue {
	return &ResizeRootFSValue{Enabled: &enabled}
}

// ResizeRootFSBackground constructs a noblock resize_rootfs value.
func ResizeRootFSBackground() *ResizeRootFSValue {
	return &ResizeRootFSValue{Mode: ResizeRootFSNoBlock}
}

// MarshalYAML implements yaml.Marshaler.
func (r ResizeRootFSValue) MarshalYAML() (any, error) {
	if r.Enabled != nil && r.Mode != "" {
		return nil, fmt.Errorf("cannot define both Enabled and Mode")
	}
	if r.Enabled != nil {
		return *r.Enabled, nil
	}
	if r.Mode != "" {
		return string(r.Mode), nil
	}
	return nil, nil
}

// SSHKeyType identifies a host key type.
type SSHKeyType string

// SSH key types supported by this package.
const (
	SSHKeyECDSA   SSHKeyType = "ecdsa"
	SSHKeyED25519 SSHKeyType = "ed25519"
	SSHKeyRSA     SSHKeyType = "rsa"
)

func validSSHKeyType(t SSHKeyType) bool {
	return slices.Contains([]SSHKeyType{SSHKeyECDSA, SSHKeyED25519, SSHKeyRSA}, t)
}
