package cloudinit

import "fmt"

// GroupEntry represents one groups item.
type GroupEntry struct {
	Name    string
	Members []string
}

// Group returns a group entry with the provided members.
func Group(name string, members ...string) GroupEntry {
	return GroupEntry{Name: name, Members: members}
}

// MarshalYAML implements yaml.Marshaler.
func (g GroupEntry) MarshalYAML() (any, error) {
	if g.Name == "" {
		return nil, fmt.Errorf("group name is required")
	}
	if len(g.Members) == 0 {
		return g.Name, nil
	}
	return map[string]any{g.Name: g.Members}, nil
}

// UserEntry represents one user or user reference.
type UserEntry struct {
	Name              string   `yaml:"name,omitempty"`
	SnapUser          string   `yaml:"snapuser,omitempty"`
	Reference         string   `yaml:"-"`
	Doas              []string `yaml:"doas,omitempty"`
	ExpireDate        string   `yaml:"expiredate,omitempty"`
	GECOS             string   `yaml:"gecos,omitempty"`
	Groups            []string `yaml:"groups,omitempty"`
	HomeDir           string   `yaml:"homedir,omitempty"`
	Inactive          string   `yaml:"inactive,omitempty"`
	LockPasswd        *bool    `yaml:"lock_passwd,omitempty"`
	NoCreateHome      *bool    `yaml:"no_create_home,omitempty"`
	NoLogInit         *bool    `yaml:"no_log_init,omitempty"`
	NoUserGroup       *bool    `yaml:"no_user_group,omitempty"`
	Passwd            string   `yaml:"passwd,omitempty"`
	HashedPasswd      string   `yaml:"hashed_passwd,omitempty"`
	PlainTextPasswd   string   `yaml:"plain_text_passwd,omitempty"`
	CreateGroups      *bool    `yaml:"create_groups,omitempty"`
	PrimaryGroup      string   `yaml:"primary_group,omitempty"`
	SELinuxUser       string   `yaml:"selinux_user,omitempty"`
	Shell             string   `yaml:"shell,omitempty"`
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys,omitempty"`
	SSHImportID       []string `yaml:"ssh_import_id,omitempty"`
	SSHRedirectUser   *bool    `yaml:"ssh_redirect_user,omitempty"`
	System            *bool    `yaml:"system,omitempty"`
	Sudo              []string `yaml:"sudo,omitempty"`
	UID               *int     `yaml:"uid,omitempty"`
}

// DefaultUserReference returns a pointer to a reference-style user entry.
func DefaultUserReference(name string) *UserEntry {
	return &UserEntry{Reference: name}
}

// UserReference returns a reference-style user entry.
func UserReference(name string) UserEntry {
	return UserEntry{Reference: name}
}

// MarshalYAML implements yaml.Marshaler.
func (u UserEntry) MarshalYAML() (any, error) {
	if u.Reference != "" {
		if referenceUserHasExtras(&u) {
			return nil, fmt.Errorf("reference users cannot define additional fields")
		}
		return u.Reference, nil
	}
	if u.Name == "" && u.SnapUser == "" {
		return nil, fmt.Errorf("user entry requires name, snapuser, or Reference")
	}
	if u.Name != "" && u.SnapUser != "" {
		return nil, fmt.Errorf("user entry cannot define both name and snapuser")
	}
	type alias UserEntry
	return alias(u), nil
}

// ChPasswd configures the chpasswd module.
type ChPasswd struct {
	Expire *bool          `yaml:"expire,omitempty"`
	Users  []PasswordUser `yaml:"users,omitempty"`
}

// PasswordType identifies the kind of password value supplied for a user.
type PasswordType string

// PasswordType values supported by chpasswd.users.
const (
	PasswordTypeHash   PasswordType = "hash"
	PasswordTypeText   PasswordType = "text"
	PasswordTypeRandom PasswordType = "RANDOM"
)

// PasswordUser represents one chpasswd.users entry.
type PasswordUser struct {
	Name     string       `yaml:"name,omitempty"`
	Type     PasswordType `yaml:"type,omitempty"`
	Password string       `yaml:"password,omitempty"`
}
