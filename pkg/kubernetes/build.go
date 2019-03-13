package kubernetes

import (
	"fmt"
)

const kubeVersion = "v1.13.0"
const dockerFileContents = `RUN apt-get update && \
apt-get install -y apt-transport-https curl && \
curl -o kubectl https://storage.googleapis.com/kubernetes-release/release/%s/bin/linux/amd64/kubectl && \
mv kubectl /usr/local/bin && \
chmod a+x /usr/local/bin/kubectl
`

// Build generates the relevant Dockerfile output for this mixin
func (m *Mixin) Build() error {
	_, err := fmt.Fprintf(m.Out, dockerFileContents, kubeVersion)
	return err
}
