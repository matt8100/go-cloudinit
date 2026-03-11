package cloudinit

import "fmt"

// PackageManager identifies a package manager in a managed package entry.
type PackageManager string

// Package managers supported by ManagedPackageGroup.
const (
	PackageManagerAPT  PackageManager = "apt"
	PackageManagerSnap PackageManager = "snap"
)

// PackageSpec describes a package name with an optional version.
type PackageSpec struct {
	Name    string
	Version string
}

// MarshalYAML implements yaml.Marshaler.
func (p PackageSpec) MarshalYAML() (any, error) {
	if p.Name == "" {
		return nil, fmt.Errorf("package name is required")
	}
	if p.Version == "" {
		return p.Name, nil
	}
	return []string{p.Name, p.Version}, nil
}

// ManagedPackageGroup groups packages under a specific manager key.
type ManagedPackageGroup struct {
	APT  []PackageSpec
	Snap []PackageSpec
}

// MarshalYAML implements yaml.Marshaler.
func (g ManagedPackageGroup) MarshalYAML() (any, error) {
	switch {
	case len(g.APT) > 0 && len(g.Snap) == 0:
		return map[string]any{"apt": g.APT}, nil
	case len(g.Snap) > 0 && len(g.APT) == 0:
		return map[string]any{"snap": g.Snap}, nil
	default:
		return nil, fmt.Errorf("managed package group must define exactly one manager")
	}
}

// PackageEntry represents one item in the packages list.
type PackageEntry struct {
	Inline  *PackageSpec
	Managed *ManagedPackageGroup
}

// Package returns a package entry for a single package name.
func Package(name string) PackageEntry {
	return PackageEntry{Inline: &PackageSpec{Name: name}}
}

// VersionedPackage returns a package entry with an explicit version.
func VersionedPackage(name, version string) PackageEntry {
	return PackageEntry{Inline: &PackageSpec{Name: name, Version: version}}
}

// APTPackages returns a managed package entry under the apt key.
func APTPackages(specs ...PackageSpec) PackageEntry {
	return PackageEntry{Managed: &ManagedPackageGroup{APT: specs}}
}

// SnapPackages returns a managed package entry under the snap key.
func SnapPackages(specs ...PackageSpec) PackageEntry {
	return PackageEntry{Managed: &ManagedPackageGroup{Snap: specs}}
}

// MarshalYAML implements yaml.Marshaler.
func (e PackageEntry) MarshalYAML() (any, error) {
	switch {
	case e.Inline != nil && e.Managed == nil:
		return e.Inline.MarshalYAML()
	case e.Managed != nil && e.Inline == nil:
		return e.Managed.MarshalYAML()
	default:
		return nil, fmt.Errorf("package entry must define either Inline or Managed")
	}
}

// APTMirrorConfig represents one apt mirror object.
type APTMirrorConfig map[string]any

// APTSource represents one named entry in apt.sources.
type APTSource struct {
	Source    string `yaml:"source,omitempty"`
	KeyID     string `yaml:"keyid,omitempty"`
	Key       string `yaml:"key,omitempty"`
	KeyServer string `yaml:"keyserver,omitempty"`
	Filename  string `yaml:"filename,omitempty"`
	Append    *bool  `yaml:"append,omitempty"`
}

// APTConfig configures the apt module.
type APTConfig struct {
	PreserveSourcesList *bool                `yaml:"preserve_sources_list,omitempty"`
	GenerateMirrorLists *bool                `yaml:"generate_mirrorlists,omitempty"`
	DisableSuites       []string             `yaml:"disable_suites,omitempty"`
	Primary             []APTMirrorConfig    `yaml:"primary,omitempty"`
	Security            []APTMirrorConfig    `yaml:"security,omitempty"`
	AddAPTRepoMatch     string               `yaml:"add_apt_repo_match,omitempty"`
	DebconfSelections   map[string]string    `yaml:"debconf_selections,omitempty"`
	SourcesList         string               `yaml:"sources_list,omitempty"`
	Conf                string               `yaml:"conf,omitempty"`
	HTTPSProxy          string               `yaml:"https_proxy,omitempty"`
	HTTPProxy           string               `yaml:"http_proxy,omitempty"`
	Proxy               string               `yaml:"proxy,omitempty"`
	FTPProxy            string               `yaml:"ftp_proxy,omitempty"`
	Sources             map[string]APTSource `yaml:"sources,omitempty"`
}

// AlpineRepo configures the alpine_repo section of apk_repos.
type AlpineRepo struct {
	BaseURL          string `yaml:"base_url,omitempty"`
	CommunityEnabled *bool  `yaml:"community_enabled,omitempty"`
	TestingEnabled   *bool  `yaml:"testing_enabled,omitempty"`
	Version          string `yaml:"version,omitempty"`
}

// APKReposConfig configures the apk_repos module.
type APKReposConfig struct {
	PreserveRepositories *bool       `yaml:"preserve_repositories,omitempty"`
	AlpineRepo           *AlpineRepo `yaml:"alpine_repo,omitempty"`
	LocalRepoBaseURL     string      `yaml:"local_repo_base_url,omitempty"`
}

// SnapConfig configures snap assertions and commands.
type SnapConfig struct {
	Assertions *StringCollection  `yaml:"assertions,omitempty"`
	Commands   *CommandCollection `yaml:"commands,omitempty"`
}

// YumRepo represents one yum repository definition.
type YumRepo struct {
	BaseURL    string         `yaml:"baseurl,omitempty"`
	Metalink   string         `yaml:"metalink,omitempty"`
	MirrorList string         `yaml:"mirrorlist,omitempty"`
	Name       string         `yaml:"name,omitempty"`
	Enabled    *bool          `yaml:"enabled,omitempty"`
	Options    map[string]any `yaml:",inline,omitempty"`
}

// YumRepos maps yum repository IDs to their configuration.
type YumRepos map[string]YumRepo

// ZypperRepo represents one zypper repository entry.
type ZypperRepo struct {
	ID      string         `yaml:"id,omitempty"`
	BaseURL string         `yaml:"baseurl,omitempty"`
	Options map[string]any `yaml:",inline,omitempty"`
}

// ZypperConfig configures zypper repos and global settings.
type ZypperConfig struct {
	Repos  []ZypperRepo   `yaml:"repos,omitempty"`
	Config map[string]any `yaml:"config,omitempty"`
}
