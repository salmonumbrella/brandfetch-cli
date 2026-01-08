package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"name": "GitHub", "domain": "github.com"}

	err := PrintJSON(&buf, data)
	if err != nil {
		t.Fatalf("PrintJSON() error = %v", err)
	}

	// Should be valid JSON
	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if result["name"] != "GitHub" {
		t.Errorf("name = %v, want GitHub", result["name"])
	}
}

func TestPrintText(t *testing.T) {
	var buf bytes.Buffer
	PrintText(&buf, "Hello %s", "World")

	if got := buf.String(); got != "Hello World\n" {
		t.Errorf("PrintText() = %q, want %q", got, "Hello World\n")
	}
}

func TestFormat_String(t *testing.T) {
	tests := []struct {
		f    Format
		want string
	}{
		{FormatText, "text"},
		{FormatJSON, "json"},
	}

	for _, tt := range tests {
		if got := tt.f.String(); got != tt.want {
			t.Errorf("Format.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input string
		want  Format
		err   bool
	}{
		{"text", FormatText, false},
		{"json", FormatJSON, false},
		{"TEXT", FormatText, false},
		{"JSON", FormatJSON, false},
		{"invalid", FormatText, true},
	}

	for _, tt := range tests {
		got, err := ParseFormat(tt.input)
		if (err != nil) != tt.err {
			t.Errorf("ParseFormat(%q) error = %v, wantErr %v", tt.input, err, tt.err)
			continue
		}
		if !tt.err && got != tt.want {
			t.Errorf("ParseFormat(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestFormatLogo_Text(t *testing.T) {
	logo := &LogoResult{
		URL:    "https://example.com/logo.svg",
		Format: "svg",
		Theme:  "light",
	}

	result := FormatLogo(logo, FormatText)

	if !strings.Contains(result, "https://example.com/logo.svg") {
		t.Errorf("FormatLogo() text missing URL")
	}
}

func TestFormatLogo_JSON(t *testing.T) {
	logo := &LogoResult{
		URL:    "https://example.com/logo.svg",
		Format: "svg",
		Theme:  "light",
	}

	result := FormatLogo(logo, FormatJSON)

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("FormatLogo() JSON invalid: %v", err)
	}

	if parsed["url"] != "https://example.com/logo.svg" {
		t.Errorf("FormatLogo() JSON url = %v, want %v", parsed["url"], "https://example.com/logo.svg")
	}
}

func TestFormatBrand_Text(t *testing.T) {
	brand := &BrandResult{
		Name:        "GitHub",
		Domain:      "github.com",
		Description: "Where the world builds software",
		Logos: []LogoInfo{
			{Type: "icon", Theme: "dark", URL: "https://example.com/icon.svg", Format: "svg"},
			{Type: "logo", Theme: "light", URL: "https://example.com/logo.png", Format: "png"},
		},
		Colors: []ColorInfo{
			{Hex: "#000000", Type: "dark", Brightness: 0},
			{Hex: "#ffffff", Type: "light", Brightness: 100},
		},
		Fonts: []FontInfo{
			{Name: "Inter", Type: "body"},
			{Name: "Helvetica", Type: "heading"},
		},
		Links: []LinkInfo{
			{Name: "Website", URL: "https://github.com"},
		},
	}

	result := FormatBrand(brand, FormatText, false)

	// Check essential parts are present
	if !strings.Contains(result, "GitHub") {
		t.Errorf("FormatBrand() text missing name")
	}
	if !strings.Contains(result, "github.com") {
		t.Errorf("FormatBrand() text missing domain")
	}
	if !strings.Contains(result, "Where the world builds software") {
		t.Errorf("FormatBrand() text missing description")
	}
	if !strings.Contains(result, "Logos: 2 available") {
		t.Errorf("FormatBrand() text missing logos count")
	}
	if !strings.Contains(result, "icon (dark)") {
		t.Errorf("FormatBrand() text missing logo info")
	}
	if !strings.Contains(result, "#000000 (dark)") {
		t.Errorf("FormatBrand() text missing color info")
	}
	if !strings.Contains(result, "Inter (body)") {
		t.Errorf("FormatBrand() text missing font info")
	}
}

func TestFormatBrand_Text_Empty(t *testing.T) {
	brand := &BrandResult{
		Name:   "MinimalBrand",
		Domain: "minimal.com",
	}

	result := FormatBrand(brand, FormatText, false)

	if !strings.Contains(result, "MinimalBrand") {
		t.Errorf("FormatBrand() text missing name")
	}
	if !strings.Contains(result, "minimal.com") {
		t.Errorf("FormatBrand() text missing domain")
	}
	// Should not contain sections for empty slices
	if strings.Contains(result, "Logos:") {
		t.Errorf("FormatBrand() text should not show logos section when empty")
	}
	if strings.Contains(result, "Colors:") {
		t.Errorf("FormatBrand() text should not show colors section when empty")
	}
	if strings.Contains(result, "Fonts:") {
		t.Errorf("FormatBrand() text should not show fonts section when empty")
	}
}

func TestFormatBrand_JSON(t *testing.T) {
	brand := &BrandResult{
		Name:        "GitHub",
		Domain:      "github.com",
		Description: "Where the world builds software",
		Logos: []LogoInfo{
			{Type: "icon", Theme: "dark", URL: "https://example.com/icon.svg", Format: "svg"},
		},
		Colors: []ColorInfo{
			{Hex: "#000000", Type: "dark", Brightness: 0},
		},
		Fonts: []FontInfo{
			{Name: "Inter", Type: "body"},
		},
	}

	result := FormatBrand(brand, FormatJSON, false)

	var parsed BrandResult
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("FormatBrand() JSON invalid: %v", err)
	}

	if parsed.Name != "GitHub" {
		t.Errorf("FormatBrand() JSON name = %v, want GitHub", parsed.Name)
	}
	if parsed.Domain != "github.com" {
		t.Errorf("FormatBrand() JSON domain = %v, want github.com", parsed.Domain)
	}
	if len(parsed.Logos) != 1 {
		t.Errorf("FormatBrand() JSON logos count = %v, want 1", len(parsed.Logos))
	}
	if len(parsed.Colors) != 1 {
		t.Errorf("FormatBrand() JSON colors count = %v, want 1", len(parsed.Colors))
	}
	if len(parsed.Fonts) != 1 {
		t.Errorf("FormatBrand() JSON fonts count = %v, want 1", len(parsed.Fonts))
	}
}

func TestFormatBrand_JSON_SpecialCharacters(t *testing.T) {
	brand := &BrandResult{
		Name:        "Test & Co. \"Quotes\" <Tags>",
		Domain:      "test.com",
		Description: "Line 1\nLine 2\tTabbed",
	}

	result := FormatBrand(brand, FormatJSON, false)

	var parsed BrandResult
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("FormatBrand() JSON with special chars invalid: %v", err)
	}

	if parsed.Name != "Test & Co. \"Quotes\" <Tags>" {
		t.Errorf("FormatBrand() JSON name with special chars = %v", parsed.Name)
	}
}

func TestFormatSearch_Text(t *testing.T) {
	results := []SearchResult{
		{Name: "GitHub", Domain: "github.com", Icon: "https://example.com/icon.png"},
		{Name: "GitLab", Domain: "gitlab.com", Icon: "https://example.com/icon2.png"},
		{Name: "Bitbucket", Domain: "bitbucket.org"},
	}

	result := FormatSearch(results, FormatText, false)

	if !strings.Contains(result, "GitHub") {
		t.Errorf("FormatSearch() text missing first result name")
	}
	if !strings.Contains(result, "github.com") {
		t.Errorf("FormatSearch() text missing first result domain")
	}
	if !strings.Contains(result, "GitLab") {
		t.Errorf("FormatSearch() text missing second result")
	}
	if !strings.Contains(result, "Bitbucket") {
		t.Errorf("FormatSearch() text missing third result")
	}

	// Check that results are formatted on separate lines
	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 3 {
		t.Errorf("FormatSearch() text line count = %v, want 3", len(lines))
	}
}

func TestFormatSearch_Text_Empty(t *testing.T) {
	results := []SearchResult{}

	result := FormatSearch(results, FormatText, false)

	if result != "" {
		t.Errorf("FormatSearch() text for empty results = %q, want empty string", result)
	}
}

func TestFormatSearch_JSON(t *testing.T) {
	results := []SearchResult{
		{Name: "GitHub", Domain: "github.com", Icon: "https://example.com/icon.png"},
		{Name: "GitLab", Domain: "gitlab.com"},
	}

	result := FormatSearch(results, FormatJSON, false)

	var parsed []SearchResult
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("FormatSearch() JSON invalid: %v", err)
	}

	if len(parsed) != 2 {
		t.Errorf("FormatSearch() JSON count = %v, want 2", len(parsed))
	}
	if parsed[0].Name != "GitHub" {
		t.Errorf("FormatSearch() JSON first name = %v, want GitHub", parsed[0].Name)
	}
	if parsed[1].Domain != "gitlab.com" {
		t.Errorf("FormatSearch() JSON second domain = %v, want gitlab.com", parsed[1].Domain)
	}
}

func TestFormatSearch_JSON_Empty(t *testing.T) {
	results := []SearchResult{}

	result := FormatSearch(results, FormatJSON, false)

	var parsed []SearchResult
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("FormatSearch() JSON for empty results invalid: %v", err)
	}

	if len(parsed) != 0 {
		t.Errorf("FormatSearch() JSON empty results count = %v, want 0", len(parsed))
	}
}

