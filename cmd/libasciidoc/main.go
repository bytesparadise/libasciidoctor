package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bytesparadise/libasciidoc/renderer"
	"github.com/bytesparadise/libasciidoc/renderer/html5"
	"github.com/bytesparadise/libasciidoc/types"

	"github.com/bytesparadise/libasciidoc/parser"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := newRootCmd()
	rootCmd.AddCommand(versionCmd)
	rootCmd.SetHelpCommand(helpCommand)
	// rootCmd.SetHelpTemplate(helpTemplate)
	// rootCmd.PersistentFlags().BoolP("help", "h", false, "Print usage")
	// rootCmd.PersistentFlags().MarkShorthandDeprecated("help", "please use --help")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var helpTemplate = `
{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`

func newRootCmd() *cobra.Command {
	var source string
	rootCmd := &cobra.Command{
		Use:   "libasciidoc",
		Short: "libasciidoc is a tool to generate an html output from an asciidoc source",
		// Long: `A Fast and Flexible Static Site Generator built with
		// 			  love by spf13 and friends in Go.
		// 			  Complete documentation is available at http://hugo.spf13.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flag("source").Value.String() == "" {
				fmt.Println("flag 'source' is required")
				os.Exit(1)
			}
			source := cmd.Flag("source").Value.String()
			b, err := ioutil.ReadFile(source)
			if err != nil {
				fmt.Printf("failed to read the source file: %v\n", err)
				os.Exit(1)
			}
			doc, err := parser.Parse(source, b)
			if err != nil {
				fmt.Printf("failed to parse the source file: %v\n", err)
				os.Exit(1)
			}
			buff := bytes.NewBuffer(nil)
			actualDocument := doc.(types.Document)
			rendererCtx := renderer.Wrap(context.Background(), actualDocument)
			_, err = html5.Render(rendererCtx, buff)
			fmt.Printf("%s\n", buff.String())
			return nil
		},
	}
	flags := rootCmd.Flags()
	flags.StringVarP(&source, "source", "s", "", "the path to the asciidoc source to process")
	return rootCmd
}

var helpCommand = &cobra.Command{
	Use:               "help [command]",
	Short:             "Help about the command",
	PersistentPreRun:  func(cmd *cobra.Command, args []string) {},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {},
	RunE: func(c *cobra.Command, args []string) error {
		cmd, args, e := c.Root().Find(args)
		if cmd == nil || e != nil || len(args) > 0 {
			return errors.Errorf("unknown help topic: %v", strings.Join(args, " "))
		}

		helpFunc := cmd.HelpFunc()
		helpFunc(cmd, args)
		return nil
	},
}
