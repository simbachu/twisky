// Command iconsprite packs Inkscape multipage icons-raw.svg into a clean
// SVG symbol sheet for //go:embed. Invoked by go generate ./static.
package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	svgNS      = "http://www.w3.org/2000/svg"
	inkscapeNS = "http://www.inkscape.org/namespaces/inkscape"
	sodipodiNS = "http://sodipodi.sourceforge.net/DTD/sodipodi-0.dtd"
)

const sheetHeader = `<!-- DO NOT EDIT.
 * Generated from internal/assets/icons/svg/icons-raw.svg.
 * Regenerate with: go generate ./static
 -->
`

func main() {
	rawPath, outPath, err := resolvePaths()
	if err != nil {
		fmt.Fprintf(os.Stderr, "iconsprite: %v\n", err)
		os.Exit(1)
	}

	raw, err := os.ReadFile(rawPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "iconsprite: read %s: %v\n", rawPath, err)
		os.Exit(1)
	}

	out, err := pack(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "iconsprite: pack: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "iconsprite: mkdir: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(outPath, out, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "iconsprite: write %s: %v\n", outPath, err)
		os.Exit(1)
	}
}

func resolvePaths() (rawPath, outPath string, err error) {
	if len(os.Args) > 2 {
		return os.Args[1], os.Args[2], nil
	}
	candidates := []struct {
		raw string
		out string
	}{
		{
			filepath.Join("..", "internal", "assets", "icons", "svg", "icons-raw.svg"),
			filepath.Join("web", "icons", "icons.svg"),
		}, // go generate from ./static
		{
			filepath.Join("internal", "assets", "icons", "svg", "icons-raw.svg"),
			filepath.Join("static", "web", "icons", "icons.svg"),
		}, // go run from module root
		{
			filepath.Join("..", "..", "internal", "assets", "icons", "svg", "icons-raw.svg"),
			filepath.Join("..", "..", "static", "web", "icons", "icons.svg"),
		}, // go run from cmd/iconsprite
	}
	for _, c := range candidates {
		if _, statErr := os.Stat(c.raw); statErr == nil {
			return c.raw, c.out, nil
		}
	}
	return "", "", fmt.Errorf("could not find icons-raw.svg; pass raw and out paths as argv")
}

type element struct {
	name     xml.Name
	attrs    []xml.Attr
	children []*element
	text     string
}

