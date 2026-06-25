package bpmn

import "fmt"

// ValidationError represents a single validation issue.
type ValidationError struct {
	Severity string // "error" or "warning"
	Message  string
	Element  string // ID of the problematic element
}

func (e ValidationError) String() string {
	if e.Element != "" {
		return fmt.Sprintf("[%s] %s (element: %s)", e.Severity, e.Message, e.Element)
	}
	return fmt.Sprintf("[%s] %s", e.Severity, e.Message)
}

// Validate checks the parsed BPMN definitions for common errors.
func Validate(defs *Definitions) []ValidationError {
	var errors []ValidationError

	if len(defs.Processes) == 0 {
		errors = append(errors, ValidationError{
			Severity: "error",
			Message:  "no process defined in BPMN file",
		})
		return errors
	}

	for _, proc := range defs.Processes {
		errors = append(errors, validateProcess(&proc, defs)...)
	}

	// Validate collaboration references
	for _, collab := range defs.Collaborations {
		errors = append(errors, validateCollaboration(&collab, defs)...)
	}

	// Validate DI completeness
	errors = append(errors, validateDI(defs)...)

	return errors
}

func validateProcess(proc *Process, defs *Definitions) []ValidationError {
	var errors []ValidationError

	if proc.ID == "" {
		errors = append(errors, ValidationError{
			Severity: "error",
			Message:  "process has no ID",
		})
	}

	// Check for start events
	if len(proc.StartEvents) == 0 {
		errors = append(errors, ValidationError{
			Severity: "warning",
			Message:  fmt.Sprintf("process '%s' has no start event", proc.ID),
			Element:  proc.ID,
		})
	}

	// Check for end events
	if len(proc.EndEvents) == 0 {
		errors = append(errors, ValidationError{
			Severity: "warning",
			Message:  fmt.Sprintf("process '%s' has no end event", proc.ID),
			Element:  proc.ID,
		})
	}

	// Validate sequence flow references
	nodes := proc.GetAllFlowNodes()
	flowIDs := make(map[string]bool)
	for _, sf := range proc.SequenceFlows {
		if sf.ID == "" {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  "sequence flow has no ID",
			})
			continue
		}
		if flowIDs[sf.ID] {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  fmt.Sprintf("duplicate sequence flow ID '%s'", sf.ID),
				Element:  sf.ID,
			})
		}
		flowIDs[sf.ID] = true

		if sf.SourceRef == "" {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  fmt.Sprintf("sequence flow '%s' has no sourceRef", sf.ID),
				Element:  sf.ID,
			})
		} else if _, ok := nodes[sf.SourceRef]; !ok {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  fmt.Sprintf("sequence flow '%s' references unknown source '%s'", sf.ID, sf.SourceRef),
				Element:  sf.ID,
			})
		}
		if sf.TargetRef == "" {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  fmt.Sprintf("sequence flow '%s' has no targetRef", sf.ID),
				Element:  sf.ID,
			})
		} else if _, ok := nodes[sf.TargetRef]; !ok {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  fmt.Sprintf("sequence flow '%s' references unknown target '%s'", sf.ID, sf.TargetRef),
				Element:  sf.ID,
			})
		}
	}

	// Check for duplicate node IDs
	nodeIDCount := make(map[string]int)
	for id := range nodes {
		nodeIDCount[id]++
	}
	allElements := collectAllIDs(proc)
	seen := make(map[string]bool)
	for _, id := range allElements {
		if id == "" {
			continue
		}
		if seen[id] {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  fmt.Sprintf("duplicate element ID '%s' in process '%s'", id, proc.ID),
				Element:  id,
			})
		}
		seen[id] = true
	}

	// Validate gateway default flows
	for _, gw := range proc.ExclusiveGateways {
		if gw.Default != "" {
			if !flowIDs[gw.Default] {
				errors = append(errors, ValidationError{
					Severity: "error",
					Message:  fmt.Sprintf("exclusive gateway '%s' references unknown default flow '%s'", gw.ID, gw.Default),
					Element:  gw.ID,
				})
			}
		}
	}
	for _, gw := range proc.InclusiveGateways {
		if gw.Default != "" {
			if !flowIDs[gw.Default] {
				errors = append(errors, ValidationError{
					Severity: "error",
					Message:  fmt.Sprintf("inclusive gateway '%s' references unknown default flow '%s'", gw.ID, gw.Default),
					Element:  gw.ID,
				})
			}
		}
	}

	// Validate gateway connectivity (should have at least one incoming and one outgoing for merge/split)
	for _, gw := range proc.ExclusiveGateways {
		errors = append(errors, checkGatewayConnectivity(gw.FlowNode, "exclusive gateway", proc)...)
	}
	for _, gw := range proc.ParallelGateways {
		errors = append(errors, checkGatewayConnectivity(gw.FlowNode, "parallel gateway", proc)...)
	}
	for _, gw := range proc.InclusiveGateways {
		errors = append(errors, checkGatewayConnectivity(gw.FlowNode, "inclusive gateway", proc)...)
	}
	for _, gw := range proc.EventBasedGateways {
		errors = append(errors, checkGatewayConnectivity(gw.FlowNode, "event-based gateway", proc)...)
	}

	// Validate boundary events reference existing tasks/subprocesses
	for _, be := range proc.BoundaryEvents {
		if be.AttachedToRef == "" {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  fmt.Sprintf("boundary event '%s' has no attachedToRef", be.ID),
				Element:  be.ID,
			})
		} else if _, ok := nodes[be.AttachedToRef]; !ok {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  fmt.Sprintf("boundary event '%s' is attached to unknown element '%s'", be.ID, be.AttachedToRef),
				Element:  be.ID,
			})
		}
	}

	// Check for disconnected nodes (no incoming or outgoing sequence flows)
	errors = append(errors, checkDisconnectedNodes(proc)...)

	// Validate lane references
	if proc.LaneSet != nil {
		for _, lane := range proc.LaneSet.Lanes {
			for _, ref := range lane.FlowNodeRefs {
				if _, ok := nodes[ref]; !ok {
					errors = append(errors, ValidationError{
						Severity: "warning",
						Message:  fmt.Sprintf("lane '%s' references unknown flow node '%s'", lane.ID, ref),
						Element:  lane.ID,
					})
				}
			}
		}
	}

	return errors
}

