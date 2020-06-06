// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package commands_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/spf13/cobra"

	"github.com/greenplum-db/gpupgrade/cli/commands"
)

func TestConfig(t *testing.T) {
	cases := []struct {
		description string
		config      string
		parameter   string
		expected    string
	}{
		{
			description: "parses parameters",
			config:      "name = value",
			parameter:   "name",
			expected:    "value",
		},
		{
			description: "uses last parameter value when parameter specified multiple times",
			config:      "name = value\nname = value2",
			parameter:   "name",
			expected:    "value2",
		},
		{
			description: "parses parameters when equal sign omitted",
			config:      "name value",
			parameter:   "name",
			expected:    "value",
		},
		{
			description: "replaces _ with - in parameter names",
			config:      "name_with_dash = value",
			parameter:   "name-with-dash",
			expected:    "value",
		},
		{
			description: "parses values containing spaces",
			config:      `name = "value with spaces"`,
			parameter:   "name",
			expected:    "value with spaces",
		},
		{
			description: "parses values containing single and double quotes",
			config:      `name = "value with "double" and 'single' quotes"`,
			parameter:   "name",
			expected:    `value with "double" and 'single' quotes`,
		},
		{
			description: "ignores empty lines",
			config:      "   \n  \t",
			parameter:   "name",
			expected:    "",
		},
		{
			description: "ignores leading comments",
			config:      "# comment",
			parameter:   "name",
			expected:    "",
		},
		{
			description: "ignores inline comments",
			config:      "name = value # comment",
			parameter:   "name",
			expected:    "value",
		},
		{
			description: "ignores inline comments containing equal signs",
			config:      "name = value # comment with = sign",
			parameter:   "name",
			expected:    "value",
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			// Using reflection allows instantiation of various parameters in
			// the above cases such as "name" and "name-with-dash".
			value := reflect.Zero(reflect.TypeOf(c.parameter)).Interface().(string)

			cmd := cobra.Command{}
			cmd.Flags().StringVar(&value, c.parameter, "", "")

			var input bytes.Buffer
			input.WriteString(c.config)

			err := commands.ParseConfig(&cmd, &input)
			if err != nil {
				t.Errorf("ParseConfig returned error: %+v", err)
			}

			err = cmd.Execute()
			if err != nil {
				t.Errorf("cmd.Execute returned error: %+v", err)
			}

			if value != c.expected {
				t.Errorf("got %q want %q", value, c.expected)
			}
		})
	}

	errorCases := []struct {
		description string
		config      string
	}{
		{
			description: "errors on unknown parameter",
			config:      "unknown = value",
		},
		{
			description: "errors when parameter value is empty and equal sign omitted",
			config:      "name ",
		},
		{
			description: "errors when parameter value is empty with an inline comment",
			config:      "name # comment",
		},
		{
			description: "errors when value is missing double quotes",
			config:      "name = value with spaces # comment",
		},
	}

	for _, c := range errorCases {
		t.Run(c.description, func(t *testing.T) {
			var name string
			cmd := cobra.Command{}
			cmd.Flags().StringVar(&name, "name", "", "")

			var input bytes.Buffer
			input.WriteString(c.config)

			err := commands.ParseConfig(&cmd, &input)
			if err == nil {
				t.Errorf("ParseConfig returned error: %+v", err)
			}
		})
	}
}