func pack(raw []byte) ([]byte, error) {
	root, err := parseSVG(raw)
	if err != nil {
		return nil, err
	}

	pages, err := extractPages(root)
	if err != nil {
		return nil, err
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("no labeled inkscape:page elements found")
	}

	seen := make(map[string]struct{}, len(pages))
	for _, p := range pages {
		if _, ok := seen[p.label]; ok {
			return nil, fmt.Errorf("duplicate page label %q", p.label)
		}
		seen[p.label] = struct{}{}
	}

	layers := findLayers(root)
	if len(layers) == 0 {
		return nil, fmt.Errorf("no inkscape layer groups found")
	}

	var symbols []*element
	for _, page := range pages {
		group, layer, err := matchIconGroup(layers, page)
		if err != nil {
			return nil, err
		}
		sym, err := buildSymbol(page, layer, group)
		if err != nil {
			return nil, err
		}
		symbols = append(symbols, sym)
	}

	sheet := &element{
		name: xml.Name{Space: svgNS, Local: "svg"},
		attrs: []xml.Attr{
			{Name: xml.Name{Local: "xmlns"}, Value: svgNS},
			{Name: xml.Name{Local: "style"}, Value: "display:none"},
		},
		children: symbols,
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	buf.WriteString(sheetHeader)
	if err := writeElement(&buf, sheet, 0); err != nil {
		return nil, err
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

type page struct {
	label  string
	x, y   float64
	width  float64
	height float64
}

func extractPages(root *element) ([]page, error) {
	var pages []page
	var walk func(*element)
	walk = func(el *element) {
		if isInkscape(el, "page") {
			label := attrValue(el, inkscapeNS, "label")
			if label == "" {
				return
			}
			x, _ := parseFloat(attrValueLocal(el, "x"))
			y, _ := parseFloat(attrValueLocal(el, "y"))
			w, errW := parseFloat(attrValueLocal(el, "width"))
			h, errH := parseFloat(attrValueLocal(el, "height"))
			if errW != nil || errH != nil || w <= 0 || h <= 0 {
				return
			}
			pages = append(pages, page{label: label, x: x, y: y, width: w, height: h})
			return
		}
		for _, c := range el.children {
			walk(c)
		}
	}
	walk(root)
	return pages, nil
}

func findLayers(root *element) []*element {
	var layers []*element
	var walk func(*element)
	walk = func(el *element) {
		if isSVG(el, "g") && attrValue(el, inkscapeNS, "groupmode") == "layer" {
			layers = append(layers, el)
			return
		}
		for _, c := range el.children {
			walk(c)
		}
	}
	walk(root)
	return layers
}

func matchIconGroup(layers []*element, page page) (group, layer *element, err error) {
	const eps = 0.5

	var byGeom *element
	var byGeomLayer *element

	for _, lay := range layers {
		for _, child := range lay.children {
			if !isSVG(child, "g") {
				continue
			}
			label := attrValue(child, inkscapeNS, "label")
			if label == page.label {
				return child, lay, nil
			}
			tx, ty, ok := parseTranslate(attrValueLocal(child, "transform"))
			if !ok {
				continue
			}
			// Inkscape often offsets icons slightly within the page; treat the
			// group's translate origin as belonging to a page when it falls inside.
			if tx >= page.x-eps && tx < page.x+page.width+eps &&
				ty >= page.y-eps && ty < page.y+page.height+eps {
				byGeom = child
				byGeomLayer = lay
			}
		}
	}

	if byGeom != nil {
		return byGeom, byGeomLayer, nil
	}
	return nil, nil, fmt.Errorf("labeled page %q has no drawable content", page.label)
}

func buildSymbol(page page, layer, group *element) (*element, error) {
	cleanedGroup := cleanElement(group)
	if cleanedGroup == nil || !hasDrawable(cleanedGroup) {
		return nil, fmt.Errorf("labeled page %q has no drawable content", page.label)
	}

	inner := cleanedGroup
	if layerTx := attrValueLocal(layer, "transform"); layerTx != "" {
		inner = &element{
			name: xml.Name{Space: svgNS, Local: "g"},
			attrs: []xml.Attr{
				{Name: xml.Name{Local: "transform"}, Value: layerTx},
			},
			children: []*element{cleanedGroup},
		}
	}

	content := inner
	if page.x != 0 || page.y != 0 {
		content = &element{
			name: xml.Name{Space: svgNS, Local: "g"},
			attrs: []xml.Attr{
				{Name: xml.Name{Local: "transform"}, Value: fmt.Sprintf("translate(%s,%s)", formatNum(-page.x), formatNum(-page.y))},
			},
			children: []*element{inner},
		}
	}

	return &element{
		name: xml.Name{Space: svgNS, Local: "symbol"},
		attrs: []xml.Attr{
			{Name: xml.Name{Local: "id"}, Value: page.label},
			{Name: xml.Name{Local: "viewBox"}, Value: fmt.Sprintf("0 0 %s %s", formatNum(page.width), formatNum(page.height))},
		},
		children: []*element{content},
	}, nil
}

func hasDrawable(el *element) bool {
	switch el.name.Local {
	case "path", "rect", "circle", "ellipse", "line", "polyline", "polygon", "use", "text", "image":
		return true
	}
	for _, c := range el.children {
		if hasDrawable(c) {
			return true
		}
	}
	return false
}

func cleanElement(el *element) *element {
	if el == nil {
		return nil
	}
	if el.name.Space == inkscapeNS || el.name.Space == sodipodiNS {
		return nil
	}
	if el.name.Space != "" && el.name.Space != svgNS {
		return nil
	}

	out := &element{
		name: xml.Name{Space: svgNS, Local: el.name.Local},
		text: strings.TrimSpace(el.text),
	}
	for _, a := range el.attrs {
		if a.Name.Space == inkscapeNS || a.Name.Space == sodipodiNS {
			continue
		}
		if a.Name.Local == "xmlns" || strings.HasPrefix(a.Name.Local, "xmlns:") {
			continue
		}
		// Drop namespace declarations carried as attrs with Space.
		if a.Name.Space == "http://www.w3.org/2000/xmlns/" {
			continue
		}
		out.attrs = append(out.attrs, xml.Attr{
			Name:  xml.Name{Local: a.Name.Local},
			Value: rewritePaintToCurrentColor(a.Name.Local, a.Value),
		})
	}
	for _, c := range el.children {
		cleaned := cleanElement(c)
		if cleaned != nil {
			out.children = append(out.children, cleaned)
		}
	}
	return out
}

// rewritePaintToCurrentColor replaces solid black fills/strokes with
// currentColor so CSS color on the referencing element controls the icon.
// Values that are already none/transparent are left alone.
func rewritePaintToCurrentColor(attrName, value string) string {
	switch attrName {
	case "fill", "stroke":
		if isBlackPaint(value) {
			return "currentColor"
		}
		return value
	case "style":
		return rewriteStylePaints(value)
	default:
		return value
	}
}

func rewriteStylePaints(style string) string {
	parts := strings.Split(style, ";")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, val, ok := strings.Cut(part, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if (key == "fill" || key == "stroke") && isBlackPaint(val) {
			parts[i] = key + ":currentColor"
		}
	}
	return strings.Join(parts, ";")
}

func isBlackPaint(value string) bool {
	v := strings.TrimSpace(strings.ToLower(value))
	switch v {
	case "#000", "#000000", "black", "rgb(0,0,0)", "rgb(0, 0, 0)":
		return true
	default:
		return false
	}
}

func parseSVG(raw []byte) (*element, error) {
	dec := xml.NewDecoder(bytes.NewReader(raw))
	dec.DefaultSpace = svgNS

	var stack []*element
	var root *element

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			el := &element{name: t.Name, attrs: append([]xml.Attr(nil), t.Attr...)}
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.children = append(parent.children, el)
			} else {
				root = el
			}
			stack = append(stack, el)
		case xml.EndElement:
			if len(stack) == 0 {
				return nil, fmt.Errorf("unexpected end element")
			}
			stack = stack[:len(stack)-1]
		case xml.CharData:
			if len(stack) == 0 {
				continue
			}
			text := string(t)
			if strings.TrimSpace(text) == "" {
				continue
			}
			cur := stack[len(stack)-1]
			cur.text += text
		}
	}
	if root == nil {
		return nil, fmt.Errorf("no root element")
	}
	return root, nil
}

