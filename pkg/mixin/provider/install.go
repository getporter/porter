package mixinprovider

import (
	"github.com/deislabs/porter/pkg/mixin"
)

func (p *FileSystem) Install(opts mixin.InstallOptions) error {

	/*
		curl -fsSLo $PORTER_HOME/mixins/exec/exec $PORTER_URL/mixins/exec/$PORTER_VERSION/exec-linux-amd64
		curl -fsSLo $PORTER_HOME/mixins/exec/exec-runtime $PORTER_URL/mixins/exec/$PORTER_VERSION/exec-runtime-linux-amd64
		chmod +x $PORTER_HOME/mixins/exec/exec
		chmod +x $PORTER_HOME/mixins/exec/exec-runtime
		echo Installed `$PORTER_HOME/mixins/exec/exec version`
	*/

	return nil
}
