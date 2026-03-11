package cloudinit

// WriteFileSource describes a remote source for a write_files entry.
type WriteFileSource struct {
	URI     string            `yaml:"uri"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

// WriteFile represents one write_files item.
type WriteFile struct {
	Path        string           `yaml:"path,omitempty"`
	Content     string           `yaml:"content,omitempty"`
	Source      *WriteFileSource `yaml:"source,omitempty"`
	Owner       string           `yaml:"owner,omitempty"`
	Permissions string           `yaml:"permissions,omitempty"`
	Encoding    string           `yaml:"encoding,omitempty"`
	Append      *bool            `yaml:"append,omitempty"`
	Defer       *bool            `yaml:"defer,omitempty"`
}
