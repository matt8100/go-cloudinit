package cloudinit

// SSHKeys contains host key material for the ssh_keys module.
type SSHKeys struct {
	ECDSAPrivate       string `yaml:"ecdsa_private,omitempty"`
	ECDSAPublic        string `yaml:"ecdsa_public,omitempty"`
	ECDSACertificate   string `yaml:"ecdsa_certificate,omitempty"`
	ED25519Private     string `yaml:"ed25519_private,omitempty"`
	ED25519Public      string `yaml:"ed25519_public,omitempty"`
	ED25519Certificate string `yaml:"ed25519_certificate,omitempty"`
	RSAPrivate         string `yaml:"rsa_private,omitempty"`
	RSAPublic          string `yaml:"rsa_public,omitempty"`
	RSACertificate     string `yaml:"rsa_certificate,omitempty"`
}

// SSHPublishHostKeys configures ssh_publish_hostkeys.
type SSHPublishHostKeys struct {
	Enabled   *bool    `yaml:"enabled,omitempty"`
	Blacklist []string `yaml:"blacklist,omitempty"`
}

// SSHConsoleConfig configures the ssh console output section.
type SSHConsoleConfig struct {
	EmitKeysToConsole *bool `yaml:"emit_keys_to_console,omitempty"`
}
