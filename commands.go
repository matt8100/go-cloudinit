package cloudinit

import "fmt"

// Command represents one cloud-init command item.
type Command struct {
	String string
	Args   []string
	Null   bool
}

// ShellCommand returns a command rendered as a shell string.
func ShellCommand(command string) Command {
	return Command{String: command}
}

// ExecCommand returns a command rendered as an argument list.
func ExecCommand(args ...string) Command {
	return Command{Args: args}
}

// NullCommand returns a YAML null command entry.
func NullCommand() Command {
	return Command{Null: true}
}

// MarshalYAML implements yaml.Marshaler.
func (c Command) MarshalYAML() (any, error) {
	switch {
	case c.Null:
		return nil, nil
	case c.String != "" && len(c.Args) == 0:
		return c.String, nil
	case c.String == "" && len(c.Args) > 0:
		return c.Args, nil
	default:
		return nil, fmt.Errorf("command must define exactly one of String, Args, or Null")
	}
}
