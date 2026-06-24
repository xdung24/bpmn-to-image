package dmn

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

// Parse reads a DMN file from disk and returns the parsed Definitions.
func Parse(filePath string) (*Definitions, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	return ParseReader(f)
}

// ParseReader reads DMN XML from a reader and returns the parsed Definitions.
func ParseReader(r io.Reader) (*Definitions, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read DMN data: %w", err)
	}

	content := stripNamespacePrefixes(string(data))

	var defs Definitions
	if err := xml.Unmarshal([]byte(content), &defs); err != nil {
		return nil, fmt.Errorf("failed to parse DMN XML: %w", err)
	}

	return &defs, nil
}

// stripNamespacePrefixes removes common DMN namespace prefixes from
// element names so we can decode with simple struct tags.
func stripNamespacePrefixes(content string) string {
	prefixes := []string{
		"dmn:", "dmn11:", "dmn12:", "dmn13:", "dmn14:", "dmn15:",
		"dmndi:", "dmndi12:", "dmndi13:", "dmndi14:", "dmndi15:",
		"dc:", "di:", "feel:", "xsi:", "camunda:", "kie:", "drools:",
	}
	for _, p := range prefixes {
		content = strings.ReplaceAll(content, "<"+p, "<")
		content = strings.ReplaceAll(content, "</"+p, "</")
	}
	return content
}

// NodeKind labels a DRG element so the renderer can choose a shape.
type NodeKind string

const (
	KindDecision        NodeKind = "decision"
	KindInputData       NodeKind = "inputData"
	KindBKM             NodeKind = "businessKnowledgeModel"
	KindKnowledgeSource NodeKind = "knowledgeSource"
	KindDecisionService NodeKind = "decisionService"
	KindTextAnnotation  NodeKind = "textAnnotation"
)

// EdgeKind labels a DMN edge so the renderer can choose a line style.
type EdgeKind string

const (
	EdgeInformation EdgeKind = "informationRequirement"
	EdgeKnowledge   EdgeKind = "knowledgeRequirement"
	EdgeAuthority   EdgeKind = "authorityRequirement"
	EdgeAssociation EdgeKind = "association"
)

// NodeIndex maps element IDs to their kind and display name.
type NodeIndex map[string]NodeInfo

type NodeInfo struct {
	Kind NodeKind
	Name string
}

// BuildNodeIndex collects every DRG node by ID with its kind and name.
func (defs *Definitions) BuildNodeIndex() NodeIndex {
	idx := NodeIndex{}
	for _, d := range defs.Decisions {
		idx[d.ID] = NodeInfo{KindDecision, d.Name}
	}
	for _, d := range defs.InputDatas {
		idx[d.ID] = NodeInfo{KindInputData, d.Name}
	}
	for _, d := range defs.BusinessKnowledgeModels {
		idx[d.ID] = NodeInfo{KindBKM, d.Name}
	}
	for _, d := range defs.KnowledgeSources {
		idx[d.ID] = NodeInfo{KindKnowledgeSource, d.Name}
	}
	for _, d := range defs.DecisionServices {
		idx[d.ID] = NodeInfo{KindDecisionService, d.Name}
	}
	for _, d := range defs.TextAnnotations {
		idx[d.ID] = NodeInfo{KindTextAnnotation, d.Text}
	}
	return idx
}

// EdgeIndex maps a requirement/association ID to its edge kind.
// DMNEdge.dmnElementRef points at one of these.
type EdgeIndex map[string]EdgeKind

// BuildEdgeIndex walks the DRG and records the type of every requirement
// and association by ID, so the renderer can pick the right line style.
func (defs *Definitions) BuildEdgeIndex() EdgeIndex {
	idx := EdgeIndex{}
	for _, d := range defs.Decisions {
		for _, r := range d.InformationRequirements {
			if r.ID != "" {
				idx[r.ID] = EdgeInformation
			}
		}
		for _, r := range d.KnowledgeRequirements {
			if r.ID != "" {
				idx[r.ID] = EdgeKnowledge
			}
		}
		for _, r := range d.AuthorityRequirements {
			if r.ID != "" {
				idx[r.ID] = EdgeAuthority
			}
		}
	}
	for _, b := range defs.BusinessKnowledgeModels {
		for _, r := range b.KnowledgeRequirements {
			if r.ID != "" {
				idx[r.ID] = EdgeKnowledge
			}
		}
		for _, r := range b.AuthorityRequirements {
			if r.ID != "" {
				idx[r.ID] = EdgeAuthority
			}
		}
	}
	for _, k := range defs.KnowledgeSources {
		for _, r := range k.AuthorityRequirements {
			if r.ID != "" {
				idx[r.ID] = EdgeAuthority
			}
		}
	}
	for _, a := range defs.Associations {
		if a.ID != "" {
			idx[a.ID] = EdgeAssociation
		}
	}
	return idx
}

// StripHref returns the local ID part of a "#id" href, or the input as-is.
func StripHref(h string) string {
	if h == "" {
		return ""
	}
	if i := strings.LastIndex(h, "#"); i >= 0 {
		return h[i+1:]
	}
	return h
}
