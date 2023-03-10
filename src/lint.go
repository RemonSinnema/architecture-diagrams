package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

var readers = map[string]ModelPartReader{
	"databases":         DatabaseReader{},
	"externalSystems":   ExternalSystemReader{},
	"personas":          PersonaReader{},
	"queues":            QueueReader{},
	"services":          ServiceReader{},
	"system":            SystemReader{},
	"technologyBundles": TechnologyBundleReader{},
	"technologies":      TechnologyReader{},
	"version":           VersionReader{},
	"workflows":         WorkflowReader{},
}

var connectors = []Connector{
	// Maintain order
	TechnologyBundleConnector{},
	DatabaseConnector{},
	QueueConnector{},
	ServiceConnector{},
	ExternalSystemConnector{},
	PersonaConnector{},
	WorkflowCollector{},
}

var validators = []Validator{
	DatabaseValidator{},
	DataStoreValidator{},
	ExternalSystemValidator{},
	PersonaValidator{},
	ServiceValidator{},
}

func LintText(text string) (*ArchitectureModel, []Issue) {
	model, issues := lint(text, "")
	return model, issues
}

func lint(definition string, fileName string) (model *ArchitectureModel, issues []Issue) {
	var node yaml.Node
	_ = yaml.Unmarshal([]byte(definition), &node)
	if !node.IsZero() {
		if node.Kind != yaml.DocumentNode || node.Content[0].Kind != yaml.MappingNode {
			return nil, invalidYaml("must be a map")
		}
		node = *node.Content[0]
	}

	model = &ArchitectureModel{node: &node}
	issues = make([]Issue, 0)
	children, _ := toMap(&node)
	for tag, child := range children {
		reader, exists := readers[tag]
		if exists {
			issues = append(issues, reader.read(child, fileName, model)...)
		} else {
			issues = append(issues, *NodeWarning(fmt.Sprint("Unknown top-level element: ", tag), child))
		}
	}
	for tag, reader := range readers {
		if _, processed := children[tag]; !processed {
			issues = append(issues, reader.read(nil, fileName, model)...)
		}
	}
	for _, connector := range connectors {
		issues = append(issues, connector.connect(model)...)
	}
	if len(issues) == 0 {
		for _, validator := range validators {
			issues = append(issues, validator.validate(model)...)
		}
	}
	return
}

func invalidYaml(message string) []Issue {
	return []Issue{*FileError(fmt.Sprintf("Invalid YAML: %v", message))}
}

func LintFile(fileName string) (*ArchitectureModel, []Issue) {
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		return nil, []Issue{*FileError(fmt.Sprintf("Couldn't read file %s: %v", fileName, err))}
	}
	model, issues := lint(string(bytes), fileName)
	return model, issues
}
