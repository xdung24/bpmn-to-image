package bpmn

import "encoding/xml"

// Definitions is the root element of a BPMN 2.0 XML file.
type Definitions struct {
	XMLName        xml.Name        `xml:"definitions"`
	ID             string          `xml:"id,attr"`
	Name           string          `xml:"name,attr"`
	TargetNS       string          `xml:"targetNamespace,attr"`
	Processes      []Process       `xml:"process"`
	Collaborations []Collaboration `xml:"collaboration"`
	Diagrams       []BPMNDiagram   `xml:"BPMNDiagram"`
}

type Collaboration struct {
	XMLName      xml.Name      `xml:"collaboration"`
	ID           string        `xml:"id,attr"`
	Participants []Participant `xml:"participant"`
	MessageFlows []MessageFlow `xml:"messageFlow"`
}

type Participant struct {
	XMLName    xml.Name `xml:"participant"`
	ID         string   `xml:"id,attr"`
	Name       string   `xml:"name,attr"`
	ProcessRef string   `xml:"processRef,attr"`
}

type MessageFlow struct {
	XMLName   xml.Name `xml:"messageFlow"`
	ID        string   `xml:"id,attr"`
	Name      string   `xml:"name,attr"`
	SourceRef string   `xml:"sourceRef,attr"`
	TargetRef string   `xml:"targetRef,attr"`
}

type Process struct {
	XMLName                 xml.Name                 `xml:"process"`
	ID                      string                   `xml:"id,attr"`
	Name                    string                   `xml:"name,attr"`
	IsExecutable            bool                     `xml:"isExecutable,attr"`
	StartEvents             []StartEvent             `xml:"startEvent"`
	EndEvents               []EndEvent               `xml:"endEvent"`
	Tasks                   []Task                   `xml:"task"`
	UserTasks               []UserTask               `xml:"userTask"`
	ServiceTasks            []ServiceTask            `xml:"serviceTask"`
	ScriptTasks             []ScriptTask             `xml:"scriptTask"`
	SendTasks               []SendTask               `xml:"sendTask"`
	ReceiveTasks            []ReceiveTask            `xml:"receiveTask"`
	ManualTasks             []ManualTask             `xml:"manualTask"`
	BusinessRuleTasks       []BusinessRuleTask       `xml:"businessRuleTask"`
	SubProcesses            []SubProcess             `xml:"subProcess"`
	ExclusiveGateways       []ExclusiveGateway       `xml:"exclusiveGateway"`
	ParallelGateways        []ParallelGateway        `xml:"parallelGateway"`
	InclusiveGateways       []InclusiveGateway       `xml:"inclusiveGateway"`
	EventBasedGateways      []EventBasedGateway      `xml:"eventBasedGateway"`
	IntermediateCatchEvents []IntermediateCatchEvent `xml:"intermediateCatchEvent"`
	IntermediateThrowEvents []IntermediateThrowEvent `xml:"intermediateThrowEvent"`
	BoundaryEvents          []BoundaryEvent          `xml:"boundaryEvent"`
	SequenceFlows           []SequenceFlow           `xml:"sequenceFlow"`
	DataObjects             []DataObject             `xml:"dataObject"`
	TextAnnotations         []TextAnnotation         `xml:"textAnnotation"`
	Associations            []Association            `xml:"association"`
	LaneSet                 *LaneSet                 `xml:"laneSet"`
}

type LaneSet struct {
	XMLName xml.Name `xml:"laneSet"`
	ID      string   `xml:"id,attr"`
	Lanes   []Lane   `xml:"lane"`
}

type Lane struct {
	XMLName      xml.Name `xml:"lane"`
	ID           string   `xml:"id,attr"`
	Name         string   `xml:"name,attr"`
	FlowNodeRefs []string `xml:"flowNodeRef"`
}

// Flow node base types

type FlowNode struct {
	ID       string   `xml:"id,attr"`
	Name     string   `xml:"name,attr"`
	Incoming []string `xml:"incoming"`
	Outgoing []string `xml:"outgoing"`
}

type StartEvent struct {
	XMLName xml.Name `xml:"startEvent"`
	FlowNode
	TimerEventDefinition   *TimerEventDefinition   `xml:"timerEventDefinition"`
	MessageEventDefinition *MessageEventDefinition `xml:"messageEventDefinition"`
	SignalEventDefinition  *SignalEventDefinition  `xml:"signalEventDefinition"`
}

type EndEvent struct {
	XMLName xml.Name `xml:"endEvent"`
	FlowNode
	TerminateEventDefinition *TerminateEventDefinition `xml:"terminateEventDefinition"`
	ErrorEventDefinition     *ErrorEventDefinition     `xml:"errorEventDefinition"`
	MessageEventDefinition   *MessageEventDefinition   `xml:"messageEventDefinition"`
	SignalEventDefinition    *SignalEventDefinition    `xml:"signalEventDefinition"`
}

type Task struct {
	XMLName xml.Name `xml:"task"`
	FlowNode
}

type UserTask struct {
	XMLName xml.Name `xml:"userTask"`
	FlowNode
}

type ServiceTask struct {
	XMLName xml.Name `xml:"serviceTask"`
	FlowNode
}

type ScriptTask struct {
	XMLName xml.Name `xml:"scriptTask"`
	FlowNode
}

type SendTask struct {
	XMLName xml.Name `xml:"sendTask"`
	FlowNode
}

type ReceiveTask struct {
	XMLName xml.Name `xml:"receiveTask"`
	FlowNode
}

