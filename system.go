package cloudinit

import "fmt"

// PhoneHomeField identifies a field that may be posted by phone_home.
type PhoneHomeField string

// PhoneHomeField values supported by cloud-init.
const (
	PhoneHomeAll           PhoneHomeField = "all"
	PhoneHomePubKeyRSA     PhoneHomeField = "pub_key_rsa"
	PhoneHomePubKeyECDSA   PhoneHomeField = "pub_key_ecdsa"
	PhoneHomePubKeyED25519 PhoneHomeField = "pub_key_ed25519"
	PhoneHomeInstanceID    PhoneHomeField = "instance_id"
	PhoneHomeHostname      PhoneHomeField = "hostname"
	PhoneHomeFQDN          PhoneHomeField = "fqdn"
)

// PhoneHomePost describes the post field of phone_home.
type PhoneHomePost struct {
	All    bool
	Fields []PhoneHomeField
}

// PhoneHomeAllFields returns a PhoneHomePost that renders as "all".
func PhoneHomeAllFields() *PhoneHomePost {
	return &PhoneHomePost{All: true}
}

// MarshalYAML implements yaml.Marshaler.
func (p PhoneHomePost) MarshalYAML() (any, error) {
	if p.All && len(p.Fields) > 0 {
		return nil, fmt.Errorf("cannot define All and Fields together")
	}
	if p.All {
		return "all", nil
	}
	fields := make([]string, 0, len(p.Fields))
	for _, field := range p.Fields {
		fields = append(fields, string(field))
	}
	return fields, nil
}

// PhoneHomeConfig configures the phone_home module.
type PhoneHomeConfig struct {
	URL   string         `yaml:"url,omitempty"`
	Post  *PhoneHomePost `yaml:"post,omitempty"`
	Tries *int           `yaml:"tries,omitempty"`
}

// PowerMode identifies the action taken by power_state.
type PowerMode string

// PowerState modes supported by cloud-init.
const (
	PowerOff PowerMode = "poweroff"
	Reboot   PowerMode = "reboot"
	Halt     PowerMode = "halt"
)

// DelayValue represents either an immediate delay or a minute count.
type DelayValue struct {
	Now     bool
	Minutes *int
}

// DelayNow returns a delay value rendered as "now".
func DelayNow() *DelayValue { return &DelayValue{Now: true} }

// DelayMinutes returns a delay value rendered as a minute count.
func DelayMinutes(minutes int) *DelayValue { return &DelayValue{Minutes: &minutes} }

// MarshalYAML implements yaml.Marshaler.
func (d DelayValue) MarshalYAML() (any, error) {
	switch {
	case d.Now && d.Minutes == nil:
		return "now", nil
	case !d.Now && d.Minutes != nil:
		return *d.Minutes, nil
	case !d.Now && d.Minutes == nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("cannot define both now and minutes")
	}
}

// ConditionValue represents the union accepted by power_state.condition.
type ConditionValue struct {
	Bool    *bool
	Command *Command
	String  string
}

// ConditionAlways returns a boolean condition value.
func ConditionAlways(v bool) *ConditionValue { return &ConditionValue{Bool: &v} }

// ConditionCommand returns a command-backed condition value.
func ConditionCommand(command Command) *ConditionValue { return &ConditionValue{Command: &command} }

// ConditionString returns a string-backed condition value.
func ConditionString(command string) *ConditionValue { return &ConditionValue{String: command} }

// MarshalYAML implements yaml.Marshaler.
func (c ConditionValue) MarshalYAML() (any, error) {
	count := 0
	if c.Bool != nil {
		count++
	}
	if c.Command != nil {
		count++
	}
	if c.String != "" {
		count++
	}
	if count > 1 {
		return nil, fmt.Errorf("must define exactly one of Bool, Command, or String")
	}
	switch {
	case c.Bool != nil:
		return *c.Bool, nil
	case c.Command != nil:
		return c.Command.MarshalYAML()
	case c.String != "":
		return c.String, nil
	default:
		return nil, nil
	}
}

// PowerState configures the power_state module.
type PowerState struct {
	Delay     *DelayValue     `yaml:"delay,omitempty"`
	Mode      PowerMode       `yaml:"mode,omitempty"`
	Message   string          `yaml:"message,omitempty"`
	Timeout   *int            `yaml:"timeout,omitempty"`
	Condition *ConditionValue `yaml:"condition,omitempty"`
}

// GrowPartMode identifies the growpart implementation.
type GrowPartMode string

// GrowPart modes supported by cloud-init.
const (
	GrowPartAuto     GrowPartMode = "auto"
	GrowPartGrowPart GrowPartMode = "growpart"
	GrowPartGPart    GrowPartMode = "gpart"
	GrowPartOff      GrowPartMode = "off"
)

