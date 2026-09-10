package cloudinit

import "regexp"

var (
	validWriteFileEncodings = map[string]struct{}{
		"gz":          {},
		"gzip":        {},
		"gz+base64":   {},
		"gzip+base64": {},
		"gz+b64":      {},
		"gzip+b64":    {},
		"b64":         {},
		"base64":      {},
		"text/plain":  {},
	}
	validRandomSeedEncodings = map[string]struct{}{
		"raw":    {},
		"base64": {},
		"b64":    {},
		"gzip":   {},
		"gz":     {},
	}
	validPhoneHomeFields = map[PhoneHomeField]struct{}{
		PhoneHomePubKeyRSA:     {},
		PhoneHomePubKeyECDSA:   {},
		PhoneHomePubKeyED25519: {},
		PhoneHomeInstanceID:    {},
		PhoneHomeHostname:      {},
		PhoneHomeFQDN:          {},
	}
	sizePattern = regexp.MustCompile(`^([0-9]+)?\.?[0-9]+[BKMGT]$`)
)

// Validate checks the config for schema and runtime-shape errors.
func (c Config) Validate() error {
	var errs ValidationErrors

	validateManageEtcHosts(&errs, c.ManageEtcHosts)
	validateLocale(&errs, c.Locale)
	validatePackages(&errs, c.Packages)
	validateAPT(&errs, c.APT)
	validateAPKRepos(&errs, c.APKRepos)
	validateAPTPipelining(&errs, c.APTPipelining)
	validateSnapConfig(&errs, c.Snap)
	validateYumRepos(&errs, c.YumRepoDir, c.YumRepos)
	validateZypper(&errs, c.Zypper)
	validateCommands(&errs, "bootcmd", c.BootCmd, false)
	validateCommands(&errs, "runcmd", c.RunCmd, true)
	validateWriteFiles(&errs, c.WriteFiles)
	validateGroups(&errs, c.Groups)
	validateUserEntry(&errs, "user", c.User)
	for i := range c.Users {
		validateUserEntry(&errs, formatIndex("users", i), &c.Users[i])
	}
	validateChPasswd(&errs, c.ChPasswd)
	validateSSH(&errs, c)
	validateCACerts(&errs, c.CACerts)
	validateMounts(&errs, c.Mounts, c.MountDefaultFields, c.Swap)
	validateDiskSetup(&errs, c.DeviceAliases, c.DiskSetup, c.FSSetup)
	validatePhoneHome(&errs, c.PhoneHome)
	validatePowerState(&errs, c.PowerState)
	validateGrowPart(&errs, c.GrowPart)
	validateResizeRootFS(&errs, c.ResizeRootFS)
	validateSystemModules(&errs, c)
	validateIntegrationModules(&errs, c)
	validateConfigExtra(&errs, c.Extra)

	return errs.Err()
}

func validateConfigExtra(errs *ValidationErrors, extra RawObject) {
	if extra == nil {
		return
	}
	modules := SupportedModuleSet()
	for name := range extra {
		if name == "" {
			errs.add("extra", "extension names must be non-empty")
			continue
		}
		if _, supported := modules[name]; supported {
			errs.add("extra."+name, "conflicts with a supported module")
		}
	}
}
