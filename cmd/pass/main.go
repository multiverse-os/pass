package main

import (
	"os"

	cli "github.com/multiverse-os/cli"
)

func main() {
	cmd := cli.New(&cli.CLI{
		Name:    "pass",
		Version: cli.Version{Major: 0, Minor: 1, Patch: 0},
		Usage:   "An implementation of Unix `pass` with additional features abstracted ontop of the original protocol",
		Commands: []cli.Command{
			cli.Command{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Initialize the password store in ~/.password-store with the specified GPG identity",
				Flags: []cli.Flag{
					cli.Flag{
						Name:  "gpg",
						Usage: "Specify the gpg identity to initialize the pass-store with",
					},
				},
			},
			cli.Command{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List passwords",
			},
			cli.Command{
				Name:    "find",
				Aliases: []string{"f"},
				Usage:   "List passwod that match pass-names",
			},
			cli.Command{
				Name:    "show",
				Aliases: []string{"s"},
				Usage:   "Show passwords in the password-store",
			},
			cli.Command{
				Name:    "grep",
				Aliases: []string{"gr"},
				Usage:   "Search for password files containing search-string when decrypted",
			},
			cli.Command{
				Name:    "insert",
				Aliases: []string{"in"},
				Usage:   "Insert a password into the password store",
			},
			cli.Command{
				Name:    "edit",
				Aliases: []string{"e"},
				Usage:   "Insert a new password or edit ane xisting password using an editor",
			},
			cli.Command{
				Name:    "generate",
				Aliases: []string{"gen", "g"},
				Usage:   "Generate a new password of pass-legnth (or 32 if unspecified) with optionally no sybmols",
				Flags: []cli.Flag{
					cli.Flag{
						Name:    "no-symbols",
						Aliases: []string{"n"},
						Usage:   "Do not use symbols when generating the new password",
					},
					cli.Flag{
						Name:    "clip",
						Aliases: []string{"c"},
						Usage:   "Optionally copy the password to the clipboard and clear after 45 seconds",
					},
				},
			},
			cli.Command{
				Name:    "remove",
				Aliases: []string{"rm"},
				Usage:   "Remove existing password or directory, optionally forcefully",
			},
			cli.Command{
				Name:    "move",
				Aliases: []string{"mv"},
				Usage:   "Move existing password or directory, optionally forcefully, selectively re-encrypting",
			},
			cli.Command{
				Name:    "copy",
				Aliases: []string{"cp"},
				Usage:   "Copies old-path to new-path, optionally forcefully, selectively re-encrypting",
			},
			cli.Command{
				Name:  "git",
				Usage: "If the password store is a git repository, execute a git command",
			},
		},
		DefaultAction: func(context *cli.Context) error {
			// TODO: List all passwords using tree
			return nil
		},
	})

	cmd.Run(os.Args)
}
