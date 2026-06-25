package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xdung24/bpmn-to-image/bpmn"
	"github.com/xdung24/bpmn-to-image/dmn"
	"github.com/xdung24/bpmn-to-image/renderer"
)

var version = "0.2.2"

func main() {
	var (
		input   string
		output  string
		format  string
		scale   float64
		quality int
		padding float64
	)

	flag.StringVar(&input, "input", "", "Path to input BPMN or DMN file (required)")
	flag.StringVar(&input, "i", "", "Path to input file (shorthand)")
	flag.StringVar(&output, "output", "", "Path to output image file (default: input filename with new extension)")
	flag.StringVar(&output, "o", "", "Path to output image file (shorthand)")
	flag.StringVar(&format, "format", "", "Output format: svg, png, jpg (default: inferred from output extension)")
	flag.StringVar(&format, "f", "", "Output format (shorthand)")
	flag.Float64Var(&scale, "scale", 2.0, "Scale factor for raster output (default: 2.0)")
	flag.IntVar(&quality, "quality", 90, "JPEG quality 1-100 (default: 90)")
	flag.Float64Var(&padding, "padding", 30, "Padding around the diagram in pixels (default: 30)")
	ver := flag.Bool("version", false, "Print version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "bpmn-to-image - Convert BPMN or DMN files to images (SVG, PNG, JPG)\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  bpmn-to-image -i <input.bpmn|input.dmn> [-o <output.png>] [-f svg|png|jpg]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  bpmn-to-image -i process.bpmn -o process.svg\n")
		fmt.Fprintf(os.Stderr, "  bpmn-to-image -i process.bpmn -f png\n")
		fmt.Fprintf(os.Stderr, "  bpmn-to-image -i decisions.dmn -o decisions.png\n")
		fmt.Fprintf(os.Stderr, "  bpmn-to-image -i decisions.dmn -f svg\n")
	}

	flag.Parse()

	if *ver {
		fmt.Printf("bpmn-to-image v%s\n", version)
		os.Exit(0)
	}

	if input == "" {
		if flag.NArg() > 0 {
			input = flag.Arg(0)
		} else {
			fmt.Fprintln(os.Stderr, "Error: input file is required")
			flag.Usage()
			os.Exit(1)
		}
	}

	kind, err := detectInputKind(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Determine output format
	format = strings.ToLower(format)
	if format == "" {
		if output != "" {
			ext := strings.ToLower(filepath.Ext(output))
			switch ext {
			case ".svg":
				format = "svg"
			case ".png":
				format = "png"
			case ".jpg", ".jpeg":
				format = "jpg"
			default:
				fmt.Fprintf(os.Stderr, "Error: cannot determine format from extension '%s'; use -f to specify\n", ext)
				os.Exit(1)
			}
		} else {
			format = "svg"
		}
	}

	// Determine output path
	if output == "" {
		base := strings.TrimSuffix(input, filepath.Ext(input))
		switch format {
		case "svg":
			output = base + ".svg"
		case "png":
			output = base + ".png"
		case "jpg":
			output = base + ".jpg"
		}
	}

	switch format {
	case "svg", "png", "jpg":
	default:
		fmt.Fprintf(os.Stderr, "Error: unsupported format '%s'; use svg, png, or jpg\n", format)
		os.Exit(1)
	}

	switch kind {
	case "bpmn":
		if err := convertBPMN(input, output, format, scale, quality, padding); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "dmn":
		if err := convertDMN(input, output, format, scale, quality, padding); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

// detectInputKind returns "bpmn" or "dmn" based on the file extension.
func detectInputKind(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".bpmn", ".bpmn2":
		return "bpmn", nil
	case ".dmn":
		return "dmn", nil
	default:
		return "", fmt.Errorf("unsupported input extension %q; expected .bpmn or .dmn", ext)
	}
}

func convertBPMN(input, output, format string, scale float64, quality int, padding float64) error {
	defs, err := bpmn.Parse(input)
	if err != nil {
		return fmt.Errorf("parsing BPMN file: %w", err)
	}

	validationErrors := bpmn.Validate(defs)
	if len(validationErrors) > 0 {
		hasErrors := bpmn.HasErrors(validationErrors)
		for _, ve := range validationErrors {
			fmt.Fprintln(os.Stderr, ve.String())
		}
		if hasErrors {
			return fmt.Errorf("BPMN file has errors; cannot render")
		}
		fmt.Fprintln(os.Stderr, "")
	}

	switch format {
	case "svg":
		data, err := renderer.NewSVGRenderer(padding).Render(defs)
		if err != nil {
			return fmt.Errorf("rendering SVG: %w", err)
		}
		return os.WriteFile(output, data, 0644)
	case "png":
		return renderer.NewRasterRenderer(scale, padding).RenderPNG(defs, output)
	case "jpg":
		return renderer.NewRasterRenderer(scale, padding).RenderJPG(defs, output, quality)
	}
	return nil
}

func convertDMN(input, output, format string, scale float64, quality int, padding float64) error {
	defs, err := dmn.Parse(input)
	if err != nil {
		return fmt.Errorf("parsing DMN file: %w", err)
	}

	validationErrors := dmn.Validate(defs)
	if len(validationErrors) > 0 {
		hasErrors := dmn.HasErrors(validationErrors)
		for _, ve := range validationErrors {
			fmt.Fprintln(os.Stderr, ve.String())
		}
		if hasErrors {
			return fmt.Errorf("DMN file has errors; cannot render")
		}
		fmt.Fprintln(os.Stderr, "")
	}

	switch format {
	case "svg":
		data, err := renderer.NewDMNSVGRenderer(padding).Render(defs)
		if err != nil {
			return fmt.Errorf("rendering SVG: %w", err)
		}
		return os.WriteFile(output, data, 0644)
	case "png":
		return renderer.NewDMNRasterRenderer(scale, padding).RenderPNG(defs, output)
	case "jpg":
		return renderer.NewDMNRasterRenderer(scale, padding).RenderJPG(defs, output, quality)
	}
	return nil
}