func writeElement(w io.Writer, el *element, depth int) error {
	indent := strings.Repeat("  ", depth)
	if _, err := io.WriteString(w, indent+"<"+el.name.Local); err != nil {
		return err
	}
	for _, a := range el.attrs {
		if _, err := fmt.Fprintf(w, ` %s="%s"`, a.Name.Local, xmlEscape(a.Value)); err != nil {
			return err
		}
	}
	if len(el.children) == 0 && el.text == "" {
		_, err := io.WriteString(w, " />\n")
		return err
	}
	if _, err := io.WriteString(w, ">"); err != nil {
		return err
	}
	if el.text != "" && len(el.children) == 0 {
		if _, err := io.WriteString(w, xmlEscape(el.text)); err != nil {
			return err
		}
		_, err := io.WriteString(w, "</"+el.name.Local+">\n")
		return err
	}
	if _, err := io.WriteString(w, "\n"); err != nil {
		return err
	}
	for _, c := range el.children {
		if err := writeElement(w, c, depth+1); err != nil {
			return err
		}
	}
	if el.text != "" {
		if _, err := io.WriteString(w, indent+"  "+xmlEscape(el.text)+"\n"); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, indent+"</"+el.name.Local+">\n")
	return err
}

func xmlEscape(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}

func isInkscape(el *element, local string) bool {
	return el.name.Space == inkscapeNS && el.name.Local == local
}

func isSVG(el *element, local string) bool {
	return (el.name.Space == svgNS || el.name.Space == "") && el.name.Local == local
}

func attrValue(el *element, space, local string) string {
	for _, a := range el.attrs {
		if a.Name.Space == space && a.Name.Local == local {
			return a.Value
		}
		// Some encoders put prefixed attrs with empty Space and Local "inkscape:label".
		if a.Name.Space == "" && a.Name.Local == local {
			return a.Value
		}
		if space == inkscapeNS && a.Name.Local == "inkscape:"+local {
			return a.Value
		}
		if space == sodipodiNS && a.Name.Local == "sodipodi:"+local {
			return a.Value
		}
	}
	return ""
}

func attrValueLocal(el *element, local string) string {
	for _, a := range el.attrs {
		if a.Name.Local == local && (a.Name.Space == "" || a.Name.Space == svgNS) {
			return a.Value
		}
	}
	return ""
}

func parseFloat(s string) (float64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	return strconv.ParseFloat(s, 64)
}

func formatNum(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}

func parseTranslate(transform string) (x, y float64, ok bool) {
	transform = strings.TrimSpace(transform)
	if transform == "" {
		return 0, 0, false
	}
	const prefix = "translate("
	if !strings.HasPrefix(transform, prefix) {
		return 0, 0, false
	}
	rest := strings.TrimPrefix(transform, prefix)
	end := strings.IndexByte(rest, ')')
	if end < 0 {
		return 0, 0, false
	}
	inside := strings.TrimSpace(rest[:end])
	inside = strings.ReplaceAll(inside, ",", " ")
	fields := strings.Fields(inside)
	if len(fields) == 0 || len(fields) > 2 {
		return 0, 0, false
	}
	x, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, 0, false
	}
	if len(fields) == 2 {
		y, err = strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return 0, 0, false
		}
	}
	return x, y, true
}