func TestFormatColors_Text(t *testing.T) {
	colors := []ColorInfo{
		{Hex: "#ff0000", Type: "primary", Brightness: 50},
		{Hex: "#00ff00", Type: "secondary", Brightness: 75},
		{Hex: "#0000ff", Type: "accent", Brightness: 40},
	}

	result := FormatColors(colors, FormatText, false)

	if !strings.Contains(result, "#ff0000 (primary)") {
		t.Errorf("FormatColors() text missing first color")
	}
	if !strings.Contains(result, "#00ff00 (secondary)") {
		t.Errorf("FormatColors() text missing second color")
	}
	if !strings.Contains(result, "#0000ff (accent)") {
		t.Errorf("FormatColors() text missing third color")
	}

	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 3 {
		t.Errorf("FormatColors() text line count = %v, want 3", len(lines))
	}
}

func TestFormatColors_Text_Empty(t *testing.T) {
	colors := []ColorInfo{}

	result := FormatColors(colors, FormatText, false)

	if result != "" {
		t.Errorf("FormatColors() text for empty colors = %q, want empty string", result)
	}
}

func TestFormatColors_JSON(t *testing.T) {
	colors := []ColorInfo{
		{Hex: "#ff0000", Type: "primary", Brightness: 50},
		{Hex: "#00ff00", Type: "secondary", Brightness: 75},
	}

	result := FormatColors(colors, FormatJSON, false)

	var parsed []ColorInfo
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("FormatColors() JSON invalid: %v", err)
	}

	if len(parsed) != 2 {
		t.Errorf("FormatColors() JSON count = %v, want 2", len(parsed))
	}
	if parsed[0].Hex != "#ff0000" {
		t.Errorf("FormatColors() JSON first hex = %v, want #ff0000", parsed[0].Hex)
	}
	if parsed[0].Brightness != 50 {
		t.Errorf("FormatColors() JSON first brightness = %v, want 50", parsed[0].Brightness)
	}
	if parsed[1].Type != "secondary" {
		t.Errorf("FormatColors() JSON second type = %v, want secondary", parsed[1].Type)
	}
}

