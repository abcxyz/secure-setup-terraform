package lint

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
)

const lintCommandHelp = `
The "lint" command 
EXAMPLES
  lint-%s <file1> <file2> <directory>
FLAGS
`

// RunLinter run executes the linter for a set of files
func RunLinter(ctx context.Context, originalArgs []string, linter Linter) error {
	f := flag.NewFlagSet("", flag.ExitOnError)
	f.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n\n", strings.TrimSpace(fmt.Sprintf(lintCommandHelp, linter.ViolationType())))
		f.PrintDefaults()
	}
	showVersion := f.Bool("version", false, "display version information")

	if err := f.Parse(originalArgs); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}
	if *showVersion {
		fmt.Fprintln(os.Stderr, linter.Version())
		return nil
	}

	// The linter needs at least one file or directory
	args := f.Args()
	if got := len(args); got < 1 {
		return fmt.Errorf("expected atleast one argument, got %d", got)
	}

	violations := []ViolationInstance{}
	// Process each provided path looking for violations
	for _, path := range args {
		instances, err := lint(path, linter)
		if err != nil {
			return fmt.Errorf("error linting files: %w", err)
		}
		violations = append(violations, instances...)
	}
	for _, instance := range violations {
		fmt.Printf("'%s' detected at [%s:%d]\n", linter.ViolationType(), instance.Path, instance.Line)
	}
	if len(violations) != 0 {
		return fmt.Errorf("found %d violation(s)", len(violations))
	}

	return nil
}
