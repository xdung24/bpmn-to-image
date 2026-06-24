package renderer

import (
	"fmt"
	"math"
	"strings"

	"github.com/xdung24/bpmn-to-image/bpmn"
)

// SVGRenderer generates SVG output from BPMN diagram data.
type SVGRenderer struct {
	padding float64
	theme   Theme
}

// NewSVGRenderer creates a new SVG renderer with default settings.
func NewSVGRenderer() *SVGRenderer {
	return &SVGRenderer{
		padding: 30,
		theme:   DefaultTheme,
	}
}

// Render generates SVG content from the BPMN definitions.
func (r *SVGRenderer) Render(defs *bpmn.Definitions) ([]byte, error) {
	if len(defs.Diagrams) == 0 {
		return nil, fmt.Errorf(noDIMessage)
	}

	diagram := defs.Diagrams[0]
	plane := diagram.Plane

	// Build element type map and name map from processes
	elementTypes := buildElementTypeMap(defs)
	elementNames := BuildElementNameMap(defs)

	// Calculate canvas bounds
	minX, minY, maxX, maxY := r.calculateBounds(&plane)
	width := maxX - minX + 2*r.padding
	height := maxY - minY + 2*r.padding
	offsetX := -minX + r.padding
	offsetY := -minY + r.padding

	var sb strings.Builder
	t := r.theme

	// SVG header
	sb.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f">
<defs>
  <marker id="arrowhead" markerWidth="12" markerHeight="9" refX="9" refY="4.5" orient="auto" markerUnits="userSpaceOnUse">
    <path d="M0,0 L9,4.5 L0,9 L2.5,4.5 Z" fill="%s"/>
  </marker>
  <linearGradient id="grad-task" x1="0" y1="0" x2="0" y2="1">
    <stop offset="0%%" stop-color="#ffffff"/>
    <stop offset="100%%" stop-color="%s"/>
  </linearGradient>
  <linearGradient id="grad-subprocess" x1="0" y1="0" x2="0" y2="1">
    <stop offset="0%%" stop-color="#ffffff"/>
    <stop offset="100%%" stop-color="%s"/>
  </linearGradient>
  <linearGradient id="grad-start" x1="0" y1="0" x2="0" y2="1">
    <stop offset="0%%" stop-color="#ffffff"/>
    <stop offset="100%%" stop-color="%s"/>
  </linearGradient>
  <linearGradient id="grad-end" x1="0" y1="0" x2="0" y2="1">
    <stop offset="0%%" stop-color="#ffffff"/>
    <stop offset="100%%" stop-color="%s"/>
  </linearGradient>
  <linearGradient id="grad-intermediate" x1="0" y1="0" x2="0" y2="1">
    <stop offset="0%%" stop-color="#ffffff"/>
    <stop offset="100%%" stop-color="%s"/>
  </linearGradient>
  <linearGradient id="grad-gateway" x1="0" y1="0" x2="0" y2="1">
    <stop offset="0%%" stop-color="#ffffff"/>
    <stop offset="100%%" stop-color="%s"/>
  </linearGradient>
  <filter id="shadow" x="-20%%" y="-20%%" width="140%%" height="140%%">
    <feDropShadow dx="1.5" dy="2" stdDeviation="2" flood-color="#0f172a" flood-opacity="0.18"/>
  </filter>
</defs>
<style>
  .bpmn-task { fill: url(#grad-task); stroke: %s; stroke-width: 2; filter: url(#shadow); }
  .bpmn-subprocess { fill: url(#grad-subprocess); stroke: %s; stroke-width: 2; filter: url(#shadow); }
  .bpmn-event-start { fill: url(#grad-start); stroke: %s; stroke-width: 2; filter: url(#shadow); }
  .bpmn-event-intermediate { fill: url(#grad-intermediate); stroke: %s; stroke-width: 2; }
  .bpmn-event-end { fill: url(#grad-end); stroke: %s; stroke-width: 3.5; filter: url(#shadow); }
  .bpmn-gateway { fill: url(#grad-gateway); stroke: %s; stroke-width: 2; filter: url(#shadow); }
  .bpmn-gateway-marker { stroke: %s; fill: none; }
  .bpmn-flow { fill: none; stroke: %s; stroke-width: 1.6; marker-end: url(#arrowhead); stroke-linejoin: round; stroke-linecap: round; }
  .bpmn-label { font-family: 'Segoe UI', Arial, sans-serif; font-size: 12px; font-weight: 500; fill: %s; text-anchor: middle; dominant-baseline: central; }
  .bpmn-flow-label { font-family: 'Segoe UI', Arial, sans-serif; font-size: 11px; fill: %s; text-anchor: middle; dominant-baseline: central; }
  .bpmn-pool { fill: %s; stroke: %s; stroke-width: 2; }
  .bpmn-pool-header { fill: %s; stroke: %s; stroke-width: 2; }
  .bpmn-lane { fill: none; stroke: %s; stroke-width: 1; stroke-dasharray: 4,4; }
  .bpmn-annotation { fill: none; stroke: %s; stroke-width: 1; }
  .bpmn-icon { stroke: %s; fill: none; }
</style>
<rect x="0" y="0" width="%.0f" height="%.0f" fill="%s"/>
`, width, height, width, height,
		t.Flow,
		t.TaskFill, t.SubProcessFill, t.StartFill, t.EndFill, t.IntermediateFill, t.GatewayFill,
		t.TaskStroke, t.SubProcessStroke, t.StartStroke, t.IntermediateStroke, t.EndStroke,
		t.GatewayStroke, t.GatewayMarker, t.Flow, t.Label, t.Flow,
		t.PoolFill, t.PoolStroke, t.PoolHeader, t.PoolStroke, t.LaneStroke, t.AnnotationStroke, t.IconColor,
		width, height, t.CanvasBg))

	// Render shapes
	for _, shape := range plane.Shapes {
		elemType := elementTypes[shape.BpmnElement]
		r.renderShape(&sb, &shape, elemType, offsetX, offsetY, elementNames)
	}

	// Render edges
	for _, edge := range plane.Edges {
		r.renderEdge(&sb, &edge, offsetX, offsetY, elementNames)
	}

	sb.WriteString("</svg>\n")

	return []byte(sb.String()), nil
}

func (r *SVGRenderer) calculateBounds(plane *bpmn.BPMNPlane) (minX, minY, maxX, maxY float64) {
	minX = math.MaxFloat64
	minY = math.MaxFloat64
	maxX = -math.MaxFloat64
	maxY = -math.MaxFloat64

	for _, shape := range plane.Shapes {
		b := shape.Bounds
		if b.X < minX {
			minX = b.X
		}
		if b.Y < minY {
			minY = b.Y
		}
		if b.X+b.Width > maxX {
			maxX = b.X + b.Width
		}
		if b.Y+b.Height > maxY {
			maxY = b.Y + b.Height
		}
	}

	for _, edge := range plane.Edges {
		for _, wp := range edge.Waypoints {
			if wp.X < minX {
				minX = wp.X
			}
			if wp.Y < minY {
				minY = wp.Y
			}
			if wp.X > maxX {
				maxX = wp.X
			}
			if wp.Y > maxY {
				maxY = wp.Y
			}
		}
	}

	return
}

func (r *SVGRenderer) renderShape(sb *strings.Builder, shape *bpmn.BPMNShape, elemType string, offsetX, offsetY float64, names map[string]string) {
	b := shape.Bounds
	x := b.X + offsetX
	y := b.Y + offsetY
	w := b.Width
	h := b.Height

	name := names[shape.BpmnElement]

	switch elemType {
	case "startEvent":
		r.renderStartEvent(sb, x, y, w, h, shape, name)
	case "endEvent":
		r.renderEndEvent(sb, x, y, w, h, shape, name)
	case "task", "userTask", "serviceTask", "scriptTask", "sendTask", "receiveTask", "manualTask", "businessRuleTask":
		r.renderTask(sb, x, y, w, h, shape, elemType, name)
	case "subProcess":
		r.renderSubProcess(sb, x, y, w, h, shape, name)
	case "exclusiveGateway":
		r.renderGateway(sb, x, y, w, h, shape, "X", name)
	case "parallelGateway":
		r.renderGateway(sb, x, y, w, h, shape, "+", name)
	case "inclusiveGateway":
		r.renderGateway(sb, x, y, w, h, shape, "O", name)
	case "eventBasedGateway":
		r.renderGateway(sb, x, y, w, h, shape, "⬠", name)
	case "intermediateCatchEvent", "intermediateThrowEvent":
		r.renderIntermediateEvent(sb, x, y, w, h, shape, name)
	case "boundaryEvent":
		r.renderBoundaryEvent(sb, x, y, w, h, shape, name)
	case "participant":
		r.renderPool(sb, x, y, w, h, shape, name)
	case "lane":
		r.renderLane(sb, x, y, w, h, shape, name)
	case "textAnnotation":
		r.renderTextAnnotation(sb, x, y, w, h, shape)
	default:
		// Generic rectangle for unknown elements
		sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="5" ry="5" class="bpmn-task"/>
`, x, y, w, h))
	}
}

func (r *SVGRenderer) renderStartEvent(sb *strings.Builder, x, y, w, h float64, shape *bpmn.BPMNShape, name string) {
	cx := x + w/2
	cy := y + h/2
	radius := math.Min(w, h) / 2
	sb.WriteString(fmt.Sprintf(`  <circle cx="%.1f" cy="%.1f" r="%.1f" class="bpmn-event-start"/>
`, cx, cy, radius))

	r.renderShapeLabel(sb, shape, x, y, w, h, name)
}

func (r *SVGRenderer) renderEndEvent(sb *strings.Builder, x, y, w, h float64, shape *bpmn.BPMNShape, name string) {
	cx := x + w/2
	cy := y + h/2
	radius := math.Min(w, h) / 2
	sb.WriteString(fmt.Sprintf(`  <circle cx="%.1f" cy="%.1f" r="%.1f" class="bpmn-event-end"/>
`, cx, cy, radius))

	r.renderShapeLabel(sb, shape, x, y, w, h, name)
}

func (r *SVGRenderer) renderTask(sb *strings.Builder, x, y, w, h float64, shape *bpmn.BPMNShape, elemType string, name string) {
	sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="10" ry="10" class="bpmn-task"/>
`, x, y, w, h))

	// Task type icon (small indicator in top-left)
	switch elemType {
	case "userTask":
		iconX := x + 8
		iconY := y + 8
		sb.WriteString(fmt.Sprintf(`  <circle cx="%.1f" cy="%.1f" r="5" class="bpmn-icon" stroke-width="1.2"/>
`, iconX+5, iconY+3))
		sb.WriteString(fmt.Sprintf(`  <path d="M%.1f,%.1f Q%.1f,%.1f %.1f,%.1f" class="bpmn-icon" stroke-width="1.2"/>
`, iconX, iconY+14, iconX+5, iconY+10, iconX+10, iconY+14))
	case "serviceTask":
		iconX := x + 12
		iconY := y + 12
		sb.WriteString(fmt.Sprintf(`  <circle cx="%.1f" cy="%.1f" r="5" class="bpmn-icon" stroke-width="1.5"/>
`, iconX, iconY))
		sb.WriteString(fmt.Sprintf(`  <circle cx="%.1f" cy="%.1f" r="2" class="bpmn-icon" stroke-width="1.5"/>
`, iconX, iconY))
	case "scriptTask":
		iconX := x + 6
		iconY := y + 6
		sb.WriteString(fmt.Sprintf(`  <path d="M%.1f,%.1f L%.1f,%.1f L%.1f,%.1f L%.1f,%.1f Z" class="bpmn-icon" stroke-width="1.2"/>
`, iconX, iconY, iconX+12, iconY, iconX+10, iconY+12, iconX-2, iconY+12))
	}

	r.renderShapeLabel(sb, shape, x, y, w, h, name)
}

func (r *SVGRenderer) renderSubProcess(sb *strings.Builder, x, y, w, h float64, shape *bpmn.BPMNShape, name string) {
	sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="10" ry="10" class="bpmn-subprocess"/>
`, x, y, w, h))

	// + marker at bottom center
	cx := x + w/2
	cy := y + h - 10
	sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="12" height="12" class="bpmn-icon" stroke-width="1.2"/>
`, cx-6, cy-6))
	sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="bpmn-icon" stroke-width="1.2"/>
`, cx, cy-4, cx, cy+4))
	sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="bpmn-icon" stroke-width="1.2"/>
`, cx-4, cy, cx+4, cy))

	r.renderShapeLabel(sb, shape, x, y, w, h, name)
}

func (r *SVGRenderer) renderGateway(sb *strings.Builder, x, y, w, h float64, shape *bpmn.BPMNShape, marker string, name string) {
	cx := x + w/2
	cy := y + h/2
	hw := w / 2
	hh := h / 2

	// Diamond shape
	sb.WriteString(fmt.Sprintf(`  <polygon points="%.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f" class="bpmn-gateway"/>
`, cx, y, x+w, cy, cx, y+h, x, cy))

	// Marker
	switch marker {
	case "X":
		size := math.Min(hw, hh) * 0.4
		sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="bpmn-gateway-marker" stroke-width="2.5"/>
`, cx-size, cy-size, cx+size, cy+size))
		sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="bpmn-gateway-marker" stroke-width="2.5"/>
`, cx+size, cy-size, cx-size, cy+size))
	case "+":
		size := math.Min(hw, hh) * 0.45
		sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="bpmn-gateway-marker" stroke-width="3"/>
`, cx, cy-size, cx, cy+size))
		sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="bpmn-gateway-marker" stroke-width="3"/>
`, cx-size, cy, cx+size, cy))
	case "O":
		radius := math.Min(hw, hh) * 0.35
		sb.WriteString(fmt.Sprintf(`  <circle cx="%.1f" cy="%.1f" r="%.1f" class="bpmn-gateway-marker" stroke-width="2.5"/>
`, cx, cy, radius))
	}

	// Gateway labels sit BELOW the diamond so they don't overlap the marker.
	// External label bounds (from the BPMN file) still win when present.
	if name != "" {
		if shape.Label != nil && shape.Label.Bounds != nil {
			r.renderShapeLabel(sb, shape, x, y, w, h, name)
		} else {
			r.renderLabelBelow(sb, x+w/2, y+h+12, math.Max(w*1.6, 90), name)
		}
	}
}

func (r *SVGRenderer) renderIntermediateEvent(sb *strings.Builder, x, y, w, h float64, shape *bpmn.BPMNShape, name string) {
	cx := x + w/2
	cy := y + h/2
	radius := math.Min(w, h) / 2
	sb.WriteString(fmt.Sprintf(`  <circle cx="%.1f" cy="%.1f" r="%.1f" class="bpmn-event-intermediate"/>
`, cx, cy, radius))
	sb.WriteString(fmt.Sprintf(`  <circle cx="%.1f" cy="%.1f" r="%.1f" fill="none" stroke="%s" stroke-width="1.5"/>
`, cx, cy, radius*0.78, r.theme.IntermediateStroke))

	r.renderShapeLabel(sb, shape, x, y, w, h, name)
}

func (r *SVGRenderer) renderBoundaryEvent(sb *strings.Builder, x, y, w, h float64, shape *bpmn.BPMNShape, name string) {
	cx := x + w/2
	cy := y + h/2
	radius := math.Min(w, h) / 2
	sb.WriteString(fmt.Sprintf(`  <circle cx="%.1f" cy="%.1f" r="%.1f" class="bpmn-event-intermediate"/>
`, cx, cy, radius))
	sb.WriteString(fmt.Sprintf(`  <circle cx="%.1f" cy="%.1f" r="%.1f" fill="none" stroke="%s" stroke-width="1.5"/>
`, cx, cy, radius*0.78, r.theme.IntermediateStroke))

	r.renderShapeLabel(sb, shape, x, y, w, h, name)
}

func (r *SVGRenderer) renderPool(sb *strings.Builder, x, y, w, h float64, shape *bpmn.BPMNShape, name string) {
	sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" class="bpmn-pool"/>
`, x, y, w, h))

	if shape.IsHorizontal == nil || *shape.IsHorizontal {
		// Header strip on the left
		sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="30" height="%.1f" class="bpmn-pool-header"/>
`, x, y, h))
		labelX := x + 15
		labelY := y + h/2
		sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" transform="rotate(-90,%.1f,%.1f)" class="bpmn-label">%s</text>
`, labelX, labelY, labelX, labelY, escapeXML(name)))
	}
}

