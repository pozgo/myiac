package secret

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/iac-io/myiac/testutil"
	"github.com/stretchr/testify/assert"
)

func TestCreateSecret(t *testing.T) {
	// setup
	cmdLine := testutil.FakeCommandRunner("test-output")
	kubernetesRunner := NewKubernetesRunner(cmdLine)
	secretManager := NewKubernetesSecretManager("default", kubernetesRunner)

	// given
	filePath := "/tmp/filepath"
	secretName := "test-secret-name"
	_, err := os.Create(filePath)

	if err != nil {
		log.Fatalf("error: creating file")
	}

	// when
	secretManager.CreateFileSecret(secretName, filePath)

	// then
	//TODO: should validate snake case in the secret name (camelCase failures)
	expectedDeleteSecretCmdLine := "kubectl delete secret test-secret-name -n default"
	expectedCreateSecretCmdLine :=
		fmt.Sprintf("kubectl create secret generic %s "+
			"--from-file=%s.json=%s -n default", secretName, secretName, filePath)
	actualDeleteSecretCmdLine := cmdLine.CmdLines[0]
	actualCreateSecretCmdLine := cmdLine.CmdLines[1]

	assert.Equal(t, expectedCreateSecretCmdLine, actualCreateSecretCmdLine)
	assert.Equal(t, expectedDeleteSecretCmdLine, actualDeleteSecretCmdLine)
}