func TestFormatFonts_Text(t *testing.T) {
	fonts := []FontInfo{
		{Name: "Inter", Type: "body"},
		{Name: "Helvetica Neue", Type: "heading"},
		{Name: "Monaco", Type: "monospace"},
	}

	result := FormatFonts(fonts, FormatText, false)

	if !strings.Contains(result, "Inter (body)") {
		t.Errorf("FormatFonts() text missing first font")
	}
	if !strings.Contains(result, "Helvetica Neue (heading)") {
		t.Errorf("FormatFonts() text missing second font")
	}
	if !strings.Contains(result, "Monaco (monospace)") {
		t.Errorf("FormatFonts() text missing third font")
	}

	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 3 {
		t.Errorf("FormatFonts() text line count = %v, want 3", len(lines))
	}
}

func TestFormatFonts_Text_Empty(t *testing.T) {
	fonts := []FontInfo{}

	result := FormatFonts(fonts, FormatText, false)

	if result != "" {
		t.Errorf("FormatFonts() text for empty fonts = %q, want empty string", result)
	}
}

func TestFormatFonts_JSON(t *testing.T) {
	fonts := []FontInfo{
		{Name: "Inter", Type: "body"},
		{Name: "Helvetica", Type: "heading"},
	}

	result := FormatFonts(fonts, FormatJSON, false)

	var parsed []FontInfo
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("FormatFonts() JSON invalid: %v", err)
	}

	if len(parsed) != 2 {
		t.Errorf("FormatFonts() JSON count = %v, want 2", len(parsed))
	}
	if parsed[0].Name != "Inter" {
		t.Errorf("FormatFonts() JSON first name = %v, want Inter", parsed[0].Name)
	}
	if parsed[0].Type != "body" {
		t.Errorf("FormatFonts() JSON first type = %v, want body", parsed[0].Type)
	}
	if parsed[1].Name != "Helvetica" {
		t.Errorf("FormatFonts() JSON second name = %v, want Helvetica", parsed[1].Name)
	}
}

