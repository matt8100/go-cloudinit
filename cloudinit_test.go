package cloudinit

import (
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestConfigRender(t *testing.T) {
	cfg := Config{
		Hostname:               "web-01",
		FQDN:                   "web-01.example.internal",
		PreserveHostname:       Bool(false),
		PreferFQDNOverHostname: Bool(true),
		CreateHostnameFile:     Bool(true),
		ManageEtcHosts:         ManageEtcHostsLocalhostValue(),
		Locale:                 Locale("en_US.UTF-8"),
		Timezone:               "UTC",
		PackageUpdate:          Bool(true),
		PackageUpgrade:         Bool(true),
		Packages: []PackageEntry{
			Package("curl"),
			VersionedPackage("jq", "1.6-2"),
			APTPackages(PackageSpec{Name: "git"}),
			SnapPackages(PackageSpec{Name: "microk8s"}),
		},
		BootCmd: []Command{
			ExecCommand("cloud-init-per", "always", "hello", "echo", "booting"),
		},
		RunCmd: []Command{
			ShellCommand("echo ready"),
			ExecCommand("systemctl", "restart", "sshd"),
			NullCommand(),
		},
		WriteFiles: []WriteFile{
			{
				Path:        "/etc/motd",
				Content:     "managed by go-cloudinit\n",
				Permissions: "0644",
			},
		},
		Groups: []GroupEntry{
			Group("developers", "alice"),
		},
		User: DefaultUserReference("default"),
		Users: []UserEntry{
			UserReference("default"),
			{
				Name:              "alice",
				Groups:            []string{"developers", "sudo"},
				SSHAuthorizedKeys: []string{"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIexample alice@example"},
				Shell:             "/bin/bash",
				LockPasswd:        Bool(false),
			},
		},
		SSHPWAuth:          BoolValue(false),
		SSHAuthorizedKeys:  []string{"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIroot root@example"},
		SSHImportID:        []string{"gh:alice"},
		SSHDeleteKeys:      Bool(true),
		SSHGenKeyTypes:     []SSHKeyType{SSHKeyRSA, SSHKeyED25519},
		DisableRoot:        Bool(true),
		AllowPublicSSHKeys: Bool(true),
		SSHQuietKeygen:     Bool(true),
		SSHPublishHostKeys: &SSHPublishHostKeys{
			Enabled:   Bool(true),
			Blacklist: []string{"rsa"},
		},
		CACerts: &CACertsConfig{
			Trusted: []string{"-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----"},
		},
		FinalMessage: "cloud-init finished",
		Mounts: []MountEntry{
			{Spec: "ephemeral0", File: "/mnt", VFSType: "auto", Options: "defaults,nofail", Freq: "0", PassNo: "2"},
		},
		MountDefaultFields: &MountDefaults{
			nil,
			nil,
			StringPtr("auto"),
			StringPtr("defaults,nofail,x-systemd.after=cloud-init-network.service"),
			StringPtr("0"),
			StringPtr("2"),
		},
		Swap: &SwapDefinition{
			Filename: "/swapfile",
			Size:     HumanSize("1G"),
			MaxSize:  BytesSize(2147483648),
		},
		PhoneHome: &PhoneHomeConfig{
			URL:   "https://example.internal/phone-home",
			Post:  PhoneHomeAllFields(),
			Tries: Int(3),
		},
		PowerState: &PowerState{
			Mode:      Reboot,
			Delay:     DelayNow(),
			Timeout:   Int(30),
			Condition: ConditionCommand(ExecCommand("test", "-f", "/var/run/reboot-required")),
		},
		GrowPart: &GrowPartConfig{
			Mode:    GrowPartAuto,
			Devices: []string{"/"},
		},
		ResizeRootFS: ResizeRootFSBackground(),
	}

	rendered, err := cfg.Render()
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	if !strings.HasPrefix(string(rendered), Header) {
		t.Fatalf("rendered config missing %q prefix: %s", Header, rendered)
	}

	body := strings.TrimPrefix(string(rendered), Header)
	var parsed map[string]any
	if err := yaml.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("rendered YAML failed to parse: %v", err)
	}

	for _, needle := range []string{
		"hostname: web-01",
		"manage_etc_hosts: localhost",
		"package_update: true",
		"- - jq",
		"  - 1.6-2",
		"ssh_pwauth: false",
		"resize_rootfs: noblock",
		"url: https://example.internal/phone-home",
		"mode: reboot",
	} {
		if !strings.Contains(string(rendered), needle) {
			t.Fatalf("rendered config missing %q\n%s", needle, rendered)
		}
	}
}

func TestConfigRenderExtensionsAndExplicitEmptyDisableRootOpts(t *testing.T) {
	cfg := Config{
		DisableRootOpts: StringPtr(""),
		Extra: RawObject{
			"bluecat_license": RawObject{
				"id":  "0014000000L0yAK",
				"key": "license-key",
			},
			"bluecat_netconf": RawObject{
				"ipaddr":  "192.0.2.20",
				"cidr":    "24",
				"gateway": "192.0.2.2",
			},
			"bluecat_service_config": RawObject{
				"payload": `{"version":"1.3.0"}`,
			},
		},
	}

	rendered, err := cfg.Render()
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	for _, needle := range []string{
		"disable_root_opts: \"\"",
		"bluecat_license:",
		"bluecat_netconf:",
		"bluecat_service_config:",
	} {
		if !strings.Contains(string(rendered), needle) {
			t.Fatalf("rendered config missing %q\n%s", needle, rendered)
		}
	}
}

