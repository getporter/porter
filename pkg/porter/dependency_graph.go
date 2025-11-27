package porter

import (
	"fmt"
	"regexp"
	"strings"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/manifest"
	"github.com/hashicorp/go-multierror"
)

// DependencyGraph represents the resolved dependency graph with execution order
type DependencyGraph struct {
	// Nodes maps dependency name to graph node
	Nodes map[string]*DependencyNode

	// ExecutionOrder contains dependency names in topologically sorted order
	ExecutionOrder []string

	// Root is the parent bundle (executed last)
	Root *DependencyNode
}

// DependencyNode represents a single dependency in the graph
type DependencyNode struct {
	// Name is the dependency alias from manifest
	Name string

	// Lock contains the resolved bundle reference
	Lock cnab.DependencyLock

	// Dependencies are other nodes this node depends on
	Dependencies []*DependencyNode

	// Dependents are nodes that depend on this node
	Dependents []*DependencyNode

	// OutputsUsed tracks which outputs this node consumes from dependencies
	OutputsUsed map[string]OutputReference // param/cred name -> output reference
}

// OutputReference describes an output consumed from another dependency
type OutputReference struct {
	DependencyName string // Which dependency provides this output
	OutputName     string // Name of the output
}

// Regular expression to match dependency output references
// Matches: ${bundle.dependencies.DEP_NAME.outputs.OUTPUT_NAME}
var outputReferenceRegex = regexp.MustCompile(`\$\{bundle\.dependencies\.([^.]+)\.outputs\.([^}]+)\}`)

// parseOutputReference extracts dependency output references from template strings
// Matches: ${bundle.dependencies.DEP_NAME.outputs.OUTPUT_NAME}
// Returns nil if the value is not an output reference
func parseOutputReference(value string) *OutputReference {
	// Trim whitespace
	value = strings.TrimSpace(value)

	// Try to match the full template syntax
	matches := outputReferenceRegex.FindStringSubmatch(value)
	if matches == nil {
		return nil
	}

	if len(matches) != 3 {
		return nil
	}

	return &OutputReference{
		DependencyName: matches[1],
		OutputName:     matches[2],
	}
}

// buildDependencyGraph constructs a dependency graph from manifest dependencies
func (p *Porter) buildDependencyGraph(m *manifest.Manifest) (*DependencyGraph, error) {
	graph := &DependencyGraph{
		Nodes: make(map[string]*DependencyNode),
	}

	// Step 1: Create nodes for all dependencies
	for _, dep := range m.Dependencies.Requires {
		node := &DependencyNode{
			Name:        dep.Name,
			OutputsUsed: make(map[string]OutputReference),
		}
		graph.Nodes[dep.Name] = node
	}

	// Step 2: Parse parameter/credential sources to find output references
	for _, dep := range m.Dependencies.Requires {
		node := graph.Nodes[dep.Name]

		// Check parameters for dependency output references
		for paramName, paramValue := range dep.Parameters {
			if ref := parseOutputReference(paramValue); ref != nil {
				node.OutputsUsed[paramName] = *ref
				// Create edge: node depends on ref.DependencyName
				if refNode, exists := graph.Nodes[ref.DependencyName]; exists {
					node.Dependencies = append(node.Dependencies, refNode)
					refNode.Dependents = append(refNode.Dependents, node)
				} else {
					return nil, fmt.Errorf("dependency %s references output from non-existent dependency %s", dep.Name, ref.DependencyName)
				}
			}
		}

		// Check credentials for dependency output references
		for credName, credValue := range dep.Credentials {
			if ref := parseOutputReference(credValue); ref != nil {
				node.OutputsUsed[credName] = *ref
				// Create edge: node depends on ref.DependencyName
				if refNode, exists := graph.Nodes[ref.DependencyName]; exists {
					node.Dependencies = append(node.Dependencies, refNode)
					refNode.Dependents = append(refNode.Dependents, node)
				} else {
					return nil, fmt.Errorf("dependency %s references output from non-existent dependency %s", dep.Name, ref.DependencyName)
				}
			}
		}
	}

	return graph, nil
}

// computeExecutionOrder performs topological sort to determine execution order
// Uses Kahn's algorithm for topological sorting
func (g *DependencyGraph) computeExecutionOrder() error {
	// Step 1: Calculate in-degrees
	inDegree := make(map[string]int)
	for name, node := range g.Nodes {
		inDegree[name] = len(node.Dependencies)
	}

	// Step 2: Queue all nodes with no dependencies
	queue := []string{}
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// Step 3: Process queue
	executionOrder := []string{}
	for len(queue) > 0 {
		// Dequeue
		current := queue[0]
		queue = queue[1:]
		executionOrder = append(executionOrder, current)

		// Reduce in-degree of dependents
		node := g.Nodes[current]
		for _, dependent := range node.Dependents {
			inDegree[dependent.Name]--
			if inDegree[dependent.Name] == 0 {
				queue = append(queue, dependent.Name)
			}
		}
	}

	// Step 4: Check for cycles
	if len(executionOrder) != len(g.Nodes) {
		// Try to detect the cycle for a better error message
		if err := g.detectCycles(); err != nil {
			return err
		}
		return fmt.Errorf("circular dependency detected in bundle dependencies")
	}

	g.ExecutionOrder = executionOrder
	return nil
}

// detectCycles returns detailed error about circular dependencies
func (g *DependencyGraph) detectCycles() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var cycle []string

	var detectCycleHelper func(name string) error
	detectCycleHelper = func(name string) error {
		visited[name] = true
		recStack[name] = true
		cycle = append(cycle, name)

		node := g.Nodes[name]
		for _, dep := range node.Dependencies {
			if !visited[dep.Name] {
				if err := detectCycleHelper(dep.Name); err != nil {
					return err
				}
			} else if recStack[dep.Name] {
				// Found cycle
				cycleStart := 0
				for i, n := range cycle {
					if n == dep.Name {
						cycleStart = i
						break
					}
				}
				cycleStr := strings.Join(cycle[cycleStart:], " -> ")
				return fmt.Errorf("circular dependency detected: %s -> %s", cycleStr, dep.Name)
			}
		}

		recStack[name] = false
		cycle = cycle[:len(cycle)-1]
		return nil
	}

	for name := range g.Nodes {
		if !visited[name] {
			if err := detectCycleHelper(name); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateOutputReferences ensures all referenced outputs exist
func (g *DependencyGraph) validateOutputReferences(m *manifest.Manifest) error {
	var errors error

	for _, node := range g.Nodes {
		for _, outRef := range node.OutputsUsed {
			// Check if referenced dependency exists
			_, exists := g.Nodes[outRef.DependencyName]
			if !exists {
				errors = multierror.Append(errors, fmt.Errorf(
					"dependency %s references output from non-existent dependency %s",
					node.Name, outRef.DependencyName))
				continue
			}

			// For pinned versions (this issue's scope), we could validate outputs
			// by inspecting the bundle definition, but that requires fetching the bundle
			// which happens later in the resolution process. For now, we'll rely on
			// runtime validation when the output is actually resolved.
		}
	}

	return errors
}