func (r *SVGRenderer) renderLane(sb *strings.Builder, x, y, w, h float64, shape *bpmn.BPMNShape, name string) {
	sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" class="bpmn-lane"/>
`, x, y, w, h))
}

func (r *SVGRenderer) renderTextAnnotation(sb *strings.Builder, x, y, w, h float64, shape *bpmn.BPMNShape) {
	// Open bracket shape
	sb.WriteString(fmt.Sprintf(`  <path d="M%.1f,%.1f L%.1f,%.1f L%.1f,%.1f L%.1f,%.1f" class="bpmn-annotation"/>
`, x+10, y, x, y, x, y+h, x+10, y+h))
}

func (r *SVGRenderer) renderEdge(sb *strings.Builder, edge *bpmn.BPMNEdge, offsetX, offsetY float64, names map[string]string) {
	if len(edge.Waypoints) < 2 {
		return
	}

	var points []string
	for _, wp := range edge.Waypoints {
		points = append(points, fmt.Sprintf("%.1f,%.1f", wp.X+offsetX, wp.Y+offsetY))
	}

	sb.WriteString(fmt.Sprintf(`  <polyline points="%s" class="bpmn-flow"/>
`, strings.Join(points, " ")))

	// Render edge label if present
	name := names[edge.BpmnElement]
	if name != "" && edge.Label != nil && edge.Label.Bounds != nil {
		lb := edge.Label.Bounds
		lx := lb.X + offsetX + lb.Width/2
		ly := lb.Y + offsetY + lb.Height/2
		// Background pill for readability
		sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="16" rx="3" fill="%s" opacity="0.85"/>
`, lx-float64(len(name))*3.2-3, ly-8, float64(len(name))*6.4+6, r.theme.CanvasBg))
		sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" class="bpmn-flow-label">%s</text>
