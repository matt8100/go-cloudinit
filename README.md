# go-cloudinit

`go-cloudinit` is a Go library for generating `#cloud-config` YAML documents programmatically.

The package is intended for other Go programs that need to produce cloud-init configuration on the fly without templating strings by hand. Rendering is pure Go. Validation is also pure Go at runtime, and the test suite can optionally verify rendered fixtures with `cloud-init schema`.

## Status

This repository starts with typed support for a broad initial cloud-config surface:

- host identity and `/etc/hosts` management
- locale and timezone
- package update/upgrade/install
- apt/apk/yum/zypper/snap configuration
- `bootcmd` and `runcmd`
- `write_files`
- users and groups
- SSH configuration
- CA certificates
- mounts, swap, disk setup, and wireguard
- phone-home, power-state, growpart, resize_rootfs, random seeding, keyboard, and resolv.conf
- ansible, chef, puppet, salt, landscape, lxd, rsyslog, spacewalk, and subscription-style integrations
- Ubuntu autoinstall, Ubuntu drivers, and Ubuntu Pro
- final_message

## Example

```go
package main

import (
	"fmt"

	"go-cloudinit"
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

## Integration Testing

If `cloud-init` is installed, `go test ./...` runs an integration test that validates a rendered fixture with:

```sh
cloud-init schema -c <rendered-file> -t cloud-config
```
