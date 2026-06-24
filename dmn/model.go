// Package dmn provides parsing, validation and a data model for DMN 1.x
// (Decision Model and Notation) files, focused on the Decision
// Requirements Diagram (DRD). Decision tables are parsed structurally
// but their rule rendering is out of scope for this package.
package dmn

import "encoding/xml"

// Definitions is the root element of a DMN XML file.
type Definitions struct {
	XMLName                 xml.Name                 `xml:"definitions"`
	ID                      string                   `xml:"id,attr"`
	Name                    string                   `xml:"name,attr"`
	Namespace               string                   `xml:"namespace,attr"`
	Decisions               []Decision               `xml:"decision"`
	InputDatas              []InputData              `xml:"inputData"`
	BusinessKnowledgeModels []BusinessKnowledgeModel `xml:"businessKnowledgeModel"`
	KnowledgeSources        []KnowledgeSource        `xml:"knowledgeSource"`
	DecisionServices        []DecisionService        `xml:"decisionService"`
	TextAnnotations         []TextAnnotation         `xml:"textAnnotation"`
	Associations            []Association            `xml:"association"`
	DMNDI                   *DMNDI                   `xml:"DMNDI"`
}

// Decision is a single DRG decision node.
type Decision struct {
	XMLName                 xml.Name                 `xml:"decision"`
	ID                      string                   `xml:"id,attr"`
	Name                    string                   `xml:"name,attr"`
	Question                string                   `xml:"question"`
	AllowedAnswers          string                   `xml:"allowedAnswers"`
	InformationRequirements []InformationRequirement `xml:"informationRequirement"`
	KnowledgeRequirements   []KnowledgeRequirement   `xml:"knowledgeRequirement"`
	AuthorityRequirements   []AuthorityRequirement   `xml:"authorityRequirement"`
}

// InputData is data entered into the model (no logic, just a value).
type InputData struct {
	XMLName xml.Name `xml:"inputData"`
	ID      string   `xml:"id,attr"`
	Name    string   `xml:"name,attr"`
}

// BusinessKnowledgeModel represents reusable decision logic invoked
// from one or more decisions.
type BusinessKnowledgeModel struct {
	XMLName               xml.Name               `xml:"businessKnowledgeModel"`
	ID                    string                 `xml:"id,attr"`
	Name                  string                 `xml:"name,attr"`
	KnowledgeRequirements []KnowledgeRequirement `xml:"knowledgeRequirement"`
	AuthorityRequirements []AuthorityRequirement `xml:"authorityRequirement"`
}

// KnowledgeSource cites a policy/regulation/expert that authorizes a decision.
type KnowledgeSource struct {
	XMLName               xml.Name               `xml:"knowledgeSource"`
	ID                    string                 `xml:"id,attr"`
	Name                  string                 `xml:"name,attr"`
	AuthorityRequirements []AuthorityRequirement `xml:"authorityRequirement"`
}

// DecisionService bundles a set of decisions exposed as a service.
type DecisionService struct {
	XMLName          xml.Name `xml:"decisionService"`
	ID               string   `xml:"id,attr"`
	Name             string   `xml:"name,attr"`
	OutputDecisions  []Href   `xml:"outputDecision"`
	EncapsulatedDecs []Href   `xml:"encapsulatedDecision"`
	InputDecisions   []Href   `xml:"inputDecision"`
	InputData        []Href   `xml:"inputData"`
}

// TextAnnotation is an unstyled note.
type TextAnnotation struct {
	XMLName xml.Name `xml:"textAnnotation"`
	ID      string   `xml:"id,attr"`
	Text    string   `xml:"text"`
}

// Association connects a TextAnnotation to another element.
type Association struct {
	XMLName   xml.Name `xml:"association"`
	ID        string   `xml:"id,attr"`
	SourceRef Href     `xml:"sourceRef"`
	TargetRef Href     `xml:"targetRef"`
}

// Href wraps a single href="#id" reference child element.
type Href struct {
	Href string `xml:"href,attr"`
}

// InformationRequirement: another decision or input data feeds this decision.
type InformationRequirement struct {
	XMLName          xml.Name `xml:"informationRequirement"`
	ID               string   `xml:"id,attr"`
	RequiredDecision *Href    `xml:"requiredDecision"`
	RequiredInput    *Href    `xml:"requiredInput"`
}

// KnowledgeRequirement: a BKM is invoked by this decision or BKM.
type KnowledgeRequirement struct {
	XMLName           xml.Name `xml:"knowledgeRequirement"`
	ID                string   `xml:"id,attr"`
	RequiredKnowledge *Href    `xml:"requiredKnowledge"`
}

// AuthorityRequirement: a KnowledgeSource authorizes this element.
type AuthorityRequirement struct {
	XMLName           xml.Name `xml:"authorityRequirement"`
	ID                string   `xml:"id,attr"`
	RequiredAuthority *Href    `xml:"requiredAuthority"`
	RequiredDecision  *Href    `xml:"requiredDecision"`
	RequiredInput     *Href    `xml:"requiredInput"`
}

// DMN Diagram Interchange (DMNDI)

type DMNDI struct {
	XMLName  xml.Name     `xml:"DMNDI"`
	Diagrams []DMNDiagram `xml:"DMNDiagram"`
}

type DMNDiagram struct {
	XMLName xml.Name   `xml:"DMNDiagram"`
	ID      string     `xml:"id,attr"`
	Name    string     `xml:"name,attr"`
	Shapes  []DMNShape `xml:"DMNShape"`
	Edges   []DMNEdge  `xml:"DMNEdge"`
}

type DMNShape struct {
	XMLName       xml.Name `xml:"DMNShape"`
	ID            string   `xml:"id,attr"`
	DMNElementRef string   `xml:"dmnElementRef,attr"`
	IsCollapsed   *bool    `xml:"isCollapsed,attr"`
	Bounds        Bounds   `xml:"Bounds"`
	Label         *Label   `xml:"DMNLabel"`
}

type DMNEdge struct {
	XMLName       xml.Name   `xml:"DMNEdge"`
	ID            string     `xml:"id,attr"`
	DMNElementRef string     `xml:"dmnElementRef,attr"`
	Waypoints     []Waypoint `xml:"waypoint"`
	Label         *Label     `xml:"DMNLabel"`
}

type Bounds struct {
	XMLName xml.Name `xml:"Bounds"`
	X       float64  `xml:"x,attr"`
	Y       float64  `xml:"y,attr"`
	Width   float64  `xml:"width,attr"`
	Height  float64  `xml:"height,attr"`
}

type Waypoint struct {
	XMLName xml.Name `xml:"waypoint"`
	X       float64  `xml:"x,attr"`
	Y       float64  `xml:"y,attr"`
}

type Label struct {
	Bounds *Bounds `xml:"Bounds"`
	Text   string  `xml:"Text"`
}
