package renderer

import (
	"fmt"
	"math"
	"strings"

	"github.com/xdung24/bpmn-to-image/dmn"
)

// DMNSVGRenderer generates SVG output from a DMN Decision Requirements
// Diagram (DRD). Decision-table contents are not rendered.
type DMNSVGRenderer struct {
	padding float64
	theme   Theme
}

// NewDMNSVGRenderer returns an SVG renderer for DMN files.
func NewDMNSVGRenderer() *DMNSVGRenderer {
	return &DMNSVGRenderer{
		padding: 30,
		theme:   DefaultTheme,
	}
}

// Render produces an SVG document for the first DMN diagram found in defs.
func (r *DMNSVGRenderer) Render(defs *dmn.Definitions) ([]byte, error) {
	if defs.DMNDI == nil || len(defs.DMNDI.Diagrams) == 0 {
		return nil, fmt.Errorf(noDMNDIMessage)
	}
	diagram := defs.DMNDI.Diagrams[0]
	nodes := defs.BuildNodeIndex()
	edges := defs.BuildEdgeIndex()

	minX, minY, maxX, maxY := r.calculateBounds(&diagram)
	width := maxX - minX + 2*r.padding
	height := maxY - minY + 2*r.padding
	offsetX := -minX + r.padding
	offsetY := -minY + r.padding

	var sb strings.Builder
	t := r.theme

	sb.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f">
<defs>
  <marker id="dmn-arrow-filled" markerWidth="12" markerHeight="9" refX="9" refY="4.5" orient="auto" markerUnits="userSpaceOnUse">
    <path d="M0,0 L9,4.5 L0,9 L2.5,4.5 Z" fill="%s"/>
  </marker>
  <marker id="dmn-arrow-open" markerWidth="12" markerHeight="9" refX="9" refY="4.5" orient="auto" markerUnits="userSpaceOnUse">
    <path d="M0,0 L9,4.5 L0,9" fill="none" stroke="%s" stroke-width="1.4"/>
  </marker>
  <marker id="dmn-dot" markerWidth="8" markerHeight="8" refX="4" refY="4" orient="auto" markerUnits="userSpaceOnUse">
    <circle cx="4" cy="4" r="3.2" fill="%s"/>
  </marker>
  <linearGradient id="dmn-grad-decision" x1="0" y1="0" x2="0" y2="1">
    <stop offset="0%%" stop-color="#ffffff"/><stop offset="100%%" stop-color="%s"/>
  </linearGradient>
  <linearGradient id="dmn-grad-input" x1="0" y1="0" x2="0" y2="1">
    <stop offset="0%%" stop-color="#ffffff"/><stop offset="100%%" stop-color="%s"/>
  </linearGradient>
  <linearGradient id="dmn-grad-bkm" x1="0" y1="0" x2="0" y2="1">
    <stop offset="0%%" stop-color="#ffffff"/><stop offset="100%%" stop-color="%s"/>
  </linearGradient>
  <linearGradient id="dmn-grad-knowledge" x1="0" y1="0" x2="0" y2="1">
    <stop offset="0%%" stop-color="#ffffff"/><stop offset="100%%" stop-color="%s"/>
  </linearGradient>
  <filter id="dmn-shadow" x="-20%%" y="-20%%" width="140%%" height="140%%">
    <feDropShadow dx="1.5" dy="2" stdDeviation="2" flood-color="#0f172a" flood-opacity="0.18"/>
  </filter>
</defs>
<style>
  .dmn-decision { fill: url(#dmn-grad-decision); stroke: %s; stroke-width: 2; filter: url(#dmn-shadow); }
  .dmn-input    { fill: url(#dmn-grad-input);    stroke: %s; stroke-width: 2; filter: url(#dmn-shadow); }
  .dmn-bkm      { fill: url(#dmn-grad-bkm);      stroke: %s; stroke-width: 2; filter: url(#dmn-shadow); }
  .dmn-knowledge{ fill: url(#dmn-grad-knowledge); stroke: %s; stroke-width: 2; filter: url(#dmn-shadow); }
  .dmn-service  { fill: none; stroke: %s; stroke-width: 2; stroke-dasharray: 8,4; }
  .dmn-annotation { fill: none; stroke: %s; stroke-width: 1; }
  .dmn-information { fill: none; stroke: %s; stroke-width: 1.6; marker-end: url(#dmn-arrow-filled); stroke-linejoin: round; stroke-linecap: round; }
  .dmn-knowledge-edge { fill: none; stroke: %s; stroke-width: 1.4; stroke-dasharray: 6,4; marker-end: url(#dmn-arrow-open); stroke-linejoin: round; stroke-linecap: round; }
  .dmn-authority { fill: none; stroke: %s; stroke-width: 1.4; stroke-dasharray: 3,3; marker-end: url(#dmn-dot); stroke-linejoin: round; stroke-linecap: round; }
  .dmn-association { fill: none; stroke: %s; stroke-width: 1; stroke-dasharray: 2,3; }
  .dmn-label { font-family: 'Segoe UI', Arial, sans-serif; font-size: 12px; font-weight: 500; fill: %s; text-anchor: middle; dominant-baseline: central; }
  .dmn-edge-label { font-family: 'Segoe UI', Arial, sans-serif; font-size: 11px; fill: %s; text-anchor: middle; dominant-baseline: central; }
</style>
<rect x="0" y="0" width="%.0f" height="%.0f" fill="%s"/>
`,
		width, height, width, height,
		t.Flow, // filled arrow color
		t.Flow, // open arrow stroke
		t.Flow, // dot fill
		t.DecisionFill, t.InputDataFill, t.BKMFill, t.KnowledgeSourceFill,
		t.DecisionStroke, t.InputDataStroke, t.BKMStroke, t.KnowledgeSourceStroke,
		t.DecisionServiceStroke,
		t.AnnotationStroke,
		t.Flow, t.Flow, t.Flow, t.AnnotationStroke,
		t.Label, t.Label,
		width, height, t.CanvasBg,
	))

	// 1. Decision services first (act as containers behind other shapes).
	for _, shape := range diagram.Shapes {
		info := nodes[shape.DMNElementRef]
		if info.Kind == dmn.KindDecisionService {
			r.renderShape(&sb, &shape, info, offsetX, offsetY)
		}
	}
	// 2. All other shapes on top.
	for _, shape := range diagram.Shapes {
		info := nodes[shape.DMNElementRef]
		if info.Kind == "" || info.Kind == dmn.KindDecisionService {
			continue
		}
		r.renderShape(&sb, &shape, info, offsetX, offsetY)
	}
	// 3. Edges last so arrowheads sit above shape fills.
	for _, edge := range diagram.Edges {
		r.renderEdge(&sb, &edge, edges[edge.DMNElementRef], offsetX, offsetY)
	}

	sb.WriteString("</svg>\n")
	return []byte(sb.String()), nil
}

func (r *DMNSVGRenderer) calculateBounds(d *dmn.DMNDiagram) (minX, minY, maxX, maxY float64) {
	minX, minY = math.MaxFloat64, math.MaxFloat64
	maxX, maxY = -math.MaxFloat64, -math.MaxFloat64
	have := false

	for _, s := range d.Shapes {
		b := s.Bounds
		if b.Width == 0 && b.Height == 0 {
			continue
		}
		have = true
		minX = math.Min(minX, b.X)
		minY = math.Min(minY, b.Y)
		maxX = math.Max(maxX, b.X+b.Width)
		maxY = math.Max(maxY, b.Y+b.Height)
		if s.Label != nil && s.Label.Bounds != nil {
			lb := s.Label.Bounds
			minX = math.Min(minX, lb.X)
			minY = math.Min(minY, lb.Y)
			maxX = math.Max(maxX, lb.X+lb.Width)
			maxY = math.Max(maxY, lb.Y+lb.Height)
		}
	}
	for _, e := range d.Edges {
		for _, wp := range e.Waypoints {
			have = true
			minX = math.Min(minX, wp.X)
			minY = math.Min(minY, wp.Y)
			maxX = math.Max(maxX, wp.X)
			maxY = math.Max(maxY, wp.Y)
		}
	}
	if !have {
		return 0, 0, 200, 200
	}
	return
}

func (r *DMNSVGRenderer) renderShape(sb *strings.Builder, s *dmn.DMNShape, info dmn.NodeInfo, ox, oy float64) {
	x := s.Bounds.X + ox
	y := s.Bounds.Y + oy
	w := s.Bounds.Width
	h := s.Bounds.Height
	name := info.Name

	switch info.Kind {
	case dmn.KindDecision:
		// Rounded rectangle
		sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="6" ry="6" class="dmn-decision"/>
`, x, y, w, h))
		r.renderLabel(sb, s, x, y, w, h, name, "dmn-label", true)

	case dmn.KindInputData:
		// Stadium (fully rounded) shape
		rx := h / 2
		sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="%.1f" ry="%.1f" class="dmn-input"/>
`, x, y, w, h, rx, rx))
		r.renderLabel(sb, s, x, y, w, h, name, "dmn-label", true)

	case dmn.KindBKM:
		// Rectangle with top-left & top-right corners clipped
		c := math.Min(12, math.Min(w, h)*0.25)
		path := fmt.Sprintf("M %.1f %.1f L %.1f %.1f L %.1f %.1f L %.1f %.1f L %.1f %.1f L %.1f %.1f Z",
			x+c, y,
			x+w-c, y,
			x+w, y+c,
			x+w, y+h,
			x, y+h,
			x, y+c)
		sb.WriteString(fmt.Sprintf(`  <path d="%s" class="dmn-bkm"/>
`, path))
		r.renderLabel(sb, s, x, y, w, h, name, "dmn-label", true)

	case dmn.KindKnowledgeSource:
		// Rectangle with wavy bottom
		sb.WriteString(fmt.Sprintf(`  <path d="%s" class="dmn-knowledge"/>
`, knowledgeSourcePath(x, y, w, h)))
		r.renderLabel(sb, s, x, y, w, h*0.85, name, "dmn-label", true)

	case dmn.KindDecisionService:
		sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="14" ry="14" class="dmn-service"/>
`, x, y, w, h))
		// Header line ~25% from top
		yh := y + h*0.25
		sb.WriteString(fmt.Sprintf(`  <line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="1.5" stroke-dasharray="6,3"/>
`, x, yh, x+w, yh, r.theme.DecisionServiceStroke))
		// Service name in the header band
		if name != "" {
			cx := x + w/2
			cy := y + h*0.125
			sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" class="dmn-label">%s</text>
`, cx, cy, escapeXML(name)))
		}

	case dmn.KindTextAnnotation:
		// Open-left bracket with text inside
		bracket := fmt.Sprintf("M %.1f %.1f L %.1f %.1f L %.1f %.1f L %.1f %.1f",
			x+10, y,
			x, y,
			x, y+h,
			x+10, y+h)
		sb.WriteString(fmt.Sprintf(`  <path d="%s" class="dmn-annotation"/>
`, bracket))
		if name != "" {
			lines := wrapText(name, int(w-14))
			startY := y + h/2 - float64(len(lines)-1)*7
			for i, line := range lines {
				sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" class="dmn-label" text-anchor="start">%s</text>
`, x+8, startY+float64(i)*14, escapeXML(line)))
			}
		}
	}
}

// knowledgeSourcePath returns a path describing a rectangle with a
// 3-cycle sine-wavy bottom edge.
func knowledgeSourcePath(x, y, w, h float64) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "M %.1f %.1f ", x, y)
	fmt.Fprintf(&sb, "L %.1f %.1f ", x+w, y)
	fmt.Fprintf(&sb, "L %.1f %.1f ", x+w, y+h*0.85)
	// Wavy bottom: 3 cubic Bezier curves alternating up/down
	cycles := 3
	dx := w / float64(cycles)
	amp := h * 0.10
	yBase := y + h*0.85
	for i := 0; i < cycles; i++ {
		x0 := x + w - float64(i)*dx
		x1 := x0 - dx
		// Wave: go from (x0, yBase) to (x1, yBase) with a dip down then up
		mid := (x0 + x1) / 2
		fmt.Fprintf(&sb, "Q %.1f %.1f %.1f %.1f ", mid+dx/4, yBase+amp, mid, yBase)
		fmt.Fprintf(&sb, "Q %.1f %.1f %.1f %.1f ", mid-dx/4, yBase-amp, x1, yBase)
	}
	fmt.Fprintf(&sb, "L %.1f %.1f Z", x, y)
	return sb.String()
}

func (r *DMNSVGRenderer) renderLabel(sb *strings.Builder, s *dmn.DMNShape, x, y, w, h float64, name, class string, useExternal bool) {
	if name == "" {
		return
	}
	labelX := x + w/2
	labelY := y + h/2
	wrapW := w - 12
	if useExternal && s.Label != nil && s.Label.Bounds != nil {
		// External labels live in their own bounds.
		lb := s.Label.Bounds
		labelX = lb.X + (x - s.Bounds.X) + lb.Width/2
		labelY = lb.Y + (y - s.Bounds.Y) + lb.Height/2
		if lb.Width > 0 {
			wrapW = lb.Width
		}
	}
	if wrapW < 30 {
		wrapW = 30
	}
	lines := wrapText(name, int(wrapW))
	startY := labelY - float64(len(lines)-1)*7
	for i, line := range lines {
		sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" class="%s">%s</text>
`, labelX, startY+float64(i)*14, class, escapeXML(line)))
	}
}

func (r *DMNSVGRenderer) renderEdge(sb *strings.Builder, e *dmn.DMNEdge, kind dmn.EdgeKind, ox, oy float64) {
	if len(e.Waypoints) < 2 {
		return
	}
	class := "dmn-information"
	switch kind {
	case dmn.EdgeKnowledge:
		class = "dmn-knowledge-edge"
	case dmn.EdgeAuthority:
		class = "dmn-authority"
	case dmn.EdgeAssociation:
		class = "dmn-association"
	}

	var path strings.Builder
	for i, wp := range e.Waypoints {
		if i == 0 {
			fmt.Fprintf(&path, "M %.1f %.1f ", wp.X+ox, wp.Y+oy)
		} else {
			fmt.Fprintf(&path, "L %.1f %.1f ", wp.X+ox, wp.Y+oy)
		}
	}
	sb.WriteString(fmt.Sprintf(`  <path d="%s" class="%s"/>
`, strings.TrimSpace(path.String()), class))

	if e.Label != nil && e.Label.Bounds != nil && strings.TrimSpace(e.Label.Text) != "" {
		lb := e.Label.Bounds
		cx := lb.X + ox + lb.Width/2
		cy := lb.Y + oy + lb.Height/2
		sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f" class="dmn-edge-label">%s</text>
`, cx, cy, escapeXML(e.Label.Text)))
	}
}
