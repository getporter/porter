package manifest

import "get.porter.sh/porter/pkg/config"

func MakeCNABCompatible(def *ParameterDefinition) bool {
	if v, ok := def.Type.(string); ok {
		if t, ok := config.PorterParamMap[v]; ok {
			def.Type = t
			def.ContentEncoding = "base64"

			return ok
		}
	}

	return false
}
