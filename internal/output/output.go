package output

import (
	"encoding/json"
	"fmt"
	"io"
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
	URL    string `json:"url"`
	Format string `json:"format"`
	Theme  string `json:"theme"`
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
	Name        string      `json:"name"`
	Domain      string      `json:"domain"`
	Description string      `json:"description,omitempty"`
	Logos       []LogoInfo  `json:"logos,omitempty"`
	Colors      []ColorInfo `json:"colors,omitempty"`
	Fonts       []FontInfo  `json:"fonts,omitempty"`
	Links       []LinkInfo  `json:"links,omitempty"`
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
func FormatBrand(brand *BrandResult, format Format) string {
	if format == FormatJSON {
		data, _ := json.MarshalIndent(brand, "", "  ")
		return string(data)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s (%s)\n", brand.Name, brand.Domain))
	if brand.Description != "" {
		sb.WriteString(fmt.Sprintf("\nDescription: %s\n", brand.Description))
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
			sb.WriteString(fmt.Sprintf("  %s (%s)\n", c.Hex, c.Type))
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
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Icon   string `json:"icon,omitempty"`
}

// FormatSearch formats search results.
func FormatSearch(results []SearchResult, format Format) string {
	if format == FormatJSON {
		data, _ := json.MarshalIndent(results, "", "  ")
		return string(data)
	}

	var sb strings.Builder
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("%-30s %s\n", r.Name, r.Domain))
	}
	return sb.String()
}

// FormatColors formats color palette.
func FormatColors(colors []ColorInfo, format Format) string {
	if format == FormatJSON {
		data, _ := json.MarshalIndent(colors, "", "  ")
		return string(data)
	}

	var sb strings.Builder
	for _, c := range colors {
		sb.WriteString(fmt.Sprintf("%s (%s)\n", c.Hex, c.Type))
	}
	return sb.String()
}

// FormatFonts formats font list.
func FormatFonts(fonts []FontInfo, format Format) string {
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
func FormatQuick(result *QuickResult, format Format) string {
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
			sb.WriteString(fmt.Sprintf("  %s (%s)\n", c.Hex, c.Type))
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
