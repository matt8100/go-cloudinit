package cloudinit

// Header is the prefix written before rendered cloud-config YAML.
const Header = "#cloud-config\n"

// SupportedModules lists the top-level cloud-config keys supported by this package.
var SupportedModules = []string{
	"allow_public_ssh_keys",
	"ansible",
	"apk_repos",
	"apt",
	"apt_pipelining",
	"authkey_hash",
	"autoinstall",
	"bootcmd",
	"byobu_by_default",
	"ca_certs",
	"chef",
	"chpasswd",
	"create_hostname_file",
	"device_aliases",
	"disable_ec2_metadata",
	"disable_root",
	"disable_root_opts",
	"disk_setup",
	"drivers",
	"fan",
	"final_message",
	"fqdn",
	"fs_setup",
	"grub_dpkg",
	"groups",
	"growpart",
	"hostname",
	"keyboard",
	"landscape",
	"locale",
	"locale_configfile",
	"lxd",
	"manage_etc_hosts",
	"manage_resolv_conf",
	"mcollective",
	"mount_default_fields",
	"mounts",
	"no_ssh_fingerprints",
	"ntp",
	"package_reboot_if_required",
	"package_update",
	"package_upgrade",
	"packages",
	"password",
	"phone_home",
	"power_state",
	"prefer_fqdn_over_hostname",
	"preserve_hostname",
	"puppet",
	"random_seed",
	"resolv_conf",
	"resize_rootfs",
	"rh_subscription",
	"rsyslog",
	"runcmd",
	"salt_minion",
	"swap",
	"spacewalk",
	"snap",
	"ssh",
	"ssh_authorized_keys",
	"ssh_deletekeys",
	"ssh_fp_console_blacklist",
	"ssh_genkeytypes",
	"ssh_import_id",
	"ssh_key_console_blacklist",
	"ssh_keys",
	"ssh_publish_hostkeys",
	"ssh_pwauth",
	"ssh_quiet_keygen",
	"timezone",
	"ubuntu_pro",
	"updates",
	"user",
	"users",
	"wireguard",
	"write_files",
	"yum_repo_dir",
	"yum_repos",
	"zypper",
}

