package porter

import "get.porter.sh/porter/pkg/storage"

type DisplayWorkflow struct {
	storage.WorkflowSpec
}

what if i had a custom function that serialized to a map, I tweak what we got, then to json/yaml
so that I don't have to duplicate the darn structure like we do with installation'