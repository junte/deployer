package deployer

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
			[]string{"/bin/bash", "-c", "./deploy --tag={{.Args.tag}}"},
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

	deployer := ComponentDeployer{}

	for _, testCase := range tests {
		t.Run(fmt.Sprintf("%s,%s", testCase.command, testCase.args), func(t *testing.T) {
			command, err := deployer.prepareCommand(testCase.command, testCase.args)
			if err != nil {
				t.Errorf("err: %s", err)
			} else if !reflect.DeepEqual(command, testCase.want) {
				t.Errorf("got %s, want %s", command, testCase.want)
			}
		})
	}
}