`, lx, ly, escapeXML(name)))
	}
}

func (r *SVGRenderer) renderShapeLabel(sb *strings.Builder, shape *bpmn.BPMNShape, x, y, w, h float64, name string) {
	if name == "" {
		return
	}

	var labelX, labelY, wrapWidth float64

	if shape.Label != nil && shape.Label.Bounds != nil {
		// External label
		lb := shape.Label.Bounds
		labelX = lb.X + (x - shape.Bounds.X) + lb.Width/2
		labelY = lb.Y + (y - shape.Bounds.Y) + lb.Height/2
		wrapWidth = math.Max(lb.Width, 60)
	} else {
		// Centered label inside shape
		labelX = x + w/2
		labelY = y + h/2
		wrapWidth = w - 12
	}

	// Word wrap for long labels
	lines := wrapText(name, int(wrapWidth))
	if len(lines) == 1 {
		sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" class="bpmn-label">%s</text>
`, labelX, labelY, escapeXML(lines[0])))
	} else {
		startY := labelY - float64(len(lines)-1)*7
		for i, line := range lines {
			sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" class="bpmn-label">%s</text>
`, labelX, startY+float64(i)*14, escapeXML(line)))
		}
	}
}

// renderLabelBelow writes a wrapped, centered label whose first line
// starts at topY (used for gateway/event labels that should sit
// outside the shape so they don't overlap inner markers).
func (r *SVGRenderer) renderLabelBelow(sb *strings.Builder, cx, topY, wrapWidth float64, name string) {
	if name == "" {
		return
	}
	if wrapWidth < 50 {
		wrapWidth = 50
	}
	lines := wrapText(name, int(wrapWidth))
	for i, line := range lines {
		sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" class="bpmn-label">%s</text>
`, cx, topY+float64(i)*14, escapeXML(line)))
	}
}

