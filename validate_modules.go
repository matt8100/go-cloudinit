package cloudinit

import (
	"fmt"
	"net"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

func formatIndex(path string, index int) string {
	return fmt.Sprintf("%s[%d]", path, index)
}

func mapKeyPath(path, key string) string {
	if key == "" {
		return path
	}
	return path + "." + key
}

var rsyslogRemotePattern = regexp.MustCompile(`^(@{0,2})(?:\[([^\]]*)\]|([^:]*))(?::([0-9]+))?$`)
var landscapeTagsPattern = regexp.MustCompile(`^[-_0-9a-zA-Z]+(,[-_0-9a-zA-Z]+)*$`)
var yumRepoIDPattern = regexp.MustCompile(`^[0-9a-zA-Z _-]+$`)
var yumOptionNamePattern = regexp.MustCompile(`^[0-9a-zA-Z_]+$`)
var validNTPClients = map[string]struct{}{
	"":                  {},
	"auto":              {},
	"chrony":            {},
	"ntp":               {},
	"ntpdate":           {},
	"openntpd":          {},
	"systemd-timesyncd": {},
}

func validateStringOrStringSlice(errs *ValidationErrors, path string, value *StringOrStringSlice) (present bool, valid bool) {
	if value == nil {
		return false, false
	}
	switch {
	case value.String != "" && len(value.List) == 0:
		return true, true
	case value.String == "" && len(value.List) > 0:
		return true, true
	case value.String == "" && len(value.List) == 0:
		errs.add(path, "must define either String or List")
		return true, false
	default:
		errs.add(path, "must define exactly one of String or List")
		return true, false
	}
}

func validateStringOrInt(errs *ValidationErrors, path string, value *StringOrInt) bool {
	if value == nil {
		return false
	}
	switch {
	case value.String != "" && value.Int == nil:
		return true
	case value.String == "" && value.Int != nil:
		return true
	case value.String == "" && value.Int == nil:
		errs.add(path, "must define either String or Int")
		return false
	default:
		errs.add(path, "must define exactly one of String or Int")
		return false
	}
}

func validateBoolOrStringDeprecated(errs *ValidationErrors, path string, value *BoolOrStringDeprecated) {
	if value == nil {
		return
	}
	switch {
	case value.Bool != nil && value.String == "":
		return
	case value.Bool == nil && value.String != "":
		return
	case value.Bool == nil && value.String == "":
		errs.add(path, "must define either Bool or String")
	default:
		errs.add(path, "must define exactly one of Bool or String")
	}
}

func asStringKeyMap(value any) (map[string]any, bool) {
	if value == nil {
		return nil, false
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Map || rv.Type().Key().Kind() != reflect.String {
		return nil, false
	}
	out := make(map[string]any, rv.Len())
	for _, key := range rv.MapKeys() {
		out[key.String()] = rv.MapIndex(key).Interface()
	}
	return out, true
}

func validateInlineMapKeys(errs *ValidationErrors, path string, value any, fixedKeys ...string) {
	if value == nil {
		return
	}
	m, ok := asStringKeyMap(value)
	if !ok {
		errs.add(path, "must be an object")
		return
	}
	for _, key := range fixedKeys {
		if _, ok := m[key]; ok {
			errs.add(path+"."+key, "conflicts with an inlined field")
		}
	}
}

func validateScalarMapValues(errs *ValidationErrors, path string, value any) {
	if value == nil {
		return
	}
	m, ok := asStringKeyMap(value)
	if !ok {
		errs.add(path, "must be an object")
		return
	}
	for key, item := range m {
		if key == "" {
			errs.add(path, "property names must be non-empty")
			continue
		}
		rv := reflect.ValueOf(item)
		if !rv.IsValid() {
			errs.add(path+"."+key, "must be a boolean, integer, or string")
			continue
		}
		switch rv.Kind() {
		case reflect.Bool, reflect.String,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			continue
		default:
			errs.add(path+"."+key, "must be a boolean, integer, or string")
		}
	}
}

func validateStringCollection(errs *ValidationErrors, path string, value *StringCollection) bool {
	if value == nil {
		return false
	}
	switch {
	case len(value.List) > 0 && len(value.Map) == 0:
		validateUniqueStrings(errs, path, value.List)
		return true
	case len(value.Map) > 0 && len(value.List) == 0:
		return true
	case len(value.List) == 0 && len(value.Map) == 0:
		errs.add(path, "must define either List or Map entries")
	default:
		errs.add(path, "must define exactly one of List or Map")
	}
	return false
}

func validateCommandCollection(errs *ValidationErrors, path string, value *CommandCollection) bool {
	if value == nil {
		return false
	}
	switch {
	case len(value.List) > 0 && len(value.Map) == 0:
		validateCommands(errs, path, value.List, false)
		return true
	case len(value.Map) > 0 && len(value.List) == 0:
		for name, command := range value.Map {
			validateCommands(errs, path+"."+name, []Command{command}, false)
		}
		return true
	case len(value.List) == 0 && len(value.Map) == 0:
		errs.add(path, "must define either List or Map entries")
	default:
		errs.add(path, "must define exactly one of List or Map")
	}
	return false
}

func validNTPHostname(value string) bool {
	if value == "" || strings.Contains(value, "://") || strings.ContainsAny(value, "/?# ") {
		return false
	}
	if ip := net.ParseIP(value); ip != nil {
		return true
	}
	if strings.HasSuffix(value, ".") {
		value = strings.TrimSuffix(value, ".")
	}
	if value == "" || len(value) > 253 {
		return false
	}
	for _, label := range strings.Split(value, ".") {
		if label == "" || len(label) > 63 {
			return false
		}
		for i := range label {
			c := label[i]
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' {
				continue
			}
			return false
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
	}
	return true
}

func validateStringListValue(errs *ValidationErrors, path string, value any, minItems int) []string {
	var items []string
	switch v := value.(type) {
	case []string:
		items = v
	case []any:
		items = make([]string, 0, len(v))
		for i, item := range v {
			s, ok := item.(string)
			if !ok {
				errs.add(formatIndex(path, i), "must be a string")
				continue
			}
			items = append(items, s)
		}
	default:
		errs.add(path, "must be an array of strings")
		return nil
	}

	if len(items) < minItems {
		errs.add(path, fmt.Sprintf("must define at least %d item(s)", minItems))
	}
	for i, item := range items {
		if item == "" {
			errs.add(formatIndex(path, i), "must be a non-empty string")
		}
	}
	return items
}

func validateUniqueStrings(errs *ValidationErrors, path string, items []string) {
	seen := make(map[string]int, len(items))
	for i, item := range items {
		if first, ok := seen[item]; ok {
			errs.add(formatIndex(path, i), fmt.Sprintf("must be unique; duplicate of index %d", first))
			continue
		}
		seen[item] = i
	}
}

func validateManageEtcHosts(errs *ValidationErrors, value *ManageEtcHostsValue) {
	if value == nil {
		return
	}
	switch value.Mode {
	case ManageEtcHostsTrue, ManageEtcHostsFalse, ManageEtcHostsLocalhost, ManageEtcHostsTemplate:
	default:
		errs.add("manage_etc_hosts", fmt.Sprintf("unsupported mode %q", value.Mode))
	}
}

func validateLocale(errs *ValidationErrors, value *LocaleValue) {
	if value == nil {
		return
	}
	if value.Enabled != nil && value.Name != "" {
		errs.add("locale", "cannot define both Name and Enabled")
	}
	if value.Enabled == nil && value.Name == "" {
		errs.add("locale", "must define either Name or Enabled")
	}
}

func validatePackages(errs *ValidationErrors, packages []PackageEntry) {
	for i, entry := range packages {
		path := formatIndex("packages", i)
		switch {
		case entry.Inline != nil && entry.Managed == nil:
			if entry.Inline.Name == "" {
				errs.add(path, "package name is required")
			}
		case entry.Managed != nil && entry.Inline == nil:
			switch {
			case len(entry.Managed.APT) > 0 && len(entry.Managed.Snap) == 0:
				for j, spec := range entry.Managed.APT {
					if spec.Name == "" {
						errs.add(formatIndex(path+".apt", j), "package name is required")
					}
				}
			case len(entry.Managed.Snap) > 0 && len(entry.Managed.APT) == 0:
				for j, spec := range entry.Managed.Snap {
					if spec.Name == "" {
						errs.add(formatIndex(path+".snap", j), "package name is required")
					}
				}
			default:
				errs.add(path, "managed packages must define exactly one manager")
			}
		default:
			errs.add(path, "must define either Inline or Managed")
		}
	}
}

func validateAPT(errs *ValidationErrors, config *APTConfig) {
	if config == nil {
		return
	}
	if config.PreserveSourcesList == nil &&
		config.GenerateMirrorLists == nil &&
		len(config.DisableSuites) == 0 &&
		len(config.Primary) == 0 &&
		len(config.Security) == 0 &&
		config.AddAPTRepoMatch == "" &&
		len(config.DebconfSelections) == 0 &&
		config.SourcesList == "" &&
		config.Conf == "" &&
		config.HTTPSProxy == "" &&
		config.HTTPProxy == "" &&
		config.Proxy == "" &&
		config.FTPProxy == "" &&
		len(config.Sources) == 0 {
		errs.add("apt", "must define at least one property")
	}
	for i, suite := range config.DisableSuites {
		if suite == "" {
			errs.add(formatIndex("apt.disable_suites", i), "must be a non-empty string")
		}
	}
	validateUniqueStrings(errs, "apt.disable_suites", config.DisableSuites)
	if config.DebconfSelections != nil && len(config.DebconfSelections) == 0 {
		errs.add("apt.debconf_selections", "must define at least one entry")
	}
	for name := range config.DebconfSelections {
		if name == "" {
			errs.add("apt.debconf_selections", "keys must be non-empty")
		}
	}
	validateAPTMirrors(errs, "apt.primary", config.Primary)
	validateAPTMirrors(errs, "apt.security", config.Security)
	for name, source := range config.Sources {
		if name == "" {
			errs.add("apt.sources", "keys must be non-empty")
		}
		if source.Source == "" && source.KeyID == "" && source.Key == "" && source.KeyServer == "" && source.Filename == "" && source.Append == nil {
			errs.add(mapKeyPath("apt.sources", name), "source definition must not be empty")
		}
	}
}

func validateAPTMirrors(errs *ValidationErrors, path string, mirrors []APTMirrorConfig) {
	for i, mirror := range mirrors {
		itemPath := formatIndex(path, i)
		for key, value := range mirror {
			switch key {
			case "arches":
				validateStringListValue(errs, itemPath+".arches", value, 1)
			case "uri":
				uri, ok := value.(string)
				if !ok {
					errs.add(itemPath+".uri", "must be a string")
				} else if uri == "" || !validateURI(uri) {
					errs.add(itemPath+".uri", "must be an absolute URI")
				}
			case "search":
				search := validateStringListValue(errs, itemPath+".search", value, 1)
				for j, uri := range search {
					if uri != "" && !validateURI(uri) {
						errs.add(formatIndex(itemPath+".search", j), "must be an absolute URI")
					}
				}
			case "search_dns":
				if _, ok := value.(bool); !ok {
					errs.add(itemPath+".search_dns", "must be a boolean")
				}
			case "keyid", "key", "keyserver":
				if _, ok := value.(string); !ok {
					errs.add(itemPath+"."+key, "must be a string")
				}
			default:
				errs.add(itemPath+"."+key, "unsupported property")
			}
		}
		if _, ok := mirror["arches"]; !ok {
			errs.add(itemPath+".arches", "arches is required")
		}
	}
}

func validateAPKRepos(errs *ValidationErrors, config *APKReposConfig) {
	if config == nil {
		return
	}
	if config.PreserveRepositories == nil && config.AlpineRepo == nil && config.LocalRepoBaseURL == "" {
		errs.add("apk_repos", "must define at least one property")
	}
	if config.AlpineRepo != nil && config.AlpineRepo.Version == "" {
		errs.add("apk_repos.alpine_repo.version", "version is required")
	}
}

func validateAPTPipelining(errs *ValidationErrors, value *IntBoolStringValue) {
	if value == nil {
		return
	}
	count := 0
	if value.Int != nil {
		count++
	}
	if value.Bool != nil {
		count++
	}
	if value.String != "" {
		count++
		switch value.String {
		case "os", "none", "unchanged":
		default:
			errs.add("apt_pipelining", fmt.Sprintf("unsupported string value %q", value.String))
		}
	}
	if count != 1 {
		errs.add("apt_pipelining", "must define exactly one of Int, Bool, or String")
	}
}

func validateSnapConfig(errs *ValidationErrors, config *SnapConfig) {
	if config == nil {
		return
	}
	if config.Assertions == nil && config.Commands == nil {
		errs.add("snap", "must define assertions or commands")
	}
	validateStringCollection(errs, "snap.assertions", config.Assertions)
	validateCommandCollection(errs, "snap.commands", config.Commands)
}

func validateYumRepos(errs *ValidationErrors, repoDir string, repos YumRepos) {
	if repoDir != "" && len(repos) == 0 {
		errs.add("yum_repo_dir", "yum_repo_dir requires yum_repos")
	}
	for name, repo := range repos {
		path := mapKeyPath("yum_repos", name)
		if name == "" || !yumRepoIDPattern.MatchString(name) {
			errs.add(path, "repo id must match ^[0-9a-zA-Z _-]+$")
		}
		validateInlineMapKeys(errs, path, repo.Options, "baseurl", "metalink", "mirrorlist", "name", "enabled")
		validateScalarMapValues(errs, path, repo.Options)
		for optionName := range repo.Options {
			if optionName == "" || !yumOptionNamePattern.MatchString(optionName) {
				errs.add(mapKeyPath(path, optionName), "option name must match ^[0-9a-zA-Z_]+$")
			}
		}
		if repo.BaseURL == "" && repo.Metalink == "" && repo.MirrorList == "" {
			errs.add(path, "must define one of baseurl, metalink, or mirrorlist")
		}
		if repo.BaseURL != "" && !validateURI(repo.BaseURL) {
			errs.add(path+".baseurl", "must be an absolute URI")
		}
		if repo.Metalink != "" && !validateURI(repo.Metalink) {
			errs.add(path+".metalink", "must be an absolute URI")
		}
		if repo.MirrorList != "" && !validateURI(repo.MirrorList) {
			errs.add(path+".mirrorlist", "must be an absolute URI")
		}
	}
}

func validateZypper(errs *ValidationErrors, config *ZypperConfig) {
	if config == nil {
		return
	}
	if len(config.Repos) == 0 && len(config.Config) == 0 {
		errs.add("zypper", "must define repos or config")
	}
	for i, repo := range config.Repos {
		path := formatIndex("zypper.repos", i)
		validateInlineMapKeys(errs, path, repo.Options, "id", "baseurl")
		if repo.ID == "" {
			errs.add(path+".id", "id is required")
		}
		if repo.BaseURL == "" {
			errs.add(path+".baseurl", "baseurl is required")
		} else if !validateURI(repo.BaseURL) {
			errs.add(path+".baseurl", "must be an absolute URI")
		}
	}
}

func validateCommands(errs *ValidationErrors, path string, commands []Command, allowNull bool) {
	for i, cmd := range commands {
		itemPath := formatIndex(path, i)
		switch {
		case cmd.Null:
			if cmd.String != "" || len(cmd.Args) > 0 {
				errs.add(itemPath, "must define exactly one of String, Args, or Null")
				continue
			}
			if !allowNull {
				errs.add(itemPath, "null commands are not allowed here")
			}
		case cmd.String != "" && len(cmd.Args) == 0:
		case cmd.String == "" && len(cmd.Args) > 0:
			for j, arg := range cmd.Args {
				if arg == "" {
					errs.add(formatIndex(itemPath, j), "command arguments must be non-empty")
				}
			}
		default:
			errs.add(itemPath, "must define exactly one of String, Args, or Null")
		}
	}
}

func validateWriteFiles(errs *ValidationErrors, files []WriteFile) {
	for i, file := range files {
		path := formatIndex("write_files", i)
		if file.Path == "" {
			errs.add(path+".path", "path is required")
		}
		if file.Encoding != "" {
			if _, ok := validWriteFileEncodings[file.Encoding]; !ok {
				errs.add(path+".encoding", "unsupported encoding")
			}
		}
		if file.Source != nil {
			if file.Source.URI == "" {
				errs.add(path+".source.uri", "uri is required")
			} else if !validateURI(file.Source.URI) {
				errs.add(path+".source.uri", "must be an absolute URI")
			}
		}
	}
}

func validateGroups(errs *ValidationErrors, groups []GroupEntry) {
	for i, group := range groups {
		path := formatIndex("groups", i)
		if group.Name == "" {
			errs.add(path, "group name is required")
		}
		for j, member := range group.Members {
			if member == "" {
				errs.add(formatIndex(path+".members", j), "member names must be non-empty")
			}
		}
	}
}

func validateUserEntry(errs *ValidationErrors, path string, entry *UserEntry) {
	if entry == nil {
		return
	}
	if entry.Reference != "" {
		if referenceUserHasExtras(entry) {
			errs.add(path, "reference users cannot define additional fields")
		}
		return
	}
	if entry.Name == "" && entry.SnapUser == "" {
		errs.add(path, "user must define name, snapuser, or Reference")
	}
	if entry.Name != "" && entry.SnapUser != "" {
		errs.add(path, "name and snapuser are mutually exclusive")
	}
	if valueTrue(entry.SSHRedirectUser) {
		if len(entry.SSHAuthorizedKeys) > 0 {
			errs.add(path+".ssh_authorized_keys", "cannot be combined with ssh_redirect_user")
		}
		if len(entry.SSHImportID) > 0 {
			errs.add(path+".ssh_import_id", "cannot be combined with ssh_redirect_user")
		}
	}
	if valueTrue(entry.NoCreateHome) || valueTrue(entry.System) {
		if len(entry.SSHAuthorizedKeys) > 0 {
			errs.add(path+".ssh_authorized_keys", "requires a home directory")
		}
		if len(entry.SSHImportID) > 0 {
			errs.add(path+".ssh_import_id", "requires a home directory")
		}
		if valueTrue(entry.SSHRedirectUser) {
			errs.add(path+".ssh_redirect_user", "requires a home directory")
		}
	}
}

func referenceUserHasExtras(entry *UserEntry) bool {
	return entry.Name != "" ||
		entry.SnapUser != "" ||
		len(entry.Doas) > 0 ||
		entry.ExpireDate != "" ||
		entry.GECOS != "" ||
		len(entry.Groups) > 0 ||
		entry.HomeDir != "" ||
		entry.Inactive != "" ||
		entry.LockPasswd != nil ||
		entry.NoCreateHome != nil ||
		entry.NoLogInit != nil ||
		entry.NoUserGroup != nil ||
		entry.Passwd != "" ||
		entry.HashedPasswd != "" ||
		entry.PlainTextPasswd != "" ||
		entry.CreateGroups != nil ||
		entry.PrimaryGroup != "" ||
		entry.SELinuxUser != "" ||
		entry.Shell != "" ||
		len(entry.SSHAuthorizedKeys) > 0 ||
		len(entry.SSHImportID) > 0 ||
		entry.SSHRedirectUser != nil ||
		entry.System != nil ||
		len(entry.Sudo) > 0 ||
		entry.UID != nil
}

func validateChPasswd(errs *ValidationErrors, config *ChPasswd) {
	if config == nil {
		return
	}
	for i, user := range config.Users {
		path := formatIndex("chpasswd.users", i)
		if user.Name == "" {
			errs.add(path+".name", "name is required")
		}
		switch user.Type {
		case "", PasswordTypeHash, PasswordTypeText:
			if user.Password == "" {
				errs.add(path+".password", "password is required when type is hash or text")
			}
		case PasswordTypeRandom:
			if user.Password != "" {
				errs.add(path+".password", "password must be omitted when type is RANDOM")
			}
		default:
			errs.add(path+".type", "unsupported password type")
		}
	}
}

func validateSSH(errs *ValidationErrors, cfg Config) {
	for i, keyType := range cfg.SSHGenKeyTypes {
		if !validSSHKeyType(keyType) {
			errs.add(formatIndex("ssh_genkeytypes", i), fmt.Sprintf("unsupported SSH key type %q", keyType))
		}
	}
	for i, key := range cfg.SSHAuthorizedKeys {
		if key == "" {
			errs.add(formatIndex("ssh_authorized_keys", i), "SSH keys must be non-empty")
		}
	}
	for i, value := range cfg.SSHImportID {
		if value == "" {
			errs.add(formatIndex("ssh_import_id", i), "SSH import IDs must be non-empty")
		}
	}
	if cfg.SSHPWAuth != nil {
		if cfg.SSHPWAuth.Bool == nil && cfg.SSHPWAuth.String == "" {
			errs.add("ssh_pwauth", "must define either Bool or String")
		}
		if cfg.SSHPWAuth.Bool != nil && cfg.SSHPWAuth.String != "" {
			errs.add("ssh_pwauth", "cannot define both Bool and String")
		}
	}
	if cfg.SSHKeys != nil {
		keys := []string{
			cfg.SSHKeys.ECDSAPrivate, cfg.SSHKeys.ECDSAPublic, cfg.SSHKeys.ECDSACertificate,
			cfg.SSHKeys.ED25519Private, cfg.SSHKeys.ED25519Public, cfg.SSHKeys.ED25519Certificate,
			cfg.SSHKeys.RSAPrivate, cfg.SSHKeys.RSAPublic, cfg.SSHKeys.RSACertificate,
		}
		hasValue := false
		for _, key := range keys {
			hasValue = hasValue || key != ""
		}
		if !hasValue {
			errs.add("ssh_keys", "must define at least one host key value")
		}
	}
	if cfg.SSHConsole != nil && cfg.SSHConsole.EmitKeysToConsole == nil {
		errs.add("ssh.emit_keys_to_console", "emit_keys_to_console is required")
	}
	validateUniqueStrings(errs, "ssh_key_console_blacklist", cfg.SSHKeyConsoleBlacklist)
	validateUniqueStrings(errs, "ssh_fp_console_blacklist", cfg.SSHFPConsoleBlacklist)
}

func validateCACerts(errs *ValidationErrors, ca *CACertsConfig) {
	if ca == nil {
		return
	}
	if !valueTrue(ca.RemoveDefaults) && len(ca.Trusted) == 0 {
		errs.add("ca_certs", "must define trusted certificates or remove_defaults")
	}
	for i, cert := range ca.Trusted {
		if cert == "" {
			errs.add(formatIndex("ca_certs.trusted", i), "trusted certificates must be non-empty")
		}
	}
}

func validateMounts(errs *ValidationErrors, mounts []MountEntry, defaults *MountDefaults, swap *SwapDefinition) {
	for i, mount := range mounts {
		path := formatIndex("mounts", i)
		if mount.Spec == "" {
			errs.add(path, "spec is required")
		}
	}
	if defaults != nil {
		for i, value := range defaults {
			if value != nil && *value == "" {
				errs.add(formatIndex("mount_default_fields", i), "must be null or non-empty string")
			}
		}
	}
	if swap == nil {
		return
	}
	validateSizeOrValue(errs, "swap.size", swap.Size, true)
	validateSizeOrValue(errs, "swap.maxsize", swap.MaxSize, false)
}

func validateDiskSetup(errs *ValidationErrors, aliases map[string]string, setups DiskSetupMap, filesystems []FilesystemSetup) {
	for name, path := range aliases {
		if name == "" {
			errs.add("device_aliases", "keys must be non-empty")
		}
		if path == "" {
			errs.add(mapKeyPath("device_aliases", name), "alias path must be non-empty")
		}
	}
	for name, setup := range setups {
		path := mapKeyPath("disk_setup", name)
		if name == "" {
			errs.add("disk_setup", "keys must be non-empty")
		}
		if setup.Layout == nil {
			errs.add(path+".layout", "layout is required")
			continue
		}
		if setup.TableType != "" && setup.TableType != "mbr" && setup.TableType != "gpt" {
			errs.add(path+".table_type", "must be mbr or gpt")
		}
		modeCount := 0
		if setup.Layout.Remove {
			modeCount++
		}
		if setup.Layout.Enabled != nil {
			modeCount++
		}
		if len(setup.Layout.Parts) > 0 {
			modeCount++
			for i, part := range setup.Layout.Parts {
				if part.Size <= 0 {
					errs.add(formatIndex(path+".layout.parts", i)+".size", "size must be > 0")
				}
			}
		}
		if modeCount != 1 {
			errs.add(path+".layout", "must define exactly one of Remove, Enabled, or Parts")
		}
	}
	for i, fs := range filesystems {
		path := formatIndex("fs_setup", i)
		if fs.Device == "" {
			errs.add(path+".device", "device is required")
		}
		cmdPresent, cmdValid := validateStringOrStringSlice(errs, path+".cmd", fs.Cmd)
		validateStringOrStringSlice(errs, path+".extra_opts", fs.ExtraOpts)
		if fs.Filesystem == "" && !cmdValid {
			if !cmdPresent {
				errs.add(path+".filesystem", "filesystem or cmd is required")
			}
		}
		if validateStringOrInt(errs, path+".partition", fs.Partition) && fs.Partition.String != "" {
			switch fs.Partition.String {
			case "auto", "any", "none":
			default:
				errs.add(path+".partition", "string value must be auto, any, or none")
			}
		}
	}
}

func validatePhoneHome(errs *ValidationErrors, config *PhoneHomeConfig) {
	if config == nil {
		return
	}
	if config.URL == "" {
		errs.add("phone_home.url", "url is required")
	} else if !validateURI(config.URL) {
		errs.add("phone_home.url", "must be an absolute URI")
	}
	if config.Tries != nil && *config.Tries < 0 {
		errs.add("phone_home.tries", "must be >= 0")
	}
	if config.Post == nil {
		return
	}
	if config.Post.All && len(config.Post.Fields) > 0 {
		errs.add("phone_home.post", "cannot set All and Fields together")
	}
	if !config.Post.All && len(config.Post.Fields) == 0 {
		errs.add("phone_home.post", "must define All or at least one field")
	}
	for i, field := range config.Post.Fields {
		if _, ok := validPhoneHomeFields[field]; !ok {
			errs.add(formatIndex("phone_home.post", i), fmt.Sprintf("unsupported phone_home field %q", field))
		}
	}
}

func validatePowerState(errs *ValidationErrors, power *PowerState) {
	if power == nil {
		return
	}
	switch power.Mode {
	case PowerOff, Reboot, Halt:
	case "":
		errs.add("power_state.mode", "mode is required")
	default:
		errs.add("power_state.mode", "must be one of poweroff, reboot, halt")
	}
	if power.Delay != nil {
		if power.Delay.Now && power.Delay.Minutes != nil {
			errs.add("power_state.delay", "cannot define both now and minutes")
		}
		if !power.Delay.Now && power.Delay.Minutes == nil {
			errs.add("power_state.delay", "must define now or minutes")
		}
		if power.Delay.Minutes != nil && *power.Delay.Minutes < 0 {
			errs.add("power_state.delay", "minutes must be >= 0")
		}
	}
	if power.Timeout != nil && *power.Timeout < 0 {
		errs.add("power_state.timeout", "must be >= 0")
	}
	if power.Condition != nil {
		count := 0
		if power.Condition.Bool != nil {
			count++
		}
		if power.Condition.Command != nil {
			count++
		}
		if power.Condition.String != "" {
			count++
		}
		if count != 1 {
			errs.add("power_state.condition", "must define exactly one of Bool, Command, or String")
		}
		if power.Condition.Command != nil {
			validateCommands(errs, "power_state.condition", []Command{*power.Condition.Command}, false)
		}
	}
}

func validateGrowPart(errs *ValidationErrors, config *GrowPartConfig) {
	if config == nil {
		return
	}
	switch config.Mode {
	case "", GrowPartAuto, GrowPartGrowPart, GrowPartGPart, GrowPartOff:
	default:
		errs.add("growpart.mode", "unsupported mode")
	}
	for i, device := range config.Devices {
		if device == "" {
			errs.add(formatIndex("growpart.devices", i), "devices must be non-empty")
		}
	}
}

func validateResizeRootFS(errs *ValidationErrors, value *ResizeRootFSValue) {
	if value == nil {
		return
	}
	if value.Enabled != nil && value.Mode != "" {
		errs.add("resize_rootfs", "cannot define both Enabled and Mode")
	}
	if value.Enabled == nil && value.Mode == "" {
		errs.add("resize_rootfs", "must define Enabled or Mode")
	}
	if value.Mode != "" && value.Mode != ResizeRootFSNoBlock {
		errs.add("resize_rootfs", fmt.Sprintf("unsupported mode %q", value.Mode))
	}
}

func validateSizeOrValue(errs *ValidationErrors, path string, value *SizeOrValue, allowAuto bool) {
	if value == nil {
		return
	}
	count := 0
	if value.Auto {
		count++
	}
	if value.Integer != nil {
		count++
	}
	if value.String != "" {
		count++
	}
	if count != 1 {
		errs.add(path, "must define exactly one of Auto, Integer, or String")
		return
	}
	if value.Auto && !allowAuto {
		errs.add(path, "auto is not supported here")
	}
	if value.Integer != nil && *value.Integer < 0 {
		errs.add(path, "must be >= 0")
	}
	if value.String != "" && !sizePattern.MatchString(value.String) {
		errs.add(path, "must match the cloud-init size format")
	}
}

func validateSystemModules(errs *ValidationErrors, cfg Config) {
	switch cfg.ByobuByDefault {
	case "", "enable-system", "enable-user", "disable-system", "disable-user", "enable", "disable", "user", "system":
	default:
		errs.add("byobu_by_default", fmt.Sprintf("unsupported byobu mode %q", cfg.ByobuByDefault))
	}
	if cfg.Fan != nil && cfg.Fan.Config == "" {
		errs.add("fan.config", "config is required")
	}
	if cfg.GrubDpkg != nil {
		validateBoolOrStringDeprecated(errs, "grub_dpkg.grub-pc/install_devices_empty", cfg.GrubDpkg.GrubPCInstallDevicesEmpty)
	}
	if cfg.Updates != nil && cfg.Updates.Network != nil {
		if len(cfg.Updates.Network.When) == 0 {
			errs.add("updates.network.when", "when is required")
		}
		for i, when := range cfg.Updates.Network.When {
			switch when {
			case "boot-new-instance", "boot-legacy", "boot", "hotplug":
			default:
				errs.add(formatIndex("updates.network.when", i), fmt.Sprintf("unsupported value %q", when))
			}
		}
	}
	if cfg.Keyboard != nil && cfg.Keyboard.Layout == "" {
		errs.add("keyboard.layout", "layout is required")
	}
	if cfg.ResolvConf != nil && !valueTrue(cfg.ManageResolvConf) {
		errs.add("manage_resolv_conf", "must be true when resolv_conf is provided")
	}
	if cfg.RandomSeed != nil {
		if cfg.RandomSeed.Encoding != "" {
			if _, ok := validRandomSeedEncodings[cfg.RandomSeed.Encoding]; !ok {
				errs.add("random_seed.encoding", "unsupported encoding")
			}
		}
		if valueTrue(cfg.RandomSeed.CommandRequired) && len(cfg.RandomSeed.Command) == 0 {
			errs.add("random_seed.command", "command is required when command_required is true")
		}
		for i, arg := range cfg.RandomSeed.Command {
			if arg == "" {
				errs.add(formatIndex("random_seed.command", i), "command entries must be non-empty")
			}
		}
	}
	if cfg.Drivers != nil && cfg.Drivers.Nvidia != nil && cfg.Drivers.Nvidia.LicenseAccepted == nil {
		errs.add("drivers.nvidia.license-accepted", "license-accepted is required")
	}
	if cfg.Autoinstall != nil && cfg.Autoinstall.Version == 0 {
		errs.add("autoinstall.version", "version is required")
	}
	if cfg.Autoinstall != nil {
		validateInlineMapKeys(errs, "autoinstall", cfg.Autoinstall.Extra, "version")
	}
	validateNTP(errs, cfg.NTP)
	if cfg.WireGuard != nil {
		if len(cfg.WireGuard.Interfaces) == 0 {
			errs.add("wireguard.interfaces", "at least one interface is required")
		}
		for i, iface := range cfg.WireGuard.Interfaces {
			path := formatIndex("wireguard.interfaces", i)
			if iface.Name == "" {
				errs.add(path+".name", "name is required")
			}
			if iface.ConfigPath == "" {
				errs.add(path+".config_path", "config_path is required")
			}
			if iface.Content == "" {
				errs.add(path+".content", "content is required")
			}
		}
		validateUniqueStrings(errs, "wireguard.readinessprobe", cfg.WireGuard.ReadinessProbe)
	}
}

func validateIntegrationModules(errs *ValidationErrors, cfg Config) {
	if cfg.Ansible != nil {
		switch cfg.Ansible.InstallMethod {
		case AnsibleInstallMethodDistro, AnsibleInstallMethodPip:
		case "":
			errs.add("ansible.install_method", "install_method is required")
		default:
			errs.add("ansible.install_method", fmt.Sprintf("unsupported install_method %q", cfg.Ansible.InstallMethod))
		}
		if cfg.Ansible.PackageName == "" {
			errs.add("ansible.package_name", "package_name is required")
		}
		if cfg.Ansible.Pull != nil {
			if cfg.Ansible.Pull.URL == "" {
				errs.add("ansible.pull.url", "url is required")
			}
			if cfg.Ansible.Pull.PlaybookName == "" {
				errs.add("ansible.pull.playbook_name", "playbook_name is required")
			}
		}
		if cfg.Ansible.Galaxy != nil && len(cfg.Ansible.Galaxy.Actions) == 0 {
			errs.add("ansible.galaxy.actions", "actions is required")
		}
		if cfg.Ansible.SetupController != nil {
			if len(cfg.Ansible.SetupController.Repositories) == 0 && len(cfg.Ansible.SetupController.RunAnsible) == 0 {
				errs.add("ansible.setup_controller", "repositories or run_ansible is required")
			}
			for i, repo := range cfg.Ansible.SetupController.Repositories {
				path := formatIndex("ansible.setup_controller.repositories", i)
				if repo.Path == "" {
					errs.add(path+".path", "path is required")
				}
				if repo.Source == "" {
					errs.add(path+".source", "source is required")
				}
			}
			for i, run := range cfg.Ansible.SetupController.RunAnsible {
				path := formatIndex("ansible.setup_controller.run_ansible", i)
				validateInlineMapKeys(errs, path, run.Extra,
					"playbook_name", "playbook_dir", "become_password_file", "connection_password_file",
					"list_hosts", "syntax_check", "timeout", "vault_id", "vault_password_file",
					"background", "check", "diff", "module_path", "poll", "args", "extra_vars",
					"forks", "inventory", "scp_extra_args", "sftp_extra_args", "private_key",
					"connection", "module_name", "sleep", "tags", "skip_tags")
				if run.PlaybookDir == "" {
					errs.add(path+".playbook_dir", "playbook_dir is required")
				}
				if run.PlaybookName == "" {
					errs.add(path+".playbook_name", "playbook_name is required")
				}
				if run.Timeout != nil && *run.Timeout < 0 {
					errs.add(path+".timeout", "timeout must be greater than or equal to 0")
				}
				if run.Background != nil && *run.Background < 0 {
					errs.add(path+".background", "background must be greater than or equal to 0")
				}
				if run.Poll != nil && *run.Poll < 0 {
					errs.add(path+".poll", "poll must be greater than or equal to 0")
				}
				if run.Forks != nil && *run.Forks < 0 {
					errs.add(path+".forks", "forks must be greater than or equal to 0")
				}
			}
		}
	}
	if cfg.Chef != nil {
		if isEmptyChef(*cfg.Chef) {
			errs.add("chef", "must define at least one property")
		}
		if cfg.Chef.ServerURL == "" {
			errs.add("chef.server_url", "server_url is required")
		}
		if cfg.Chef.ValidationName == "" {
			errs.add("chef.validation_name", "validation_name is required")
		}
		switch cfg.Chef.InstallType {
		case "", ChefInstallTypePackages, ChefInstallTypeGems, ChefInstallTypeOmnibus:
		default:
			errs.add("chef.install_type", fmt.Sprintf("unsupported install_type %q", cfg.Chef.InstallType))
		}
		switch cfg.Chef.ChefLicense {
		case "", ChefLicenseAccept, ChefLicenseAcceptSilent, ChefLicenseAcceptNoPersist:
		default:
			errs.add("chef.chef_license", fmt.Sprintf("unsupported chef_license %q", cfg.Chef.ChefLicense))
		}
		validateUniqueStrings(errs, "chef.directories", cfg.Chef.Directories)
	}
	if cfg.Landscape != nil {
		if cfg.Landscape.Client == nil {
			errs.add("landscape.client", "client is required")
		} else {
			validateInlineMapKeys(errs, "landscape.client", cfg.Landscape.Client.Extra,
				"url", "ping_url", "data_path", "log_level", "computer_title", "account_name",
				"registration_key", "tags", "http_proxy", "https_proxy")
			if cfg.Landscape.Client.AccountName == "" {
				errs.add("landscape.client.account_name", "account_name is required")
			}
			if cfg.Landscape.Client.ComputerTitle == "" {
				errs.add("landscape.client.computer_title", "computer_title is required")
			}
			switch cfg.Landscape.Client.LogLevel {
			case "", "debug", "info", "warning", "error", "critical":
			default:
				errs.add("landscape.client.log_level", fmt.Sprintf("unsupported log_level %q", cfg.Landscape.Client.LogLevel))
			}
			if cfg.Landscape.Client.Tags != "" && !landscapeTagsPattern.MatchString(cfg.Landscape.Client.Tags) {
				errs.add("landscape.client.tags", "tags must be a comma-separated list of letters, digits, hyphens, or underscores")
			}
		}
	}
	if cfg.LXD != nil {
		if cfg.LXD.Init == nil && cfg.LXD.Bridge == nil && cfg.LXD.Preseed == "" {
			errs.add("lxd", "must define init, bridge, or preseed")
		}
		if cfg.LXD.Preseed != "" && (cfg.LXD.Init != nil || cfg.LXD.Bridge != nil) {
			errs.add("lxd", "preseed cannot be combined with init or bridge")
		}
		if cfg.LXD.Init != nil {
			switch cfg.LXD.Init.StorageBackend {
			case "", LXDStorageBackendZFS, LXDStorageBackendDir, LXDStorageBackendLVM, LXDStorageBackendBTRFS:
			default:
				errs.add("lxd.init.storage_backend", fmt.Sprintf("unsupported storage_backend %q", cfg.LXD.Init.StorageBackend))
			}
		}
		if cfg.LXD.Bridge != nil {
			switch cfg.LXD.Bridge.Mode {
			case LXDBridgeModeNone, LXDBridgeModeExisting, LXDBridgeModeNew:
			case "":
				errs.add("lxd.bridge.mode", "mode is required")
			default:
				errs.add("lxd.bridge.mode", fmt.Sprintf("unsupported mode %q", cfg.LXD.Bridge.Mode))
			}
			if cfg.LXD.Bridge.MTU != nil && *cfg.LXD.Bridge.MTU < -1 {
				errs.add("lxd.bridge.mtu", "mtu must be greater than or equal to -1")
			}
			if cfg.LXD.Bridge.IPv4Address != "" && cfg.LXD.Bridge.IPv4Netmask == nil {
				errs.add("lxd.bridge.ipv4_netmask", "ipv4_netmask is required when ipv4_address is set")
			}
			if cfg.LXD.Bridge.IPv6Address != "" && cfg.LXD.Bridge.IPv6Netmask == nil {
				errs.add("lxd.bridge.ipv6_netmask", "ipv6_netmask is required when ipv6_address is set")
			}
		}
	}
	if cfg.MCollective != nil && cfg.MCollective.Conf != nil {
		validateInlineMapKeys(errs, "mcollective.conf", cfg.MCollective.Conf.Extra, "public-cert", "private-cert")
		validateScalarMapValues(errs, "mcollective.conf", cfg.MCollective.Conf.Extra)
	}
	if cfg.Puppet != nil {
		switch cfg.Puppet.InstallType {
		case "", PuppetInstallTypePackages, PuppetInstallTypeAIO:
		default:
			errs.add("puppet.install_type", fmt.Sprintf("unsupported install_type %q", cfg.Puppet.InstallType))
		}
	}
	if cfg.RHSubscription != nil {
		hasUserPass := cfg.RHSubscription.Username != "" || cfg.RHSubscription.Password != ""
		hasActivation := cfg.RHSubscription.ActivationKey != "" || cfg.RHSubscription.Org != ""
		if hasUserPass && hasActivation {
			errs.add("rh_subscription", "username/password cannot be combined with activation-key/org")
		}
		if (cfg.RHSubscription.Username == "") != (cfg.RHSubscription.Password == "") {
			errs.add("rh_subscription", "username and password must be provided together")
		}
		if (cfg.RHSubscription.ActivationKey == "") != (cfg.RHSubscription.Org == "") {
			errs.add("rh_subscription", "activation-key and org must be provided together")
		}
		if cfg.RHSubscription.ServiceLevel != "" && !valueTrue(cfg.RHSubscription.AutoAttach) {
			errs.add("rh_subscription.service-level", "service-level requires auto-attach to be true")
		}
	}
	if cfg.Rsyslog != nil {
		if _, valid := validateStringOrStringSlice(errs, "rsyslog.service_reload_command", cfg.Rsyslog.ServiceReloadCommand); valid &&
			cfg.Rsyslog.ServiceReloadCommand != nil &&
			cfg.Rsyslog.ServiceReloadCommand.String != "" &&
			cfg.Rsyslog.ServiceReloadCommand.String != "auto" {
			errs.add("rsyslog.service_reload_command", "string value must be auto")
		}
		for i, entry := range cfg.Rsyslog.Configs {
			if entry.Content == "" {
				errs.add(formatIndex("rsyslog.configs", i)+".content", "content is required")
			}
		}
		for name, remote := range cfg.Rsyslog.Remotes {
			if err := validateRsyslogRemote(remote); err != nil {
				errs.add("rsyslog.remotes."+name, err.Error())
			}
		}
		validateUniqueStrings(errs, "rsyslog.packages", cfg.Rsyslog.Packages)
	}
	if cfg.UbuntuPro != nil && cfg.UbuntuPro.Config != nil {
		validateInlineMapKeys(errs, "ubuntu_pro.config", cfg.UbuntuPro.Config.Extra,
			"http_proxy", "https_proxy", "global_apt_http_proxy", "global_apt_https_proxy",
			"ua_apt_http_proxy", "ua_apt_https_proxy")
		validateNullableHTTPURI(errs, "ubuntu_pro.config.http_proxy", cfg.UbuntuPro.Config.HTTPProxy)
		validateNullableHTTPURI(errs, "ubuntu_pro.config.https_proxy", cfg.UbuntuPro.Config.HTTPSProxy)
		validateNullableHTTPURI(errs, "ubuntu_pro.config.global_apt_http_proxy", cfg.UbuntuPro.Config.GlobalAPTHTTPProxy)
		validateNullableHTTPURI(errs, "ubuntu_pro.config.global_apt_https_proxy", cfg.UbuntuPro.Config.GlobalAPTHTTPSProxy)
		validateNullableHTTPURI(errs, "ubuntu_pro.config.ua_apt_http_proxy", cfg.UbuntuPro.Config.UAAPTHTTPProxy)
		validateNullableHTTPURI(errs, "ubuntu_pro.config.ua_apt_https_proxy", cfg.UbuntuPro.Config.UAAPTHTTPSProxy)
	}
}

func validateRsyslogRemote(line string) error {
	if line == "" {
		return fmt.Errorf("remote target is required")
	}

	data := line
	if idx := strings.Index(data, "#"); idx >= 0 {
		data = data[:idx]
	}

	toks := strings.Fields(strings.TrimSpace(data))
	var hostPort string
	switch len(toks) {
	case 1:
		hostPort = toks[0]
	case 2:
		hostPort = toks[1]
	default:
		return fmt.Errorf("invalid remote target")
	}

	match := rsyslogRemotePattern.FindStringSubmatch(hostPort)
	if match == nil {
		return fmt.Errorf("invalid remote target")
	}

	addr := match[2]
	if addr == "" {
		addr = match[3]
	}
	if addr == "" {
		return fmt.Errorf("address is required")
	}

	return nil
}

func validateNTP(errs *ValidationErrors, cfg *NTPConfig) {
	if cfg == nil {
		return
	}
	for key, value := range *cfg {
		switch key {
		case "pools", "servers", "peers":
			items := validateStringListValue(errs, "ntp."+key, value, 0)
			validateUniqueStrings(errs, "ntp."+key, items)
			for i, item := range items {
				if item != "" && !validNTPHostname(item) {
					errs.add(formatIndex("ntp."+key, i), "must be a valid hostname")
				}
			}
		case "allow":
			items := validateStringListValue(errs, "ntp.allow", value, 0)
			validateUniqueStrings(errs, "ntp.allow", items)
		case "ntp_client":
			client, ok := value.(string)
			if !ok {
				errs.add("ntp.ntp_client", "must be a string")
			} else if _, ok := validNTPClients[client]; !ok {
				errs.add("ntp.ntp_client", fmt.Sprintf("unsupported ntp client %q", client))
			}
		case "enabled":
			if _, ok := value.(bool); !ok {
				errs.add("ntp.enabled", "must be a boolean")
			}
		case "config":
			validateNTPClientConfig(errs, "ntp.config", value)
		default:
			errs.add("ntp."+key, "unsupported property")
		}
	}
}

func validateNTPClientConfig(errs *ValidationErrors, path string, value any) {
	config, ok := asStringKeyMap(value)
	if !ok {
		errs.add(path, "must be an object")
		return
	}
	if len(config) == 0 {
		errs.add(path, "must define at least one property")
		return
	}
	for key, item := range config {
		switch key {
		case "confpath", "check_exe", "service_name", "template":
			if _, ok := item.(string); !ok {
				errs.add(path+"."+key, "must be a string")
			}
		case "packages":
			items := validateStringListValue(errs, path+".packages", item, 0)
			validateUniqueStrings(errs, path+".packages", items)
		default:
			errs.add(path+"."+key, "unsupported property")
		}
	}
}

func validateNullableHTTPURI(errs *ValidationErrors, path string, value *NullableString) {
	if value == nil {
		return
	}
	if value.Null && value.String != nil {
		errs.add(path, "must define either String or Null")
		return
	}
	if value.Null || value.String == nil {
		return
	}
	if !validateURI(*value.String) {
		errs.add(path, "must be an absolute URI")
		return
	}
	parsed, err := url.Parse(*value.String)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		errs.add(path, "must use http or https")
	}
}

func isEmptyChef(cfg ChefConfig) bool {
	return len(cfg.Directories) == 0 &&
		cfg.ConfigPath == "" &&
		cfg.ValidationCert == "" &&
		cfg.ValidationKey == "" &&
		cfg.FirstbootPath == "" &&
		cfg.Exec == nil &&
		cfg.ClientKey == "" &&
		cfg.EncryptedDataBagSecret == "" &&
		cfg.Environment == "" &&
		cfg.FileBackupPath == "" &&
		cfg.FileCachePath == "" &&
		cfg.JSONAttribs == "" &&
		cfg.LogLevel == "" &&
		cfg.LogLocation == "" &&
		cfg.NodeName == "" &&
		cfg.OmnibusURL == "" &&
		cfg.OmnibusURLRetries == nil &&
		cfg.OmnibusVersion == "" &&
		cfg.PIDFile == "" &&
		cfg.ServerURL == "" &&
		cfg.ShowTime == nil &&
		cfg.SSLVerifyMode == "" &&
		cfg.ValidationName == "" &&
		cfg.ForceInstall == nil &&
		len(cfg.InitialAttributes) == 0 &&
		cfg.InstallType == "" &&
		len(cfg.RunList) == 0 &&
		cfg.ChefLicense == ""
}
