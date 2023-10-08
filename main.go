package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Date    = "unknown"
)

type RaycastManifest struct {
	Name        string           `json:"name"`
	Author      string           `json:"author"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Commands    []RaycastCommand `json:"commands"`
}

type RaycastCommand struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Mode        string `json:"mode"`
	Arguments   []struct {
		Name        string       `json:"name"`
		Type        ArgumentType `json:"type"`
		Placeholder string       `json:"placeholder"`
		Required    bool         `json:"required"`
	} `json:"arguments"`
}

type ArgumentType string

const (
	ArgumentTypeText     ArgumentType = "text"
	ArgumentTypePassword ArgumentType = "password"
)

func NewCmdCommand(manifest RaycastManifest, command RaycastCommand) *cobra.Command {
	flags := struct {
		Print bool
		Copy  bool
	}{}

	cmd := &cobra.Command{
		Short: command.Title,
		Args:  cobra.NoArgs,
		Long:  command.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			deeplinkArgs := make(map[string]string)
			for i, argument := range command.Arguments {
				if len(args) > i {
					deeplinkArgs[argument.Name] = args[i]
				} else if argument.Required {
					return fmt.Errorf("missing required argument: %s", argument.Name)
				}
			}

			query := make(url.Values)
			if command.Mode != "view" {
				query.Set("launchType", "background")
			}

			if len(deeplinkArgs) > 0 {
				jsonArgs, err := json.Marshal(deeplinkArgs)
				if err != nil {
					return err
				}

				query.Set("arguments", string(jsonArgs))
			}

			if !isatty.IsTerminal(os.Stdin.Fd()) {
				context, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}

				if len(context) > 0 {
					query.Set("context", string(context))
				}
			}

			deeplink := fmt.Sprintf("raycast://extensions/%s/%s/%s", manifest.Author, manifest.Name, command.Name)
			deeplinkUrl, err := url.Parse(deeplink)
			if err != nil {
				return err
			}
			deeplinkUrl.RawQuery = query.Encode()

			if flags.Print {
				fmt.Print(deeplinkUrl.String())
				return nil
			}

			if flags.Copy {
				copyCmd := exec.Command("pbcopy")
				copyCmd.Stdin = strings.NewReader(deeplinkUrl.String())

				return copyCmd.Run()
			}

			openCmd := exec.Command("open", deeplinkUrl.String())
			return openCmd.Run()
		},
	}

	cmd.Flags().BoolVar(&flags.Print, "print", false, "Print the deeplink to the clipboard instead of opening it")
	cmd.Flags().BoolVar(&flags.Copy, "copy", false, "Copy the deeplink to the clipboard instead of opening it")

	cmd.MarkFlagsMutuallyExclusive("print", "copy")

	use := command.Name
	minArgs, maxArgs := 0, 0
	for _, argument := range command.Arguments {
		if argument.Required {
			use += fmt.Sprintf(" <%s>", argument.Name)
			minArgs++
			maxArgs++
		} else {
			use += fmt.Sprintf(" [%s]", argument.Name)
			maxArgs++
		}
	}

	cmd.Use = use
	cmd.Args = cobra.MatchAll(cobra.MinimumNArgs(minArgs), cobra.MaximumNArgs(maxArgs))

	return cmd

}

func NewCmdExtension(manifest RaycastManifest) *cobra.Command {
	cmd := &cobra.Command{
		Use:   manifest.Name,
		Short: manifest.Title,
		Long:  manifest.Description,
	}

	for _, command := range manifest.Commands {
		cmd.AddCommand(NewCmdCommand(manifest, command))
	}

	return cmd
}

func NewCmdRoot() (*cobra.Command, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	rootCmd := &cobra.Command{
		Use:     "ray",
		Short:   "A command line interface for Raycast",
		Version: fmt.Sprintf("%s (%s)", Version, Date),
	}

	extensionRoot := filepath.Join(homedir, ".config", "raycast", "extensions")
	extensionManifests, err := filepath.Glob(filepath.Join(extensionRoot, "*", "package.json"))
	if err != nil {
		return nil, err
	}

	for _, manifestPath := range extensionManifests {
		f, err := os.Open(manifestPath)
		if err != nil {
			return nil, err
		}

		var manifest RaycastManifest
		if err := json.NewDecoder(f).Decode(&manifest); err != nil {
			return nil, err
		}

		extensionCmd := NewCmdExtension(manifest)
		rootCmd.AddCommand(extensionCmd)
	}

	return rootCmd, nil
}

func main() {
	rootCmd, err := NewCmdRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
