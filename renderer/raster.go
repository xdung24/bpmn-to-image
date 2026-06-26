package renderer

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/xdung24/bpmn-to-image/bpmn"
	"golang.org/x/image/font/gofont/goregular"
)

// RasterRenderer generates PNG/JPG output from BPMN diagram data.
type RasterRenderer struct {
	padding float64
	scale   float64
	theme   Theme
}

// NewRasterRenderer creates a new raster renderer.
func NewRasterRenderer(scale, padding float64) *RasterRenderer {
	if scale <= 0 {
		scale = 2.0
	}
	if padding < 0 {
		padding = 30
	}
	return &RasterRenderer{
		padding: padding,
		scale:   scale,
		theme:   DefaultTheme,
	}
}

// RenderPNG generates a PNG image from BPMN definitions.
func (r *RasterRenderer) RenderPNG(defs *bpmn.Definitions, outputPath string) error {
	img, err := r.render(defs)
	if err != nil {
		return err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	return png.Encode(f, img)
}

// RenderJPG generates a JPG image from BPMN definitions.
func (r *RasterRenderer) RenderJPG(defs *bpmn.Definitions, outputPath string, quality int) error {
	img, err := r.render(defs)
	if err != nil {
		return err
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if quality <= 0 || quality > 100 {
		quality = 90
	}

	return jpeg.Encode(f, img, &jpeg.Options{Quality: quality})
}

func (r *RasterRenderer) render(defs *bpmn.Definitions) (image.Image, error) {
	if len(defs.Diagrams) == 0 {
		return nil, fmt.Errorf(noDIMessage)
	}

	diagram := defs.Diagrams[0]
	plane := diagram.Plane

	elementTypes := buildElementTypeMap(defs)
	elementNames := BuildElementNameMap(defs)

	// Calculate canvas bounds
	minX, minY, maxX, maxY := calculateBoundsFromPlane(&plane)
	width := (maxX - minX + 2*r.padding) * r.scale
	height := (maxY - minY + 2*r.padding) * r.scale
	offsetX := (-minX + r.padding) * r.scale
	offsetY := (-minY + r.padding) * r.scale

	dc := gg.NewContext(int(width), int(height))

	// Load a scalable font face for crisp labels.
	if font, err := truetype.Parse(goregular.TTF); err == nil {
		dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 12 * r.scale}))
	}

	// Background
	dc.SetColor(hexColor(r.theme.CanvasBg))
	dc.Clear()

	// Render edges first (below shapes)
	for _, edge := range plane.Edges {
		name := elementNames[edge.BpmnElement]
		edgeType := elementTypes[edge.BpmnElement]
		r.drawEdge(dc, &edge, edgeType, offsetX, offsetY, name)
	}

	// Render shapes
	for _, shape := range plane.Shapes {
		elemType := elementTypes[shape.BpmnElement]
		name := elementNames[shape.BpmnElement]
		r.drawShape(dc, &shape, elemType, offsetX, offsetY, name)
	}

	return dc.Image(), nil
}