func TestFormatFonts_Text_SpecialCharacters(t *testing.T) {
	fonts := []FontInfo{
		{Name: "Font \"With\" Quotes", Type: "special"},
		{Name: "Font & Symbols <test>", Type: "test"},
	}

	result := FormatFonts(fonts, FormatText, false)

	if !strings.Contains(result, "Font \"With\" Quotes (special)") {
		t.Errorf("FormatFonts() text not handling special characters in name")
	}
	if !strings.Contains(result, "Font & Symbols <test> (test)") {
		t.Errorf("FormatFonts() text not handling special characters")
	}
}

func TestFormatQuickCSS_Basic(t *testing.T) {
	result := &QuickResult{
		Name:   "Stripe",
		Domain: "stripe.com",
		Colors: []ColorInfo{
			{Hex: "#635BFF", Type: "accent", Brightness: 50},
			{Hex: "#0A2540", Type: "dark", Brightness: 10},
			{Hex: "#FFFFFF", Type: "light", Brightness: 100},
		},
		Fonts: []FontInfo{
			{Name: "Sohne Var", Type: "title"},
			{Name: "Sohne Var", Type: "body"},
		},
	}

	output := FormatQuickCSS(result)

	// Check structure
	if !strings.Contains(output, ":root {") {
		t.Errorf("output should start with :root {")
	}
	if !strings.HasSuffix(output, "}") {
		t.Errorf("output should end with }")
	}

	// Check comments
	if !strings.Contains(output, "/* Colors */") {
		t.Errorf("output should contain Colors comment")
	}
	if !strings.Contains(output, "/* Fonts */") {
		t.Errorf("output should contain Fonts comment")
	}

	// Check color variables
	if !strings.Contains(output, "--color-accent: #635BFF;") {
		t.Errorf("output should contain accent color")
	}
	if !strings.Contains(output, "--color-dark: #0A2540;") {
		t.Errorf("output should contain dark color")
	}
	if !strings.Contains(output, "--color-light: #FFFFFF;") {
		t.Errorf("output should contain light color")
	}

	// Check font variables with fallback
	if !strings.Contains(output, "--font-title: 'Sohne Var', sans-serif;") {
		t.Errorf("output should contain title font with fallback")
	}
	if !strings.Contains(output, "--font-body: 'Sohne Var', sans-serif;") {
		t.Errorf("output should contain body font with fallback")
	}
}

func TestFormatQuickCSS_DuplicateColorTypes(t *testing.T) {
	result := &QuickResult{
		Colors: []ColorInfo{
			{Hex: "#FF0000", Type: "brand"},
			{Hex: "#00FF00", Type: "brand"},
			{Hex: "#0000FF", Type: "brand"},
			{Hex: "#FFFFFF", Type: "light"},
		},
	}

	output := FormatQuickCSS(result)

	// Duplicate types should be numbered
	if !strings.Contains(output, "--color-brand-1: #FF0000;") {
		t.Errorf("output should contain --color-brand-1")
	}
	if !strings.Contains(output, "--color-brand-2: #00FF00;") {
		t.Errorf("output should contain --color-brand-2")
	}
	if !strings.Contains(output, "--color-brand-3: #0000FF;") {
		t.Errorf("output should contain --color-brand-3")
	}

	// Non-duplicate should NOT be numbered
	if !strings.Contains(output, "--color-light: #FFFFFF;") {
		t.Errorf("output should contain --color-light without number")
	}
	if strings.Contains(output, "--color-light-1") {
		t.Errorf("output should not number non-duplicate types")
	}
}

func TestFormatQuickCSS_DuplicateFontTypes(t *testing.T) {
	result := &QuickResult{
		Fonts: []FontInfo{
			{Name: "Roboto", Type: "body"},
			{Name: "Open Sans", Type: "body"},
			{Name: "Inter", Type: "title"},
		},
	}

	output := FormatQuickCSS(result)

	// Duplicate types should be numbered
	if !strings.Contains(output, "--font-body-1: 'Roboto', sans-serif;") {
		t.Errorf("output should contain --font-body-1")
	}
	if !strings.Contains(output, "--font-body-2: 'Open Sans', sans-serif;") {
		t.Errorf("output should contain --font-body-2")
	}

	// Non-duplicate should NOT be numbered
	if !strings.Contains(output, "--font-title: 'Inter', sans-serif;") {
		t.Errorf("output should contain --font-title without number")
	}
}

