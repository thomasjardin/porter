package extensions

import (
	"encoding/json"
	"fmt"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/pkg/errors"
)

const (
	// DependenciesKey represents the full key for the Dependencies Extension
	DependenciesKey = "io.cnab.dependencies"
	// DependenciesSchema represents the schema for the Dependencies Extension
	DependenciesSchema = "https://cnab.io/v1/dependencies.schema.json"
)

// DependenciesExtension represents the required extension to enable dependencies
var DependenciesExtension = RequiredExtension{
	Shorthand: "dependencies",
	Key:       DependenciesKey,
	Schema:    DependenciesSchema,
	Reader:    DependencyReader,
}

// Dependencies describes the set of custom extension metadata associated with the dependencies spec
// https://github.com/cnabio/cnab-spec/blob/master/500-CNAB-dependencies.md
type Dependencies struct {
	// Sequence is a list to order the dependencies
	Sequence []string `json:"sequence,omitempty" mapstructure:"sequence"`

	// Requires is a list of bundles required by this bundle
	Requires map[string]Dependency `json:"requires,omitempty" mapstructure:"requires"`
}

// Dependency describes a dependency on another bundle
type Dependency struct {
	// Name of the dependency
	Name string `json:"name" mapstructure:"name"`

	// Bundle is the location of the bundle in a registry, for example REGISTRY/NAME:TAG
	Bundle string `json:"bundle" mapstructure:"bundle"`

	// Version is a set of allowed versions
	Version *DependencyVersion `json:"version,omitempty" mapstructure:"version"`
}

// DependencyVersion is a set of allowed versions for a dependency
type DependencyVersion struct {
	// Ranges of semantic versions, with or without the leading v prefix, allowed by the dependency
	Ranges []string `json:"ranges,omitempty" mapstructure:"ranges"`

	// AllowPrereleases specifies if prerelease versions can satisfy the dependency
	AllowPrereleases bool `json:"prereleases" mapstructure:"prereleases"`
}

// ReadDependencies is a convenience method for returning a bonafide
// Dependencies reference after reading from the applicable section from
// the provided bundle
func ReadDependencies(bun bundle.Bundle) (Dependencies, error) {
	raw, err := DependencyReader(bun)
	if err != nil {
		return Dependencies{}, err
	}

	deps, ok := raw.(Dependencies)
	if !ok {
		return Dependencies{}, errors.New("unable to read dependencies extension data")
	}

	// Return the dependencies
	return deps, nil
}

// DependencyReader is a Reader for the DependenciesExtension, which reads
// from the applicable section in the provided bundle and returns a the raw
// data in the form of an interface
func DependencyReader(bun bundle.Bundle) (interface{}, error) {
	data, ok := bun.Custom[DependenciesKey]
	if !ok {
		return nil, errors.Errorf("attempted to read dependencies from bundle but none are defined")
	}

	dataB, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshal the untyped dependencies extension data %q", string(dataB))
	}

	deps := Dependencies{}
	err = json.Unmarshal(dataB, &deps)
	if err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal the dependencies extension %q", string(dataB))
	}

	return deps, nil
}

// HasDependencies returns whether or not the bundle has dependencies defined.
func HasDependencies(bun bundle.Bundle) bool {
	_, ok := bun.Custom[DependenciesKey]
	return ok
}

// ListBySequence returns the dependencies by the defined sequence,
// if none is specified, they are unsorted.
func (d Dependencies) ListBySequence() []Dependency {
	deps := make([]Dependency, 0, len(d.Requires))
	if len(d.Sequence) > 0 && len(d.Sequence) == len(d.Requires) {
		for _, depName := range d.Sequence {
			dep := d.Requires[depName]
			dep.Name = depName
			deps = append(deps, dep)
		}
	} else {
		for depName, dep := range d.Requires {
			dep.Name = depName
			deps = append(deps, dep)
		}
	}
	return deps
}

// BuildPrerequisiteInstallationName generates the name of a prerequisite dependency installation.
func BuildPrerequisiteInstallationName(installation string, dependency string) string {
	return fmt.Sprintf("%s-%s", installation, dependency)

}
