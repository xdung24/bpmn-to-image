# bpmn-to-image

Convert BPMN and DMN diagrams to high-quality images (SVG, PNG, JPG).

## Features

- **BPMN 2.0 Support**: Parse and render Business Process Model and Notation diagrams with full XML namespace handling
- **DMN 1.3+ Support**: Render Decision Requirements Diagrams (DRD) from Decision Model and Notation files
- **Multiple Output Formats**: SVG (vector), PNG, and JPEG (raster)
- **Styled Rendering**: Modern flat design with semantic colors, gradients, shadows, and icons
- **Scalable**: Configurable scale factor for raster output
- **Pure Go**: No external dependencies for core rendering; uses `gg` for raster graphics

## Supported Diagrams

### BPMN
- Flow nodes: start/end events, tasks (generic, user, service, script, send, receive, manual, business rule), intermediate events, boundary events, subprocesses, gateways (exclusive, parallel, inclusive, event-based)
- Connectors: sequence flows (with conditional expressions), associations, data flows
- Containers: pools, lanes
- Annotations: text annotations
- DI coordinates automatically applied from BPMN Diagram Interchange

### DMN
- DRD elements: decisions, input data, business knowledge models, knowledge sources, decision services, text annotations
- Requirements: information requirements (solid arrow), knowledge requirements (dashed arrow), authority requirements (dashed with dot)
- Associations: text annotation links
- DI coordinates automatically applied from DMN Diagram Interchange

## Installation

### From Source

```bash
git clone https://github.com/xdung24/bpmn-to-image.git
cd bpmn-to-image
go build ./...
```

The binary `bpmn-to-image` (or `bpmn-to-image.exe` on Windows) will be created in the current directory.

## Usage

### Basic Command

```bash
bpmn-to-image -i <input.bpmn|input.dmn> [-o <output.png>] [-f svg|png|jpg]
```

### Examples

**Convert BPMN to SVG** (default format):
```bash
bpmn-to-image -i process.bpmn -o process.svg
```

**Convert BPMN to PNG with scale 3**:
```bash
bpmn-to-image -i process.bpmn -f png --scale 3
```

**Convert DMN to PNG**:
```bash
bpmn-to-image -i decisions.dmn -o decisions.png
```

**Convert DMN to JPEG with quality 85**:
```bash
bpmn-to-image -i decisions.dmn -f jpg --quality 85
```

### Flags

- `-i, --input` *(required)* — Path to input BPMN or DMN file
- `-o, --output` — Path to output image file. If omitted, defaults to input filename with the output format extension
- `-f, --format` — Output format: `svg`, `png`, or `jpg`. If omitted, inferred from output file extension (default: `svg`)
- `--scale` — Scale factor for raster output; applies to PNG and JPG (default: `2.0`)
- `--quality` — JPEG quality 1–100 (default: `90`)
- `--version` — Print version and exit
- `-h, --help` — Print help

## Features

### BPMN Rendering

- **Icons**: User tasks show a person icon; service tasks show a cog; script tasks show a script icon
- **Shapes**: Start/end events (circles), tasks (rectangles), gateways (diamonds with markers X/+/O), subprocesses (rounded rectangles with optional expansion indicator)
- **Connectors**: Arrows with labels, conditional flow branches
- **Styling**: Gradient fills, soft shadows, semantic colors (green for start, red for end, blue for tasks, yellow for gateways)
- **Text Wrapping**: Metrics-based label wrapping ensures labels fit within shapes at any scale

### DMN Rendering

- **Shapes**: Decisions (rounded rectangles), input data (stadium), BKM (rectangle with clipped corners), knowledge sources (rectangle with wavy bottom), decision services (rounded boxes with dashed border), text annotations (bracket)
- **Requirements**: Solid arrows (information), dashed arrows (knowledge), dashed + dot (authority), dashed + thin (association)
- **Styling**: Same color palette and design language as BPMN
- **Nested Containers**: Decision services rendered as background containers

## Architecture

### Packages

