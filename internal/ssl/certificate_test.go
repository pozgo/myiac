package ssl

import (
	"testing"

	"github.com/iac-io/myiac/internal/secret"
	"github.com/iac-io/myiac/internal/util"
	"github.com/iac-io/myiac/testutil"
	"github.com/stretchr/testify/assert"
)

func TestCreateTlsCertificate(t *testing.T) {
	// setup
	domain := "test-domain"
	cmdLine := testutil.FakeCommandRunner(domain)
	kubernetesRunner := secret.NewKubernetesRunner(cmdLine)
	secretManager := secret.NewKubernetesSecretManager("default", kubernetesRunner)

	// given
	certPath := "/tmp/cert.pem"
	keyPath := "/tmp/cert.key"

	_ = util.WriteStringToFile("testCert", certPath)
	_ = util.WriteStringToFile("testKey", keyPath)

	// when
	certificate := NewCertificate(domain, certPath, keyPath)
	certStore := NewSecretCertStore(secretManager)
	certStore.Register(certificate)

	// then
	expectedDeleteSecretCmdLine := "kubectl delete secret test-domain -n default"
	expectedCreateSecretCmdLine :=
		"kubectl -n default create secret tls test-domain --key=/tmp/tls.key --cert=/tmp/tls.crt"
	actualDeleteSecretCmdLine := cmdLine.CmdLines[0]
	actualCreateSecretCmdLine := cmdLine.CmdLines[1]

	createdSecretName := kubernetesRunner.FindSecret(domain, "default")
	assert.Contains(t, createdSecretName, domain)
	assert.Equal(t, expectedCreateSecretCmdLine, actualCreateSecretCmdLine)
	assert.Equal(t, expectedDeleteSecretCmdLine, actualDeleteSecretCmdLine)
}
