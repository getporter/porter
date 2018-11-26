package helm

import (
	"fmt"
)

const helmClientVersion = "v2.11.0"
const dockerfileLines = `RUN apt-get update && \
 apt-get install -y curl && \
 curl -o helm.tgz https://storage.googleapis.com/kubernetes-helm/helm-%s-linux-amd64.tar.gz && \
 tar -xzf helm.tgz && \
 mv linux-amd64/helm /usr/local/bin && \
 rm helm.tgz
RUN helm init --client-only`

func (m *Mixin) Build() error {
	fmt.Fprintf(m.Out, dockerfileLines, helmClientVersion)
	return nil
}
