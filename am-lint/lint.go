package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

var readers = map[string]ModelPartReader{
	"version":         VersionReader{},
	"system":          SystemReader{},
	"personas":        PersonaReader{},
	"externalSystems": ExternalSystemReader{},
}

func LintText(text string) (*ArchitectureModel, []Issue) {
	model, issues := lint(text, "")
	return model, issues
}

func lint(definition string, fileName string) (model *ArchitectureModel, issues []Issue) {
	var node yaml.Node
	err := yaml.Unmarshal([]byte(definition), &node)
	if err != nil {
		return nil, invalidYaml()
	}
	if !node.IsZero() {
		if node.Kind != yaml.DocumentNode || node.Content[0].Kind != yaml.MappingNode {
			return nil, invalidYaml()
		}
		node = *node.Content[0]
	}

	model = &ArchitectureModel{}
	issues = make([]Issue, 0)
	children, issue := toMap(&node)
	if issue != nil {
		issues = append(issues, *issue)
		return
	}
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
	return
}

func invalidYaml() []Issue {
	return []Issue{*FileError("Invalid YAML")}
}

func LintFile(fileName string) (*ArchitectureModel, []Issue) {
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		return nil, []Issue{*FileError(fmt.Sprintf("Couldn't read file %s: %v", fileName, err))}
	}
	model, issues := lint(string(bytes), fileName)
	return model, issues
}