func TestFormatQuickCSS_Empty(t *testing.T) {
	result := &QuickResult{
		Name:   "Empty",
		Domain: "empty.com",
	}

	output := FormatQuickCSS(result)

	// Should still have valid structure
	if !strings.Contains(output, ":root {") {
		t.Errorf("output should contain :root {")
	}
	if !strings.HasSuffix(output, "}") {
		t.Errorf("output should end with }")
	}

	// Should NOT have comments for empty sections
	if strings.Contains(output, "/* Colors */") {
		t.Errorf("output should not contain Colors comment when no colors")
	}
	if strings.Contains(output, "/* Fonts */") {
		t.Errorf("output should not contain Fonts comment when no fonts")
	}
}

func TestFormatQuickCSS_OnlyColors(t *testing.T) {
	result := &QuickResult{
		Colors: []ColorInfo{
			{Hex: "#FF0000", Type: "primary"},
		},
	}

	output := FormatQuickCSS(result)

	if !strings.Contains(output, "/* Colors */") {
		t.Errorf("output should contain Colors comment")
	}
	if strings.Contains(output, "/* Fonts */") {
		t.Errorf("output should not contain Fonts comment when no fonts")
	}
	if !strings.Contains(output, "--color-primary: #FF0000;") {
		t.Errorf("output should contain primary color")
	}
}

func TestFormatQuickCSS_OnlyFonts(t *testing.T) {
	result := &QuickResult{
		Fonts: []FontInfo{
			{Name: "Arial", Type: "body"},
		},
	}

	output := FormatQuickCSS(result)

	if strings.Contains(output, "/* Colors */") {
		t.Errorf("output should not contain Colors comment when no colors")
	}
	if !strings.Contains(output, "/* Fonts */") {
		t.Errorf("output should contain Fonts comment")
	}
	if !strings.Contains(output, "--font-body: 'Arial', sans-serif;") {
		t.Errorf("output should contain body font")
	}
}

func TestFormatQuickCSS_FontsWithSpecialChars(t *testing.T) {
	result := &QuickResult{
		Fonts: []FontInfo{
			{Name: "Sohne Var", Type: "title"},
			{Name: "SF Pro Display", Type: "body"},
		},
	}

	output := FormatQuickCSS(result)

	// Font names should be quoted
	if !strings.Contains(output, "'Sohne Var'") {
		t.Errorf("output should quote font name with space")
	}
	if !strings.Contains(output, "'SF Pro Display'") {
		t.Errorf("output should quote font name with spaces")
	}
}

func TestFormatQuickTailwind_Basic(t *testing.T) {
	result := &QuickResult{
		Name:   "Stripe",
		Domain: "stripe.com",
		Colors: []ColorInfo{
			{Hex: "#635BFF", Type: "accent", Brightness: 50},
			{Hex: "#0A2540", Type: "dark", Brightness: 10},
			{Hex: "#FFFFFF", Type: "light", Brightness: 100},
		},
		Fonts: []FontInfo{
			{Name: "Sohne Var", Type: "title"},
			{Name: "Sohne Var", Type: "body"},
		},
	}

	output := FormatQuickTailwind(result)

	// Check header comments
	if !strings.Contains(output, "// Tailwind CSS config for Stripe") {
		t.Errorf("output should contain brand name in comment")
	}
	if !strings.Contains(output, "// Add to your tailwind.config.js theme.extend") {
		t.Errorf("output should contain usage hint comment")
	}

	// Check module.exports structure
	if !strings.Contains(output, "module.exports = {") {
		t.Errorf("output should contain module.exports = {")
	}
	if !strings.HasSuffix(output, "}") {
		t.Errorf("output should end with }")
	}

	// Check colors section
	if !strings.Contains(output, "colors: {") {
		t.Errorf("output should contain colors: {")
	}
	if !strings.Contains(output, "accent: '#635BFF',") {
		t.Errorf("output should contain accent color")
	}
	if !strings.Contains(output, "dark: '#0A2540',") {
		t.Errorf("output should contain dark color")
	}
	if !strings.Contains(output, "light: '#FFFFFF',") {
		t.Errorf("output should contain light color")
	}

	// Check fontFamily section
	if !strings.Contains(output, "fontFamily: {") {
		t.Errorf("output should contain fontFamily: {")
	}
	if !strings.Contains(output, `title: ['"Sohne Var"', 'sans-serif'],`) {
		t.Errorf("output should contain title font with double quotes and fallback")
	}
	if !strings.Contains(output, `body: ['"Sohne Var"', 'sans-serif'],`) {
		t.Errorf("output should contain body font with double quotes and fallback")
	}
}