// GrowPartConfig configures the growpart module.
type GrowPartConfig struct {
	Mode                   GrowPartMode `yaml:"mode,omitempty"`
	Devices                []string     `yaml:"devices,omitempty"`
	IgnoreGrowrootDisabled *bool        `yaml:"ignore_growroot_disabled,omitempty"`
}

// ByobuMode is the string value accepted by byobu_by_default.
type ByobuMode string

// FanConfig configures the fan module.
type FanConfig struct {
	Config     string `yaml:"config,omitempty"`
	ConfigPath string `yaml:"config_path,omitempty"`
}

// BoolOrStringDeprecated represents deprecated bool-or-string YAML forms.
type BoolOrStringDeprecated struct {
	Bool   *bool
	String string
}

// MarshalYAML implements yaml.Marshaler.
func (v BoolOrStringDeprecated) MarshalYAML() (any, error) {
	if v.Bool != nil && v.String != "" {
		return nil, fmt.Errorf("must define either Bool or String")
	}
	if v.Bool != nil {
		return *v.Bool, nil
	}
	if v.String != "" {
		return v.String, nil
	}
	return nil, fmt.Errorf("must define either Bool or String")
}

// GrubDpkgConfig configures the grub_dpkg module.
type GrubDpkgConfig struct {
	Enabled                   *bool                   `yaml:"enabled,omitempty"`
	GrubPCInstallDevices      string                  `yaml:"grub-pc/install_devices,omitempty"`
	GrubPCInstallDevicesEmpty *BoolOrStringDeprecated `yaml:"grub-pc/install_devices_empty,omitempty"`
	GrubEFIInstallDevices     string                  `yaml:"grub-efi/install_devices,omitempty"`
}

// HotplugNetworkConfig configures the updates.network section.
type HotplugNetworkConfig struct {
	When []string `yaml:"when,omitempty"`
}

// HotplugUpdates configures the updates module.
type HotplugUpdates struct {
	Network *HotplugNetworkConfig `yaml:"network,omitempty"`
}

// KeyboardConfig configures the keyboard module.
type KeyboardConfig struct {
	Layout  string `yaml:"layout,omitempty"`
	Model   string `yaml:"model,omitempty"`
	Variant string `yaml:"variant,omitempty"`
	Options string `yaml:"options,omitempty"`
}

// NTPConfig stores ntp configuration as a raw object.
type NTPConfig map[string]any

// ResolvConfConfig configures the resolv_conf module.
type ResolvConfConfig struct {
	Nameservers   []string       `yaml:"nameservers,omitempty"`
	SearchDomains []string       `yaml:"searchdomains,omitempty"`
	Domain        string         `yaml:"domain,omitempty"`
	SortList      []string       `yaml:"sortlist,omitempty"`
	Options       map[string]any `yaml:"options,omitempty"`
}

// RandomSeedConfig configures the random_seed module.
type RandomSeedConfig struct {
	File            string   `yaml:"file,omitempty"`
	Data            string   `yaml:"data,omitempty"`
	Encoding        string   `yaml:"encoding,omitempty"`
	Command         []string `yaml:"command,omitempty"`
	CommandRequired *bool    `yaml:"command_required,omitempty"`
}

// NVIDIADrivers configures the drivers.nvidia section.
type NVIDIADrivers struct {
	LicenseAccepted *bool  `yaml:"license-accepted,omitempty"`
	Version         string `yaml:"version,omitempty"`
}

// UbuntuDriversConfig configures the drivers module.
type UbuntuDriversConfig struct {
	Nvidia *NVIDIADrivers `yaml:"nvidia,omitempty"`
}

// UbuntuAutoinstallConfig configures the autoinstall module.
type UbuntuAutoinstallConfig struct {
	Version int            `yaml:"version,omitempty"`
	Extra   map[string]any `yaml:",inline,omitempty"`
}

// WireGuardInterface represents one wireguard interface definition.
type WireGuardInterface struct {
	Name       string `yaml:"name,omitempty"`
	ConfigPath string `yaml:"config_path,omitempty"`
	Content    string `yaml:"content,omitempty"`
}

// WireGuardConfig configures the wireguard module.
type WireGuardConfig struct {
	Interfaces     []WireGuardInterface `yaml:"interfaces,omitempty"`
	ReadinessProbe []string             `yaml:"readinessprobe,omitempty"`
}

// SpacewalkConfig configures the spacewalk module.
type SpacewalkConfig struct {
	Server        string `yaml:"server,omitempty"`
	Proxy         string `yaml:"proxy,omitempty"`
	ActivationKey string `yaml:"activation_key,omitempty"`
}
