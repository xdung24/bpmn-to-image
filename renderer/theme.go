package renderer

import "image/color"

// Theme defines the color palette used by both the SVG and raster renderers.
// Colors are inspired by a modern flat design with semantic coloring:
// green for start, red for end, blue for activities, amber for gateways.
type Theme struct {
	CanvasBg string

	TaskStroke string
	TaskFill   string

	SubProcessStroke string
	SubProcessFill   string

	StartStroke string
	StartFill   string

	EndStroke string
	EndFill   string

	IntermediateStroke string
	IntermediateFill   string

	GatewayStroke string
	GatewayFill   string
	GatewayMarker string

	Flow  string
	Label string

	PoolStroke string
	PoolFill   string
	PoolHeader string

	LaneStroke string

	AnnotationStroke string
	IconColor        string

	// DMN-specific colors
	DecisionStroke        string
	DecisionFill          string
	InputDataStroke       string
	InputDataFill         string
	BKMStroke             string
	BKMFill               string
	KnowledgeSourceStroke string
	KnowledgeSourceFill   string
	DecisionServiceStroke string
	DecisionServiceFill   string
}

// DefaultTheme is the standard colored theme.
var DefaultTheme = Theme{
	CanvasBg: "#ffffff",

	TaskStroke: "#2563eb", // blue-600
	TaskFill:   "#eff6ff", // blue-50

	SubProcessStroke: "#7c3aed", // violet-600
	SubProcessFill:   "#f5f3ff", // violet-50

	StartStroke: "#16a34a", // green-600
	StartFill:   "#dcfce7", // green-100

	EndStroke: "#dc2626", // red-600
	EndFill:   "#fee2e2", // red-100

	IntermediateStroke: "#d97706", // amber-600
	IntermediateFill:   "#fef3c7", // amber-100

	GatewayStroke: "#ca8a04", // yellow-600
	GatewayFill:   "#fefce8", // yellow-50
	GatewayMarker: "#a16207", // yellow-700

	Flow:  "#475569", // slate-600
	Label: "#1e293b", // slate-800

	PoolStroke: "#475569", // slate-600
	PoolFill:   "#ffffff",
	PoolHeader: "#f1f5f9", // slate-100

	LaneStroke: "#94a3b8", // slate-400

	AnnotationStroke: "#94a3b8", // slate-400
	IconColor:        "#334155", // slate-700

	// DMN palette
	DecisionStroke:        "#2563eb", // blue-600 (decisions = primary "task" color)
	DecisionFill:          "#eff6ff", // blue-50
	InputDataStroke:       "#0891b2", // cyan-600
	InputDataFill:         "#ecfeff", // cyan-50
	BKMStroke:             "#7c3aed", // violet-600
	BKMFill:               "#f5f3ff", // violet-50
	KnowledgeSourceStroke: "#d97706", // amber-600
	KnowledgeSourceFill:   "#fffbeb", // amber-50
	DecisionServiceStroke: "#475569", // slate-600
	DecisionServiceFill:   "#f8fafc", // slate-50
}

// hexColor converts a "#rrggbb" string to a color.Color. Falls back to black.
func hexColor(hex string) color.Color {
	if len(hex) == 7 && hex[0] == '#' {
		r := hexByte(hex[1], hex[2])
		g := hexByte(hex[3], hex[4])
		b := hexByte(hex[5], hex[6])
		return color.RGBA{R: r, G: g, B: b, A: 255}
	}
	return color.Black
}

func hexByte(hi, lo byte) uint8 {
	return hexNibble(hi)<<4 | hexNibble(lo)
}

func hexNibble(c byte) uint8 {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}
