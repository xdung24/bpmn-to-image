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

// Validate performs a light structural check on the DMN definitions.
func Validate(defs *Definitions) []ValidationError {
	var out []ValidationError

	nodes := defs.BuildNodeIndex()
	if len(nodes) == 0 {
		out = append(out, ValidationError{Severity: "error", Message: "DMN file contains no DRG elements (decision/inputData/BKM/knowledgeSource)"})
		return out
	}

	// Validate each requirement references a node that exists.
	for _, d := range defs.Decisions {
		if d.ID == "" {
			out = append(out, ValidationError{Severity: "error", Message: "decision has no ID"})
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
	for _, b := range defs.BusinessKnowledgeModels {
		if b.ID == "" {
			out = append(out, ValidationError{Severity: "error", Message: "BKM has no ID"})
		}
		for _, r := range b.KnowledgeRequirements {
			out = append(out, checkRef(b.ID, "knowledgeRequirement", r.ID, nodes, r.RequiredKnowledge)...)
		}
		for _, r := range b.AuthorityRequirements {
			out = append(out, checkRef(b.ID, "authorityRequirement", r.ID, nodes, r.RequiredAuthority, r.RequiredDecision, r.RequiredInput)...)
		}
	}
	for _, k := range defs.KnowledgeSources {
		for _, r := range k.AuthorityRequirements {
			out = append(out, checkRef(k.ID, "authorityRequirement", r.ID, nodes, r.RequiredAuthority, r.RequiredDecision, r.RequiredInput)...)
		}
	}

	if defs.DMNDI == nil || len(defs.DMNDI.Diagrams) == 0 {
		out = append(out, ValidationError{
			Severity: "warning",
			Message:  "no diagram interchange (DMNDI) data found; rendering may not be possible",
		})
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
