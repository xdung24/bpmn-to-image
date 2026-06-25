package renderer

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"strings"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"github.com/xdung24/bpmn-to-image/dmn"
	"golang.org/x/image/font/gofont/goregular"
)

// DMNRasterRenderer generates PNG/JPG output from a DMN DRD.
type DMNRasterRenderer struct {
	padding float64
	scale   float64
	theme   Theme
}

// NewDMNRasterRenderer returns a raster renderer for DMN at the given scale.
func NewDMNRasterRenderer(scale, padding float64) *DMNRasterRenderer {
	if scale <= 0 {
		scale = 2.0
	}
	if padding < 0 {
		padding = 30
	}
	return &DMNRasterRenderer{
		padding: padding,
		scale:   scale,
		theme:   DefaultTheme,
	}
}

// RenderPNG renders the first DMN diagram as a PNG file.
func (r *DMNRasterRenderer) RenderPNG(defs *dmn.Definitions, outputPath string) error {
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

// RenderJPG renders the first DMN diagram as a JPEG file.
func (r *DMNRasterRenderer) RenderJPG(defs *dmn.Definitions, outputPath string, quality int) error {
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

func (r *DMNRasterRenderer) render(defs *dmn.Definitions) (image.Image, error) {
	if defs.DMNDI == nil || len(defs.DMNDI.Diagrams) == 0 {
		return nil, fmt.Errorf(noDMNDIMessage)
	}
	diagram := defs.DMNDI.Diagrams[0]
	nodes := defs.BuildNodeIndex()
	edges := defs.BuildEdgeIndex()

	minX, minY, maxX, maxY := dmnPlaneBounds(&diagram)
	width := (maxX - minX + 2*r.padding) * r.scale
	height := (maxY - minY + 2*r.padding) * r.scale
	offsetX := (-minX + r.padding) * r.scale
	offsetY := (-minY + r.padding) * r.scale

	dc := gg.NewContext(int(width), int(height))
	if font, err := truetype.Parse(goregular.TTF); err == nil {
		dc.SetFontFace(truetype.NewFace(font, &truetype.Options{Size: 12 * r.scale}))
	}

	dc.SetColor(hexColor(r.theme.CanvasBg))
	dc.Clear()

	// 1. Decision services first (background containers).
	for _, s := range diagram.Shapes {
		info := nodes[s.DMNElementRef]
		if info.Kind == dmn.KindDecisionService {
			r.drawShape(dc, &s, info, offsetX, offsetY)
		}
	}
	// 2. Edges below the foreground shapes so arrowheads sit at the boundary.
	for _, e := range diagram.Edges {
		r.drawEdge(dc, &e, edges[e.DMNElementRef], offsetX, offsetY)
	}
	// 3. Foreground shapes.
	for _, s := range diagram.Shapes {
		info := nodes[s.DMNElementRef]
		if info.Kind == "" || info.Kind == dmn.KindDecisionService {
			continue
		}
		r.drawShape(dc, &s, info, offsetX, offsetY)
	}

	return dc.Image(), nil
}

func dmnPlaneBounds(d *dmn.DMNDiagram) (minX, minY, maxX, maxY float64) {
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

func (r *DMNRasterRenderer) drawShape(dc *gg.Context, s *dmn.DMNShape, info dmn.NodeInfo, ox, oy float64) {
	x := s.Bounds.X*r.scale + ox
	y := s.Bounds.Y*r.scale + oy
	w := s.Bounds.Width * r.scale
	h := s.Bounds.Height * r.scale
	name := info.Name

	switch info.Kind {
	case dmn.KindDecision:
		rx := 6 * r.scale
		r.softShadowRect(dc, x, y, w, h, rx)
		dc.SetColor(hexColor(r.theme.DecisionFill))
		dc.DrawRoundedRectangle(x, y, w, h, rx)
		dc.Fill()
		dc.SetColor(hexColor(r.theme.DecisionStroke))
		dc.SetLineWidth(2 * r.scale)
		dc.DrawRoundedRectangle(x, y, w, h, rx)
		dc.Stroke()
		r.drawShapeLabel(dc, s, x, y, w, h, name)

	case dmn.KindInputData:
		rx := h / 2
		r.softShadowRect(dc, x, y, w, h, rx)
		dc.SetColor(hexColor(r.theme.InputDataFill))
		dc.DrawRoundedRectangle(x, y, w, h, rx)
		dc.Fill()
		dc.SetColor(hexColor(r.theme.InputDataStroke))
		dc.SetLineWidth(2 * r.scale)
		dc.DrawRoundedRectangle(x, y, w, h, rx)
		dc.Stroke()
		r.drawShapeLabel(dc, s, x, y, w, h, name)

	case dmn.KindBKM:
		c := math.Min(12*r.scale, math.Min(w, h)*0.25)
		drawBKMPath(dc, x, y, w, h, c)
		dc.SetColor(hexColor(r.theme.BKMFill))
		dc.FillPreserve()
		dc.SetColor(hexColor(r.theme.BKMStroke))
		dc.SetLineWidth(2 * r.scale)
		dc.Stroke()
		r.drawShapeLabel(dc, s, x, y, w, h, name)

	case dmn.KindKnowledgeSource:
		drawKnowledgeSourcePath(dc, x, y, w, h)
		dc.SetColor(hexColor(r.theme.KnowledgeSourceFill))
		dc.FillPreserve()
		dc.SetColor(hexColor(r.theme.KnowledgeSourceStroke))
		dc.SetLineWidth(2 * r.scale)
		dc.Stroke()
		// Center label in the rectangular portion (top 85%)
		r.drawShapeLabel(dc, s, x, y, w, h*0.85, name)

	case dmn.KindDecisionService:
		rx := 14 * r.scale
		dc.SetColor(hexColor(r.theme.DecisionServiceFill))
		dc.DrawRoundedRectangle(x, y, w, h, rx)
		dc.Fill()
		dc.SetColor(hexColor(r.theme.DecisionServiceStroke))
		dc.SetLineWidth(2 * r.scale)
		dc.SetDash(8*r.scale, 4*r.scale)
		dc.DrawRoundedRectangle(x, y, w, h, rx)
		dc.Stroke()
		// Header divider line
		yh := y + h*0.25
		dc.SetDash(6*r.scale, 3*r.scale)
		dc.SetLineWidth(1.5 * r.scale)
		dc.DrawLine(x, yh, x+w, yh)
		dc.Stroke()
		dc.SetDash() // reset
		if name != "" {
			dc.SetColor(hexColor(r.theme.Label))
			dc.DrawStringAnchored(name, x+w/2, y+h*0.125, 0.5, 0.5)
		}

	case dmn.KindTextAnnotation:
		dc.SetColor(hexColor(r.theme.AnnotationStroke))
		dc.SetLineWidth(1 * r.scale)
		dc.MoveTo(x+10*r.scale, y)
		dc.LineTo(x, y)
		dc.LineTo(x, y+h)
		dc.LineTo(x+10*r.scale, y+h)
		dc.Stroke()
		if name != "" {
			dc.SetColor(hexColor(r.theme.Label))
			lines := r.wrapMeasured(dc, name, w-14*r.scale)
			lineH := dc.FontHeight() * 1.15
			startY := y + h/2 - (float64(len(lines)-1)*lineH)/2
			for i, line := range lines {
				dc.DrawStringAnchored(line, x+8*r.scale, startY+float64(i)*lineH, 0, 0.5)
			}
		}
	}
}

// drawBKMPath draws the rectangle-with-clipped-top-corners shape (BKM).
func drawBKMPath(dc *gg.Context, x, y, w, h, c float64) {
	dc.NewSubPath()
	dc.MoveTo(x+c, y)
	dc.LineTo(x+w-c, y)
	dc.LineTo(x+w, y+c)
	dc.LineTo(x+w, y+h)
	dc.LineTo(x, y+h)
	dc.LineTo(x, y+c)
	dc.ClosePath()
}

// drawKnowledgeSourcePath draws a rectangle with a 3-cycle wavy bottom edge.
func drawKnowledgeSourcePath(dc *gg.Context, x, y, w, h float64) {
	dc.NewSubPath()
	dc.MoveTo(x, y)
	dc.LineTo(x+w, y)
	dc.LineTo(x+w, y+h*0.85)
	yBase := y + h*0.85
	amp := h * 0.10
	cycles := 3
	dx := w / float64(cycles)
	for i := 0; i < cycles; i++ {
		x0 := x + w - float64(i)*dx
		x1 := x0 - dx
		mid := (x0 + x1) / 2
		dc.QuadraticTo(mid+dx/4, yBase+amp, mid, yBase)
		dc.QuadraticTo(mid-dx/4, yBase-amp, x1, yBase)
	}
	dc.LineTo(x, y)
	dc.ClosePath()
}

func (r *DMNRasterRenderer) drawShapeLabel(dc *gg.Context, s *dmn.DMNShape, x, y, w, h float64, name string) {
	if name == "" {
		return
	}
	dc.SetColor(hexColor(r.theme.Label))
	if s.Label != nil && s.Label.Bounds != nil {
		lb := s.Label.Bounds
		lx := lb.X*r.scale + (x - s.Bounds.X*r.scale) + lb.Width*r.scale/2
		ly := lb.Y*r.scale + (y - s.Bounds.Y*r.scale) + lb.Height*r.scale/2
		r.drawWrappedText(dc, lx, ly, math.Max(lb.Width*r.scale, 60*r.scale), name)
		return
	}
	r.drawWrappedText(dc, x+w/2, y+h/2, w-12*r.scale, name)
}

func (r *DMNRasterRenderer) softShadowRect(dc *gg.Context, x, y, w, h, rx float64) {
	for i := 3; i >= 1; i-- {
		off := float64(i) * r.scale * 0.6
		dc.SetRGBA(0.06, 0.09, 0.16, 0.05)
		dc.DrawRoundedRectangle(x+off, y+off, w, h, rx)
		dc.Fill()
	}
}

func (r *DMNRasterRenderer) wrapMeasured(dc *gg.Context, text string, maxWidth float64) []string {
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

func (r *DMNRasterRenderer) drawWrappedText(dc *gg.Context, cx, cy, maxWidth float64, text string) {
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

func (r *DMNRasterRenderer) drawEdge(dc *gg.Context, e *dmn.DMNEdge, kind dmn.EdgeKind, ox, oy float64) {
	if len(e.Waypoints) < 2 {
		return
	}
	pts := make([][2]float64, len(e.Waypoints))
	for i, wp := range e.Waypoints {
		pts[i] = [2]float64{wp.X*r.scale + ox, wp.Y*r.scale + oy}
	}

	dc.SetColor(hexColor(r.theme.Flow))
	switch kind {
	case dmn.EdgeKnowledge:
		dc.SetLineWidth(1.4 * r.scale)
		dc.SetDash(6*r.scale, 4*r.scale)
	case dmn.EdgeAuthority:
		dc.SetLineWidth(1.4 * r.scale)
		dc.SetDash(3*r.scale, 3*r.scale)
	case dmn.EdgeAssociation:
		dc.SetColor(hexColor(r.theme.AnnotationStroke))
		dc.SetLineWidth(1 * r.scale)
		dc.SetDash(2*r.scale, 3*r.scale)
	default:
		dc.SetLineWidth(1.6 * r.scale)
	}

	dc.MoveTo(pts[0][0], pts[0][1])
	for _, p := range pts[1:] {
		dc.LineTo(p[0], p[1])
	}
	dc.Stroke()
	dc.SetDash()

	// Terminators
	last := pts[len(pts)-1]
	prev := pts[len(pts)-2]
	switch kind {
	case dmn.EdgeInformation, "":
		r.drawFilledArrow(dc, prev[0], prev[1], last[0], last[1])
	case dmn.EdgeKnowledge:
		r.drawOpenArrow(dc, prev[0], prev[1], last[0], last[1])
	case dmn.EdgeAuthority:
		r.drawDot(dc, last[0], last[1])
	}

	if e.Label != nil && e.Label.Bounds != nil && strings.TrimSpace(e.Label.Text) != "" {
		lb := e.Label.Bounds
		cx := lb.X*r.scale + ox + lb.Width*r.scale/2
		cy := lb.Y*r.scale + oy + lb.Height*r.scale/2
		dc.SetColor(hexColor(r.theme.Label))
		dc.DrawStringAnchored(e.Label.Text, cx, cy, 0.5, 0.5)
	}
}

func (r *DMNRasterRenderer) drawFilledArrow(dc *gg.Context, fromX, fromY, toX, toY float64) {
	angle := math.Atan2(toY-fromY, toX-fromX)
	l := 10 * r.scale
	a := math.Pi / 7
	x1 := toX - l*math.Cos(angle-a)
	y1 := toY - l*math.Sin(angle-a)
	x2 := toX - l*math.Cos(angle+a)
	y2 := toY - l*math.Sin(angle+a)
	dc.SetColor(hexColor(r.theme.Flow))
	dc.MoveTo(toX, toY)
	dc.LineTo(x1, y1)
	dc.LineTo(x2, y2)
	dc.ClosePath()
	dc.Fill()
}

func (r *DMNRasterRenderer) drawOpenArrow(dc *gg.Context, fromX, fromY, toX, toY float64) {
	angle := math.Atan2(toY-fromY, toX-fromX)
	l := 10 * r.scale
	a := math.Pi / 6
	x1 := toX - l*math.Cos(angle-a)
	y1 := toY - l*math.Sin(angle-a)
	x2 := toX - l*math.Cos(angle+a)
	y2 := toY - l*math.Sin(angle+a)
	dc.SetColor(hexColor(r.theme.Flow))
	dc.SetLineWidth(1.4 * r.scale)
	dc.MoveTo(x1, y1)
	dc.LineTo(toX, toY)
	dc.LineTo(x2, y2)
	dc.Stroke()
}

func (r *DMNRasterRenderer) drawDot(dc *gg.Context, x, y float64) {
	dc.SetColor(hexColor(r.theme.Flow))
	dc.DrawCircle(x, y, 3.2*r.scale)
	dc.Fill()
}