func calculateBoundsFromPlane(plane *bpmn.BPMNPlane) (minX, minY, maxX, maxY float64) {
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

func (r *RasterRenderer) drawShape(dc *gg.Context, shape *bpmn.BPMNShape, elemType string, offsetX, offsetY float64, name string) {
	b := shape.Bounds
	x := b.X*r.scale + offsetX
	y := b.Y*r.scale + offsetY
	w := b.Width * r.scale
	h := b.Height * r.scale

	switch elemType {
	case "startEvent":
		r.drawEvent(dc, x, y, w, h, 2, hexColor(r.theme.StartStroke), hexColor(r.theme.StartFill), false, name, shape)
	case "endEvent":
		r.drawEvent(dc, x, y, w, h, 3.5, hexColor(r.theme.EndStroke), hexColor(r.theme.EndFill), false, name, shape)
	case "task", "userTask", "serviceTask", "scriptTask", "sendTask", "receiveTask", "manualTask", "businessRuleTask", "callActivity":
		r.drawTask(dc, x, y, w, h, elemType, name)
	case "subProcess":
		r.drawSubProcess(dc, x, y, w, h, name, shape)
	case "exclusiveGateway":
		r.drawGateway(dc, x, y, w, h, "X", name)
	case "parallelGateway":
		r.drawGateway(dc, x, y, w, h, "+", name)
	case "inclusiveGateway":
		r.drawGateway(dc, x, y, w, h, "O", name)
	case "eventBasedGateway":
		r.drawGateway(dc, x, y, w, h, "E", name)
	case "intermediateCatchEvent", "intermediateThrowEvent":
		r.drawIntermediateEvent(dc, x, y, w, h, name, shape)
	case "boundaryEvent":
		r.drawIntermediateEvent(dc, x, y, w, h, name, shape)
	case "participant":
		r.drawPool(dc, x, y, w, h, name, shape)
	case "lane":
		r.drawLane(dc, x, y, w, h, name, shape)
	case "textAnnotation":
		r.drawTextAnnotation(dc, x, y, w, h, name)
	case "dataObjectReference":
		r.drawDataObjectReference(dc, x, y, w, h, name)
	case "dataStoreReference":
		r.drawDataStoreReference(dc, x, y, w, h, name)
	case "group":
		r.drawGroup(dc, x, y, w, h, name)
	default:
		r.drawTask(dc, x, y, w, h, "task", name)
	}
}

// drawSoftShadow draws a subtle offset translucent shadow for a rounded rectangle.
func (r *RasterRenderer) drawSoftShadowRect(dc *gg.Context, x, y, w, h, rx float64) {
	for i := 3; i >= 1; i-- {
		off := float64(i) * r.scale * 0.6
		dc.SetRGBA(0.06, 0.09, 0.16, 0.05)
		dc.DrawRoundedRectangle(x+off, y+off, w, h, rx)
		dc.Fill()
	}
}

// drawSoftShadowCircle draws a subtle offset translucent shadow for a circle.
func (r *RasterRenderer) drawSoftShadowCircle(dc *gg.Context, cx, cy, radius float64) {
	for i := 3; i >= 1; i-- {
		off := float64(i) * r.scale * 0.6
		dc.SetRGBA(0.06, 0.09, 0.16, 0.05)
		dc.DrawCircle(cx+off, cy+off, radius)
		dc.Fill()
	}
}

func (r *RasterRenderer) drawEvent(dc *gg.Context, x, y, w, h, strokeWidth float64, stroke, fill color.Color, shadowSkip bool, name string, shape *bpmn.BPMNShape) {
	cx := x + w/2
	cy := y + h/2
	radius := math.Min(w, h) / 2

	if !shadowSkip {
		r.drawSoftShadowCircle(dc, cx, cy, radius)
	}

	dc.SetColor(fill)
	dc.DrawCircle(cx, cy, radius)
	dc.Fill()

	dc.SetColor(stroke)
	dc.SetLineWidth(strokeWidth * r.scale)
	dc.DrawCircle(cx, cy, radius)
	dc.Stroke()

	if name != "" {
		r.drawLabel(dc, shape, x, y, w, h, name)
	}
}

func (r *RasterRenderer) drawTask(dc *gg.Context, x, y, w, h float64, elemType, name string) {
	rx := 10 * r.scale

	r.drawSoftShadowRect(dc, x, y, w, h, rx)

	dc.SetColor(hexColor(r.theme.TaskFill))
	dc.DrawRoundedRectangle(x, y, w, h, rx)
	dc.Fill()

	dc.SetColor(hexColor(r.theme.TaskStroke))
	dc.SetLineWidth(2 * r.scale)
	dc.DrawRoundedRectangle(x, y, w, h, rx)
	dc.Stroke()

	r.drawTaskIcon(dc, x, y, elemType)

	if name != "" {
		r.drawCenteredText(dc, x+w/2, y+h/2, w-10*r.scale, name)
	}
}

func (r *RasterRenderer) drawTaskIcon(dc *gg.Context, x, y float64, elemType string) {
	dc.SetColor(hexColor(r.theme.IconColor))
	dc.SetLineWidth(1.2 * r.scale)
	s := r.scale
	switch elemType {
	case "userTask":
		dc.DrawCircle(x+13*s, y+11*s, 4*s)
		dc.Stroke()
		dc.DrawArc(x+13*s, y+22*s, 6*s, math.Pi, 2*math.Pi)
		dc.Stroke()
	case "serviceTask":
		dc.DrawCircle(x+13*s, y+13*s, 5*s)
		dc.Stroke()
		dc.DrawCircle(x+13*s, y+13*s, 2*s)
		dc.Stroke()
	case "scriptTask":
		dc.MoveTo(x+7*s, y+8*s)
		dc.LineTo(x+17*s, y+8*s)
		dc.MoveTo(x+7*s, y+13*s)
		dc.LineTo(x+17*s, y+13*s)
		dc.MoveTo(x+7*s, y+18*s)
		dc.LineTo(x+14*s, y+18*s)
		dc.Stroke()
	}
}

func (r *RasterRenderer) drawSubProcess(dc *gg.Context, x, y, w, h float64, name string, shape *bpmn.BPMNShape) {
	rx := 10 * r.scale

	r.drawSoftShadowRect(dc, x, y, w, h, rx)

	dc.SetColor(hexColor(r.theme.SubProcessFill))
	dc.DrawRoundedRectangle(x, y, w, h, rx)
	dc.Fill()

	dc.SetColor(hexColor(r.theme.SubProcessStroke))
	dc.SetLineWidth(2 * r.scale)
	dc.DrawRoundedRectangle(x, y, w, h, rx)
	dc.Stroke()

	expanded := shape != nil && shape.IsExpanded != nil && *shape.IsExpanded

	if expanded {
		// Expanded: no [+] marker, label across the top.
		if name != "" {
			r.drawWrappedTextTop(dc, x+w/2, y+8*r.scale, w-10*r.scale, name)
		}
		return
	}

	// Collapsed: + marker at bottom, label in middle.
	cx := x + w/2
	cy := y + h - 10*r.scale
	size := 6 * r.scale
	dc.SetColor(hexColor(r.theme.IconColor))
	dc.SetLineWidth(1.2 * r.scale)
	dc.DrawRectangle(cx-size, cy-size, size*2, size*2)
	dc.Stroke()
	dc.DrawLine(cx, cy-size*0.6, cx, cy+size*0.6)
	dc.Stroke()
	dc.DrawLine(cx-size*0.6, cy, cx+size*0.6, cy)
	dc.Stroke()

	if name != "" {
		r.drawCenteredText(dc, x+w/2, y+h/2-10*r.scale, w-10*r.scale, name)
	}
}

func (r *RasterRenderer) drawGateway(dc *gg.Context, x, y, w, h float64, marker string, name string) {
	cx := x + w/2
	cy := y + h/2
	hw := w / 2
	hh := h / 2

	// Shadow
	for i := 3; i >= 1; i-- {
		off := float64(i) * r.scale * 0.6
		dc.SetRGBA(0.06, 0.09, 0.16, 0.05)
		dc.MoveTo(cx+off, y+off)
		dc.LineTo(x+w+off, cy+off)
		dc.LineTo(cx+off, y+h+off)
		dc.LineTo(x+off, cy+off)
		dc.ClosePath()
		dc.Fill()
	}

	// Diamond fill
	dc.MoveTo(cx, y)
	dc.LineTo(x+w, cy)
	dc.LineTo(cx, y+h)
	dc.LineTo(x, cy)
	dc.ClosePath()
	dc.SetColor(hexColor(r.theme.GatewayFill))
	dc.Fill()

	dc.MoveTo(cx, y)
	dc.LineTo(x+w, cy)
	dc.LineTo(cx, y+h)
	dc.LineTo(x, cy)
	dc.ClosePath()
	dc.SetColor(hexColor(r.theme.GatewayStroke))
	dc.SetLineWidth(2 * r.scale)
	dc.Stroke()

	// Marker
	dc.SetColor(hexColor(r.theme.GatewayMarker))
	switch marker {
	case "X":
		size := math.Min(hw, hh) * 0.35
		dc.SetLineWidth(2.5 * r.scale)
		dc.DrawLine(cx-size, cy-size, cx+size, cy+size)
		dc.Stroke()
		dc.DrawLine(cx+size, cy-size, cx-size, cy+size)
		dc.Stroke()
	case "+":
		size := math.Min(hw, hh) * 0.4
		dc.SetLineWidth(3 * r.scale)
		dc.DrawLine(cx, cy-size, cx, cy+size)
		dc.Stroke()
		dc.DrawLine(cx-size, cy, cx+size, cy)
		dc.Stroke()
	case "O":
		radius := math.Min(hw, hh) * 0.3
		dc.SetLineWidth(2.5 * r.scale)
		dc.DrawCircle(cx, cy, radius)
		dc.Stroke()
	}

	// Gateway labels go below
	if name != "" {
		dc.SetColor(hexColor(r.theme.Label))
		r.drawWrappedTextTop(dc, cx, y+h+10*r.scale, 110*r.scale, name)
	}
}

func (r *RasterRenderer) drawIntermediateEvent(dc *gg.Context, x, y, w, h float64, name string, shape *bpmn.BPMNShape) {
	cx := x + w/2
	cy := y + h/2
	radius := math.Min(w, h) / 2

	r.drawSoftShadowCircle(dc, cx, cy, radius)

	dc.SetColor(hexColor(r.theme.IntermediateFill))
	dc.DrawCircle(cx, cy, radius)
	dc.Fill()

	dc.SetColor(hexColor(r.theme.IntermediateStroke))
	dc.SetLineWidth(2 * r.scale)
	dc.DrawCircle(cx, cy, radius)
	dc.Stroke()

	dc.SetLineWidth(1.5 * r.scale)
	dc.DrawCircle(cx, cy, radius*0.78)
	dc.Stroke()

	if name != "" {
		r.drawLabel(dc, shape, x, y, w, h, name)
	}
}

func (r *RasterRenderer) drawPool(dc *gg.Context, x, y, w, h float64, name string, shape *bpmn.BPMNShape) {
	// Body
	dc.SetColor(hexColor(r.theme.PoolFill))
	dc.DrawRectangle(x, y, w, h)
	dc.Fill()

	if shape.IsHorizontal == nil || *shape.IsHorizontal {
		// Header strip on the left
		dc.SetColor(hexColor(r.theme.PoolHeader))
		dc.DrawRectangle(x, y, 30*r.scale, h)
		dc.Fill()
	}

	dc.SetColor(hexColor(r.theme.PoolStroke))
	dc.SetLineWidth(2 * r.scale)
	dc.DrawRectangle(x, y, w, h)
	dc.Stroke()

	if shape.IsHorizontal == nil || *shape.IsHorizontal {
		dc.DrawLine(x+30*r.scale, y, x+30*r.scale, y+h)
		dc.Stroke()

		if name != "" {
			dc.SetColor(hexColor(r.theme.Label))
			dc.Push()
			dc.RotateAbout(-math.Pi/2, x+15*r.scale, y+h/2)
			dc.DrawStringAnchored(name, x+15*r.scale, y+h/2, 0.5, 0.5)
			dc.Pop()
		}
	}
}

func (r *RasterRenderer) drawLane(dc *gg.Context, x, y, w, h float64, name string, shape *bpmn.BPMNShape) {
	horizontal := shape == nil || shape.IsHorizontal == nil || *shape.IsHorizontal

	// Lane outline
	dc.SetColor(hexColor(r.theme.LaneStroke))
	dc.SetLineWidth(1 * r.scale)
	dc.DrawRectangle(x, y, w, h)
	dc.Stroke()

	if name == "" {
		return
	}

	if horizontal {
		// Lane header on the left, vertical text
		dc.SetColor(hexColor(r.theme.PoolHeader))
		dc.DrawRectangle(x, y, 20*r.scale, h)
		dc.Fill()
		dc.SetColor(hexColor(r.theme.LaneStroke))
		dc.SetLineWidth(1 * r.scale)
		dc.DrawLine(x+20*r.scale, y, x+20*r.scale, y+h)
		dc.Stroke()

		dc.SetColor(hexColor(r.theme.Label))
		dc.Push()
		dc.RotateAbout(-math.Pi/2, x+10*r.scale, y+h/2)
		dc.DrawStringAnchored(name, x+10*r.scale, y+h/2, 0.5, 0.5)
		dc.Pop()
	} else {
		// Lane header on the top, horizontal text
		dc.SetColor(hexColor(r.theme.PoolHeader))
		dc.DrawRectangle(x, y, w, 20*r.scale)
		dc.Fill()
		dc.SetColor(hexColor(r.theme.LaneStroke))
		dc.SetLineWidth(1 * r.scale)
		dc.DrawLine(x, y+20*r.scale, x+w, y+20*r.scale)
		dc.Stroke()

		dc.SetColor(hexColor(r.theme.Label))
		dc.DrawStringAnchored(name, x+w/2, y+10*r.scale, 0.5, 0.5)
	}
}

// drawDataObjectReference draws a document/page shape with a folded corner.
func (r *RasterRenderer) drawDataObjectReference(dc *gg.Context, x, y, w, h float64, name string) {
	fold := 12 * r.scale
	if fold > w*0.4 {
		fold = w * 0.4
	}

	// Main shape (with folded corner)
	dc.SetColor(hexColor("#ffffff"))
	dc.MoveTo(x, y)
	dc.LineTo(x+w-fold, y)
	dc.LineTo(x+w, y+fold)
	dc.LineTo(x+w, y+h)
	dc.LineTo(x, y+h)
	dc.ClosePath()
	dc.FillPreserve()
	dc.SetColor(hexColor(r.theme.Label))
	dc.SetLineWidth(1.4 * r.scale)
	dc.Stroke()

	// Fold triangle
	dc.SetColor(hexColor("#f3f4f6"))
	dc.MoveTo(x+w-fold, y)
	dc.LineTo(x+w-fold, y+fold)
	dc.LineTo(x+w, y+fold)
	dc.ClosePath()
	dc.FillPreserve()
	dc.SetColor(hexColor(r.theme.Label))
	dc.SetLineWidth(1.4 * r.scale)
	dc.Stroke()

	if name != "" {
		dc.SetColor(hexColor(r.theme.Label))
		r.drawWrappedTextTop(dc, x+w/2, y+h+8*r.scale, math.Max(w, 80*r.scale), name)
	}
}

// drawDataStoreReference draws a cylinder (database) shape.
func (r *RasterRenderer) drawDataStoreReference(dc *gg.Context, x, y, w, h float64, name string) {
	ry := math.Min(h*0.15, 8*r.scale)

	// Body rectangle between top and bottom curves
	dc.SetColor(hexColor("#ffffff"))
	dc.DrawRectangle(x, y+ry, w, h-2*ry)
	dc.Fill()

	// Bottom ellipse
	dc.DrawEllipse(x+w/2, y+h-ry, w/2, ry)
	dc.SetColor(hexColor("#ffffff"))
	dc.FillPreserve()
	dc.SetColor(hexColor(r.theme.Label))
	dc.SetLineWidth(1.4 * r.scale)
	dc.Stroke()

	// Side lines
	dc.SetColor(hexColor(r.theme.Label))
	dc.SetLineWidth(1.4 * r.scale)
	dc.DrawLine(x, y+ry, x, y+h-ry)
	dc.Stroke()
	dc.DrawLine(x+w, y+ry, x+w, y+h-ry)
	dc.Stroke()

	// Top ellipse
	dc.DrawEllipse(x+w/2, y+ry, w/2, ry)
	dc.SetColor(hexColor("#ffffff"))
	dc.FillPreserve()
	dc.SetColor(hexColor(r.theme.Label))
	dc.SetLineWidth(1.4 * r.scale)
	dc.Stroke()

	if name != "" {
		dc.SetColor(hexColor(r.theme.Label))
		r.drawWrappedTextTop(dc, x+w/2, y+h+8*r.scale, math.Max(w+40*r.scale, 80*r.scale), name)
	}
}

// drawGroup draws a dashed rounded rectangle.
func (r *RasterRenderer) drawGroup(dc *gg.Context, x, y, w, h float64, name string) {
	dc.SetColor(hexColor(r.theme.Label))
	dc.SetLineWidth(1.5 * r.scale)
	dc.SetDash(10*r.scale, 4*r.scale, 2*r.scale, 4*r.scale)
	dc.DrawRoundedRectangle(x, y, w, h, 10*r.scale)
	dc.Stroke()
	dc.SetDash()

	if name != "" {
		dc.SetColor(hexColor(r.theme.Label))
		dc.DrawStringAnchored(name, x+10*r.scale, y-4*r.scale, 0, 0.5)
	}
}

func (r *RasterRenderer) drawTextAnnotation(dc *gg.Context, x, y, w, h float64, name string) {
	dc.SetColor(hexColor(r.theme.AnnotationStroke))
	dc.SetLineWidth(1 * r.scale)
	dc.MoveTo(x+10*r.scale, y)
	dc.LineTo(x, y)
	dc.LineTo(x, y+h)
	dc.LineTo(x+10*r.scale, y+h)
	dc.Stroke()

	if name == "" {
		return
	}
	pad := 14 * r.scale
	wrapWidth := w - pad - 4*r.scale
	if wrapWidth < 30*r.scale {
		wrapWidth = 30 * r.scale
	}
	lines := r.wrapMeasured(dc, name, wrapWidth)
	lineH := 12 * r.scale * 1.3
	totalH := float64(len(lines)) * lineH
	startY := y + (h-totalH)/2 + lineH*0.5
	dc.SetColor(hexColor(r.theme.Label))
	for i, line := range lines {
		dc.DrawStringAnchored(line, x+pad, startY+float64(i)*lineH, 0, 0.5)
	}
}

func (r *RasterRenderer) drawEdge(dc *gg.Context, edge *bpmn.BPMNEdge, edgeType string, offsetX, offsetY float64, name string) {
	if len(edge.Waypoints) < 2 {
		return
	}

	dc.SetColor(hexColor(r.theme.Flow))
	dc.SetLineCapRound()
	dc.SetLineJoinRound()

	// Edge-type-specific style.
	switch edgeType {
	case "messageFlow":
		dc.SetLineWidth(r.lineWidth(1.4))
		dc.SetDash(6*r.scale, 4*r.scale)
	case "association":
		dc.SetLineWidth(r.lineWidth(1))
		dc.SetDash(2*r.scale, 3*r.scale)
	default:
		dc.SetLineWidth(r.lineWidth(1.6))
		dc.SetDash() // explicit clear in case previous edge set a pattern
	}

	first := edge.Waypoints[0]
	dc.NewSubPath()
	dc.MoveTo(first.X*r.scale+offsetX, first.Y*r.scale+offsetY)
	for _, wp := range edge.Waypoints[1:] {
		dc.LineTo(wp.X*r.scale+offsetX, wp.Y*r.scale+offsetY)
	}
	dc.Stroke()
	dc.SetDash()

	// Draw arrowhead at the end (no arrowhead for plain associations)
	if edgeType != "association" && len(edge.Waypoints) >= 2 {
		last := edge.Waypoints[len(edge.Waypoints)-1]
		prev := edge.Waypoints[len(edge.Waypoints)-2]
		if edgeType == "messageFlow" {
			r.drawOpenArrowhead(dc, prev.X*r.scale+offsetX, prev.Y*r.scale+offsetY, last.X*r.scale+offsetX, last.Y*r.scale+offsetY)
			// Open circle at the start for message flows
			dc.SetColor(hexColor("#ffffff"))
			dc.DrawCircle(first.X*r.scale+offsetX, first.Y*r.scale+offsetY, 4*r.scale)
			dc.FillPreserve()
			dc.SetColor(hexColor(r.theme.Flow))
			dc.SetLineWidth(r.lineWidth(1.2))
			dc.Stroke()
		} else {
			r.drawArrowhead(dc, prev.X*r.scale+offsetX, prev.Y*r.scale+offsetY, last.X*r.scale+offsetX, last.Y*r.scale+offsetY)
		}
	}

	// Edge label
	if name != "" && edge.Label != nil && edge.Label.Bounds != nil {
		lb := edge.Label.Bounds
		lcx := lb.X + lb.Width/2
		lcy := lb.Y + lb.Height/2

		// If the label is far from the line, snap it close to the line so
		// the relationship is obvious.
		const maxDist = 20.0
		const targetDist = 12.0
		nx, ny, dist := nearestPointOnPolylineRaw(edge.Waypoints, lcx, lcy)
		if dist > maxDist {
			dx, dy := lcx-nx, lcy-ny
			if dist > 0 {
				lcx = nx + dx/dist*targetDist
				lcy = ny + dy/dist*targetDist
			} else {
				lcx = nx
				lcy = ny - targetDist
			}
		}

		lx := lcx*r.scale + offsetX
		ly := lcy*r.scale + offsetY

		// Background pill for readability
		tw, th := dc.MeasureString(name)
		dc.SetColor(hexColor(r.theme.CanvasBg))
		dc.DrawRoundedRectangle(lx-tw/2-3*r.scale, ly-th/2-1*r.scale, tw+6*r.scale, th+2*r.scale, 3*r.scale)
		dc.Fill()
		dc.SetColor(hexColor(r.theme.Flow))
		dc.DrawStringAnchored(name, lx, ly, 0.5, 0.5)
	}
}

// nearestPointOnPolylineRaw returns the closest point on the unscaled
// polyline to (px, py) (also unscaled), and the distance to it.
func nearestPointOnPolylineRaw(waypoints []bpmn.Waypoint, px, py float64) (nx, ny, dist float64) {
	if len(waypoints) == 0 {
		return px, py, 0
	}
	nx, ny = waypoints[0].X, waypoints[0].Y
	dist = math.Hypot(px-nx, py-ny)
	for i := 0; i+1 < len(waypoints); i++ {
		ax, ay := waypoints[i].X, waypoints[i].Y
		bx, by := waypoints[i+1].X, waypoints[i+1].Y
		dx, dy := bx-ax, by-ay
		lenSq := dx*dx + dy*dy
		var cx, cy float64
		if lenSq == 0 {
			cx, cy = ax, ay
		} else {
			t := ((px-ax)*dx + (py-ay)*dy) / lenSq
			if t < 0 {
				t = 0
			} else if t > 1 {
				t = 1
			}
			cx, cy = ax+t*dx, ay+t*dy
		}
		d := math.Hypot(px-cx, py-cy)
		if d < dist {
			dist = d
			nx, ny = cx, cy
		}
	}
	return
}

// lineWidth returns the stroke width for a line, ensuring it never falls
// below a sensible minimum so thin lines remain visible after the image is
// downscaled for display. At scales >= ~2 the natural scaled width is used.
func (r *RasterRenderer) lineWidth(base float64) float64 {
	w := base * r.scale
	// Enforce a minimum stroke width so edges/borders don't disappear when
	// the viewer downsamples the raster image. 3 device pixels is a good
	// balance between visibility and looking proportional at high scales.
	if w < 3 {
		return 3
	}
	return w
}

func (r *RasterRenderer) drawArrowhead(dc *gg.Context, fromX, fromY, toX, toY float64) {
	angle := math.Atan2(toY-fromY, toX-fromX)
	arrowLen := 10 * r.scale
	arrowAngle := math.Pi / 7

	x1 := toX - arrowLen*math.Cos(angle-arrowAngle)
	y1 := toY - arrowLen*math.Sin(angle-arrowAngle)
	x2 := toX - arrowLen*math.Cos(angle+arrowAngle)
	y2 := toY - arrowLen*math.Sin(angle+arrowAngle)

	dc.SetColor(hexColor(r.theme.Flow))
	dc.MoveTo(toX, toY)
	dc.LineTo(x1, y1)
	dc.LineTo(x2, y2)
	dc.ClosePath()
	dc.Fill()
}

// drawOpenArrowhead draws an unfilled "V" arrowhead used for message flows.
func (r *RasterRenderer) drawOpenArrowhead(dc *gg.Context, fromX, fromY, toX, toY float64) {
	angle := math.Atan2(toY-fromY, toX-fromX)
	arrowLen := 11 * r.scale
	arrowAngle := math.Pi / 6

	x1 := toX - arrowLen*math.Cos(angle-arrowAngle)
	y1 := toY - arrowLen*math.Sin(angle-arrowAngle)
	x2 := toX - arrowLen*math.Cos(angle+arrowAngle)
	y2 := toY - arrowLen*math.Sin(angle+arrowAngle)

	dc.SetColor(hexColor(r.theme.Flow))
	dc.SetLineWidth(r.lineWidth(1.4))
	dc.MoveTo(x1, y1)
	dc.LineTo(toX, toY)
	dc.LineTo(x2, y2)
	dc.Stroke()
}

func (r *RasterRenderer) drawLabel(dc *gg.Context, shape *bpmn.BPMNShape, x, y, w, h float64, name string) {
	dc.SetColor(hexColor(r.theme.Label))
	if shape.Label != nil && shape.Label.Bounds != nil {
		// External label positioned by the diagram
		lb := shape.Label.Bounds
		lx := lb.X*r.scale + (x - shape.Bounds.X*r.scale) + lb.Width*r.scale/2
		ly := lb.Y*r.scale + (y - shape.Bounds.Y*r.scale) + lb.Height*r.scale/2
		r.drawWrappedText(dc, lx, ly, math.Max(lb.Width*r.scale, 60*r.scale), name)
	} else {
		// Label below the shape (for events)
		baseY := y + h + 10*r.scale
		r.drawWrappedTextTop(dc, x+w/2, baseY, 90*r.scale, name)
	}
}

func (r *RasterRenderer) drawCenteredText(dc *gg.Context, cx, cy, maxWidth float64, text string) {
	dc.SetColor(hexColor(r.theme.Label))
	r.drawWrappedText(dc, cx, cy, maxWidth, text)
}

// wrapMeasured wraps text so each line fits within maxWidth using real font metrics.
func (r *RasterRenderer) wrapMeasured(dc *gg.Context, text string, maxWidth float64) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	cur := words[0]
	for _, word := range words[1:] {
		candidate := cur + " " + word
		if tw, _ := dc.MeasureString(candidate); tw <= maxWidth {
			cur = candidate
		} else {
			lines = append(lines, cur)
			cur = word
		}
	}
	lines = append(lines, cur)
	return lines
}

// drawWrappedText draws text wrapped to maxWidth, vertically centered on (cx, cy).
func (r *RasterRenderer) drawWrappedText(dc *gg.Context, cx, cy, maxWidth float64, text string) {
	lines := r.wrapMeasured(dc, text, maxWidth)
	if len(lines) == 0 {
		return
	}
	lineH := dc.FontHeight() * 1.15
	startY := cy - (float64(len(lines)-1)*lineH)/2
	for i, line := range lines {
		dc.DrawStringAnchored(line, cx, startY+float64(i)*lineH, 0.5, 0.5)
	}
}

// drawWrappedTextTop draws text wrapped to maxWidth, with the first line starting at topY.
func (r *RasterRenderer) drawWrappedTextTop(dc *gg.Context, cx, topY, maxWidth float64, text string) {
	lines := r.wrapMeasured(dc, text, maxWidth)
	lineH := dc.FontHeight() * 1.15
	for i, line := range lines {
		dc.DrawStringAnchored(line, cx, topY+float64(i)*lineH, 0.5, 0.5)
	}
}
