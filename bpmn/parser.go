package bpmn

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

// Parse reads a BPMN file and returns the parsed Definitions.
func Parse(filePath string) (*Definitions, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	return ParseReader(f)
}

// ParseReader reads BPMN XML from a reader and returns the parsed Definitions.
func ParseReader(r io.Reader) (*Definitions, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read BPMN data: %w", err)
	}

	// BPMN files use various namespaces; we need to handle them.
	// Strip namespace prefixes for simpler parsing.
	content := stripNamespacePrefixes(string(data))

	var defs Definitions
	if err := xml.Unmarshal([]byte(content), &defs); err != nil {
		return nil, fmt.Errorf("failed to parse BPMN XML: %w", err)
	}

	return &defs, nil
}

// stripNamespacePrefixes removes common BPMN namespace prefixes from element names
// to allow simpler XML unmarshaling.
func stripNamespacePrefixes(content string) string {
	// Common BPMN namespace prefixes
	prefixes := []string{
		"bpmn:", "bpmn2:", "bpmndi:", "dc:", "di:", "omgdi:", "omgdc:",
		"semantic:", "xsi:", "camunda:", "zeebe:",
	}

	for _, prefix := range prefixes {
		// Replace opening tags: <prefix:Element -> <Element
		content = strings.ReplaceAll(content, "<"+prefix, "<")
		// Replace closing tags: </prefix:Element -> </Element
		content = strings.ReplaceAll(content, "</"+prefix, "</")
	}

	return content
}

// GetAllFlowNodes returns all flow node IDs in a process for reference validation.
func (p *Process) GetAllFlowNodes() map[string]string {
	nodes := make(map[string]string)

	for _, e := range p.StartEvents {
		nodes[e.ID] = "startEvent"
	}
	for _, e := range p.EndEvents {
		nodes[e.ID] = "endEvent"
	}
	for _, e := range p.Tasks {
		nodes[e.ID] = "task"
	}
	for _, e := range p.UserTasks {
		nodes[e.ID] = "userTask"
	}
	for _, e := range p.ServiceTasks {
		nodes[e.ID] = "serviceTask"
	}
	for _, e := range p.ScriptTasks {
		nodes[e.ID] = "scriptTask"
	}
	for _, e := range p.SendTasks {
		nodes[e.ID] = "sendTask"
	}
	for _, e := range p.ReceiveTasks {
		nodes[e.ID] = "receiveTask"
	}
	for _, e := range p.ManualTasks {
		nodes[e.ID] = "manualTask"
	}
	for _, e := range p.BusinessRuleTasks {
		nodes[e.ID] = "businessRuleTask"
	}
	for _, e := range p.SubProcesses {
		nodes[e.ID] = "subProcess"
	}
	for _, e := range p.ExclusiveGateways {
		nodes[e.ID] = "exclusiveGateway"
	}
	for _, e := range p.ParallelGateways {
		nodes[e.ID] = "parallelGateway"
	}
	for _, e := range p.InclusiveGateways {
		nodes[e.ID] = "inclusiveGateway"
	}
	for _, e := range p.EventBasedGateways {
		nodes[e.ID] = "eventBasedGateway"
	}
	for _, e := range p.IntermediateCatchEvents {
		nodes[e.ID] = "intermediateCatchEvent"
	}
	for _, e := range p.IntermediateThrowEvents {
		nodes[e.ID] = "intermediateThrowEvent"
	}
	for _, e := range p.BoundaryEvents {
		nodes[e.ID] = "boundaryEvent"
	}

	return nodes
}
