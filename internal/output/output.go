package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Format represents output format.
type Format int

const (
	FormatText Format = iota
	FormatJSON
)

func (f Format) String() string {
	switch f {
	case FormatJSON:
		return "json"
	default:
		return "text"
	}
}

// ParseFormat parses a format string.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "text":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	default:
		return FormatText, fmt.Errorf("invalid format: %s (valid: text, json)", s)
	}
}

// ColorMode represents output color preference.
type ColorMode int

const (
	ColorAuto ColorMode = iota
	ColorAlways
	ColorNever
)

// ParseColorMode parses a color mode string.
func ParseColorMode(s string) (ColorMode, error) {
	switch strings.ToLower(s) {
	case "auto":
		return ColorAuto, nil
	case "always":
		return ColorAlways, nil
	case "never":
		return ColorNever, nil
	default:
		return ColorAuto, fmt.Errorf("invalid color mode: %s (valid: auto, always, never)", s)
	}
}

// ResolveColorMode returns whether color should be enabled.
func ResolveColorMode(mode ColorMode, format Format, noColor bool, isTTY bool) bool {
	if format == FormatJSON || noColor {
		return false
	}
	switch mode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	default:
		return isTTY
	}
}

// PrintJSON writes data as indented JSON.
func PrintJSON(w io.Writer, data interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// PrintText writes formatted text with newline.
func PrintText(w io.Writer, format string, args ...interface{}) {
	fmt.Fprintf(w, format+"\n", args...)
}

// LogoResult represents logo output data.
type LogoResult struct {
	URL        string `json:"url"`
	Identifier string `json:"identifier,omitempty"`
	Format     string `json:"format,omitempty"`
	Theme      string `json:"theme,omitempty"`
	Type       string `json:"type,omitempty"`
	Fallback   string `json:"fallback,omitempty"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
}

// FormatLogo formats logo result.
func FormatLogo(logo *LogoResult, format Format) string {
	if format == FormatJSON {
		data, _ := json.MarshalIndent(logo, "", "  ")
		return string(data)
	}
	return logo.URL
}

// BrandResult represents brand output data.
type BrandResult struct {
	ID              string      `json:"id,omitempty"`
	Name            string      `json:"name"`
	Domain          string      `json:"domain"`
	Description     string      `json:"description,omitempty"`
	LongDescription string      `json:"longDescription,omitempty"`
	Claimed         bool        `json:"claimed,omitempty"`
	QualityScore    float64     `json:"qualityScore,omitempty"`
	IsNSFW          bool        `json:"isNsfw,omitempty"`
	URN             string      `json:"urn,omitempty"`
	Logos           []LogoInfo  `json:"logos,omitempty"`
	Colors          []ColorInfo `json:"colors,omitempty"`
	Fonts           []FontInfo  `json:"fonts,omitempty"`
	Links           []LinkInfo  `json:"links,omitempty"`
}

type LogoInfo struct {
	Type   string `json:"type"`
	Theme  string `json:"theme"`
	URL    string `json:"url"`
	Format string `json:"format"`
}

type ColorInfo struct {
	Hex        string `json:"hex"`
	Type       string `json:"type"`
	Brightness int    `json:"brightness"`
}

type FontInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type LinkInfo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// FormatBrand formats brand result.
func FormatBrand(brand *BrandResult, format Format, colorize bool) string {
	if format == FormatJSON {
		data, _ := json.MarshalIndent(brand, "", "  ")
		return string(data)
	}

	var sb strings.Builder
	if brand.ID != "" {
		sb.WriteString(fmt.Sprintf("%s (%s) [%s]\n", brand.Name, brand.Domain, brand.ID))
	} else {
		sb.WriteString(fmt.Sprintf("%s (%s)\n", brand.Name, brand.Domain))
	}
	if brand.Description != "" {
		sb.WriteString(fmt.Sprintf("\nDescription: %s\n", brand.Description))
	}
	if brand.LongDescription != "" {
		sb.WriteString(fmt.Sprintf("\nAbout: %s\n", brand.LongDescription))
	}
	if brand.URN != "" {
		sb.WriteString(fmt.Sprintf("\nURN: %s\n", brand.URN))
	}
	if brand.Claimed {
		sb.WriteString("\nClaimed: yes\n")
	}
	if brand.QualityScore > 0 {
		sb.WriteString(fmt.Sprintf("Quality score: %.2f\n", brand.QualityScore))
	}
	if brand.IsNSFW {
		sb.WriteString("NSFW: yes\n")
	}

	if len(brand.Logos) > 0 {
		sb.WriteString(fmt.Sprintf("\nLogos: %d available\n", len(brand.Logos)))
		for _, l := range brand.Logos {
			sb.WriteString(fmt.Sprintf("  - %s (%s): %s\n", l.Type, l.Theme, l.URL))
		}
	}

	if len(brand.Colors) > 0 {
		sb.WriteString("\nColors:\n")
		for _, c := range brand.Colors {
			sb.WriteString(fmt.Sprintf("  %s (%s)\n", colorizeHex(c.Hex, colorize), c.Type))
		}
	}

	if len(brand.Fonts) > 0 {
		sb.WriteString("\nFonts:\n")
		for _, f := range brand.Fonts {
			sb.WriteString(fmt.Sprintf("  %s (%s)\n", f.Name, f.Type))
		}
	}

	return sb.String()
}

// SearchResult represents search output data.
type SearchResult struct {
	Name    string `json:"name"`
	Domain  string `json:"domain"`
	Icon    string `json:"icon,omitempty"`
	Claimed bool   `json:"claimed,omitempty"`
	BrandID string `json:"brandId,omitempty"`
}

// FormatSearch formats search results.
func FormatSearch(results []SearchResult, format Format, colorize bool) string {
	if format == FormatJSON {
		data, _ := json.MarshalIndent(results, "", "  ")
		return string(data)
	}

	var sb strings.Builder
	for _, r := range results {
		domain := r.Domain
		if r.Claimed {
			domain = domain + " (claimed)"
		}
		if r.BrandID != "" {
			domain = domain + " [" + r.BrandID + "]"
		}
		sb.WriteString(fmt.Sprintf("%-30s %s\n", r.Name, domain))
	}
	return sb.String()
}

// FormatColors formats color palette.
func FormatColors(colors []ColorInfo, format Format, colorize bool) string {
	if format == FormatJSON {
		data, _ := json.MarshalIndent(colors, "", "  ")
		return string(data)
	}

	var sb strings.Builder
	for _, c := range colors {
		sb.WriteString(fmt.Sprintf("%s (%s)\n", colorizeHex(c.Hex, colorize), c.Type))
	}
	return sb.String()
}

// FormatFonts formats font list.
func FormatFonts(fonts []FontInfo, format Format, colorize bool) string {
	if format == FormatJSON {
		data, _ := json.MarshalIndent(fonts, "", "  ")
		return string(data)
	}

	var sb strings.Builder
	for _, f := range fonts {
		sb.WriteString(fmt.Sprintf("%s (%s)\n", f.Name, f.Type))
	}
	return sb.String()
}

// QuickResult represents the essentials output: logos, favicon, colors, fonts.
type QuickResult struct {
	Name      string      `json:"name"`
	Domain    string      `json:"domain"`
	LogoLight string      `json:"logo_light,omitempty"`
	LogoDark  string      `json:"logo_dark,omitempty"`
	Favicon   string      `json:"favicon,omitempty"`
	Colors    []ColorInfo `json:"colors"`
	Fonts     []FontInfo  `json:"fonts"`
}

// FormatQuick formats quick result (essentials).
func FormatQuick(result *QuickResult, format Format, colorize bool) string {
	if format == FormatJSON {
		data, _ := json.MarshalIndent(result, "", "  ")
		return string(data)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s (%s)\n", result.Name, result.Domain))

	// Logos
	sb.WriteString("\nLogos (SVG):\n")
	if result.LogoLight != "" {
		sb.WriteString(fmt.Sprintf("  light: %s\n", result.LogoLight))
	}
	if result.LogoDark != "" {
		sb.WriteString(fmt.Sprintf("  dark:  %s\n", result.LogoDark))
	}
	if result.LogoLight == "" && result.LogoDark == "" {
		sb.WriteString("  (no SVG available)\n")
	}

	// Favicon
	if result.Favicon != "" {
		sb.WriteString(fmt.Sprintf("\nFavicon:\n  %s\n", result.Favicon))
	}

	// Colors
	if len(result.Colors) > 0 {
		sb.WriteString("\nColors:\n")
		for _, c := range result.Colors {
			sb.WriteString(fmt.Sprintf("  %s (%s)\n", colorizeHex(c.Hex, colorize), c.Type))
		}
	}

	// Fonts
	if len(result.Fonts) > 0 {
		sb.WriteString("\nFonts:\n")
		for _, f := range result.Fonts {
			sb.WriteString(fmt.Sprintf("  %s (%s)\n", f.Name, f.Type))
		}
	}

	return sb.String()
}

// FormatQuickCSS formats quick result as CSS custom properties.
func FormatQuickCSS(result *QuickResult) string {
	var sb strings.Builder
	sb.WriteString(":root {\n")

	// Colors
	if len(result.Colors) > 0 {
		sb.WriteString("  /* Colors */\n")
		colorVars := buildColorVariables(result.Colors)
		for _, v := range colorVars {
			sb.WriteString(fmt.Sprintf("  %s: %s;\n", v.name, v.value))
		}
	}

	// Fonts
	if len(result.Fonts) > 0 {
		if len(result.Colors) > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("  /* Fonts */\n")
		fontVars := buildFontVariables(result.Fonts)
		for _, v := range fontVars {
			sb.WriteString(fmt.Sprintf("  %s: %s;\n", v.name, v.value))
		}
	}

	sb.WriteString("}")
	return sb.String()
}

type cssVar struct {
	name  string
	value string
}

// buildColorVariables generates CSS variable names for colors, handling duplicates.
func buildColorVariables(colors []ColorInfo) []cssVar {
	// Count occurrences of each type
	typeCounts := make(map[string]int)
	for _, c := range colors {
		typeCounts[c.Type]++
	}

	// Track which types we've seen (for numbering duplicates)
	typeIndex := make(map[string]int)

	var vars []cssVar
	for _, c := range colors {
		varName := fmt.Sprintf("--color-%s", c.Type)

		// If there are duplicates of this type, append a number
		if typeCounts[c.Type] > 1 {
			typeIndex[c.Type]++
			varName = fmt.Sprintf("--color-%s-%d", c.Type, typeIndex[c.Type])
		}

		vars = append(vars, cssVar{name: varName, value: c.Hex})
	}

	return vars
}

// FormatQuickTailwind formats quick result as Tailwind CSS config JavaScript.
func FormatQuickTailwind(result *QuickResult) string {
	var sb strings.Builder

	// Header comment
	sb.WriteString(fmt.Sprintf("// Tailwind CSS config for %s\n", result.Name))
	sb.WriteString("// Add to your tailwind.config.js theme.extend\n")
	sb.WriteString("module.exports = {\n")

	// Colors
	if len(result.Colors) > 0 {
		sb.WriteString("  colors: {\n")
		colorEntries := buildTailwindColors(result.Colors)
		for _, entry := range colorEntries {
			sb.WriteString(entry)
		}
		sb.WriteString("  },\n")
	}

	// Fonts
	if len(result.Fonts) > 0 {
		sb.WriteString("  fontFamily: {\n")
		fontEntries := buildTailwindFonts(result.Fonts)
		for _, entry := range fontEntries {
			sb.WriteString(entry)
		}
		sb.WriteString("  },\n")
	}

	sb.WriteString("}")
	return sb.String()
}

// buildTailwindColors generates Tailwind color entries, handling duplicates with object nesting.
func buildTailwindColors(colors []ColorInfo) []string {
	// Group colors by type to handle duplicates properly
	typeColors := make(map[string][]string)
	typeOrder := []string{} // Preserve order of first occurrence

	for _, c := range colors {
		if _, exists := typeColors[c.Type]; !exists {
			typeOrder = append(typeOrder, c.Type)
		}
		typeColors[c.Type] = append(typeColors[c.Type], c.Hex)
	}

	var entries []string
	for _, colorType := range typeOrder {
		hexValues := typeColors[colorType]
		if len(hexValues) == 1 {
			// Single color: simple key-value
			entries = append(entries, fmt.Sprintf("    %s: '%s',\n", colorType, hexValues[0]))
		} else {
			// Multiple colors: nested object
			var nested strings.Builder
			nested.WriteString(fmt.Sprintf("    %s: {\n", colorType))
			for i, hex := range hexValues {
				nested.WriteString(fmt.Sprintf("      %d: '%s',\n", i+1, hex))
			}
			nested.WriteString("    },\n")
			entries = append(entries, nested.String())
		}
	}

	return entries
}

// buildTailwindFonts generates Tailwind fontFamily entries, handling duplicates.
func buildTailwindFonts(fonts []FontInfo) []string {
	// Count occurrences of each type
	typeCounts := make(map[string]int)
	for _, f := range fonts {
		typeCounts[f.Type]++
	}

	// Track which types we've seen (for numbering duplicates)
	typeIndex := make(map[string]int)

	// Track unique fonts for deduplication (same name + type = skip)
	seen := make(map[string]bool)

	var entries []string
	for _, f := range fonts {
		key := f.Name + "|" + f.Type
		if seen[key] {
			continue
		}
		seen[key] = true

		if typeCounts[f.Type] > 1 {
			// Duplicate types - but for fonts we typically just list them
			// The spec says to use object nesting for colors, but for fonts
			// we'll follow the same pattern as CSS and number them
			typeIndex[f.Type]++
			entries = append(entries, fmt.Sprintf("    %s%d: ['\"%s\"', 'sans-serif'],\n", f.Type, typeIndex[f.Type], f.Name))
		} else {
			entries = append(entries, fmt.Sprintf("    %s: ['\"%s\"', 'sans-serif'],\n", f.Type, f.Name))
		}
	}

	return entries
}

// buildFontVariables generates CSS variable names for fonts, handling duplicates.
func buildFontVariables(fonts []FontInfo) []cssVar {
	// Count occurrences of each type
	typeCounts := make(map[string]int)
	for _, f := range fonts {
		typeCounts[f.Type]++
	}

	// Track which types we've seen (for numbering duplicates)
	typeIndex := make(map[string]int)

	// Track unique fonts for deduplication (same name + type = skip)
	seen := make(map[string]bool)

	var vars []cssVar
	for _, f := range fonts {
		key := f.Name + "|" + f.Type
		if seen[key] {
			continue
		}
		seen[key] = true

		varName := fmt.Sprintf("--font-%s", f.Type)

		// If there are duplicates of this type (after dedup), append a number
		if typeCounts[f.Type] > 1 {
			typeIndex[f.Type]++
			varName = fmt.Sprintf("--font-%s-%d", f.Type, typeIndex[f.Type])
		}

		// Quote font name and add sans-serif fallback
		value := fmt.Sprintf("'%s', sans-serif", f.Name)
		vars = append(vars, cssVar{name: varName, value: value})
	}

	return vars
}

// FormatQuickBatch formats multiple quick results for batch output.
func FormatQuickBatch(results []*QuickResult, format Format, colorize bool) string {
	if len(results) == 0 {
		return ""
	}

	// Single result: use original format
	if len(results) == 1 {
		return FormatQuick(results[0], format, colorize)
	}

	if format == FormatJSON {
		data, _ := json.MarshalIndent(results, "", "  ")
		return string(data)
	}

	// Text format: separate each brand with blank line
	var sb strings.Builder
	for i, result := range results {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(FormatQuick(result, format, colorize))
	}
	return sb.String()
}

// FormatQuickCSSBatch formats multiple quick results as CSS with brand-prefixed variables.
func FormatQuickCSSBatch(results []*QuickResult) string {
	if len(results) == 0 {
		return ":root {\n}"
	}

	// Single result: use original format
	if len(results) == 1 {
		return FormatQuickCSS(results[0])
	}

	var sb strings.Builder
	sb.WriteString(":root {\n")

	for i, result := range results {
		if i > 0 {
			sb.WriteString("\n")
		}
		brandPrefix := sanitizeCSSName(result.Domain)
		sb.WriteString(fmt.Sprintf("  /* %s */\n", result.Name))

		// Colors
		if len(result.Colors) > 0 {
			colorVars := buildColorVariablesWithPrefix(result.Colors, brandPrefix)
			for _, v := range colorVars {
				sb.WriteString(fmt.Sprintf("  %s: %s;\n", v.name, v.value))
			}
		}

		// Fonts
		if len(result.Fonts) > 0 {
			fontVars := buildFontVariablesWithPrefix(result.Fonts, brandPrefix)
			for _, v := range fontVars {
				sb.WriteString(fmt.Sprintf("  %s: %s;\n", v.name, v.value))
			}
		}
	}

	sb.WriteString("}")
	return sb.String()
}

// sanitizeCSSName converts a domain to a valid CSS variable name prefix.
func sanitizeCSSName(domain string) string {
	// Remove common TLDs
	name := strings.TrimSuffix(domain, ".com")
	name = strings.TrimSuffix(name, ".io")
	name = strings.TrimSuffix(name, ".org")
	name = strings.TrimSuffix(name, ".net")
	name = strings.TrimSuffix(name, ".co")
	// Replace dots with hyphens
	name = strings.ReplaceAll(name, ".", "-")
	return name
}

// buildColorVariablesWithPrefix generates CSS variable names with brand prefix.
func buildColorVariablesWithPrefix(colors []ColorInfo, prefix string) []cssVar {
	typeCounts := make(map[string]int)
	for _, c := range colors {
		typeCounts[c.Type]++
	}

	typeIndex := make(map[string]int)

	var vars []cssVar
	for _, c := range colors {
		varName := fmt.Sprintf("--%s-color-%s", prefix, c.Type)

		if typeCounts[c.Type] > 1 {
			typeIndex[c.Type]++
			varName = fmt.Sprintf("--%s-color-%s-%d", prefix, c.Type, typeIndex[c.Type])
		}

		vars = append(vars, cssVar{name: varName, value: c.Hex})
	}

	return vars
}

// buildFontVariablesWithPrefix generates CSS font variable names with brand prefix.
func buildFontVariablesWithPrefix(fonts []FontInfo, prefix string) []cssVar {
	typeCounts := make(map[string]int)
	for _, f := range fonts {
		typeCounts[f.Type]++
	}

	typeIndex := make(map[string]int)
	seen := make(map[string]bool)

	var vars []cssVar
	for _, f := range fonts {
		key := f.Name + "|" + f.Type
		if seen[key] {
			continue
		}
		seen[key] = true

		varName := fmt.Sprintf("--%s-font-%s", prefix, f.Type)

		if typeCounts[f.Type] > 1 {
			typeIndex[f.Type]++
			varName = fmt.Sprintf("--%s-font-%s-%d", prefix, f.Type, typeIndex[f.Type])
		}

		value := fmt.Sprintf("'%s', sans-serif", f.Name)
		vars = append(vars, cssVar{name: varName, value: value})
	}

	return vars
}

// FormatQuickTailwindBatch formats multiple quick results as Tailwind config with nested brand objects.
func FormatQuickTailwindBatch(results []*QuickResult) string {
	if len(results) == 0 {
		return "module.exports = {\n}"
	}

	// Single result: use original format
	if len(results) == 1 {
		return FormatQuickTailwind(results[0])
	}

	var sb strings.Builder
	sb.WriteString("// Tailwind CSS config for multiple brands\n")
	sb.WriteString("// Add to your tailwind.config.js theme.extend\n")
	sb.WriteString("module.exports = {\n")

	// Colors section
	hasColors := false
	for _, result := range results {
		if len(result.Colors) > 0 {
			hasColors = true
			break
		}
	}

	if hasColors {
		sb.WriteString("  colors: {\n")
		for _, result := range results {
			if len(result.Colors) == 0 {
				continue
			}
			brandKey := sanitizeTailwindKey(result.Domain)
			sb.WriteString(fmt.Sprintf("    %s: {\n", brandKey))
			colorEntries := buildTailwindColorsNested(result.Colors)
			for _, entry := range colorEntries {
				sb.WriteString(entry)
			}
			sb.WriteString("    },\n")
		}
		sb.WriteString("  },\n")
	}

	// Fonts section
	hasFonts := false
	for _, result := range results {
		if len(result.Fonts) > 0 {
			hasFonts = true
			break
		}
	}

	if hasFonts {
		sb.WriteString("  fontFamily: {\n")
		for _, result := range results {
			if len(result.Fonts) == 0 {
				continue
			}
			brandKey := sanitizeTailwindKey(result.Domain)
			sb.WriteString(fmt.Sprintf("    %s: {\n", brandKey))
			fontEntries := buildTailwindFontsNested(result.Fonts)
			for _, entry := range fontEntries {
				sb.WriteString(entry)
			}
			sb.WriteString("    },\n")
		}
		sb.WriteString("  },\n")
	}

	sb.WriteString("}")
	return sb.String()
}

// sanitizeTailwindKey converts a domain to a valid Tailwind config key.
func sanitizeTailwindKey(domain string) string {
	name := strings.TrimSuffix(domain, ".com")
	name = strings.TrimSuffix(name, ".io")
	name = strings.TrimSuffix(name, ".org")
	name = strings.TrimSuffix(name, ".net")
	name = strings.TrimSuffix(name, ".co")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

// buildTailwindColorsNested generates Tailwind color entries for nesting inside a brand object.
func buildTailwindColorsNested(colors []ColorInfo) []string {
	typeColors := make(map[string][]string)
	typeOrder := []string{}

	for _, c := range colors {
		if _, exists := typeColors[c.Type]; !exists {
			typeOrder = append(typeOrder, c.Type)
		}
		typeColors[c.Type] = append(typeColors[c.Type], c.Hex)
	}

	var entries []string
	for _, colorType := range typeOrder {
		hexValues := typeColors[colorType]
		if len(hexValues) == 1 {
			entries = append(entries, fmt.Sprintf("      %s: '%s',\n", colorType, hexValues[0]))
		} else {
			var nested strings.Builder
			nested.WriteString(fmt.Sprintf("      %s: {\n", colorType))
			for i, hex := range hexValues {
				nested.WriteString(fmt.Sprintf("        %d: '%s',\n", i+1, hex))
			}
			nested.WriteString("      },\n")
			entries = append(entries, nested.String())
		}
	}

	return entries
}

// buildTailwindFontsNested generates Tailwind font entries for nesting inside a brand object.
func buildTailwindFontsNested(fonts []FontInfo) []string {
	typeCounts := make(map[string]int)
	for _, f := range fonts {
		typeCounts[f.Type]++
	}

	typeIndex := make(map[string]int)
	seen := make(map[string]bool)

	var entries []string
	for _, f := range fonts {
		key := f.Name + "|" + f.Type
		if seen[key] {
			continue
		}
		seen[key] = true

		if typeCounts[f.Type] > 1 {
			typeIndex[f.Type]++
			entries = append(entries, fmt.Sprintf("      %s%d: ['\"%s\"', 'sans-serif'],\n", f.Type, typeIndex[f.Type], f.Name))
		} else {
			entries = append(entries, fmt.Sprintf("      %s: ['\"%s\"', 'sans-serif'],\n", f.Type, f.Name))
		}
	}

	return entries
}

func colorizeHex(hex string, enabled bool) string {
	if !enabled {
		return hex
	}
	if len(hex) != 7 || !strings.HasPrefix(hex, "#") {
		return hex
	}

	r, err1 := strconv.ParseInt(hex[1:3], 16, 0)
	g, err2 := strconv.ParseInt(hex[3:5], 16, 0)
	b, err3 := strconv.ParseInt(hex[5:7], 16, 0)
	if err1 != nil || err2 != nil || err3 != nil {
		return hex
	}

	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", r, g, b, hex)
}
