// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package filters

import (
	"regexp"
	"strings"
)

var (
	// regex for trigger transformation
	triggerCommentRegex *regexp.Regexp
	triggerCreateRegex  *regexp.Regexp
)

func init() {
	triggerCommentRegex = regexp.MustCompile(`; Type: TRIGGER;`)
	triggerCreateRegex = regexp.MustCompile(`CREATE TRIGGER `)
}

func IsTriggerDdl(buf []string, line string) bool {
	return len(buf) > 0 && triggerCommentRegex.MatchString(strings.Join(buf, " ")) && triggerCreateRegex.MatchString(line)
}

func FormatTriggerDdl(allTokens []string) string {
	var line string
	for _, token := range allTokens {
		if line == "" {
			// processing the first element
			line = token
			continue
		}

		// by default add single space between tokens, but if a token is identified which marks a new line
		// use a new line and 4 character space indentation to match the format of old dump
		indentation := " "
		for _, identifier := range []string{"AFTER", "BEFORE", "FOR", "EXECUTE"} {
			if token == identifier {
				indentation = "\n    "
				break
			}
		}

		line = line + indentation + token
	}

	return line
}

func BuildTriggerDdl(line string, allTokens []string) (string, []string) {
	tokens := strings.Fields(line)
	allTokens = append(allTokens, tokens...)
	return FormatTriggerDdl(allTokens), allTokens
}
