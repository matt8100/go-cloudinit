package cloudinit

// AnsibleInstallMethod identifies how Ansible should be installed.
type AnsibleInstallMethod string

// Ansible install methods supported by this package.
const (
	AnsibleInstallMethodDistro AnsibleInstallMethod = "distro"
	AnsibleInstallMethodPip    AnsibleInstallMethod = "pip"
)

// AnsibleConfig configures the ansible module.
type AnsibleConfig struct {
	InstallMethod   AnsibleInstallMethod    `yaml:"install_method,omitempty"`
	RunUser         string                  `yaml:"run_user,omitempty"`
	ConfigPath      string                  `yaml:"ansible_config,omitempty"`
	SetupController *AnsibleSetupController `yaml:"setup_controller,omitempty"`
	Galaxy          *AnsibleGalaxy          `yaml:"galaxy,omitempty"`
	PackageName     string                  `yaml:"package_name,omitempty"`
	Pull            *AnsiblePull            `yaml:"pull,omitempty"`
}

// AnsibleSetupController configures ansible.setup_controller.
type AnsibleSetupController struct {
	Repositories []AnsibleRepository `yaml:"repositories,omitempty"`
	RunAnsible   []AnsibleRun        `yaml:"run_ansible,omitempty"`
}

// AnsibleRepository describes one setup_controller repository.
type AnsibleRepository struct {
	Path   string `yaml:"path,omitempty"`
	Source string `yaml:"source,omitempty"`
}

// AnsibleRun describes one setup_controller.run_ansible action.
type AnsibleRun struct {
	PlaybookName           string    `yaml:"playbook_name,omitempty"`
	PlaybookDir            string    `yaml:"playbook_dir,omitempty"`
	BecomePasswordFile     string    `yaml:"become_password_file,omitempty"`
	ConnectionPasswordFile string    `yaml:"connection_password_file,omitempty"`
	ListHosts              *bool     `yaml:"list_hosts,omitempty"`
	SyntaxCheck            *bool     `yaml:"syntax_check,omitempty"`
	Timeout                *int      `yaml:"timeout,omitempty"`
	VaultID                string    `yaml:"vault_id,omitempty"`
	VaultPasswordFile      string    `yaml:"vault_password_file,omitempty"`
	Background             *int      `yaml:"background,omitempty"`
	Check                  *bool     `yaml:"check,omitempty"`
	Diff                   *bool     `yaml:"diff,omitempty"`
	ModulePath             string    `yaml:"module_path,omitempty"`
	Poll                   *int      `yaml:"poll,omitempty"`
	Args                   string    `yaml:"args,omitempty"`
	ExtraVars              string    `yaml:"extra_vars,omitempty"`
	Forks                  *int      `yaml:"forks,omitempty"`
	Inventory              string    `yaml:"inventory,omitempty"`
	SCPExtraArgs           string    `yaml:"scp_extra_args,omitempty"`
	SFTPExtraArgs          string    `yaml:"sftp_extra_args,omitempty"`
	PrivateKey             string    `yaml:"private_key,omitempty"`
	Connection             string    `yaml:"connection,omitempty"`
	ModuleName             string    `yaml:"module_name,omitempty"`
	Sleep                  string    `yaml:"sleep,omitempty"`
	Tags                   string    `yaml:"tags,omitempty"`
	SkipTags               string    `yaml:"skip_tags,omitempty"`
	Extra                  RawObject `yaml:",inline,omitempty"`
}

// AnsibleGalaxy configures ansible.galaxy.
type AnsibleGalaxy struct {
	Actions [][]string `yaml:"actions,omitempty"`
}

// AnsiblePull configures ansible.pull.
type AnsiblePull struct {
	AcceptHostKey     *bool  `yaml:"accept_host_key,omitempty"`
	Clean             *bool  `yaml:"clean,omitempty"`
	Full              *bool  `yaml:"full,omitempty"`
	Diff              *bool  `yaml:"diff,omitempty"`
	SSHCommonArgs     string `yaml:"ssh_common_args,omitempty"`
	SCPExtraArgs      string `yaml:"scp_extra_args,omitempty"`
	SFTPExtraArgs     string `yaml:"sftp_extra_args,omitempty"`
	PrivateKey        string `yaml:"private_key,omitempty"`
	Checkout          string `yaml:"checkout,omitempty"`
	ModulePath        string `yaml:"module_path,omitempty"`
	Timeout           string `yaml:"timeout,omitempty"`
	URL               string `yaml:"url,omitempty"`
	Connection        string `yaml:"connection,omitempty"`
	VaultID           string `yaml:"vault_id,omitempty"`
	VaultPasswordFile string `yaml:"vault_password_file,omitempty"`
	VerifyCommit      *bool  `yaml:"verify_commit,omitempty"`
	Inventory         string `yaml:"inventory,omitempty"`
	ModuleName        string `yaml:"module_name,omitempty"`
	Sleep             string `yaml:"sleep,omitempty"`
	Tags              string `yaml:"tags,omitempty"`
	SkipTags          string `yaml:"skip_tags,omitempty"`
	PlaybookName      string `yaml:"playbook_name,omitempty"`
}

