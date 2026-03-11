package cloudinit

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCloudInitSchemaIntegration(t *testing.T) {
	if _, err := exec.LookPath("cloud-init"); err != nil {
		t.Skip("cloud-init binary not available")
	}

	cfg := Config{
		Hostname:          "schema-check",
		PackageUpdate:     Bool(true),
		Packages:          []PackageEntry{Package("curl")},
		WriteFiles:        []WriteFile{{Path: "/etc/test.txt", Content: "hello\n"}},
		SSHAuthorizedKeys: []string{"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIexample schema@example"},
		Users: []UserEntry{
			UserReference("default"),
			{Name: "alice", Groups: []string{"sudo"}, LockPasswd: Bool(false)},
		},
		RunCmd: []Command{
			ShellCommand("echo schema"),
		},
		PowerState: &PowerState{
			Mode:  Reboot,
			Delay: DelayNow(),
		},
	}

	rendered, err := cfg.Render()
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "cloud-config.yaml")
	if err := os.WriteFile(path, rendered, 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	cmd := exec.Command("cloud-init", "schema", "-c", path, "-t", "cloud-config")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("cloud-init schema failed: %v\n%s", err, output)
	}
}

func TestCloudInitSchemaExpandedModuleIntegration(t *testing.T) {
	if _, err := exec.LookPath("cloud-init"); err != nil {
		t.Skip("cloud-init binary not available")
	}

	cfg := Config{
		APT: &APTConfig{
			Sources: map[string]APTSource{
				"corp": {Source: "deb https://apt.example.com stable main"},
			},
		},
		APKRepos: &APKReposConfig{
			AlpineRepo: &AlpineRepo{Version: "v3.20"},
		},
		APTPipelining: IntStringValue("os"),
		Snap: &SnapConfig{
			Commands: &CommandCollection{
				List: []Command{ShellCommand("install hello-world")},
			},
		},
		YumRepos: YumRepos{
			"epel": {BaseURL: "https://yum.example.com/epel"},
		},
		Zypper: &ZypperConfig{
			Repos: []ZypperRepo{{ID: "base", BaseURL: "https://zypper.example.com/base"}},
		},
		ByobuByDefault:     "enable-system",
		DisableEC2Metadata: Bool(true),
		Fan:                &FanConfig{Config: "underlay eth0\n"},
		GrubDpkg:           &GrubDpkgConfig{Enabled: Bool(true)},
		Updates:            &HotplugUpdates{Network: &HotplugNetworkConfig{When: []string{"boot", "hotplug"}}},
		Keyboard:           &KeyboardConfig{Layout: "us"},
		ManageResolvConf:   Bool(true),
		ResolvConf:         &ResolvConfConfig{Nameservers: []string{"1.1.1.1"}},
		RandomSeed:         &RandomSeedConfig{Data: "seed-data", Encoding: "raw"},
		Drivers:            &UbuntuDriversConfig{Nvidia: &NVIDIADrivers{LicenseAccepted: Bool(true)}},
		Autoinstall:        &UbuntuAutoinstallConfig{Version: 1},
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
		LXD: &LXDConfig{Preseed: "config: {}\n"},
		MCollective: &MCollectiveConfig{
			Conf: &MCollectiveConf{PublicCert: "cert"},
		},
		Puppet:         &PuppetConfig{Install: Bool(true)},
		RHSubscription: &RHSubscriptionConfig{Username: "user", Password: "pass"},
		Rsyslog: &RsyslogConfig{
			Configs: []RsyslogConfigEntry{RsyslogContent("*.* @192.0.2.15")},
		},
		SaltMinion: &SaltMinionConfig{
			Conf: RawObject{"master": "salt.example.com"},
		},
		Spacewalk: &SpacewalkConfig{Server: "spacewalk.example.com", ActivationKey: "activation-key"},
		UbuntuPro: &UbuntuProConfig{Token: "pro-token"},
		WireGuard: &WireGuardConfig{
			Interfaces: []WireGuardInterface{{
				Name:       "wg0",
				ConfigPath: "/etc/wireguard/wg0.conf",
				Content:    "[Interface]\nPrivateKey = abc\n",
			}},
		},
	}

	rendered, err := cfg.Render()
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "expanded-cloud-config.yaml")
	if err := os.WriteFile(path, rendered, 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	cmd := exec.Command("cloud-init", "schema", "-c", path, "-t", "cloud-config")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("cloud-init schema failed: %v\n%s", err, output)
	}
}