func TestConfigExtraRejectsSupportedModuleNames(t *testing.T) {
	err := (Config{Extra: RawObject{"hostname": "vendor-host"}}).Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for extension colliding with a supported module")
	}
	if !strings.Contains(err.Error(), "extra.hostname") {
		t.Fatalf("expected collision error path, got %v", err)
	}
}

func TestValidateReportsFieldPaths(t *testing.T) {
	cfg := Config{
		ManageEtcHosts: &ManageEtcHostsValue{Mode: "invalid"},
		WriteFiles: []WriteFile{
			{
				Encoding: "rot13",
			},
		},
		Users: []UserEntry{
			{
				Name:              "alice",
				SSHRedirectUser:   Bool(true),
				SSHAuthorizedKeys: []string{"ssh-ed25519 AAAA alice@example"},
			},
		},
		PhoneHome: &PhoneHomeConfig{
			URL: "not-a-url",
			Post: &PhoneHomePost{
				Fields: []PhoneHomeField{"bad_field"},
			},
		},
		PowerState: &PowerState{},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for invalid config")
	}

	message := err.Error()
	for _, needle := range []string{
		"manage_etc_hosts",
		"write_files[0].path",
		"write_files[0].encoding",
		"users[0].ssh_authorized_keys",
		"phone_home.url",
		"phone_home.post[0]",
		"power_state.mode",
	} {
		if !strings.Contains(message, needle) {
			t.Fatalf("expected validation error to mention %q\n%s", needle, message)
		}
	}
}

func TestMountEntryMarshalYAMLPreservesEmptyOptionalColumns(t *testing.T) {
	value, err := MountEntry{
		Spec:    "ephemeral0",
		File:    "/mnt",
		Options: "defaults,nofail",
	}.MarshalYAML()
	if err != nil {
		t.Fatalf("MarshalYAML() error = %v", err)
	}

	values, ok := value.([]string)
	if !ok {
		t.Fatalf("MarshalYAML() returned %T, want []string", value)
	}

	want := []string{"ephemeral0", "/mnt", "", "defaults,nofail"}
	if len(values) != len(want) {
		t.Fatalf("MarshalYAML() returned %v, want %v", values, want)
	}
	for i := range want {
		if values[i] != want[i] {
			t.Fatalf("MarshalYAML() returned %v, want %v", values, want)
		}
	}
}

func TestUnionWrapperMarshalYAMLRejectsZeroValues(t *testing.T) {
	if _, err := (StringOrStringSlice{}).MarshalYAML(); err == nil {
		t.Fatal("StringOrStringSlice.MarshalYAML() returned nil error for zero value")
	}
	if _, err := (StringOrInt{}).MarshalYAML(); err == nil {
		t.Fatal("StringOrInt.MarshalYAML() returned nil error for zero value")
	}
	if _, err := (BoolOrStringDeprecated{}).MarshalYAML(); err == nil {
		t.Fatal("BoolOrStringDeprecated.MarshalYAML() returned nil error for zero value")
	}
}