// ChefInstallType identifies how Chef should be installed.
type ChefInstallType string

// Chef install types supported by this package.
const (
	ChefInstallTypePackages ChefInstallType = "packages"
	ChefInstallTypeGems     ChefInstallType = "gems"
	ChefInstallTypeOmnibus  ChefInstallType = "omnibus"
)

// ChefLicense identifies the Chef license acceptance mode.
type ChefLicense string

// Chef license modes supported by this package.
const (
	ChefLicenseAccept          ChefLicense = "accept"
	ChefLicenseAcceptSilent    ChefLicense = "accept-silent"
	ChefLicenseAcceptNoPersist ChefLicense = "accept-no-persist"
)

// ChefConfig configures the chef module.
type ChefConfig struct {
	Directories            []string        `yaml:"directories,omitempty"`
	ConfigPath             string          `yaml:"config_path,omitempty"`
	ValidationCert         string          `yaml:"validation_cert,omitempty"`
	ValidationKey          string          `yaml:"validation_key,omitempty"`
	FirstbootPath          string          `yaml:"firstboot_path,omitempty"`
	Exec                   *bool           `yaml:"exec,omitempty"`
	ClientKey              string          `yaml:"client_key,omitempty"`
	EncryptedDataBagSecret string          `yaml:"encrypted_data_bag_secret,omitempty"`
	Environment            string          `yaml:"environment,omitempty"`
	FileBackupPath         string          `yaml:"file_backup_path,omitempty"`
	FileCachePath          string          `yaml:"file_cache_path,omitempty"`
	JSONAttribs            string          `yaml:"json_attribs,omitempty"`
	LogLevel               string          `yaml:"log_level,omitempty"`
	LogLocation            string          `yaml:"log_location,omitempty"`
	NodeName               string          `yaml:"node_name,omitempty"`
	OmnibusURL             string          `yaml:"omnibus_url,omitempty"`
	OmnibusURLRetries      *int            `yaml:"omnibus_url_retries,omitempty"`
	OmnibusVersion         string          `yaml:"omnibus_version,omitempty"`
	PIDFile                string          `yaml:"pid_file,omitempty"`
	ServerURL              string          `yaml:"server_url,omitempty"`
	ShowTime               *bool           `yaml:"show_time,omitempty"`
	SSLVerifyMode          string          `yaml:"ssl_verify_mode,omitempty"`
	ValidationName         string          `yaml:"validation_name,omitempty"`
	ForceInstall           *bool           `yaml:"force_install,omitempty"`
	InitialAttributes      RawObject       `yaml:"initial_attributes,omitempty"`
	InstallType            ChefInstallType `yaml:"install_type,omitempty"`
	RunList                []string        `yaml:"run_list,omitempty"`
	ChefLicense            ChefLicense     `yaml:"chef_license,omitempty"`
}

// LandscapeConfig configures the landscape module.
type LandscapeConfig struct {
	Client *LandscapeClient `yaml:"client,omitempty"`
}

// LandscapeClient configures landscape.client.
type LandscapeClient struct {
	URL             string    `yaml:"url,omitempty"`
	PingURL         string    `yaml:"ping_url,omitempty"`
	DataPath        string    `yaml:"data_path,omitempty"`
	LogLevel        string    `yaml:"log_level,omitempty"`
	ComputerTitle   string    `yaml:"computer_title,omitempty"`
	AccountName     string    `yaml:"account_name,omitempty"`
	RegistrationKey string    `yaml:"registration_key,omitempty"`
	Tags            string    `yaml:"tags,omitempty"`
	HTTPProxy       string    `yaml:"http_proxy,omitempty"`
	HTTPSProxy      string    `yaml:"https_proxy,omitempty"`
	Extra           RawObject `yaml:",inline,omitempty"`
}

// LXDStorageBackend identifies the lxd storage backend.
type LXDStorageBackend string

// LXD storage backends supported by this package.
const (
	LXDStorageBackendZFS   LXDStorageBackend = "zfs"
	LXDStorageBackendDir   LXDStorageBackend = "dir"
	LXDStorageBackendLVM   LXDStorageBackend = "lvm"
	LXDStorageBackendBTRFS LXDStorageBackend = "btrfs"
)

