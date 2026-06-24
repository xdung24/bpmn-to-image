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
func NewRasterRenderer(scale float64) *RasterRenderer {
	if scale <= 0 {
		scale = 2.0
	}
	return &RasterRenderer{
		padding: 30,
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
		return nil, fmt.Errorf("no diagram data available for rendering")
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
		r.drawEdge(dc, &edge, offsetX, offsetY, name)
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
	case "task", "userTask", "serviceTask", "scriptTask", "sendTask", "receiveTask", "manualTask", "businessRuleTask":
		r.drawTask(dc, x, y, w, h, elemType, name)
	case "subProcess":
		r.drawSubProcess(dc, x, y, w, h, name)
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
		r.drawLane(dc, x, y, w, h)
	case "textAnnotation":
		r.drawTextAnnotation(dc, x, y, w, h)
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

func (r *RasterRenderer) drawSubProcess(dc *gg.Context, x, y, w, h float64, name string) {
	rx := 10 * r.scale

	r.drawSoftShadowRect(dc, x, y, w, h, rx)

	dc.SetColor(hexColor(r.theme.SubProcessFill))
	dc.DrawRoundedRectangle(x, y, w, h, rx)
	dc.Fill()

	dc.SetColor(hexColor(r.theme.SubProcessStroke))
	dc.SetLineWidth(2 * r.scale)
	dc.DrawRoundedRectangle(x, y, w, h, rx)
	dc.Stroke()

	// + marker at bottom
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

func (r *RasterRenderer) drawLane(dc *gg.Context, x, y, w, h float64) {
	dc.SetColor(hexColor(r.theme.LaneStroke))
	dc.SetLineWidth(1 * r.scale)
	dc.SetDash(4*r.scale, 4*r.scale)
	dc.DrawRectangle(x, y, w, h)
	dc.Stroke()
	dc.SetDash() // reset
}

func (r *RasterRenderer) drawTextAnnotation(dc *gg.Context, x, y, w, h float64) {
	dc.SetColor(hexColor(r.theme.AnnotationStroke))
	dc.SetLineWidth(1 * r.scale)
	dc.MoveTo(x+10*r.scale, y)
	dc.LineTo(x, y)
	dc.LineTo(x, y+h)
	dc.LineTo(x+10*r.scale, y+h)
	dc.Stroke()
}

func (r *RasterRenderer) drawEdge(dc *gg.Context, edge *bpmn.BPMNEdge, offsetX, offsetY float64, name string) {
	if len(edge.Waypoints) < 2 {
		return
	}

	dc.SetColor(hexColor(r.theme.Flow))
	dc.SetLineWidth(1.6 * r.scale)
	dc.SetLineCapRound()
	dc.SetLineJoinRound()

	first := edge.Waypoints[0]
	dc.MoveTo(first.X*r.scale+offsetX, first.Y*r.scale+offsetY)
	for _, wp := range edge.Waypoints[1:] {
		dc.LineTo(wp.X*r.scale+offsetX, wp.Y*r.scale+offsetY)
	}
	dc.Stroke()

	// Draw arrowhead at the end
	if len(edge.Waypoints) >= 2 {
		last := edge.Waypoints[len(edge.Waypoints)-1]
		prev := edge.Waypoints[len(edge.Waypoints)-2]
		r.drawArrowhead(dc, prev.X*r.scale+offsetX, prev.Y*r.scale+offsetY, last.X*r.scale+offsetX, last.Y*r.scale+offsetY)
	}

	// Edge label
	if name != "" && edge.Label != nil && edge.Label.Bounds != nil {
		lb := edge.Label.Bounds
		lx := lb.X*r.scale + offsetX + lb.Width*r.scale/2
		ly := lb.Y*r.scale + offsetY + lb.Height*r.scale/2
		// Background pill for readability
		tw, th := dc.MeasureString(name)
		dc.SetColor(hexColor(r.theme.CanvasBg))
		dc.DrawRoundedRectangle(lx-tw/2-3*r.scale, ly-th/2-1*r.scale, tw+6*r.scale, th+2*r.scale, 3*r.scale)
		dc.Fill()
		dc.SetColor(hexColor(r.theme.Flow))
		dc.DrawStringAnchored(name, lx, ly, 0.5, 0.5)
	}
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