func TestValidateAcceptsFileURIs(t *testing.T) {
	cfg := Config{
		WriteFiles: []WriteFile{
			{
				Path: "/etc/example.txt",
				Source: &WriteFileSource{
					URI: "file:///var/lib/cloud/source.txt",
				},
			},
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsEmptyAPTConfig(t *testing.T) {
	cfg := Config{
		APT: &APTConfig{},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for empty apt config")
	}
	if !strings.Contains(err.Error(), "apt") {
		t.Fatalf("expected validation error to mention apt\n%s", err)
	}
}

func TestValidateRejectsSnapCollectionsWithBothForms(t *testing.T) {
	cfg := Config{
		Snap: &SnapConfig{
			Assertions: &StringCollection{
				List: []string{"assertion"},
				Map:  map[string]string{"a": "assertion"},
			},
			Commands: &CommandCollection{
				List: []Command{ShellCommand("install hello-world")},
				Map:  CommandMap{"cmd": ShellCommand("remove hello-world")},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for snap collections with both forms")
	}
	for _, needle := range []string{"snap.assertions", "snap.commands"} {
		if !strings.Contains(err.Error(), needle) {
			t.Fatalf("expected validation error to mention %q\n%s", needle, err)
		}
	}
}

func TestValidateRejectsDuplicateSnapAssertions(t *testing.T) {
	cfg := Config{
		Snap: &SnapConfig{
			Assertions: &StringCollection{
				List: []string{"assertion", "assertion"},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for duplicate snap assertions")
	}
	if !strings.Contains(err.Error(), "snap.assertions[1]") {
		t.Fatalf("expected validation error to mention snap.assertions[1]\n%s", err)
	}
}

func TestValidateRejectsInvalidAPTMirrorConfig(t *testing.T) {
	cfg := Config{
		APT: &APTConfig{
			Primary: []APTMirrorConfig{
				{"uri": 1},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for invalid apt mirror config")
	}
	for _, needle := range []string{
		"apt.primary[0].arches",
		"apt.primary[0].uri",
	} {
		if !strings.Contains(err.Error(), needle) {
			t.Fatalf("expected validation error to mention %q\n%s", needle, err)
		}
	}
}

func TestValidateRejectsDuplicateAPTDisableSuites(t *testing.T) {
	cfg := Config{
		APT: &APTConfig{
			DisableSuites: []string{"updates", "updates"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for duplicate apt.disable_suites entries")
	}
	if !strings.Contains(err.Error(), "apt.disable_suites[1]") {
		t.Fatalf("expected validation error to mention apt.disable_suites[1]\n%s", err)
	}
}

func TestValidateRejectsInvalidAPTMapKeys(t *testing.T) {
	cfg := Config{
		APT: &APTConfig{
			DebconfSelections: map[string]string{
				"": "debconf data",
			},
			Sources: map[string]APTSource{
				"": {Source: "deb https://apt.example.com stable main"},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for invalid apt map keys")
	}
	for _, needle := range []string{"apt.debconf_selections", "apt.sources"} {
		if !strings.Contains(err.Error(), needle) {
			t.Fatalf("expected validation error to mention %q\n%s", needle, err)
		}
	}
}

func TestValidateAcceptsPathBasedAbsoluteURI(t *testing.T) {
	cfg := Config{
		APT: &APTConfig{
			Primary: []APTMirrorConfig{
				{
					"arches": []string{"default"},
					"uri":    "mirror+file:///etc/apt/mirrors/ubuntu.list",
				},
			},
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsInlineMapKeyConflicts(t *testing.T) {
	cfg := Config{
		YumRepos: YumRepos{
			"epel": {
				BaseURL: "https://yum.example.com/epel",
				Options: map[string]any{"enabled": "yes"},
			},
		},
		Zypper: &ZypperConfig{
			Repos: []ZypperRepo{
				{
					ID:      "base",
					BaseURL: "https://zypper.example.com/base",
					Options: map[string]any{"id": "shadow"},
				},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for inline map key conflicts")
	}
	for _, needle := range []string{
		"yum_repos.epel.enabled",
		"zypper.repos[0].id",
	} {
		if !strings.Contains(err.Error(), needle) {
			t.Fatalf("expected validation error to mention %q\n%s", needle, err)
		}
	}
}

func TestValidateRejectsNonScalarYumRepoOption(t *testing.T) {
	cfg := Config{
		YumRepos: YumRepos{
			"epel": {
				BaseURL: "https://yum.example.com/epel",
				Options: map[string]any{"gpgcheck": []int{1}},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for non-scalar yum repo option")
	}
	if !strings.Contains(err.Error(), "yum_repos.epel.gpgcheck") {
		t.Fatalf("expected validation error to mention yum_repos.epel.gpgcheck\n%s", err)
	}
}

func TestValidateRejectsInvalidYumRepoNames(t *testing.T) {
	cfg := Config{
		YumRepos: YumRepos{
			"": {
				BaseURL: "https://yum.example.com/empty",
			},
			"epel": {
				BaseURL: "https://yum.example.com/epel",
				Options: map[string]any{"bad-opt": true},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for invalid yum repo names")
	}
	for _, needle := range []string{"yum_repos", "yum_repos.epel.bad-opt"} {
		if !strings.Contains(err.Error(), needle) {
			t.Fatalf("expected validation error to mention %q\n%s", needle, err)
		}
	}
}

func TestRenderReturnsErrorInsteadOfPanickingOnInlineMapConflicts(t *testing.T) {
	cfg := Config{
		YumRepos: YumRepos{
			"epel": {
				BaseURL: "https://yum.example.com/epel",
				Options: map[string]any{"enabled": "yes"},
			},
		},
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("Render() panicked: %v", recovered)
		}
	}()

	if _, err := cfg.Render(); err == nil {
		t.Fatal("Render() returned nil error for inline map key conflict")
	}
}

func TestValidateRejectsEmptyAPKRepos(t *testing.T) {
	cfg := Config{
		APKRepos: &APKReposConfig{},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for empty apk_repos")
	}
	if !strings.Contains(err.Error(), "apk_repos") {
		t.Fatalf("expected validation error to mention apk_repos\n%s", err)
	}
}

func TestValidateRejectsReferenceUserWithAdditionalFields(t *testing.T) {
	cfg := Config{
		Users: []UserEntry{
			{
				Reference: "default",
				Groups:    []string{"sudo"},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for invalid reference user")
	}
	if !strings.Contains(err.Error(), "users[0]") {
		t.Fatalf("expected validation error to mention users[0]\n%s", err)
	}
}

func TestValidateRejectsSSHSettingsWithoutHomeDirectory(t *testing.T) {
	cfg := Config{
		Users: []UserEntry{
			{
				Name:              "alice",
				NoCreateHome:      Bool(true),
				SSHAuthorizedKeys: []string{"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIexample alice@example"},
			},
			{
				Name:        "daemon",
				System:      Bool(true),
				SSHImportID: []string{"gh:alice"},
			},
			{
				Name:            "redirect",
				NoCreateHome:    Bool(true),
				SSHRedirectUser: Bool(true),
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for SSH settings without a home directory")
	}
	for _, needle := range []string{
		"users[0].ssh_authorized_keys",
		"users[1].ssh_import_id",
		"users[2].ssh_redirect_user",
	} {
		if !strings.Contains(err.Error(), needle) {
			t.Fatalf("expected validation error to mention %q\n%s", needle, err)
		}
	}
}

func TestValidateRejectsUserWithNameAndSnapUser(t *testing.T) {
	cfg := Config{
		Users: []UserEntry{
			{
				Name:     "alice",
				SnapUser: "alice-snap",
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for conflicting user identity fields")
	}
	if !strings.Contains(err.Error(), "users[0]") {
		t.Fatalf("expected validation error to mention users[0]\n%s", err)
	}
}

func TestUserEntryMarshalYAMLRejectsConflictingIdentityFields(t *testing.T) {
	_, err := UserEntry{Name: "alice", SnapUser: "alice-snap"}.MarshalYAML()
	if err == nil {
		t.Fatal("MarshalYAML() returned nil error for conflicting user identity fields")
	}
}

func TestUserEntryMarshalYAMLRejectsReferenceWithAdditionalFields(t *testing.T) {
	_, err := UserEntry{Reference: "default", Groups: []string{"sudo"}}.MarshalYAML()
	if err == nil {
		t.Fatal("MarshalYAML() returned nil error for reference user with additional fields")
	}
}

func TestValidateRequiresManageResolvConfTrue(t *testing.T) {
	cfg := Config{
		ManageResolvConf: Bool(false),
		ResolvConf: &ResolvConfConfig{
			Nameservers: []string{"1.1.1.1"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error when manage_resolv_conf is false")
	}
	if !strings.Contains(err.Error(), "manage_resolv_conf") {
		t.Fatalf("expected validation error to mention manage_resolv_conf\n%s", err)
	}
}

func TestValidateRejectsDuplicateSSHConsoleBlacklists(t *testing.T) {
	cfg := Config{
		SSHKeyConsoleBlacklist: []string{"rsa", "rsa"},
		SSHFPConsoleBlacklist:  []string{"sha256", "sha256"},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for duplicate SSH console blacklist entries")
	}
	for _, needle := range []string{
		"ssh_key_console_blacklist[1]",
		"ssh_fp_console_blacklist[1]",
	} {
		if !strings.Contains(err.Error(), needle) {
			t.Fatalf("expected validation error to mention %q\n%s", needle, err)
		}
	}
}

func TestValidateRejectsRandomSeedCommandRequiredWithoutCommand(t *testing.T) {
	cfg := Config{
		RandomSeed: &RandomSeedConfig{
			CommandRequired: Bool(true),
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error when random_seed.command_required is true without command")
	}
	if !strings.Contains(err.Error(), "random_seed.command") {
		t.Fatalf("expected validation error to mention random_seed.command\n%s", err)
	}
}

func TestValidateRejectsRandomSeedCommandWithEmptyExecutable(t *testing.T) {
	cfg := Config{
		RandomSeed: &RandomSeedConfig{
			Command: []string{""},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for empty random_seed.command entry")
	}
	if !strings.Contains(err.Error(), "random_seed.command[0]") {
		t.Fatalf("expected validation error to mention random_seed.command[0]\n%s", err)
	}
}

func TestValidateAllowsSwapWithoutFilename(t *testing.T) {
	cfg := Config{
		Swap: &SwapDefinition{
			Size: AutoSize(),
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsInvalidFSSetupPartitionString(t *testing.T) {
	cfg := Config{
		FSSetup: []FilesystemSetup{
			{
				Filesystem: "ext4",
				Device:     "/dev/sda1",
				Partition:  &StringOrInt{String: "foo"},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for invalid fs_setup partition")
	}
	if !strings.Contains(err.Error(), "fs_setup[0].partition") {
		t.Fatalf("expected validation error to mention fs_setup[0].partition\n%s", err)
	}
}

func TestValidateAllowsCmdOnlyFSSetup(t *testing.T) {
	cfg := Config{
		FSSetup: []FilesystemSetup{
			{
				Device: "/dev/sda1",
				Cmd: &StringOrStringSlice{
					String: "mkfs.custom %(device)s",
				},
			},
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsInvalidDiskLayoutInvariants(t *testing.T) {
	cfg := Config{
		DiskSetup: DiskSetupMap{
			"empty": {
				Layout: &DiskLayout{},
			},
			"mixed": {
				Layout: &DiskLayout{
					Enabled: Bool(true),
					Parts:   []PartitionLayout{{Size: 50}},
				},
			},
			"badparts": {
				Layout: &DiskLayout{
					Parts: []PartitionLayout{{Size: 0}},
				},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for invalid disk layouts")
	}
	for _, needle := range []string{
		"disk_setup.empty.layout",
		"disk_setup.mixed.layout",
		"disk_setup.badparts.layout.parts[0].size",
	} {
		if !strings.Contains(err.Error(), needle) {
			t.Fatalf("expected validation error to mention %q\n%s", needle, err)
		}
	}
}

func TestValidateRejectsEmptyDeviceAliasAndDiskSetupKeys(t *testing.T) {
	cfg := Config{
		DeviceAliases: map[string]string{
			"": "/dev/sda",
		},
		DiskSetup: DiskSetupMap{
			"": {
				Layout: &DiskLayout{
					Parts: []PartitionLayout{{Size: 50}},
				},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for empty device_aliases or disk_setup keys")
	}
	for _, needle := range []string{"device_aliases", "disk_setup"} {
		if !strings.Contains(err.Error(), needle) {
			t.Fatalf("expected validation error to mention %q\n%s", needle, err)
		}
	}
}

func TestValidateRejectsEmptyUnionWrappers(t *testing.T) {
	cfg := Config{
		FSSetup: []FilesystemSetup{
			{
				Filesystem: "ext4",
				Device:     "/dev/sda1",
				Partition:  &StringOrInt{},
				ExtraOpts:  &StringOrStringSlice{},
			},
		},
		GrubDpkg: &GrubDpkgConfig{
			GrubPCInstallDevicesEmpty: &BoolOrStringDeprecated{},
		},
		Rsyslog: &RsyslogConfig{
			ServiceReloadCommand: &StringOrStringSlice{},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for empty union wrappers")
	}
	for _, needle := range []string{
		"fs_setup[0].partition",
		"fs_setup[0].extra_opts",
		"grub_dpkg.grub-pc/install_devices_empty",
		"rsyslog.service_reload_command",
	} {
		if !strings.Contains(err.Error(), needle) {
			t.Fatalf("expected validation error to mention %q\n%s", needle, err)
		}
	}
}

func TestValidateRejectsAnsibleMissingRuntimeRequiredKeys(t *testing.T) {
	cfg := Config{
		Ansible: &AnsibleConfig{},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for ansible config missing runtime-required keys")
	}
	for _, needle := range []string{"ansible.install_method", "ansible.package_name"} {
		if !strings.Contains(err.Error(), needle) {
			t.Fatalf("expected validation error to mention %q\n%s", needle, err)
		}
	}
}

func TestValidateRejectsIncompleteAnsibleSetupControllerRun(t *testing.T) {
	cfg := Config{
		Ansible: &AnsibleConfig{
			InstallMethod: AnsibleInstallMethodPip,
			PackageName:   "ansible",
			SetupController: &AnsibleSetupController{
				RunAnsible: []AnsibleRun{{ModuleName: "ping"}},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for incomplete ansible setup_controller run")
	}
	for _, needle := range []string{
		"ansible.setup_controller.run_ansible[0].playbook_dir",
		"ansible.setup_controller.run_ansible[0].playbook_name",
	} {
		if !strings.Contains(err.Error(), needle) {
			t.Fatalf("expected validation error to mention %q\n%s", needle, err)
		}
	}
}

func TestValidateRejectsNegativeAnsibleRunNumericOptions(t *testing.T) {
	cfg := Config{
		Ansible: &AnsibleConfig{
			InstallMethod: AnsibleInstallMethodDistro,
			PackageName:   "ansible",
			SetupController: &AnsibleSetupController{
				RunAnsible: []AnsibleRun{
					{
						PlaybookDir:  "/srv/ansible",
						PlaybookName: "site.yml",
						Timeout:      Int(-1),
						Background:   Int(-1),
						Poll:         Int(-1),
						Forks:        Int(-1),
					},
				},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for negative ansible numeric options")
	}
	for _, needle := range []string{
		"ansible.setup_controller.run_ansible[0].timeout",
		"ansible.setup_controller.run_ansible[0].background",
		"ansible.setup_controller.run_ansible[0].poll",
		"ansible.setup_controller.run_ansible[0].forks",
	} {
		if !strings.Contains(err.Error(), needle) {
			t.Fatalf("expected validation error to mention %q\n%s", needle, err)
		}
	}
}

func TestValidateRejectsDuplicateWireGuardReadinessProbeEntries(t *testing.T) {
	cfg := Config{
		WireGuard: &WireGuardConfig{
			Interfaces: []WireGuardInterface{
				{
					Name:       "wg0",
					ConfigPath: "/etc/wireguard/wg0.conf",
					Content:    "[Interface]\nPrivateKey = abc\n",
				},
			},
			ReadinessProbe: []string{"ping -c1 192.0.2.1", "ping -c1 192.0.2.1"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for duplicate wireguard readinessprobe entries")
	}
	if !strings.Contains(err.Error(), "wireguard.readinessprobe[1]") {
		t.Fatalf("expected validation error to mention wireguard.readinessprobe[1]\n%s", err)
	}
}

func TestValidateRejectsWireGuardInterfaceMissingRequiredKeys(t *testing.T) {
	cfg := Config{
		WireGuard: &WireGuardConfig{
			Interfaces: []WireGuardInterface{
				{Name: "wg0", Content: "[Interface]\nPrivateKey = abc\n"},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for incomplete wireguard interface")
	}
	if !strings.Contains(err.Error(), "wireguard.interfaces[0].config_path") {
		t.Fatalf("expected validation error to mention wireguard.interfaces[0].config_path\n%s", err)
	}
}

func TestValidateRejectsInvalidRsyslogRemote(t *testing.T) {
	cfg := Config{
		Rsyslog: &RsyslogConfig{
			Remotes: map[string]string{"bad": "@@"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for invalid rsyslog remote")
	}
	if !strings.Contains(err.Error(), "rsyslog.remotes.bad") {
		t.Fatalf("expected validation error to mention rsyslog.remotes.bad\n%s", err)
	}
}

func TestValidateRejectsEmptyRsyslogRemote(t *testing.T) {
	cfg := Config{
		Rsyslog: &RsyslogConfig{
			Remotes: map[string]string{"main": ""},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for empty rsyslog remote")
	}
	if !strings.Contains(err.Error(), "rsyslog.remotes.main") {
		t.Fatalf("expected validation error to mention rsyslog.remotes.main\n%s", err)
	}
}

func TestValidateRejectsDuplicateRsyslogPackages(t *testing.T) {
	cfg := Config{
		Rsyslog: &RsyslogConfig{
			Packages: []string{"rsyslog", "rsyslog"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for duplicate rsyslog.packages entries")
	}
	if !strings.Contains(err.Error(), "rsyslog.packages[1]") {
		t.Fatalf("expected validation error to mention rsyslog.packages[1]\n%s", err)
	}
}

func TestValidateAcceptsSpacewalkWithoutServer(t *testing.T) {
	cfg := Config{
		Spacewalk: &SpacewalkConfig{
			ActivationKey: "activation-key",
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsInvalidLandscapeTags(t *testing.T) {
	cfg := Config{
		Landscape: &LandscapeConfig{
			Client: &LandscapeClient{
				AccountName:   "account",
				ComputerTitle: "instance-1",
				Tags:          "bad tag",
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for invalid landscape tags")
	}
	if !strings.Contains(err.Error(), "landscape.client.tags") {
		t.Fatalf("expected validation error to mention landscape.client.tags\n%s", err)
	}
}

func TestValidateRejectsDuplicateChefDirectories(t *testing.T) {
	cfg := Config{
		Chef: &ChefConfig{
			ServerURL:      "https://chef.example.com",
			ValidationName: "chef-validator",
			Directories:    []string{"/etc/chef", "/etc/chef"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for duplicate chef.directories entries")
	}
	if !strings.Contains(err.Error(), "chef.directories[1]") {
		t.Fatalf("expected validation error to mention chef.directories[1]\n%s", err)
	}
}

func TestValidateAcceptsRepoOnlyRHSubscription(t *testing.T) {
	cfg := Config{
		RHSubscription: &RHSubscriptionConfig{
			EnableRepo: []string{"rhel-9-baseos-rpms"},
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsLXDMTUBelowMinusOne(t *testing.T) {
	cfg := Config{
		LXD: &LXDConfig{
			Bridge: &LXDBridge{
				Mode: LXDBridgeModeNew,
				MTU:  Int(-2),
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for invalid lxd bridge mtu")
	}
	if !strings.Contains(err.Error(), "lxd.bridge.mtu") {
		t.Fatalf("expected validation error to mention lxd.bridge.mtu\n%s", err)
	}
}

func TestValidateRejectsNonScalarMCollectiveConfExtra(t *testing.T) {
	cfg := Config{
		MCollective: &MCollectiveConfig{
			Conf: &MCollectiveConf{
				PublicCert: "cert",
				Extra:      RawObject{"foo": []string{"bar"}},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for non-scalar mcollective.conf extra")
	}
	if !strings.Contains(err.Error(), "mcollective.conf.foo") {
		t.Fatalf("expected validation error to mention mcollective.conf.foo\n%s", err)
	}
}

func TestValidateRejectsNonHTTPUbuntuProProxy(t *testing.T) {
	cfg := Config{
		UbuntuPro: &UbuntuProConfig{
			Config: &UbuntuProSettings{
				HTTPProxy: StringSetting("ftp://proxy.example:21"),
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for invalid ubuntu pro proxy scheme")
	}
	if !strings.Contains(err.Error(), "ubuntu_pro.config.http_proxy") {
		t.Fatalf("expected validation error to mention ubuntu_pro.config.http_proxy\n%s", err)
	}
}

func TestValidateRejectsIncompleteChefConfig(t *testing.T) {
	cfg := Config{
		Chef: &ChefConfig{
			InstallType: ChefInstallTypeOmnibus,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for incomplete chef config")
	}
	for _, needle := range []string{"chef.server_url", "chef.validation_name"} {
		if !strings.Contains(err.Error(), needle) {
			t.Fatalf("expected validation error to mention %q\n%s", needle, err)
		}
	}
}

func TestSupportedModulesIncludesSwap(t *testing.T) {
	if _, ok := SupportedModuleSet()["swap"]; !ok {
		t.Fatal("SupportedModuleSet() missing swap")
	}
}

func TestValidateRejectsInvalidHotplugWhenValue(t *testing.T) {
	cfg := Config{
		Updates: &HotplugUpdates{
			Network: &HotplugNetworkConfig{When: []string{"bogus"}},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for invalid updates.network.when value")
	}
	if !strings.Contains(err.Error(), "updates.network.when[0]") {
		t.Fatalf("expected validation error to mention updates.network.when[0]\n%s", err)
	}
}

func TestValidateRejectsInvalidNTPConfig(t *testing.T) {
	cfg := Config{
		NTP: &NTPConfig{
			"enabled": "no",
			"servers": []any{1},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for invalid ntp config")
	}
	for _, needle := range []string{
		"ntp.enabled",
		"ntp.servers[0]",
	} {
		if !strings.Contains(err.Error(), needle) {
			t.Fatalf("expected validation error to mention %q\n%s", needle, err)
		}
	}
}

func TestValidateAcceptsNamedMapTypesForNTPConfig(t *testing.T) {
	cfg := Config{
		NTP: &NTPConfig{
			"config": map[string]string{
				"check_exe": "ntpd",
			},
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsUnsupportedNTPClient(t *testing.T) {
	cfg := Config{
		NTP: &NTPConfig{
			"ntp_client": "bogus",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for unsupported ntp_client")
	}
	if !strings.Contains(err.Error(), "ntp.ntp_client") {
		t.Fatalf("expected validation error to mention ntp.ntp_client\n%s", err)
	}
}

func TestValidateRejectsInvalidNTPHostname(t *testing.T) {
	cfg := Config{
		NTP: &NTPConfig{
			"servers": []string{"http://bad"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for invalid ntp hostname")
	}
	if !strings.Contains(err.Error(), "ntp.servers[0]") {
		t.Fatalf("expected validation error to mention ntp.servers[0]\n%s", err)
	}
}

func TestValidateRejectsDuplicateNTPStringArrays(t *testing.T) {
	cfg := Config{
		NTP: &NTPConfig{
			"servers": []string{"ntp.example.com", "ntp.example.com"},
			"pools":   []string{"pool.example.com", "pool.example.com"},
			"peers":   []string{"peer.example.com", "peer.example.com"},
			"allow":   []string{"192.0.2.0/24", "192.0.2.0/24"},
			"config": map[string]any{
				"packages": []string{"chrony", "chrony"},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for duplicate NTP string-array entries")
	}
	for _, needle := range []string{
		"ntp.servers[1]",
		"ntp.pools[1]",
		"ntp.peers[1]",
		"ntp.allow[1]",
		"ntp.config.packages[1]",
	} {
		if !strings.Contains(err.Error(), needle) {
			t.Fatalf("expected validation error to mention %q\n%s", needle, err)
		}
	}
}

func TestValidateRejectsConflictingUbuntuProNullableString(t *testing.T) {
	proxy := "https://proxy.example"
	cfg := Config{
		UbuntuPro: &UbuntuProConfig{
			Config: &UbuntuProSettings{
				HTTPProxy: &NullableString{String: &proxy, Null: true},
			},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for conflicting ubuntu pro nullable string")
	}
	if !strings.Contains(err.Error(), "ubuntu_pro.config.http_proxy") {
		t.Fatalf("expected validation error to mention ubuntu_pro.config.http_proxy\n%s", err)
	}
}

func TestValidateRejectsConflictingNullCommand(t *testing.T) {
	cfg := Config{
		RunCmd: []Command{
			{Null: true, String: "echo hi"},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for conflicting null command")
	}
	if !strings.Contains(err.Error(), "runcmd[0]") {
		t.Fatalf("expected validation error to mention runcmd[0]\n%s", err)
	}
}

func TestUnionRenderers(t *testing.T) {
	cfg := Config{
		Locale:       LocaleEnabled(false),
		SSHPWAuth:    StringValue("yes"),
		ResizeRootFS: ResizeRootFSEnabled(true),
		RunCmd: []Command{
			NullCommand(),
			ExecCommand("echo", "hi"),
		},
	}

	rendered, err := cfg.Render()
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, needle := range []string{
		"locale: false",
		"ssh_pwauth: \"yes\"",
		"resize_rootfs: true",
		"- null",
		"- - echo",
		"  - hi",
	} {
		if !strings.Contains(string(rendered), needle) {
			t.Fatalf("rendered config missing %q\n%s", needle, rendered)
		}
	}
}

func TestAdditionalModulesRenderAndValidate(t *testing.T) {
	cfg := Config{
		APT: &APTConfig{
			PreserveSourcesList: Bool(true),
			Sources: map[string]APTSource{
				"corp": {
					Source:   "deb https://apt.example.com stable main",
					Filename: "corp.list",
				},
			},
		},
		APKRepos: &APKReposConfig{
			AlpineRepo: &AlpineRepo{
				Version: "v3.20",
			},
		},
		APTPipelining: IntStringValue("os"),
		Snap: &SnapConfig{
			Commands: &CommandCollection{
				List: []Command{ShellCommand("install hello-world")},
			},
		},
		YumRepoDir: "/etc/yum.repos.d",
		YumRepos: YumRepos{
			"epel": {
				BaseURL: "https://yum.example.com/epel",
				Name:    "EPEL",
			},
		},
		Zypper: &ZypperConfig{
			Repos: []ZypperRepo{
				{ID: "base", BaseURL: "https://zypper.example.com/base"},
			},
		},
		ByobuByDefault:     "enable-system",
		DisableEC2Metadata: Bool(true),
		Fan: &FanConfig{
			Config: "underlay eth0\n",
		},
		GrubDpkg: &GrubDpkgConfig{
			Enabled: Bool(true),
		},
		Updates: &HotplugUpdates{
			Network: &HotplugNetworkConfig{When: []string{"boot", "hotplug"}},
		},
		Keyboard: &KeyboardConfig{
			Layout: "us",
		},
		ManageResolvConf: Bool(true),
		ResolvConf: &ResolvConfConfig{
			Nameservers:   []string{"1.1.1.1"},
			SearchDomains: []string{"example.com"},
		},
		RandomSeed: &RandomSeedConfig{
			Data:     "seed-data",
			Encoding: "raw",
		},
		Drivers: &UbuntuDriversConfig{
			Nvidia: &NVIDIADrivers{
				LicenseAccepted: Bool(true),
				Version:         "550",
			},
		},
		Autoinstall: &UbuntuAutoinstallConfig{
			Version: 1,
		},
		NTP: &NTPConfig{
			"enabled": true,
			"servers": []string{"0.pool.ntp.org"},
		},
		Ansible: &AnsibleConfig{
			InstallMethod: AnsibleInstallMethodPip,
			PackageName:   "ansible",
			Pull: &AnsiblePull{
				URL:          "https://git.example.com/ops.git",
				PlaybookName: "site.yml",
			},
		},
		Chef: &ChefConfig{
			ServerURL:      "https://chef.example.com",
			ValidationName: "chef-validator",
		},
		Landscape: &LandscapeConfig{
			Client: &LandscapeClient{
				AccountName:   "acme",
				ComputerTitle: "node-01",
			},
		},
		LXD: &LXDConfig{
			Preseed: "config: {}\n",
		},
		MCollective: &MCollectiveConfig{
			Conf: &MCollectiveConf{
				PublicCert: "cert",
			},
		},
		Puppet: &PuppetConfig{
			Install: Bool(true),
		},
		RHSubscription: &RHSubscriptionConfig{
			Username: "user",
			Password: "pass",
		},
		Rsyslog: &RsyslogConfig{
			Configs: []RsyslogConfigEntry{RsyslogContent("*.* @192.0.2.15")},
		},
		SaltMinion: &SaltMinionConfig{
			Conf: RawObject{
				"master": "salt.example.com",
			},
		},
		Spacewalk: &SpacewalkConfig{
			Server:        "spacewalk.example.com",
			ActivationKey: "activation-key",
		},
		UbuntuPro: &UbuntuProConfig{
			Token: "pro-token",
		},
		WireGuard: &WireGuardConfig{
			Interfaces: []WireGuardInterface{
				{Name: "wg0", ConfigPath: "/etc/wireguard/wg0.conf", Content: "[Interface]\nPrivateKey = abc\n"},
			},
		},
	}

	rendered, err := cfg.Render()
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	for _, needle := range []string{
		"apt:",
		"apk_repos:",
		"apt_pipelining: os",
		"snap:",
		"yum_repos:",
		"zypper:",
		"disable_ec2_metadata: true",
		"manage_resolv_conf: true",
		"random_seed:",
		"ansible:",
		"landscape:",
		"wireguard:",
	} {
		if !strings.Contains(string(rendered), needle) {
			t.Fatalf("rendered config missing %q\n%s", needle, rendered)
		}
	}
}

func TestAdditionalModuleValidationErrors(t *testing.T) {
	cfg := Config{
		APKRepos: &APKReposConfig{
			AlpineRepo: &AlpineRepo{},
		},
		APTPipelining: IntStringValue("invalid"),
		YumRepos: YumRepos{
			"bad": {},
		},
		Zypper: &ZypperConfig{
			Repos: []ZypperRepo{{}},
		},
		ManageResolvConf: nil,
		ResolvConf:       &ResolvConfConfig{},
		Drivers: &UbuntuDriversConfig{
			Nvidia: &NVIDIADrivers{},
		},
		Autoinstall: &UbuntuAutoinstallConfig{},
		Ansible: &AnsibleConfig{
			Pull: &AnsiblePull{
				URL: "https://git.example.com/ops.git",
			},
		},
		Landscape: &LandscapeConfig{},
		WireGuard: &WireGuardConfig{
			Interfaces: []WireGuardInterface{{}},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() returned nil error for invalid additional-module config")
	}

	message := err.Error()
	for _, needle := range []string{
		"apk_repos.alpine_repo.version",
		"apt_pipelining",
		"yum_repos.bad",
		"zypper.repos[0].id",
		"manage_resolv_conf",
		"drivers.nvidia.license-accepted",
		"autoinstall.version",
		"ansible.install_method",
		"ansible.package_name",
		"ansible.pull.playbook_name",
		"landscape.client",
		"wireguard.interfaces[0].name",
		"wireguard.interfaces[0].config_path",
		"wireguard.interfaces[0].content",
	} {
		if !strings.Contains(message, needle) {
			t.Fatalf("expected validation error to mention %q\n%s", needle, message)
		}
	}
}