// LXDBridgeMode identifies the bridge mode.
type LXDBridgeMode string

// LXD bridge modes supported by this package.
const (
	LXDBridgeModeNone     LXDBridgeMode = "none"
	LXDBridgeModeExisting LXDBridgeMode = "existing"
	LXDBridgeModeNew      LXDBridgeMode = "new"
)

// LXDConfig configures the lxd module.
type LXDConfig struct {
	Init    *LXDInit   `yaml:"init,omitempty"`
	Bridge  *LXDBridge `yaml:"bridge,omitempty"`
	Preseed string     `yaml:"preseed,omitempty"`
}

// LXDInit configures lxd.init.
type LXDInit struct {
	NetworkAddress      string            `yaml:"network_address,omitempty"`
	NetworkPort         *int              `yaml:"network_port,omitempty"`
	StorageBackend      LXDStorageBackend `yaml:"storage_backend,omitempty"`
	StorageCreateDevice string            `yaml:"storage_create_device,omitempty"`
	StorageCreateLoop   *int              `yaml:"storage_create_loop,omitempty"`
	StoragePool         string            `yaml:"storage_pool,omitempty"`
	TrustPassword       string            `yaml:"trust_password,omitempty"`
}

// LXDBridge configures lxd.bridge.
type LXDBridge struct {
	Mode           LXDBridgeMode `yaml:"mode,omitempty"`
	Name           string        `yaml:"name,omitempty"`
	MTU            *int          `yaml:"mtu,omitempty"`
	IPv4Address    string        `yaml:"ipv4_address,omitempty"`
	IPv4Netmask    *int          `yaml:"ipv4_netmask,omitempty"`
	IPv4DHCPFirst  string        `yaml:"ipv4_dhcp_first,omitempty"`
	IPv4DHCPLast   string        `yaml:"ipv4_dhcp_last,omitempty"`
	IPv4DHCPLeases *int          `yaml:"ipv4_dhcp_leases,omitempty"`
	IPv4NAT        *bool         `yaml:"ipv4_nat,omitempty"`
	IPv6Address    string        `yaml:"ipv6_address,omitempty"`
	IPv6Netmask    *int          `yaml:"ipv6_netmask,omitempty"`
	IPv6NAT        *bool         `yaml:"ipv6_nat,omitempty"`
	Domain         string        `yaml:"domain,omitempty"`
}

// MCollectiveConfig configures the mcollective module.
type MCollectiveConfig struct {
	Conf *MCollectiveConf `yaml:"conf,omitempty"`
}

// MCollectiveConf configures mcollective.conf.
type MCollectiveConf struct {
	PublicCert  string    `yaml:"public-cert,omitempty"`
	PrivateCert string    `yaml:"private-cert,omitempty"`
	Extra       RawObject `yaml:",inline,omitempty"`
}

// PuppetInstallType identifies how Puppet should be installed.
type PuppetInstallType string

// Puppet install types supported by this package.
const (
	PuppetInstallTypePackages PuppetInstallType = "packages"
	PuppetInstallTypeAIO      PuppetInstallType = "aio"
)

// PuppetConfig configures the puppet module.
type PuppetConfig struct {
	Install           *bool                `yaml:"install,omitempty"`
	Version           string               `yaml:"version,omitempty"`
	InstallType       PuppetInstallType    `yaml:"install_type,omitempty"`
	Collection        string               `yaml:"collection,omitempty"`
	AIOInstallURL     string               `yaml:"aio_install_url,omitempty"`
	Cleanup           *bool                `yaml:"cleanup,omitempty"`
	ConfFile          string               `yaml:"conf_file,omitempty"`
	SSLDir            string               `yaml:"ssl_dir,omitempty"`
	CSRAttributesPath string               `yaml:"csr_attributes_path,omitempty"`
	PackageName       string               `yaml:"package_name,omitempty"`
	Exec              *bool                `yaml:"exec,omitempty"`
	ExecArgs          []string             `yaml:"exec_args,omitempty"`
	StartService      *bool                `yaml:"start_service,omitempty"`
	Conf              *PuppetConf          `yaml:"conf,omitempty"`
	CSRAttributes     *PuppetCSRAttributes `yaml:"csr_attributes,omitempty"`
}

// PuppetConf configures puppet.conf sections.
type PuppetConf struct {
	Main   RawObject `yaml:"main,omitempty"`
	Server RawObject `yaml:"server,omitempty"`
	Agent  RawObject `yaml:"agent,omitempty"`
	User   RawObject `yaml:"user,omitempty"`
	CACert string    `yaml:"ca_cert,omitempty"`
}

