package deploy

import (
	"encoding/json"
	"testing"

	"github.com/iac-io/myiac/internal/commandline"
)

const ExistingReleasesOutput = `
{
	"Next": "",
	"Releases": [{
		"Name": "esteemed-peacock",
		"Revision": 2,
		"Updated": "Mon Dec  2 18:26:30 2019",
		"Status": "DEPLOYED",
		"Chart": "moneycolfrontend-1.0.0",
		"AppVersion": "0.1.0",
		"Namespace": "default"
	}, {
		"Name": "opining-frog",
		"Revision": 36,
		"Updated": "Fri Dec  6 13:41:17 2019",
		"Status": "DEPLOYED",
		"Chart": "traefik-1.78.4",
		"AppVersion": "1.7.14",
		"Namespace": "default"
	}, {
		"Name": "ponderous-lion",
		"Revision": 3,
		"Updated": "Mon Dec  2 18:26:30 2019",
		"Status": "DEPLOYED",
		"Chart": "moneycolserver-1.0.0",
		"AppVersion": "1.0.0",
		"Namespace": "default"
	}, {
		"Name": "solitary-ragdoll",
		"Revision": 2,
		"Updated": "Thu Dec  5 12:48:25 2019",
		"Status": "DEPLOYED",
		"Chart": "elasticsearch-1.0.0",
		"AppVersion": "6.5.0",
		"Namespace": "default"
	}]
}
`

// Here we implement the CommandRunner interface with a testing mock
type mockCommandRunner struct {
	executable     string
	arguments      []string
	output         string
	suppressOutput bool
}

func (mcr *mockCommandRunner) SetSuppressOutput(suppressOutput bool) {
	mcr.suppressOutput = suppressOutput
}

func (mcr *mockCommandRunner) SetOutput(output string) {
	mcr.output = output
}

func (mcr mockCommandRunner) RunVoid() {}

func (mcr *mockCommandRunner) Output() string {
	return mcr.output
}

func (mcr mockCommandRunner) Setup(executable string, args []string) {
	mcr.executable = executable
	mcr.arguments = args
}

func (mcr mockCommandRunner) SetupWithoutOutput(executable string, args []string) {
	mcr.executable = executable
	mcr.arguments = args
}

func (mcr mockCommandRunner) IgnoreError(ignoreError bool) {}

func (mcr mockCommandRunner) Run() commandline.CommandOutput {
	return commandline.CommandOutput{Output: mcr.output}
}

func (mcr mockCommandRunner) SetupCmdLine(cmdLine string) {
	// ignored
}

// https://quii.gitbook.io/learn-go-with-tests/
// To run: go test -v
func TestReleaseDeployed(t *testing.T) {
	commandRunner := &mockCommandRunner{output: ExistingReleasesOutput}
	d := NewHelmDeployer("charts", commandRunner)

	if !d.DeployedReleasesExistsFor("traefik") {
		t.Errorf("The release is deployed was incorrect, got: %v, want: %v.", false, true)
	}
}

func TestReleaseHasFailed(t *testing.T) {
	commandRunner := &mockCommandRunner{output: ""}
	d := NewHelmDeployer("charts", commandRunner)

	// Given: a release (2nd one) has failed status
	releasesList := d.ParseReleasesList(ExistingReleasesOutput)
	release := releasesList.Releases[1]
	release.Status = "FAILED"

	existingReleasesModified, err := json.Marshal(releasesList)

	if err != nil {
		t.Errorf("Failure: error marshalling %v\n %v\n", releasesList, err)
	}

	commandRunner.SetOutput(string(existingReleasesModified))

	// When: checking if it has been deployed
	deployed := d.DeployedReleasesExistsFor("traefik")

	// Then: it shouldn't be deployed by failed
	if deployed {
		t.Errorf("The release is failed but got deployed\n")
	}
}
