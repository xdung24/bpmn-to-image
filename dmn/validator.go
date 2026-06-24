package dmn

import "fmt"

// ValidationError is a single validation issue.
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

// HasErrors returns true if any item has severity "error".
func HasErrors(errs []ValidationError) bool {
	for _, e := range errs {
		if e.Severity == "error" {
			return true
		}
	}
	return false
}

// Validate performs a structural check on the DMN definitions.
func Validate(defs *Definitions) []ValidationError {
	var out []ValidationError

	nodes := defs.BuildNodeIndex()
	if len(nodes) == 0 {
		out = append(out, ValidationError{Severity: "error", Message: "DMN file contains no DRG elements (decision/inputData/BKM/knowledgeSource)"})
		return out
	}

	// Check for duplicate IDs
	out = append(out, checkDuplicateIDs(defs)...)

	// Validate decisions
	for _, d := range defs.Decisions {
		if d.ID == "" {
			out = append(out, ValidationError{Severity: "error", Message: "decision has no ID"})
			continue
		}
		if d.Name == "" {
			out = append(out, ValidationError{
				Severity: "warning",
				Message:  fmt.Sprintf("decision '%s' has no name", d.ID),
				Element:  d.ID,
			})
		}
		for _, r := range d.InformationRequirements {
			out = append(out, checkRef(d.ID, "informationRequirement", r.ID, nodes, r.RequiredDecision, r.RequiredInput)...)
		}
		for _, r := range d.KnowledgeRequirements {
			out = append(out, checkRef(d.ID, "knowledgeRequirement", r.ID, nodes, r.RequiredKnowledge)...)
		}
		for _, r := range d.AuthorityRequirements {
			out = append(out, checkRef(d.ID, "authorityRequirement", r.ID, nodes, r.RequiredAuthority, r.RequiredDecision, r.RequiredInput)...)
		}
	}

	// Validate input data
	for _, inp := range defs.InputDatas {
		if inp.ID == "" {
			out = append(out, ValidationError{Severity: "error", Message: "inputData has no ID"})
			continue
		}
		if inp.Name == "" {
			out = append(out, ValidationError{
				Severity: "warning",
				Message:  fmt.Sprintf("inputData '%s' has no name", inp.ID),
				Element:  inp.ID,
			})
		}
	}

	// Validate BKMs
	for _, b := range defs.BusinessKnowledgeModels {
		if b.ID == "" {
			out = append(out, ValidationError{Severity: "error", Message: "BKM has no ID"})
			continue
		}
		if b.Name == "" {
			out = append(out, ValidationError{
				Severity: "warning",
				Message:  fmt.Sprintf("businessKnowledgeModel '%s' has no name", b.ID),
				Element:  b.ID,
			})
		}
		for _, r := range b.KnowledgeRequirements {
			out = append(out, checkRef(b.ID, "knowledgeRequirement", r.ID, nodes, r.RequiredKnowledge)...)
		}
		for _, r := range b.AuthorityRequirements {
			out = append(out, checkRef(b.ID, "authorityRequirement", r.ID, nodes, r.RequiredAuthority, r.RequiredDecision, r.RequiredInput)...)
		}
	}

	// Validate knowledge sources
	for _, k := range defs.KnowledgeSources {
		if k.ID == "" {
			out = append(out, ValidationError{Severity: "error", Message: "knowledgeSource has no ID"})
			continue
		}
		if k.Name == "" {
			out = append(out, ValidationError{
				Severity: "warning",
				Message:  fmt.Sprintf("knowledgeSource '%s' has no name", k.ID),
				Element:  k.ID,
			})
		}
		for _, r := range k.AuthorityRequirements {
			out = append(out, checkRef(k.ID, "authorityRequirement", r.ID, nodes, r.RequiredAuthority, r.RequiredDecision, r.RequiredInput)...)
		}
	}

	// Validate decision services
	for _, ds := range defs.DecisionServices {
		if ds.ID == "" {
			out = append(out, ValidationError{Severity: "error", Message: "decisionService has no ID"})
			continue
		}
		// Validate output decision refs
		for _, href := range ds.OutputDecisions {
			target := StripHref(href.Href)
			if target != "" {
				if info, ok := nodes[target]; !ok {
					out = append(out, ValidationError{
						Severity: "error",
						Message:  fmt.Sprintf("decisionService '%s' outputDecision references unknown element '%s'", ds.ID, target),
						Element:  ds.ID,
					})
				} else if info.Kind != KindDecision {
					out = append(out, ValidationError{
						Severity: "warning",
						Message:  fmt.Sprintf("decisionService '%s' outputDecision references '%s' which is a %s, not a decision", ds.ID, target, info.Kind),
						Element:  ds.ID,
					})
				}
			}
		}
		// Validate encapsulated decision refs
		for _, href := range ds.EncapsulatedDecs {
			target := StripHref(href.Href)
			if target != "" {
				if _, ok := nodes[target]; !ok {
					out = append(out, ValidationError{
						Severity: "error",
						Message:  fmt.Sprintf("decisionService '%s' encapsulatedDecision references unknown element '%s'", ds.ID, target),
						Element:  ds.ID,
					})
				}
			}
		}
		// Validate input decision refs
		for _, href := range ds.InputDecisions {
			target := StripHref(href.Href)
			if target != "" {
				if _, ok := nodes[target]; !ok {
					out = append(out, ValidationError{
						Severity: "error",
						Message:  fmt.Sprintf("decisionService '%s' inputDecision references unknown element '%s'", ds.ID, target),
						Element:  ds.ID,
					})
				}
			}
		}
		// Validate input data refs
		for _, href := range ds.InputData {
			target := StripHref(href.Href)
			if target != "" {
				if _, ok := nodes[target]; !ok {
					out = append(out, ValidationError{
						Severity: "error",
						Message:  fmt.Sprintf("decisionService '%s' inputData references unknown element '%s'", ds.ID, target),
						Element:  ds.ID,
					})
				}
			}
		}
	}

	// Validate associations
	for _, a := range defs.Associations {
		if a.ID == "" {
			out = append(out, ValidationError{Severity: "warning", Message: "association has no ID"})
			continue
		}
		srcID := StripHref(a.SourceRef.Href)
		tgtID := StripHref(a.TargetRef.Href)
		if srcID == "" {
			out = append(out, ValidationError{
				Severity: "error",
				Message:  fmt.Sprintf("association '%s' has no sourceRef", a.ID),
				Element:  a.ID,
			})
		} else if _, ok := nodes[srcID]; !ok {
			out = append(out, ValidationError{
				Severity: "error",
				Message:  fmt.Sprintf("association '%s' sourceRef references unknown element '%s'", a.ID, srcID),
				Element:  a.ID,
			})
		}
		if tgtID == "" {
			out = append(out, ValidationError{
				Severity: "error",
				Message:  fmt.Sprintf("association '%s' has no targetRef", a.ID),
				Element:  a.ID,
			})
		} else if _, ok := nodes[tgtID]; !ok {
			out = append(out, ValidationError{
				Severity: "error",
				Message:  fmt.Sprintf("association '%s' targetRef references unknown element '%s'", a.ID, tgtID),
				Element:  a.ID,
			})
		}
	}

	// Check for unreferenced input data (no decision depends on them)
	out = append(out, checkUnreferencedInputs(defs, nodes)...)

	// Validate DI
	out = append(out, validateDMNDI(defs, nodes)...)

	return out
}

