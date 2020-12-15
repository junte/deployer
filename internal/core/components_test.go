package core

import (
	"fmt"
	"reflect"
	"testing"
)

func TestPrepareCommand(t *testing.T) {
	var tests = []struct {
		command []string
		args    map[string]string
		want    []string
	}{
		{
			[]string{"/bin/bash", "-c", "./deploy --tag=${arg_tag}"},
			map[string]string{
				"tag":     "124",
				"command": "rm -rf /",
			},
			[]string{"/bin/bash", "-c", "./deploy --tag=124"}},
		{
			[]string{"/bin/bash", "-c", "echo Hello World"},
			map[string]string{"command": "rm -rf /"},
			[]string{"/bin/bash", "-c", "echo Hello World"},
		},
		{
			[]string{"/run.sh"},
			map[string]string{},
			[]string{"/run.sh"},
		},
	}

	for _, testCase := range tests {
		testname := fmt.Sprintf("%s,%s", testCase.command, testCase.args)
		t.Run(testname, func(t *testing.T) {
			command := prepareCommand(testCase.command, testCase.args)
			if !reflect.DeepEqual(command, testCase.want) {
				t.Errorf("got %s, want %s", command, testCase.want)
			}
		})
	}
}