func TestFormatQuickTailwind_DuplicateColorTypes(t *testing.T) {
	result := &QuickResult{
		Name: "Test",
		Colors: []ColorInfo{
			{Hex: "#FF0000", Type: "brand"},
			{Hex: "#00FF00", Type: "brand"},
			{Hex: "#0000FF", Type: "brand"},
			{Hex: "#FFFFFF", Type: "light"},
		},
	}

	output := FormatQuickTailwind(result)

	// Duplicate types should use nested object format with all values grouped
	if !strings.Contains(output, "brand: {") {
		t.Errorf("output should contain brand nested object")
	}
	if !strings.Contains(output, "1: '#FF0000',") {
		t.Errorf("output should contain 1: '#FF0000'")
	}
	if !strings.Contains(output, "2: '#00FF00',") {
		t.Errorf("output should contain 2: '#00FF00'")
	}
	if !strings.Contains(output, "3: '#0000FF',") {
		t.Errorf("output should contain 3: '#0000FF'")
	}

	// Non-duplicate should NOT use nested object
	if !strings.Contains(output, "light: '#FFFFFF',") {
		t.Errorf("output should contain light color without nesting")
	}
	if strings.Contains(output, "light: {") {
		t.Errorf("output should not nest non-duplicate types")
	}
}

func TestFormatQuickTailwind_DuplicateFontTypes(t *testing.T) {
	result := &QuickResult{
		Name: "Test",
		Fonts: []FontInfo{
			{Name: "Roboto", Type: "body"},
			{Name: "Open Sans", Type: "body"},
			{Name: "Inter", Type: "title"},
		},
	}

	output := FormatQuickTailwind(result)

	// Duplicate types should be numbered
	if !strings.Contains(output, `body1: ['"Roboto"', 'sans-serif'],`) {
		t.Errorf("output should contain body1 for first body font")
	}
	if !strings.Contains(output, `body2: ['"Open Sans"', 'sans-serif'],`) {
		t.Errorf("output should contain body2 for second body font")
	}

	// Non-duplicate should NOT be numbered
	if !strings.Contains(output, `title: ['"Inter"', 'sans-serif'],`) {
		t.Errorf("output should contain title without number")
	}
}

func TestFormatQuickTailwind_Empty(t *testing.T) {
	result := &QuickResult{
		Name:   "Empty",
		Domain: "empty.com",
	}

	output := FormatQuickTailwind(result)

	// Should have valid structure
	if !strings.Contains(output, "module.exports = {") {
		t.Errorf("output should contain module.exports = {")
	}
	if !strings.HasSuffix(output, "}") {
		t.Errorf("output should end with }")
	}

	// Should NOT have colors or fontFamily sections
	if strings.Contains(output, "colors: {") {
		t.Errorf("output should not contain colors section when no colors")
	}
	if strings.Contains(output, "fontFamily: {") {
		t.Errorf("output should not contain fontFamily section when no fonts")
	}
}

func TestFormatQuickTailwind_OnlyColors(t *testing.T) {
	result := &QuickResult{
		Name: "Test",
		Colors: []ColorInfo{
			{Hex: "#FF0000", Type: "primary"},
		},
	}

	output := FormatQuickTailwind(result)

	if !strings.Contains(output, "colors: {") {
		t.Errorf("output should contain colors section")
	}
	if strings.Contains(output, "fontFamily: {") {
		t.Errorf("output should not contain fontFamily section when no fonts")
	}
	if !strings.Contains(output, "primary: '#FF0000',") {
		t.Errorf("output should contain primary color")
	}
}

func TestFormatQuickTailwind_OnlyFonts(t *testing.T) {
	result := &QuickResult{
		Name: "Test",
		Fonts: []FontInfo{
			{Name: "Arial", Type: "body"},
		},
	}

	output := FormatQuickTailwind(result)

	if strings.Contains(output, "colors: {") {
		t.Errorf("output should not contain colors section when no colors")
	}
	if !strings.Contains(output, "fontFamily: {") {
		t.Errorf("output should contain fontFamily section")
	}
	if !strings.Contains(output, `body: ['"Arial"', 'sans-serif'],`) {
		t.Errorf("output should contain body font")
	}
}

func TestFormatQuickTailwind_FontsWithSpaces(t *testing.T) {
	result := &QuickResult{
		Name: "Test",
		Fonts: []FontInfo{
			{Name: "Sohne Var", Type: "title"},
			{Name: "SF Pro Display", Type: "body"},
		},
	}

	output := FormatQuickTailwind(result)

	// Font names should be in double quotes inside the array
	if !strings.Contains(output, `"Sohne Var"`) {
		t.Errorf("output should contain font name with space in double quotes")
	}
	if !strings.Contains(output, `"SF Pro Display"`) {
		t.Errorf("output should contain font name with spaces in double quotes")
	}
}

