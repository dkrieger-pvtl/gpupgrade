// Copyright (c) 2017-2020 VMware, Inc. or its affiliates
// SPDX-License-Identifier: Apache-2.0

package filters

import (
	"reflect"
	"testing"
)

func TestBuildTriggerDdl(t *testing.T) {
	type args struct {
		line      string
		allTokens []string
	}
	type result struct {
		line               string
		finishedFormatting bool
		resultTokens       []string
	}
	tests := []struct {
		name   string
		args   args
		result result
	}{
		{
			name: "returns completed ddl",
			args: args{
				line:      "CREATE TRIGGER before_trigger BEFORE INSERT OR DELETE ON public.foo FOR EACH ROW EXECUTE PROCEDURE public.foofunc();",
				allTokens: []string{},
			},
			result: result{
				line:               "CREATE TRIGGER before_trigger\n    BEFORE INSERT OR DELETE ON public.foo\n    FOR EACH ROW\n    EXECUTE PROCEDURE public.foofunc();",
				resultTokens:       []string{"CREATE", "TRIGGER", "before_trigger", "BEFORE", "INSERT", "OR", "DELETE", "ON", "public.foo", "FOR", "EACH", "ROW", "EXECUTE", "PROCEDURE", "public.foofunc();"},
				finishedFormatting: true,
			},
		},
		{
			name: "still processing trigger ddl",
			args: args{
				line:      "AFTER INSERT OR DELETE ON public.foo",
				allTokens: []string{"CREATE", "TRIGGER", "mytrigger"},
			},
			result: result{
				finishedFormatting: false,
				resultTokens: []string{"CREATE", "TRIGGER", "mytrigger", "AFTER", "INSERT", "OR", "DELETE", "ON",
					"public.foo"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completeDdl, resultTokens, finishedFormatting := BuildTriggerDdl(tt.args.line, tt.args.allTokens)
			if completeDdl != tt.result.line {
				t.Errorf("got %v, want %v", completeDdl, tt.result.line)
			}
			if finishedFormatting != tt.result.finishedFormatting {
				t.Errorf("got %t, want %t", finishedFormatting, tt.result.finishedFormatting)
			}
			if !reflect.DeepEqual(resultTokens, tt.result.resultTokens) {
				t.Errorf("got %q, want %q", resultTokens, tt.result.resultTokens)
			}
		})
	}
}

func TestFormatTriggerDdl(t *testing.T) {
	tests := []struct {
		name      string
		allTokens []string
		want      string
	}{
		{
			name: "formats create trigger statement and body in to one line",
			allTokens: []string{"CREATE", "TRIGGER", "after_trigger", "AFTER", "INSERT", "OR", "DELETE", "ON",
				"public.foo", "FOR", "EACH", "ROW", "EXECUTE", "PROCEDURE", "public.bfv_dml_error_func();"},
			want: "CREATE TRIGGER after_trigger\n    AFTER INSERT OR DELETE ON public.foo\n    FOR EACH ROW\n    EXECUTE PROCEDURE public.bfv_dml_error_func();",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatTriggerDdl(tt.allTokens); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