// checkDuplicateIDs verifies all DRG element IDs are unique.
func checkDuplicateIDs(defs *Definitions) []ValidationError {
	var out []ValidationError
	seen := make(map[string]string) // id -> kind

	type idEntry struct {
		id   string
		kind string
	}
	var entries []idEntry
	for _, d := range defs.Decisions {
		entries = append(entries, idEntry{d.ID, "decision"})
	}
	for _, d := range defs.InputDatas {
		entries = append(entries, idEntry{d.ID, "inputData"})
	}
	for _, d := range defs.BusinessKnowledgeModels {
		entries = append(entries, idEntry{d.ID, "businessKnowledgeModel"})
	}
	for _, d := range defs.KnowledgeSources {
		entries = append(entries, idEntry{d.ID, "knowledgeSource"})
	}
	for _, d := range defs.DecisionServices {
		entries = append(entries, idEntry{d.ID, "decisionService"})
	}
	for _, d := range defs.TextAnnotations {
		entries = append(entries, idEntry{d.ID, "textAnnotation"})
	}

	for _, e := range entries {
		if e.id == "" {
			continue
		}
		if prevKind, exists := seen[e.id]; exists {
			out = append(out, ValidationError{
				Severity: "error",
				Message:  fmt.Sprintf("duplicate element ID '%s' (found as %s and %s)", e.id, prevKind, e.kind),
				Element:  e.id,
			})
		}
		seen[e.id] = e.kind
	}
	return out
}