// Batch mode tests

func TestFormatQuickBatch_SingleResult(t *testing.T) {
	results := []*QuickResult{
		{
			Name:   "Stripe",
			Domain: "stripe.com",
			Colors: []ColorInfo{{Hex: "#635BFF", Type: "accent"}},
		},
	}

	// Single result should use original format (not array for JSON)
	output := FormatQuickBatch(results, FormatJSON, false)
	if strings.HasPrefix(output, "[") {
		t.Errorf("single result should not be an array: %s", output)
	}

	var result QuickResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("output should be valid single JSON object: %v", err)
	}
	if result.Name != "Stripe" {
		t.Errorf("expected Stripe, got %s", result.Name)
	}
}

func TestFormatQuickBatch_MultipleResults_JSON(t *testing.T) {
	results := []*QuickResult{
		{Name: "Stripe", Domain: "stripe.com", Colors: []ColorInfo{{Hex: "#635BFF", Type: "accent"}}},
		{Name: "GitHub", Domain: "github.com", Colors: []ColorInfo{{Hex: "#24292f", Type: "dark"}}},
	}

	output := FormatQuickBatch(results, FormatJSON, false)

	var parsed []QuickResult
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("output should be valid JSON array: %v", err)
	}

	if len(parsed) != 2 {
		t.Errorf("expected 2 results, got %d", len(parsed))
	}
	if parsed[0].Name != "Stripe" {
		t.Errorf("first result should be Stripe")
	}
	if parsed[1].Name != "GitHub" {
		t.Errorf("second result should be GitHub")
	}
}

func TestFormatQuickBatch_MultipleResults_Text(t *testing.T) {
	results := []*QuickResult{
		{Name: "Stripe", Domain: "stripe.com"},
		{Name: "GitHub", Domain: "github.com"},
	}

	output := FormatQuickBatch(results, FormatText, false)

	if !strings.Contains(output, "Stripe") {
		t.Errorf("output should contain Stripe")
	}
	if !strings.Contains(output, "GitHub") {
		t.Errorf("output should contain GitHub")
	}

	// Should have blank line between results (the separator)
	if !strings.Contains(output, "\n\n") {
		t.Errorf("output should have blank line between results")
	}
}

func TestFormatQuickBatch_Empty(t *testing.T) {
	var results []*QuickResult

	output := FormatQuickBatch(results, FormatText, false)
	if output != "" {
		t.Errorf("empty results should return empty string")
	}
}

func TestFormatQuickCSSBatch_SingleResult(t *testing.T) {
	results := []*QuickResult{
		{
			Name:   "Stripe",
			Domain: "stripe.com",
			Colors: []ColorInfo{{Hex: "#635BFF", Type: "accent"}},
		},
	}

	output := FormatQuickCSSBatch(results)

	// Single result should NOT have brand prefix
	if !strings.Contains(output, "--color-accent: #635BFF;") {
		t.Errorf("single result should not have brand prefix: %s", output)
	}
	if strings.Contains(output, "--stripe-") {
		t.Errorf("single result should not have brand prefix")
	}
}

func TestFormatQuickCSSBatch_MultipleResults(t *testing.T) {
	results := []*QuickResult{
		{Name: "Stripe", Domain: "stripe.com", Colors: []ColorInfo{{Hex: "#635BFF", Type: "accent"}}},
		{Name: "GitHub", Domain: "github.com", Colors: []ColorInfo{{Hex: "#24292f", Type: "dark"}}},
	}

	output := FormatQuickCSSBatch(results)

	// Should have brand-prefixed variables
	if !strings.Contains(output, "--stripe-color-accent: #635BFF;") {
		t.Errorf("output should have stripe-prefixed variable: %s", output)
	}
	if !strings.Contains(output, "--github-color-dark: #24292f;") {
		t.Errorf("output should have github-prefixed variable: %s", output)
	}

	// Should have brand comments
	if !strings.Contains(output, "/* Stripe */") {
		t.Errorf("output should have Stripe comment")
	}
	if !strings.Contains(output, "/* GitHub */") {
		t.Errorf("output should have GitHub comment")
	}
}

