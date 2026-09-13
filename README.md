# go-cloudinit

A Go library for building, validating, and rendering `#cloud-config` documents.

It gives you a typed `Config` struct that mirrors cloud-init's schema, so programs can produce
cloud-config on the fly instead of templating YAML strings by hand. Rendering and validation are
both pure Go — no external tooling is needed at runtime.

## Contents

- [Installation](#installation)
- [Quick start](#quick-start)
- [Core API](#core-api)
- [Working with optional and union values](#working-with-optional-and-union-values)
- [Validation](#validation)
- [Supported modules](#supported-modules)
- [Development](#development)
- [Project layout](#project-layout)

## Installation

```sh
go get github.com/matt8100/go-cloudinit
```

```go
import "github.com/matt8100/go-cloudinit"   // package name is cloudinit
```

Requires Go 1.24.4 or newer. The only dependency is `gopkg.in/yaml.v3`.

## Quick start

```go
package main

import (
	"fmt"

	"github.com/matt8100/go-cloudinit"
)

func main() {
	cfg := cloudinit.Config{
		Hostname:      "web-01",
		PackageUpdate: cloudinit.Bool(true),
		Packages: []cloudinit.PackageEntry{
			cloudinit.Package("curl"),
			cloudinit.VersionedPackage("jq", "1.6-2"),
		},
		WriteFiles: []cloudinit.WriteFile{
			{
				Path:    "/etc/motd",
				Content: "managed by go-cloudinit\n",
			},
		},
		RunCmd: []cloudinit.Command{
			cloudinit.ShellCommand("echo ready"),
		},
	}

	out, err := cfg.Render()
	if err != nil {
		panic(err)
	}

	fmt.Print(string(out))
}
```

Output:

```yaml
#cloud-config
hostname: web-01
package_update: true
packages:
    - curl
    - - jq
      - 1.6-2
runcmd:
    - echo ready
write_files:
    - path: /etc/motd
      content: |
        managed by go-cloudinit
```

## Core API

| Symbol | Purpose |
| --- | --- |
| `Config` | The whole document. Every cloud-config module is a field. |
| `(Config) Render() ([]byte, error)` | Validates, then returns the full document including the `#cloud-config` header. |
| `(Config) String() (string, error)` | Same as `Render`, as a `string`. |
| `(Config) Validate() error` | Validates without rendering. Returns `ValidationErrors`. |
| `(Config) MarshalYAML()` | Validates, then marshals — so embedding a `Config` in a larger YAML document stays safe. |
| `Header` | The `"#cloud-config\n"` prefix constant. |
| `TrimmedYAML([]byte) string` | Strips surrounding whitespace from rendered YAML; handy in tests. |
| `SupportedModules` / `SupportedModuleSet()` | The list (and set form) of top-level keys this package models. |

`Render` always validates first, so a successful render is always a schema-shaped document.

## Working with optional and union values

Cloud-init distinguishes "unset" from "set to false" or "set to zero", and many keys accept more
than one YAML type. The package handles both with small constructors rather than raw literals.

**Pointer helpers** for optional scalars — `Bool`, `Int`, `Int64`, `StringPtr`:

```go
cfg.PackageUpgrade = cloudinit.Bool(false) // emits `package_upgrade: false`, not omitted
```

**Union helpers** for keys that accept multiple shapes:

| Area | Constructors |
| --- | --- |
| Commands | `ShellCommand`, `ExecCommand`, `NullCommand` |
| Packages | `Package`, `VersionedPackage`, `APTPackages`, `SnapPackages` |
| Users and groups | `Group`, `UserReference`, `DefaultUserReference` |
| Scalars | `BoolValue`, `StringValue`, `IntValue`, `IntBoolValue`, `IntStringValue` |
| Nullable strings | `StringSetting`, `NullSetting` |
| Hosts and locale | `ManageEtcHostsEnabled`, `ManageEtcHostsDisabled`, `ManageEtcHostsLocalhostValue`, `Locale`, `LocaleEnabled` |
| Storage | `AutoSize`, `BytesSize`, `HumanSize`, `RemoveLayout`, `SimpleLayout` |
| Power state | `DelayNow`, `DelayMinutes`, `ConditionAlways`, `ConditionCommand`, `ConditionString` |
| Misc | `ResizeRootFSEnabled`, `ResizeRootFSBackground`, `PhoneHomeAllFields`, `RsyslogContent`, `RsyslogFile` |

For example, `ShellCommand("echo ready")` renders as a string and `ExecCommand("echo", "ready")`
renders as a list, matching how cloud-init reads each form.

## Validation

`Validate` collects *all* problems rather than stopping at the first, and reports each one with a
precise field path:

```go
cfg := cloudinit.Config{
	Packages:   []cloudinit.PackageEntry{cloudinit.Package("")},
	WriteFiles: []cloudinit.WriteFile{{Path: "/etc/motd", Encoding: "rot13"}},
}

if err := cfg.Validate(); err != nil {
	fmt.Println(err)
}
```

```
packages[0]: package name is required
write_files[0].encoding: unsupported encoding
```

To inspect failures individually, unwrap to the typed collection:

```go
var errs cloudinit.ValidationErrors
if errors.As(err, &errs) {
	for _, e := range errs {
		fmt.Println(e.Path, "->", e.Message)
	}
}
```

`ValidationErrors` is a `[]ValidationError`, each with a `Path` and `Message`.

## Supported modules

The package models a broad slice of the cloud-config surface. `SupportedModules` is the
authoritative list; by area:

- **Host identity** — `hostname`, `fqdn`, `prefer_fqdn_over_hostname`, `preserve_hostname`,
  `create_hostname_file`, `manage_etc_hosts`
- **Localization** — `locale`, `locale_configfile`, `timezone`, `keyboard`
- **Packages** — `package_update`, `package_upgrade`, `package_reboot_if_required`, `packages`,
  `apt`, `apt_pipelining`, `apk_repos`, `snap`, `yum_repos`, `yum_repo_dir`, `zypper`
- **Commands and files** — `bootcmd`, `runcmd`, `write_files`
- **Users and auth** — `users`, `user`, `groups`, `password`, `chpasswd`, `disable_root`,
  `disable_root_opts`
- **SSH** — `ssh`, `ssh_keys`, `ssh_authorized_keys`, `ssh_import_id`, `ssh_pwauth`,
  `ssh_genkeytypes`, `ssh_deletekeys`, `ssh_publish_hostkeys`, `ssh_quiet_keygen`,
  `allow_public_ssh_keys`, `no_ssh_fingerprints`, `authkey_hash`, `ssh_key_console_blacklist`,
  `ssh_fp_console_blacklist`
- **Storage** — `mounts`, `mount_default_fields`, `swap`, `disk_setup`, `fs_setup`,
  `device_aliases`, `growpart`, `resize_rootfs`
- **System and lifecycle** — `ca_certs`, `ntp`, `resolv_conf`, `manage_resolv_conf`, `random_seed`,
  `phone_home`, `power_state`, `final_message`, `updates`, `fan`, `grub_dpkg`, `byobu_by_default`,
  `disable_ec2_metadata`, `wireguard`
- **Configuration management** — `ansible`, `chef`, `puppet`, `salt_minion`, `mcollective`,
  `landscape`, `lxd`, `rsyslog`, `spacewalk`, `rh_subscription`
- **Ubuntu-specific** — `autoinstall`, `drivers`, `ubuntu_pro`

## Development

```sh
gofmt -w *.go      # format
go test ./...      # build and verify — the main development check
go test ./... -run TestName
```

This repo ships no binary, so `go test ./...` is the compile-and-verify command.

### Integration testing

When the `cloud-init` binary is on `PATH`, `go test ./...` additionally renders fixtures and checks
them against the real schema:

```sh
cloud-init schema -c <rendered-file> -t cloud-config
```

Without the binary, those tests skip. You can also run the check by hand on any rendered document.

## Project layout

| Path | Contents |
| --- | --- |
| `config.go` | `Config` struct and `SupportedModules` |
| `render.go` | Rendering and YAML marshaling |
| `validate.go`, `validate_modules.go` | Validation entry point and per-module rules |
| `errors.go` | `ValidationError` / `ValidationErrors` |
| `helpers.go` | Pointer helpers and union value types |
| `commands.go`, `files.go`, `users.go`, `ssh.go`, `certs.go` | Command, file, user, and auth types |
| `packages.go`, `storage.go`, `system.go`, `integrations.go` | Module types by feature area |
| `cloudinit_test.go` | Unit and regression tests |
| `integration_test.go` | Optional `cloud-init schema` verification |

See `AGENTS.md` for contribution conventions.