type ManualTask struct {
	XMLName xml.Name `xml:"manualTask"`
	FlowNode
}

type BusinessRuleTask struct {
	XMLName xml.Name `xml:"businessRuleTask"`
	FlowNode
}

type SubProcess struct {
	XMLName xml.Name `xml:"subProcess"`
	FlowNode
}

type ExclusiveGateway struct {
	XMLName xml.Name `xml:"exclusiveGateway"`
	FlowNode
	Default string `xml:"default,attr"`
}

type ParallelGateway struct {
	XMLName xml.Name `xml:"parallelGateway"`
	FlowNode
}

type InclusiveGateway struct {
	XMLName xml.Name `xml:"inclusiveGateway"`
	FlowNode
	Default string `xml:"default,attr"`
}

type EventBasedGateway struct {
	XMLName xml.Name `xml:"eventBasedGateway"`
	FlowNode
}

type IntermediateCatchEvent struct {
	XMLName xml.Name `xml:"intermediateCatchEvent"`
	FlowNode
	TimerEventDefinition   *TimerEventDefinition   `xml:"timerEventDefinition"`
	MessageEventDefinition *MessageEventDefinition `xml:"messageEventDefinition"`
	SignalEventDefinition  *SignalEventDefinition  `xml:"signalEventDefinition"`
}

type IntermediateThrowEvent struct {
	XMLName xml.Name `xml:"intermediateThrowEvent"`
	FlowNode
	MessageEventDefinition *MessageEventDefinition `xml:"messageEventDefinition"`
	SignalEventDefinition  *SignalEventDefinition  `xml:"signalEventDefinition"`
}

type BoundaryEvent struct {
	XMLName xml.Name `xml:"boundaryEvent"`
	FlowNode
	AttachedToRef          string                  `xml:"attachedToRef,attr"`
	CancelActivity         bool                    `xml:"cancelActivity,attr"`
	TimerEventDefinition   *TimerEventDefinition   `xml:"timerEventDefinition"`
	ErrorEventDefinition   *ErrorEventDefinition   `xml:"errorEventDefinition"`
	MessageEventDefinition *MessageEventDefinition `xml:"messageEventDefinition"`
	SignalEventDefinition  *SignalEventDefinition  `xml:"signalEventDefinition"`
}

// Event definitions

type TimerEventDefinition struct {
	XMLName xml.Name `xml:"timerEventDefinition"`
}

type MessageEventDefinition struct {
	XMLName xml.Name `xml:"messageEventDefinition"`
}

type SignalEventDefinition struct {
	XMLName xml.Name `xml:"signalEventDefinition"`
}

type TerminateEventDefinition struct {
	XMLName xml.Name `xml:"terminateEventDefinition"`
}

type ErrorEventDefinition struct {
	XMLName xml.Name `xml:"errorEventDefinition"`
}

// Sequence flows and data

type SequenceFlow struct {
	XMLName             xml.Name             `xml:"sequenceFlow"`
	ID                  string               `xml:"id,attr"`
	Name                string               `xml:"name,attr"`
	SourceRef           string               `xml:"sourceRef,attr"`
	TargetRef           string               `xml:"targetRef,attr"`
	ConditionExpression *ConditionExpression `xml:"conditionExpression"`
}

type ConditionExpression struct {
	XMLName xml.Name `xml:"conditionExpression"`
	Type    string   `xml:"type,attr"`
	Body    string   `xml:",chardata"`
}

type DataObject struct {
	XMLName xml.Name `xml:"dataObject"`
	ID      string   `xml:"id,attr"`
	Name    string   `xml:"name,attr"`
}

type TextAnnotation struct {
	XMLName xml.Name `xml:"textAnnotation"`
	ID      string   `xml:"id,attr"`
	Text    string   `xml:"text"`
}

type Association struct {
	XMLName   xml.Name `xml:"association"`
	ID        string   `xml:"id,attr"`
	SourceRef string   `xml:"sourceRef,attr"`
	TargetRef string   `xml:"targetRef,attr"`
}

// BPMN Diagram Interchange (DI)

type BPMNDiagram struct {
	XMLName xml.Name  `xml:"BPMNDiagram"`
	ID      string    `xml:"id,attr"`
	Plane   BPMNPlane `xml:"BPMNPlane"`
}

type BPMNPlane struct {
	XMLName     xml.Name    `xml:"BPMNPlane"`
	ID          string      `xml:"id,attr"`
	BpmnElement string      `xml:"bpmnElement,attr"`
	Shapes      []BPMNShape `xml:"BPMNShape"`
	Edges       []BPMNEdge  `xml:"BPMNEdge"`
}

type BPMNShape struct {
	XMLName      xml.Name `xml:"BPMNShape"`
	ID           string   `xml:"id,attr"`
	BpmnElement  string   `xml:"bpmnElement,attr"`
	IsHorizontal *bool    `xml:"isHorizontal,attr"`
	IsExpanded   *bool    `xml:"isExpanded,attr"`
	Bounds       Bounds   `xml:"Bounds"`
	Label        *Label   `xml:"BPMNLabel"`
}

type BPMNEdge struct {
	XMLName     xml.Name   `xml:"BPMNEdge"`
	ID          string     `xml:"id,attr"`
	BpmnElement string     `xml:"bpmnElement,attr"`
	Waypoints   []Waypoint `xml:"waypoint"`
	Label       *Label     `xml:"BPMNLabel"`
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
}