func TestFormatQuickCSSBatch_WithFonts(t *testing.T) {
	results := []*QuickResult{
		{
			Name:   "Stripe",
			Domain: "stripe.com",
			Colors: []ColorInfo{{Hex: "#635BFF", Type: "accent"}},
			Fonts:  []FontInfo{{Name: "Sohne Var", Type: "title"}},
		},
		{
			Name:   "GitHub",
			Domain: "github.com",
			Fonts:  []FontInfo{{Name: "Mona Sans", Type: "body"}},
		},
	}

	output := FormatQuickCSSBatch(results)

	if !strings.Contains(output, "--stripe-font-title: 'Sohne Var', sans-serif;") {
		t.Errorf("output should have stripe-prefixed font: %s", output)
	}
	if !strings.Contains(output, "--github-font-body: 'Mona Sans', sans-serif;") {
		t.Errorf("output should have github-prefixed font: %s", output)
	}
}

func TestFormatQuickCSSBatch_Empty(t *testing.T) {
	var results []*QuickResult

	output := FormatQuickCSSBatch(results)

	if output != ":root {\n}" {
		t.Errorf("empty results should return valid empty CSS: %s", output)
	}
}

func TestFormatQuickTailwindBatch_SingleResult(t *testing.T) {
	results := []*QuickResult{
		{
			Name:   "Stripe",
			Domain: "stripe.com",
			Colors: []ColorInfo{{Hex: "#635BFF", Type: "accent"}},
		},
	}

	output := FormatQuickTailwindBatch(results)

	// Single result should use original format (no nesting)
	if !strings.Contains(output, "accent: '#635BFF',") {
		t.Errorf("single result should not have brand nesting: %s", output)
	}
	if strings.Contains(output, "stripe: {") {
		t.Errorf("single result should not have brand object")
	}
}

func TestFormatQuickTailwindBatch_MultipleResults(t *testing.T) {
	results := []*QuickResult{
		{Name: "Stripe", Domain: "stripe.com", Colors: []ColorInfo{{Hex: "#635BFF", Type: "accent"}}},
		{Name: "GitHub", Domain: "github.com", Colors: []ColorInfo{{Hex: "#24292f", Type: "dark"}}},
	}

	output := FormatQuickTailwindBatch(results)

	// Should have nested brand objects
	if !strings.Contains(output, "stripe: {") {
		t.Errorf("output should have stripe nested object: %s", output)
	}
	if !strings.Contains(output, "github: {") {
		t.Errorf("output should have github nested object: %s", output)
	}

	// Should have header comment for multiple brands
	if !strings.Contains(output, "// Tailwind CSS config for multiple brands") {
		t.Errorf("output should have multiple brands comment")
	}
}

func TestFormatQuickTailwindBatch_WithFonts(t *testing.T) {
	results := []*QuickResult{
		{
			Name:   "Stripe",
			Domain: "stripe.com",
			Fonts:  []FontInfo{{Name: "Sohne Var", Type: "title"}},
		},
		{
			Name:   "GitHub",
			Domain: "github.com",
			Fonts:  []FontInfo{{Name: "Mona Sans", Type: "body"}},
		},
	}

	output := FormatQuickTailwindBatch(results)

	// Should have fontFamily section with nested brand objects
	if !strings.Contains(output, "fontFamily: {") {
		t.Errorf("output should have fontFamily section")
	}
	if !strings.Contains(output, "stripe: {") {
		t.Errorf("output should have stripe nested object in fontFamily")
	}
	if !strings.Contains(output, "github: {") {
		t.Errorf("output should have github nested object in fontFamily")
	}
}

func TestFormatQuickTailwindBatch_Empty(t *testing.T) {
	var results []*QuickResult

	output := FormatQuickTailwindBatch(results)

	if output != "module.exports = {\n}" {
		t.Errorf("empty results should return valid empty Tailwind config: %s", output)
	}
}

func TestSanitizeCSSName(t *testing.T) {
	tests := []struct {
		domain string
		want   string
	}{
		{"stripe.com", "stripe"},
		{"github.com", "github"},
		{"example.io", "example"},
		{"api.stripe.com", "api-stripe"},
		{"my-app.org", "my-app"},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			got := sanitizeCSSName(tt.domain)
			if got != tt.want {
				t.Errorf("sanitizeCSSName(%q) = %q, want %q", tt.domain, got, tt.want)
			}
		})
	}
}

func TestSanitizeTailwindKey(t *testing.T) {
	tests := []struct {
		domain string
		want   string
	}{
		{"stripe.com", "stripe"},
		{"github.com", "github"},
		{"my-app.io", "my_app"},
		{"api.stripe.com", "api_stripe"},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			got := sanitizeTailwindKey(tt.domain)
			if got != tt.want {
				t.Errorf("sanitizeTailwindKey(%q) = %q, want %q", tt.domain, got, tt.want)
			}
		})
	}
}
