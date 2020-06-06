// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"bufio"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
)

func ParseConfig(cmd *cobra.Command, config io.Reader) (err error) {
	scanner := bufio.NewScanner(config)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		if err := parseLine(cmd, scanner.Text()); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return xerrors.Errorf("scanning config: %w", err)
	}

	return nil
}

// parseLine follows much of upstream's config file format. We differ in terms
// of quoting non-simple values which must be double quoted.
// One parameter is specified per line. The equal sign between name and value is
// optional. Whitespace is insignificant (except within a quoted parameter
// value) and blank lines are ignored. Hash marks (#) designate the remainder
// of the line as a comment.
// Upstream's documentation can be found here:
// https://www.postgresql.org/docs/current/config-setting.html#CONFIG-SETTING-CONFIGURATION-FILE
func parseLine(cmd *cobra.Command, line string) error {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}

	parts := strings.SplitN(line, "=", 2)
	if !strings.Contains(line, "=") {
		parts = strings.Fields(line)
	}

	name := strings.TrimSpace(parts[0])
	if len(parts) != 2 {
		return xerrors.Errorf("found no value for parameter %q", name)
	}

	value := strings.TrimSpace(parts[1])
	value = strings.TrimSpace(strings.SplitN(value, "#", 2)[0]) // remove inline comments

	if value == "" {
		return xerrors.Errorf("found no value for parameter %q", name)
	}

	if value[0] != '"' && value[len(value)-1] != '"' && len(strings.Fields(value)) != 1 {
		return xerrors.Errorf("found parameter %q with value %q containing spaces not enclosed in double quotes", name, value)
	}

	value = strings.TrimPrefix(value, "\"") // trim enclosing quotes
	value = strings.TrimSuffix(value, "\"") // trim enclosing quotes

	// Transpose config file parameters with underscores to dashes such that
	// the correct flag can be found.
	name = strings.ReplaceAll(name, "_", "-")
	flag := cmd.Flag(name)
	if flag == nil {
		return xerrors.Errorf("%q not found", name)
	}

	err := flag.Value.Set(value)
	if err != nil {
		return xerrors.Errorf("set %q to %q: %w", name, value, err)
	}

	cmd.Flag(name).Changed = true

	return nil
}