// checkGatewayConnectivity validates that a gateway has incoming and outgoing flows.
func checkGatewayConnectivity(node FlowNode, kind string, proc *Process) []ValidationError {
	var errors []ValidationError

	incoming := 0
	outgoing := 0
	for _, sf := range proc.SequenceFlows {
		if sf.TargetRef == node.ID {
			incoming++
		}
		if sf.SourceRef == node.ID {
			outgoing++
		}
	}

	if incoming == 0 {
		errors = append(errors, ValidationError{
			Severity: "warning",
			Message:  fmt.Sprintf("%s '%s' has no incoming sequence flows", kind, node.ID),
			Element:  node.ID,
		})
	}
	if outgoing == 0 {
		errors = append(errors, ValidationError{
			Severity: "warning",
			Message:  fmt.Sprintf("%s '%s' has no outgoing sequence flows", kind, node.ID),
			Element:  node.ID,
		})
	}

	return errors
}

// checkDisconnectedNodes finds flow nodes with no incoming and no outgoing flows
// (excluding start/end events which naturally lack one side).
func checkDisconnectedNodes(proc *Process) []ValidationError {
	var errors []ValidationError

	// Build sets of nodes that are sources or targets of flows
	hasIncoming := make(map[string]bool)
	hasOutgoing := make(map[string]bool)
	for _, sf := range proc.SequenceFlows {
		hasIncoming[sf.TargetRef] = true
		hasOutgoing[sf.SourceRef] = true
	}

	// Tasks, gateways, intermediate events must have at least one connection
	type namedNode struct {
		id   string
		kind string
	}
	var checkNodes []namedNode
	for _, t := range proc.Tasks {
		checkNodes = append(checkNodes, namedNode{t.ID, "task"})
	}
	for _, t := range proc.UserTasks {
		checkNodes = append(checkNodes, namedNode{t.ID, "user task"})
	}
	for _, t := range proc.ServiceTasks {
		checkNodes = append(checkNodes, namedNode{t.ID, "service task"})
	}
	for _, t := range proc.ScriptTasks {
		checkNodes = append(checkNodes, namedNode{t.ID, "script task"})
	}
	for _, t := range proc.SendTasks {
		checkNodes = append(checkNodes, namedNode{t.ID, "send task"})
	}
	for _, t := range proc.ReceiveTasks {
		checkNodes = append(checkNodes, namedNode{t.ID, "receive task"})
	}
	for _, t := range proc.ManualTasks {
		checkNodes = append(checkNodes, namedNode{t.ID, "manual task"})
	}
	for _, t := range proc.BusinessRuleTasks {
		checkNodes = append(checkNodes, namedNode{t.ID, "business rule task"})
	}
	for _, t := range proc.SubProcesses {
		checkNodes = append(checkNodes, namedNode{t.ID, "subprocess"})
	}
	for _, t := range proc.IntermediateCatchEvents {
		checkNodes = append(checkNodes, namedNode{t.ID, "intermediate catch event"})
	}
	for _, t := range proc.IntermediateThrowEvents {
		checkNodes = append(checkNodes, namedNode{t.ID, "intermediate throw event"})
	}

	for _, n := range checkNodes {
		if n.id == "" {
			continue
		}
		if !hasIncoming[n.id] && !hasOutgoing[n.id] {
			errors = append(errors, ValidationError{
				Severity: "warning",
				Message:  fmt.Sprintf("%s '%s' is completely disconnected (no sequence flows)", n.kind, n.id),
				Element:  n.id,
			})
		}
	}

	return errors
}