// checkUnreferencedInputs finds inputData elements not referenced by any decision.
func checkUnreferencedInputs(defs *Definitions, nodes NodeIndex) []ValidationError {
	var out []ValidationError

	// Collect all referenced input IDs
	referenced := make(map[string]bool)
	for _, d := range defs.Decisions {
		for _, r := range d.InformationRequirements {
			if r.RequiredInput != nil && r.RequiredInput.Href != "" {
				referenced[StripHref(r.RequiredInput.Href)] = true
			}
		}
	}
	for _, ds := range defs.DecisionServices {
		for _, href := range ds.InputData {
			referenced[StripHref(href.Href)] = true
		}
	}

	for _, inp := range defs.InputDatas {
		if inp.ID != "" && !referenced[inp.ID] {
			out = append(out, ValidationError{
				Severity: "warning",
				Message:  fmt.Sprintf("inputData '%s' is not referenced by any decision or decision service", inp.ID),
				Element:  inp.ID,
			})
		}
	}
	return out
}

// validateDMNDI checks the diagram interchange data for completeness and correctness.
func validateDMNDI(defs *Definitions, nodes NodeIndex) []ValidationError {
	var out []ValidationError

	if defs.DMNDI == nil || len(defs.DMNDI.Diagrams) == 0 {
		out = append(out, ValidationError{
			Severity: "warning",
			Message:  "no diagram interchange (DMNDI) data found; rendering may not be possible",
		})
		return out
	}

	// Collect DI shape references
	diShapeRefs := make(map[string]bool)
	diEdgeRefs := make(map[string]bool)
	for _, diag := range defs.DMNDI.Diagrams {
		for _, shape := range diag.Shapes {
			diShapeRefs[shape.DMNElementRef] = true
			// Check for zero-size bounds
			if shape.Bounds.Width <= 0 || shape.Bounds.Height <= 0 {
				out = append(out, ValidationError{
					Severity: "warning",
					Message:  fmt.Sprintf("DI shape for element '%s' has zero or negative dimensions (w=%.0f, h=%.0f)", shape.DMNElementRef, shape.Bounds.Width, shape.Bounds.Height),
					Element:  shape.DMNElementRef,
				})
			}
		}
		for _, edge := range diag.Edges {
			diEdgeRefs[edge.DMNElementRef] = true
			if len(edge.Waypoints) < 2 {
				out = append(out, ValidationError{
					Severity: "warning",
					Message:  fmt.Sprintf("DI edge for element '%s' has fewer than 2 waypoints (%d)", edge.DMNElementRef, len(edge.Waypoints)),
					Element:  edge.DMNElementRef,
				})
			}
		}
	}

	// Check each DRG node has a corresponding DI shape
	for id, info := range nodes {
		if !diShapeRefs[id] {
			out = append(out, ValidationError{
				Severity: "warning",
				Message:  fmt.Sprintf("%s '%s' has no diagram shape (DI); it will not be rendered", info.Kind, id),
				Element:  id,
			})
		}
	}

	return out
}

func checkRef(ownerID, kind, reqID string, nodes NodeIndex, hrefs ...*Href) []ValidationError {
	var out []ValidationError
	found := false
	for _, h := range hrefs {
		if h == nil || h.Href == "" {
			continue
		}
		found = true
		target := StripHref(h.Href)
		if _, ok := nodes[target]; !ok {
			out = append(out, ValidationError{
				Severity: "error",
				Message:  fmt.Sprintf("%s of '%s' references unknown element '%s'", kind, ownerID, target),
				Element:  reqID,
			})
		}
	}
	if !found {
		out = append(out, ValidationError{
			Severity: "error",
			Message:  fmt.Sprintf("%s on '%s' has no required* reference", kind, ownerID),
			Element:  reqID,
		})
	}
	return out
}
