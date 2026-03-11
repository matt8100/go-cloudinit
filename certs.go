package cloudinit

// CACertsConfig configures the ca_certs module.
type CACertsConfig struct {
	RemoveDefaults *bool    `yaml:"remove_defaults,omitempty"`
	Trusted        []string `yaml:"trusted,omitempty"`
}