// validateCollaboration checks participant and message flow references.
func validateCollaboration(collab *Collaboration, defs *Definitions) []ValidationError {
	var errors []ValidationError

	processIDs := make(map[string]bool)
	for _, p := range defs.Processes {
		processIDs[p.ID] = true
	}

	for _, part := range collab.Participants {
		if part.ID == "" {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  "participant has no ID",
			})
			continue
		}
		if part.ProcessRef != "" && !processIDs[part.ProcessRef] {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  fmt.Sprintf("participant '%s' references unknown process '%s'", part.ID, part.ProcessRef),
				Element:  part.ID,
			})
		}
	}

	// Validate message flows reference valid participants or flow nodes
	allNodes := make(map[string]bool)
	for _, p := range defs.Processes {
		for id := range p.GetAllFlowNodes() {
			allNodes[id] = true
		}
	}
	for _, part := range collab.Participants {
		allNodes[part.ID] = true
	}
	for _, mf := range collab.MessageFlows {
		if mf.ID == "" {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  "message flow has no ID",
			})
			continue
		}
		if mf.SourceRef == "" {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  fmt.Sprintf("message flow '%s' has no sourceRef", mf.ID),
				Element:  mf.ID,
			})
		} else if !allNodes[mf.SourceRef] {
			errors = append(errors, ValidationError{
				Severity: "warning",
				Message:  fmt.Sprintf("message flow '%s' references unknown source '%s'", mf.ID, mf.SourceRef),
				Element:  mf.ID,
			})
		}
		if mf.TargetRef == "" {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  fmt.Sprintf("message flow '%s' has no targetRef", mf.ID),
				Element:  mf.ID,
			})
		} else if !allNodes[mf.TargetRef] {
			errors = append(errors, ValidationError{
				Severity: "warning",
				Message:  fmt.Sprintf("message flow '%s' references unknown target '%s'", mf.ID, mf.TargetRef),
				Element:  mf.ID,
			})
		}
	}

	return errors
}

// validateDI checks that diagram interchange data covers all process elements.
func validateDI(defs *Definitions) []ValidationError {
	var errors []ValidationError

	if len(defs.Diagrams) == 0 {
		errors = append(errors, ValidationError{
			Severity: "warning",
			Message:  "no diagram interchange (DI) data found; rendering may not be possible",
		})
		return errors
	}

	// Collect all shape references from DI
	diShapeRefs := make(map[string]bool)
	diEdgeRefs := make(map[string]bool)
	for _, diag := range defs.Diagrams {
		for _, shape := range diag.Plane.Shapes {
			diShapeRefs[shape.BpmnElement] = true
			// Check for zero-size bounds
			if shape.Bounds.Width <= 0 || shape.Bounds.Height <= 0 {
				errors = append(errors, ValidationError{
					Severity: "warning",
					Message:  fmt.Sprintf("DI shape for element '%s' has zero or negative dimensions (w=%.0f, h=%.0f)", shape.BpmnElement, shape.Bounds.Width, shape.Bounds.Height),
					Element:  shape.BpmnElement,
				})
			}
		}
		for _, edge := range diag.Plane.Edges {
			diEdgeRefs[edge.BpmnElement] = true
			// Edges need at least 2 waypoints
			if len(edge.Waypoints) < 2 {
				errors = append(errors, ValidationError{
					Severity: "warning",
					Message:  fmt.Sprintf("DI edge for element '%s' has fewer than 2 waypoints (%d)", edge.BpmnElement, len(edge.Waypoints)),
					Element:  edge.BpmnElement,
				})
			}
		}
	}

	// Check each flow node has a corresponding DI shape
	for _, proc := range defs.Processes {
		nodes := proc.GetAllFlowNodes()
		for id, kind := range nodes {
			if !diShapeRefs[id] {
				errors = append(errors, ValidationError{
					Severity: "warning",
					Message:  fmt.Sprintf("%s '%s' has no diagram shape (DI); it will not be rendered", kind, id),
					Element:  id,
				})
			}
		}
		// Check each sequence flow has a corresponding DI edge
		for _, sf := range proc.SequenceFlows {
			if sf.ID != "" && !diEdgeRefs[sf.ID] {
				errors = append(errors, ValidationError{
					Severity: "warning",
					Message:  fmt.Sprintf("sequence flow '%s' has no diagram edge (DI); it will not be rendered", sf.ID),
					Element:  sf.ID,
				})
			}
		}
	}

	return errors
}

// collectAllIDs returns all element IDs in a process for duplicate detection.
func collectAllIDs(proc *Process) []string {
	var ids []string
	for _, e := range proc.StartEvents {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.EndEvents {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.Tasks {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.UserTasks {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.ServiceTasks {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.ScriptTasks {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.SendTasks {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.ReceiveTasks {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.ManualTasks {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.BusinessRuleTasks {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.SubProcesses {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.CallActivities {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.ExclusiveGateways {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.ParallelGateways {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.InclusiveGateways {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.EventBasedGateways {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.IntermediateCatchEvents {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.IntermediateThrowEvents {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.BoundaryEvents {
		ids = append(ids, e.ID)
	}
	for _, e := range proc.SequenceFlows {
		ids = append(ids, e.ID)
	}
	return ids
}

// HasErrors returns true if any validation error has severity "error".
func HasErrors(errs []ValidationError) bool {
	for _, e := range errs {
		if e.Severity == "error" {
			return true
		}
	}
	return false
}