// Config represents a single #cloud-config document.
type Config struct {
	PreserveHostname       *bool                `yaml:"preserve_hostname,omitempty"`
	Hostname               string               `yaml:"hostname,omitempty"`
	FQDN                   string               `yaml:"fqdn,omitempty"`
	PreferFQDNOverHostname *bool                `yaml:"prefer_fqdn_over_hostname,omitempty"`
	CreateHostnameFile     *bool                `yaml:"create_hostname_file,omitempty"`
	ManageEtcHosts         *ManageEtcHostsValue `yaml:"manage_etc_hosts,omitempty"`
	Locale                 *LocaleValue         `yaml:"locale,omitempty"`
	LocaleConfigFile       string               `yaml:"locale_configfile,omitempty"`
	Timezone               string               `yaml:"timezone,omitempty"`

	PackageUpdate           *bool               `yaml:"package_update,omitempty"`
	PackageUpgrade          *bool               `yaml:"package_upgrade,omitempty"`
	PackageRebootIfRequired *bool               `yaml:"package_reboot_if_required,omitempty"`
	Packages                []PackageEntry      `yaml:"packages,omitempty"`
	APT                     *APTConfig          `yaml:"apt,omitempty"`
	APKRepos                *APKReposConfig     `yaml:"apk_repos,omitempty"`
	APTPipelining           *IntBoolStringValue `yaml:"apt_pipelining,omitempty"`
	Snap                    *SnapConfig         `yaml:"snap,omitempty"`
	YumRepoDir              string              `yaml:"yum_repo_dir,omitempty"`
	YumRepos                YumRepos            `yaml:"yum_repos,omitempty"`
	Zypper                  *ZypperConfig       `yaml:"zypper,omitempty"`

	BootCmd []Command `yaml:"bootcmd,omitempty"`
	RunCmd  []Command `yaml:"runcmd,omitempty"`

	WriteFiles []WriteFile `yaml:"write_files,omitempty"`

	Groups   []GroupEntry `yaml:"groups,omitempty"`
	User     *UserEntry   `yaml:"user,omitempty"`
	Users    []UserEntry  `yaml:"users,omitempty"`
	Password string       `yaml:"password,omitempty"`
	ChPasswd *ChPasswd    `yaml:"chpasswd,omitempty"`

	SSHPWAuth              *BoolOrStringValue  `yaml:"ssh_pwauth,omitempty"`
	SSHAuthorizedKeys      []string            `yaml:"ssh_authorized_keys,omitempty"`
	SSHImportID            []string            `yaml:"ssh_import_id,omitempty"`
	SSHKeys                *SSHKeys            `yaml:"ssh_keys,omitempty"`
	SSHDeleteKeys          *bool               `yaml:"ssh_deletekeys,omitempty"`
	SSHGenKeyTypes         []SSHKeyType        `yaml:"ssh_genkeytypes,omitempty"`
	DisableRoot            *bool               `yaml:"disable_root,omitempty"`
	DisableRootOpts        *string             `yaml:"disable_root_opts,omitempty"`
	AllowPublicSSHKeys     *bool               `yaml:"allow_public_ssh_keys,omitempty"`
	SSHQuietKeygen         *bool               `yaml:"ssh_quiet_keygen,omitempty"`
	SSHPublishHostKeys     *SSHPublishHostKeys `yaml:"ssh_publish_hostkeys,omitempty"`
	NoSSHFingerprints      *bool               `yaml:"no_ssh_fingerprints,omitempty"`
	AuthKeyHash            string              `yaml:"authkey_hash,omitempty"`
	SSHConsole             *SSHConsoleConfig   `yaml:"ssh,omitempty"`
	SSHKeyConsoleBlacklist []string            `yaml:"ssh_key_console_blacklist,omitempty"`
	SSHFPConsoleBlacklist  []string            `yaml:"ssh_fp_console_blacklist,omitempty"`

	CACerts      *CACertsConfig `yaml:"ca_certs,omitempty"`
	FinalMessage string         `yaml:"final_message,omitempty"`

	Mounts             []MountEntry      `yaml:"mounts,omitempty"`
	MountDefaultFields *MountDefaults    `yaml:"mount_default_fields,omitempty"`
	Swap               *SwapDefinition   `yaml:"swap,omitempty"`
	DeviceAliases      map[string]string `yaml:"device_aliases,omitempty"`
	DiskSetup          DiskSetupMap      `yaml:"disk_setup,omitempty"`
	FSSetup            []FilesystemSetup `yaml:"fs_setup,omitempty"`

	PhoneHome          *PhoneHomeConfig         `yaml:"phone_home,omitempty"`
	PowerState         *PowerState              `yaml:"power_state,omitempty"`
	GrowPart           *GrowPartConfig          `yaml:"growpart,omitempty"`
	ResizeRootFS       *ResizeRootFSValue       `yaml:"resize_rootfs,omitempty"`
	ByobuByDefault     ByobuMode                `yaml:"byobu_by_default,omitempty"`
	DisableEC2Metadata *bool                    `yaml:"disable_ec2_metadata,omitempty"`
	Fan                *FanConfig               `yaml:"fan,omitempty"`
	GrubDpkg           *GrubDpkgConfig          `yaml:"grub_dpkg,omitempty"`
	Updates            *HotplugUpdates          `yaml:"updates,omitempty"`
	Keyboard           *KeyboardConfig          `yaml:"keyboard,omitempty"`
	NTP                *NTPConfig               `yaml:"ntp,omitempty"`
	ManageResolvConf   *bool                    `yaml:"manage_resolv_conf,omitempty"`
	ResolvConf         *ResolvConfConfig        `yaml:"resolv_conf,omitempty"`
	RandomSeed         *RandomSeedConfig        `yaml:"random_seed,omitempty"`
	Drivers            *UbuntuDriversConfig     `yaml:"drivers,omitempty"`
	Autoinstall        *UbuntuAutoinstallConfig `yaml:"autoinstall,omitempty"`

	Ansible        *AnsibleConfig        `yaml:"ansible,omitempty"`
	Chef           *ChefConfig           `yaml:"chef,omitempty"`
	Landscape      *LandscapeConfig      `yaml:"landscape,omitempty"`
	LXD            *LXDConfig            `yaml:"lxd,omitempty"`
	MCollective    *MCollectiveConfig    `yaml:"mcollective,omitempty"`
	Puppet         *PuppetConfig         `yaml:"puppet,omitempty"`
	RHSubscription *RHSubscriptionConfig `yaml:"rh_subscription,omitempty"`
	Rsyslog        *RsyslogConfig        `yaml:"rsyslog,omitempty"`
	SaltMinion     *SaltMinionConfig     `yaml:"salt_minion,omitempty"`
	Spacewalk      *SpacewalkConfig      `yaml:"spacewalk,omitempty"`
	UbuntuPro      *UbuntuProConfig      `yaml:"ubuntu_pro,omitempty"`
	WireGuard      *WireGuardConfig      `yaml:"wireguard,omitempty"`

	// Extra holds vendor-specific top-level cloud-config keys. Keys must not
	// duplicate modules supported directly by Config.
	Extra RawObject `yaml:",inline,omitempty"`
}