func buildElementTypeMap(defs *bpmn.Definitions) map[string]string {
	types := make(map[string]string)

	for _, proc := range defs.Processes {
		nodes := proc.GetAllFlowNodes()
		for id, t := range nodes {
			types[id] = t
		}
		for _, sf := range proc.SequenceFlows {
			types[sf.ID] = "sequenceFlow"
		}
		for _, ta := range proc.TextAnnotations {
			types[ta.ID] = "textAnnotation"
		}
		for _, a := range proc.Associations {
			types[a.ID] = "association"
		}
	}

	for _, collab := range defs.Collaborations {
		for _, p := range collab.Participants {
			types[p.ID] = "participant"
		}
		for _, mf := range collab.MessageFlows {
			types[mf.ID] = "messageFlow"
		}
	}

	return types
}

// buildElementNameMap creates a map of element ID -> name for label rendering.
func BuildElementNameMap(defs *bpmn.Definitions) map[string]string {
	names := make(map[string]string)

	for _, proc := range defs.Processes {
		if proc.Name != "" {
			names[proc.ID] = proc.Name
		}
		for _, e := range proc.StartEvents {
			if e.Name != "" {
				names[e.ID] = e.Name
			}
		}
		for _, e := range proc.EndEvents {
			if e.Name != "" {
				names[e.ID] = e.Name
			}
		}
		for _, e := range proc.Tasks {
			if e.Name != "" {
				names[e.ID] = e.Name
			}
		}
		for _, e := range proc.UserTasks {
			if e.Name != "" {
				names[e.ID] = e.Name
			}
		}
		for _, e := range proc.ServiceTasks {
			if e.Name != "" {
				names[e.ID] = e.Name
			}
		}
		for _, e := range proc.ScriptTasks {
			if e.Name != "" {
				names[e.ID] = e.Name
			}
		}
		for _, e := range proc.SendTasks {
			if e.Name != "" {
				names[e.ID] = e.Name
			}
		}
		for _, e := range proc.ReceiveTasks {
			if e.Name != "" {
				names[e.ID] = e.Name
			}
		}
		for _, e := range proc.ManualTasks {
			if e.Name != "" {
				names[e.ID] = e.Name
			}
		}
		for _, e := range proc.BusinessRuleTasks {
			if e.Name != "" {
				names[e.ID] = e.Name
			}
		}
		for _, e := range proc.SubProcesses {
			if e.Name != "" {
				names[e.ID] = e.Name
			}
		}
		for _, e := range proc.ExclusiveGateways {
			if e.Name != "" {
				names[e.ID] = e.Name
			}
		}
		for _, e := range proc.ParallelGateways {
			if e.Name != "" {
				names[e.ID] = e.Name
			}
		}
		for _, e := range proc.InclusiveGateways {
			if e.Name != "" {
				names[e.ID] = e.Name
			}
		}
		for _, e := range proc.IntermediateCatchEvents {
			if e.Name != "" {
				names[e.ID] = e.Name
			}
		}
		for _, e := range proc.IntermediateThrowEvents {
			if e.Name != "" {
				names[e.ID] = e.Name
			}
		}
		for _, e := range proc.BoundaryEvents {
			if e.Name != "" {
				names[e.ID] = e.Name
			}
		}
		for _, sf := range proc.SequenceFlows {
			if sf.Name != "" {
				names[sf.ID] = sf.Name
			}
		}
	}

	for _, collab := range defs.Collaborations {
		for _, p := range collab.Participants {
			if p.Name != "" {
				names[p.ID] = p.Name
			}
		}
		for _, mf := range collab.MessageFlows {
			if mf.Name != "" {
				names[mf.ID] = mf.Name
			}
		}
	}

	return names
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func wrapText(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		maxWidth = 80
	}

	// Approximate character width at 12px font ~ 6.2px per char (Segoe UI/Arial avg)
	maxChars := int(float64(maxWidth) / 6.2)
	if maxChars < 6 {
		maxChars = 6
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= maxChars {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)

	return lines
}