- **`bpmn/`** — BPMN 2.0 data model, parser, and validator
  - `model.go` — XML struct definitions
  - `parser.go` — Read and parse BPMN files, strip namespaces
  - `validator.go` — Validate structure (check for missing start/end events, broken references)

- **`dmn/`** — DMN 1.3+ data model, parser, and validator
  - `model.go` — XML struct definitions
  - `parser.go` — Read and parse DMN files, build node/edge indexes
  - `validator.go` — Validate structure

- **`renderer/`** — Rendering backends
  - `theme.go` — Shared color palette and color utilities
  - `svg.go` — SVG renderer for BPMN (produces vector XML)
  - `raster.go` — PNG/JPG renderer for BPMN (produces raster images via `gg` + `freetype`)
  - `dmn_svg.go` — SVG renderer for DMN
  - `dmn_raster.go` — PNG/JPG renderer for DMN

- **`main.go`** — CLI entry point; auto-detects input type by extension and dispatches to the appropriate converter

## Dependencies

- `github.com/fogleman/gg` v1.3.0 — 2D graphics for raster rendering
- `github.com/golang/freetype/truetype` — TrueType font parsing
- `golang.org/x/image/font/gofont/goregular` — Embedded TrueType font (Segoe UI Regular)

## Examples

### BPMN File

A typical BPMN file with diagram interchange:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="http://www.omg.org/spec/BPMN/20100524/MODEL">
  <process id="Process_1" name="Order Process">
    <startEvent id="start" name="Order Received"/>
    <task id="task1" name="Validate Order"/>
    <exclusiveGateway id="gw1" name="Valid?"/>
    <endEvent id="end" name="Order Processed"/>
    <sequenceFlow sourceRef="start" targetRef="task1"/>
    <sequenceFlow sourceRef="task1" targetRef="gw1"/>
  </process>
  <bpmndi:BPMNDiagram>
    <bpmndi:BPMNPlane>
      <bpmndi:BPMNShape bpmnElement="start">
        <dc:Bounds x="100" y="100" width="50" height="50"/>
      </bpmndi:BPMNShape>
      <!-- ... more shapes and edges ... -->
    </bpmndi:BPMNPlane>
  </bpmndi:BPMNDiagram>
</definitions>
```

### DMN File

A typical DMN file with a DRD:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<definitions xmlns="https://www.omg.org/spec/DMN/20191111/MODEL/">
  <inputData id="age" name="Applicant Age"/>
  <decision id="eligible" name="Eligibility">
    <informationRequirement>
      <requiredInput href="#age"/>
    </informationRequirement>
  </decision>
  <dmndi:DMNDI>
    <dmndi:DMNDiagram>
      <dmndi:DMNShape dmnElementRef="age">
        <dc:Bounds x="100" y="100" width="140" height="44"/>
      </dmndi:DMNShape>
      <!-- ... more shapes and edges ... -->
    </dmndi:DMNDiagram>
  </dmndi:DMNDI>
</definitions>
```

## Validation

The tool validates input files and prints warnings/errors before rendering:

- **Errors**: Block rendering (missing process, broken sequence flow references, invalid gateway defaults)
- **Warnings**: Printed but do not block rendering (no start/end events, missing diagram interchange)

Example output:

```
[warning] process 'Process_1' has no start event (element: Process_1)
[error] sequence flow 'Flow_1' references unknown target 'NonExistent_id' (element: Flow_1)
```

## Performance

- Parsing: ~50ms for a typical 100-node diagram
- Rendering (SVG): ~100ms
- Rendering (PNG at scale 2): ~200ms

Times vary with complexity and scale.

## Limitations

- **Decision Tables**: DMN decision tables are parsed structurally but their rule content and tabular display are not rendered (only the DRD is rendered)
- **Subprocess Expansion**: Collapsed/expanded state is respected but subprocess internals are not drawn
- **Custom Extensions**: Vendor extensions (Camunda, Trisotech, etc.) are ignored but do not cause errors

## License

MIT License. See [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please fork the repository, create a feature branch, and submit a pull request.

## Version

`v0.2.3`
