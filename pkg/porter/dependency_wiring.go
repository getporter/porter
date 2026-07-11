package porter

import (
	"sort"

	v2 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
)

// wiringRef is a candidate wiring edge discovered while scanning one
// dependency's Parameters/Credentials/Outputs maps, identified by alias
// within a single Requires map. The builder resolves both aliases to
// NodeKeys once it knows each dependency's resolved reference.
type wiringRef struct {
	FromAlias string
	ToAlias   string
	Detail    WiringDetail
}

// danglingWiringRef is a wiring reference that can't produce a real edge --
// either it names a sibling dependency that doesn't exist in the same
// Requires map, or it references the dependency's own alias (never valid,
// since a dependency's own output isn't known until after it runs). Either
// way it's a bundle authoring mistake, surfaced as a warning rather than a
// hard failure.
type danglingWiringRef struct {
	FromAlias     string
	ToAlias       string
	Detail        WiringDetail
	SelfReference bool
}

// invalidWiringRef is a wiring reference that isn't a valid dependency
// output reference at all, e.g. referencing the root bundle's own output
// from within a dependency's field (`${bundle.outputs.X}`), which can never
// be resolved since dependencies run before the root bundle produces
// outputs.
type invalidWiringRef struct {
	FromAlias string
	Field     string
	FieldName string
	RawMatch  string
}

// extractWiringRefs scans every v2 dependency in requires for
// parameter/credential/output values that reference a sibling dependency's
// output, returning the valid references (edges to build), any dangling or
// self references, and any invalid (root-output) references. Only v2
// dependencies carry Parameters/Credentials/Outputs wiring fields at all, so
// this has nothing to find for v1-only bundles. The returned slices are
// sorted into a stable order so callers get deterministic output regardless
// of Go's randomized map iteration.
func extractWiringRefs(requires map[string]v2.Dependency) (refs []wiringRef, dangling []danglingWiringRef, invalid []invalidWiringRef) {
	for depName, dep := range requires {
		depRefs, depDangling, depInvalid := extractDependencyWiringRefs(depName, dep, requires)
		refs = append(refs, depRefs...)
		dangling = append(dangling, depDangling...)
		invalid = append(invalid, depInvalid...)
	}

	sort.Slice(refs, func(i, j int) bool {
		return wiringRefLess(refs[i].FromAlias, refs[i].ToAlias, refs[i].Detail, refs[j].FromAlias, refs[j].ToAlias, refs[j].Detail)
	})
	sort.Slice(dangling, func(i, j int) bool {
		return wiringRefLess(dangling[i].FromAlias, dangling[i].ToAlias, dangling[i].Detail, dangling[j].FromAlias, dangling[j].ToAlias, dangling[j].Detail)
	})
	sort.Slice(invalid, func(i, j int) bool {
		if invalid[i].FromAlias != invalid[j].FromAlias {
			return invalid[i].FromAlias < invalid[j].FromAlias
		}
		if invalid[i].Field != invalid[j].Field {
			return invalid[i].Field < invalid[j].Field
		}
		return invalid[i].FieldName < invalid[j].FieldName
	})

	return refs, dangling, invalid
}

// wiringRefLess orders by (FromAlias, Field, FieldName, ToAlias,
// SourceOutput). The first three alone can tie -- a single field value can
// embed multiple cross-dependency references (ParseAllDependencySources) --
// so ToAlias/SourceOutput are included to give every distinct ref a unique
// sort position; without them, sort.Slice's lack of stability could still
// flap the order of same-tuple refs between runs.
func wiringRefLess(fromAliasA, toAliasA string, detailA WiringDetail, fromAliasB, toAliasB string, detailB WiringDetail) bool {
	if fromAliasA != fromAliasB {
		return fromAliasA < fromAliasB
	}
	if detailA.Field != detailB.Field {
		return detailA.Field < detailB.Field
	}
	if detailA.FieldName != detailB.FieldName {
		return detailA.FieldName < detailB.FieldName
	}
	if toAliasA != toAliasB {
		return toAliasA < toAliasB
	}
	return detailA.SourceOutput < detailB.SourceOutput
}

// extractDependencyWiringRefs scans a single dependency's wiring maps for
// references to a sibling dependency's output.
func extractDependencyWiringRefs(depName string, dep v2.Dependency, requires map[string]v2.Dependency) (refs []wiringRef, dangling []danglingWiringRef, invalid []invalidWiringRef) {
	fields := []struct {
		name   string
		values map[string]string
	}{
		{"parameters", dep.Parameters},
		{"credentials", dep.Credentials},
		{"outputs", dep.Outputs},
	}

	for _, field := range fields {
		for fieldName, value := range field.values {
			sources, invalidMatches := v2.ParseAllDependencySources(value)

			for _, raw := range invalidMatches {
				invalid = append(invalid, invalidWiringRef{FromAlias: depName, Field: field.name, FieldName: fieldName, RawMatch: raw})
			}

			for _, src := range sources {
				// Only a reference to another dependency's output creates a
				// wiring edge; hardcoded values and references to the
				// current bundle's own parameters/credentials/outputs don't.
				if src.Dependency == "" || src.Output == "" {
					continue
				}

				detail := WiringDetail{
					Field:        field.name,
					FieldName:    fieldName,
					SourceOutput: src.Output,
				}

				if src.Dependency == depName {
					// A dependency can't reference its own output: it isn't
					// available until after the dependency itself has run.
					dangling = append(dangling, danglingWiringRef{FromAlias: depName, ToAlias: src.Dependency, Detail: detail, SelfReference: true})
					continue
				}

				if _, ok := requires[src.Dependency]; !ok {
					dangling = append(dangling, danglingWiringRef{FromAlias: depName, ToAlias: src.Dependency, Detail: detail})
					continue
				}

				refs = append(refs, wiringRef{FromAlias: depName, ToAlias: src.Dependency, Detail: detail})
			}
		}
	}

	return refs, dangling, invalid
}
