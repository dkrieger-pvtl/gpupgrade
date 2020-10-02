// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

/*
	The filter command massages the post-upgrade SQL dump by removing known
	differences. It does this with two sets of rules -- lines and blocks.

	- Line rules are regular expressions that will cause any matching lines to
	be removed immediately.

	- Block rules are regular expressions that cause any matching lines, and any
	preceding comments or blank lines, to be removed.

	The main complication here comes from the block rules, which require us to
	use a lookahead buffer.

	filter reads from stdin and writes to stdout. Usage:

		filter < target.sql > target-filtered.sql

	Error handling is basic: any failures result in a log.Fatal() call.
*/
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/greenplum-db/gpupgrade/ci/scripts/filters"
)

// function to identify if the buf and line pattern matches a pattern
type IdentifierFunc func(buf []string, line string) bool

// function to use for formatting a block identified by IdentifierFunc
type FormatFunc func(line string, allTokens []string) (string, []string, bool)

// identifier and corresponding formatting func
type Formatting struct {
	IdentifierFunc
	FormatFunc
}

var formatting []Formatting

var lineRegexes []*regexp.Regexp
var blockRegexes []*regexp.Regexp

func init() {
	// linePatterns remove exactly what is matched, on a line-by-line basis.
	linePatterns := []string{
		`ALTER DATABASE .+ SET gp_use_legacy_hashops TO 'on';`,
		// TODO: There may be false positives because of the below
		// pattern, and we might have to do a look ahead to really identify
		// if it can be deleted.
		`START WITH \d`,
	}

	// blockPatterns remove lines that match, AND any comments or whitespace
	// immediately preceding them.
	blockPatterns := []string{
		"CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;",
		"COMMENT ON EXTENSION plpgsql IS",
		"COMMENT ON DATABASE postgres IS",
	}

	formatting = []Formatting{
		{filters.IsViewOrRuleDdl, filters.BuildViewOrRuleDdl},
		{filters.IsTriggerDdl, filters.BuildTriggerDdl},
		{filters.IsWithClause, filters.BuildWithClause},
	}

	for _, pattern := range linePatterns {
		lineRegexes = append(lineRegexes, regexp.MustCompile(pattern))
	}
	for _, pattern := range blockPatterns {
		blockRegexes = append(blockRegexes, regexp.MustCompile(pattern))
	}
}

func writeBufAndLine(out io.Writer, buf []string, line string) []string {
	// We want to keep this line. Flush and empty our buffer first.
	if len(buf) > 0 {
		write(out, buf...)
		buf = (buf)[:0]
	}

	write(out, line)

	return buf
}

func write(out io.Writer, lines ...string) {
	for _, line := range lines {
		_, err := fmt.Fprintln(out, line)
		if err != nil {
			log.Fatalf("writing output: %+v", err)
		}
	}
}

func formatStmt(allTokens []string, line string, bf FormatFunc, buf []string) (string, []string, FormatFunc) {
	// if allTokens are already populated, continue formatting
	if len(allTokens) > 0 {
		completeDDL, resultTokens, _ := bf(line, allTokens)
		return completeDDL, resultTokens, bf
	}

	for _, r := range formatting {
		// identify if any formatting is applicable and call the corresponding formatting function
		if r.IdentifierFunc(buf, line) {
			completeDDL, resultTokens, _ := r.FormatFunc(line, allTokens)
			return completeDDL, resultTokens, r.FormatFunc
		}
	}

	// nothing to format
	return "", nil, nil
}

func Filter(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	// there are lines in icw regression suite requiring buffer
	// to be atleast 10000000, so keeping it a little higher for now.
	scanner.Buffer(nil, 9800*4024)

	var buf []string // lines buffered for look-ahead

	// temporary storage for formatting statements
	var allTokens []string
	var formatFunc FormatFunc

nextline:
	for scanner.Scan() {
		line := scanner.Text()

		completeDDL, resultTokens, f := formatStmt(allTokens, line, formatFunc, buf)
		// if a formatting function is identified, continue formatting until finished
		if f != nil {
			if len(completeDDL) > 0 {
				buf = writeBufAndLine(out, buf, completeDDL)

				// reset the temporary storage
				allTokens = nil
				formatFunc = nil
			} else {
				// maintain the tokens and the formatting function identified to
				// process the upcoming lines
				allTokens = resultTokens
				formatFunc = f
			}

			continue nextline
		}

		// First filter on a line-by-line basis.
		for _, r := range lineRegexes {
			if r.MatchString(line) {
				continue nextline
			}
		}

		if strings.HasPrefix(line, "--") || len(line) == 0 {
			// A comment or an empty line. We only want to output this section
			// if the SQL it's attached to isn't filtered.
			buf = append(buf, line)
			continue nextline
		}

		for _, r := range blockRegexes {
			if r.MatchString(line) {
				// Discard this line and any buffered comment block.
				buf = buf[:0]
				continue nextline
			}
		}

		buf = writeBufAndLine(out, buf, line)
	}

	if scanner.Err() != nil {
		log.Fatalf("scanning stdin: %+v", scanner.Err())
	}

	// Flush our buffer.
	if len(buf) > 0 {
		write(out, buf...)
	}
}

func main() {
	Filter(os.Stdin, os.Stdout)
}
