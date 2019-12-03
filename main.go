// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:generate $GOBIN/mdtogo docs/tutorials internal/docs/generated/tutorials --full=true --license=none
//go:generate $GOBIN/mdtogo docs/commands internal/docs/generated/commands --license=none
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/GoogleContainerTools/kpt/commands"
	"github.com/GoogleContainerTools/kpt/internal/cmdcomplete"
	docs "github.com/GoogleContainerTools/kpt/internal/docs/generated/commands"
	"github.com/GoogleContainerTools/kpt/internal/util/cmdutil"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:     "kpt",
		Short:   docs.KptShort,
		Long:    docs.KptLong,
		Example: docs.KptExamples,
	}

	// help and documentation
	cmd.InitDefaultHelpCmd()
	cmd.AddCommand(commands.GetCommands("kpt")...)

	// enable stack traces
	cmd.PersistentFlags().BoolVar(&cmdutil.StackOnError, "stack-trace", false,
		"print a stack-trace on failure")

	// exit on an error
	cmdutil.ExitOnError = true

	// bash shell completion passes the command name as the first argument
	// do this after configuring cmd so it has all the subcommands
	if len(os.Args) > 1 {
		// use the base name in case kpt is called with an absolute path
		name := filepath.Base(os.Args[1])
		if name == "kpt" {
			// complete calls kpt with itself as an argument
			cmdcomplete.Complete(cmd).Complete("kpt")
			os.Exit(0)
		}
	}

	if val, found := os.LookupEnv("KPT_NO_PAGER_HELP"); !found || val != "1" {
		// use a pager for printing tutorials
		e, found := os.LookupEnv("PAGER")
		var err error
		if !found {
			e, err = exec.LookPath("pager")
			if err != nil {
				e, err = exec.LookPath("less")
				if err != nil {
					return
				}
			}
		}
		for i := range cmd.Commands() {
			replace(cmd.Commands()[i], e)
		}
	}

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func replace(c *cobra.Command, e string) {
	for i := range c.Commands() {
		replace(c.Commands()[i], e)
	}
	c.SetHelpFunc(newHelp(e, c))
}

func newHelp(e string, c *cobra.Command) func(command *cobra.Command, strings []string) {
	fn := c.HelpFunc()
	return func(command *cobra.Command, strings []string) {
		b := &bytes.Buffer{}
		pager := exec.Command(e)
		pager.Stdin = b
		pager.Stdout = c.OutOrStdout()
		c.SetOut(b)
		fn(command, strings)
		if err := pager.Run(); err != nil {
			fmt.Fprintf(c.ErrOrStderr(), "%v", err)
			os.Exit(1)
		}
	}
}