// PuppetCSRAttributes configures puppet CSR attributes.
type PuppetCSRAttributes struct {
	CustomAttributes  RawObject `yaml:"custom_attributes,omitempty"`
	ExtensionRequests RawObject `yaml:"extension_requests,omitempty"`
}

// RHSubscriptionConfig configures the rh_subscription module.
type RHSubscriptionConfig struct {
	Username       string   `yaml:"username,omitempty"`
	Password       string   `yaml:"password,omitempty"`
	ActivationKey  string   `yaml:"activation-key,omitempty"`
	Org            string   `yaml:"org,omitempty"`
	AutoAttach     *bool    `yaml:"auto-attach,omitempty"`
	ServiceLevel   string   `yaml:"service-level,omitempty"`
	AddPool        []string `yaml:"add-pool,omitempty"`
	EnableRepo     []string `yaml:"enable-repo,omitempty"`
	DisableRepo    []string `yaml:"disable-repo,omitempty"`
	RHSMBaseURL    string   `yaml:"rhsm-baseurl,omitempty"`
	ServerHostname string   `yaml:"server-hostname,omitempty"`
}

// RsyslogConfig configures the rsyslog module.
type RsyslogConfig struct {
	ConfigDir            string               `yaml:"config_dir,omitempty"`
	ConfigFilename       string               `yaml:"config_filename,omitempty"`
	Configs              []RsyslogConfigEntry `yaml:"configs,omitempty"`
	Remotes              map[string]string    `yaml:"remotes,omitempty"`
	ServiceReloadCommand *StringOrStringSlice `yaml:"service_reload_command,omitempty"`
	InstallRsyslog       *bool                `yaml:"install_rsyslog,omitempty"`
	CheckExe             string               `yaml:"check_exe,omitempty"`
	Packages             []string             `yaml:"packages,omitempty"`
}

// RsyslogConfigEntry represents one rsyslog config snippet.
type RsyslogConfigEntry struct {
	Content  string
	Filename string
}

// RsyslogContent returns an inline rsyslog config entry.
func RsyslogContent(content string) RsyslogConfigEntry {
	return RsyslogConfigEntry{Content: content}
}

// RsyslogFile returns a file-backed rsyslog config entry.
func RsyslogFile(filename, content string) RsyslogConfigEntry {
	return RsyslogConfigEntry{Filename: filename, Content: content}
}

// MarshalYAML implements yaml.Marshaler.
func (e RsyslogConfigEntry) MarshalYAML() (any, error) {
	if e.Content == "" {
		return nil, nil
	}
	if e.Filename == "" {
		return e.Content, nil
	}
	return map[string]any{
		"filename": e.Filename,
		"content":  e.Content,
	}, nil
}

// SaltMinionConfig configures the salt_minion module.
type SaltMinionConfig struct {
	PkgName     string    `yaml:"pkg_name,omitempty"`
	ServiceName string    `yaml:"service_name,omitempty"`
	ConfigDir   string    `yaml:"config_dir,omitempty"`
	Conf        RawObject `yaml:"conf,omitempty"`
	Grains      RawObject `yaml:"grains,omitempty"`
	PublicKey   string    `yaml:"public_key,omitempty"`
	PrivateKey  string    `yaml:"private_key,omitempty"`
	PKIDir      string    `yaml:"pki_dir,omitempty"`
}

// UbuntuProConfig configures the ubuntu_pro module.
type UbuntuProConfig struct {
	Enable     []string           `yaml:"enable,omitempty"`
	EnableBeta []string           `yaml:"enable_beta,omitempty"`
	Token      string             `yaml:"token,omitempty"`
	Features   *UbuntuProFeatures `yaml:"features,omitempty"`
	Config     *UbuntuProSettings `yaml:"config,omitempty"`
}

// UbuntuProFeatures configures ubuntu_pro.features.
type UbuntuProFeatures struct {
	DisableAutoAttach *bool `yaml:"disable_auto_attach,omitempty"`
}

// UbuntuProSettings configures ubuntu_pro.config.
type UbuntuProSettings struct {
	HTTPProxy           *NullableString `yaml:"http_proxy,omitempty"`
	HTTPSProxy          *NullableString `yaml:"https_proxy,omitempty"`
	GlobalAPTHTTPProxy  *NullableString `yaml:"global_apt_http_proxy,omitempty"`
	GlobalAPTHTTPSProxy *NullableString `yaml:"global_apt_https_proxy,omitempty"`
	UAAPTHTTPProxy      *NullableString `yaml:"ua_apt_http_proxy,omitempty"`
	UAAPTHTTPSProxy     *NullableString `yaml:"ua_apt_https_proxy,omitempty"`
	Extra               RawObject       `yaml:",inline,omitempty"`
}
