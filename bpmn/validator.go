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
		errors = append(errors, validateProcess(&proc)...)
	}

	if len(defs.Diagrams) == 0 {
		errors = append(errors, ValidationError{
			Severity: "warning",
			Message:  "no diagram interchange (DI) data found; rendering may not be possible",
		})
	}

	return errors
}

func validateProcess(proc *Process) []ValidationError {
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
	for _, sf := range proc.SequenceFlows {
		if sf.ID == "" {
			errors = append(errors, ValidationError{
				Severity: "error",
				Message:  "sequence flow has no ID",
			})
			continue
		}
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

	// Validate gateway default flows
	for _, gw := range proc.ExclusiveGateways {
		if gw.Default != "" {
			found := false
			for _, sf := range proc.SequenceFlows {
				if sf.ID == gw.Default {
					found = true
					break
				}
			}
			if !found {
				errors = append(errors, ValidationError{
					Severity: "error",
					Message:  fmt.Sprintf("exclusive gateway '%s' references unknown default flow '%s'", gw.ID, gw.Default),
					Element:  gw.ID,
				})
			}
		}
	}

	return errors
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
