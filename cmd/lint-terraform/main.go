// Copyright 2022 Google LLC
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

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/abcxyz/secure-setup-terraform/pkg/linter"
	"github.com/abcxyz/secure-setup-terraform/pkg/version"
)

const lintCommandHelp = `
The "lint" command 
EXAMPLES
  lint-terraform <file1> <file2> <directory>
FLAGS
`

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func realMain() error {
	ctx, done := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer done()

	f := flag.NewFlagSet("", flag.ExitOnError)
	f.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n\n", strings.TrimSpace(lintCommandHelp))
		f.PrintDefaults()
	}
	showVersion := f.Bool("version", false, "display version information")

	if err := f.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}
	if *showVersion {
		fmt.Fprintln(os.Stderr, version.HumanVersion)
		return nil
	}

	// The linter needs at least one file or directory
	args := f.Args()
	if got := len(args); got < 1 {
		return fmt.Errorf("expected at least one argument, got %d", got)
	}

	if err := linter.RunLinter(ctx, args, &linter.TerraformLinter{}); err != nil {
		return fmt.Errorf("error running linter %w", err)
	}
	return nil
}
