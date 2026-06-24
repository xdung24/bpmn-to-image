package renderer

// noDIMessage is the detailed error shown when a BPMN file has no diagram
// interchange (BPMNDiagram) section. This means the file defines process
// logic but lacks the visual coordinates needed for rendering.
const noDIMessage = `no diagram layout data found — cannot render

Your BPMN file defines the process logic (tasks, gateways, sequence flows)
but is missing the <bpmndi:BPMNDiagram> section which provides the visual
layout (x/y coordinates, dimensions, and edge waypoints for every element).

Without this data, the renderer does not know where to place elements on the
canvas.

How to fix:
  1. Open the file in a BPMN editor (Camunda Modeler, bpmn.io, Signavio, etc.)
     — the editor will auto-layout the diagram and save the DI section.
  2. Or add a <bpmndi:BPMNDiagram> block manually with <bpmndi:BPMNShape>
     entries for each node and <bpmndi:BPMNEdge> entries for each flow.

Example of a minimal DI section:

  <bpmndi:BPMNDiagram id="Diagram_1">
    <bpmndi:BPMNPlane bpmnElement="Process_1">
      <bpmndi:BPMNShape bpmnElement="StartEvent_1">
        <dc:Bounds x="100" y="200" width="36" height="36"/>
      </bpmndi:BPMNShape>
      <bpmndi:BPMNEdge bpmnElement="Flow_1">
        <di:waypoint x="136" y="218"/>
        <di:waypoint x="200" y="218"/>
      </bpmndi:BPMNEdge>
    </bpmndi:BPMNPlane>
  </bpmndi:BPMNDiagram>`

// noDMNDIMessage is the detailed error shown when a DMN file has no DMNDI section.
const noDMNDIMessage = `no diagram layout data found — cannot render

Your DMN file defines the decision logic (decisions, input data, BKMs,
knowledge sources) but is missing the <dmndi:DMNDI> section which provides
the visual layout (x/y coordinates, dimensions, and edge waypoints).

Without this data, the renderer does not know where to place elements on the
canvas.

How to fix:
  1. Open the file in a DMN editor (Camunda Modeler, Trisotech, etc.)
     — the editor will auto-layout the DRD and save the DMNDI section.
  2. Or add a <dmndi:DMNDI> block manually with <dmndi:DMNShape> entries
     for each node and <dmndi:DMNEdge> entries for each requirement.

Example of a minimal DMNDI section:

  <dmndi:DMNDI>
    <dmndi:DMNDiagram id="Diagram_1">
      <dmndi:DMNShape dmnElementRef="Decision_1">
        <dc:Bounds x="200" y="100" width="180" height="80"/>
      </dmndi:DMNShape>
      <dmndi:DMNEdge dmnElementRef="InformationRequirement_1">
        <di:waypoint x="290" y="300"/>
        <di:waypoint x="290" y="180"/>
      </dmndi:DMNEdge>
    </dmndi:DMNDiagram>
  </dmndi:DMNDI>`
