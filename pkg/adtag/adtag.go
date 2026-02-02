// Package adtag provides client-side ad tag generation for direct publisher integration
package adtag

import (
	"fmt"
	"html/template"
	"net/url"
	"strings"
)

// AdTagConfig represents configuration for generating ad tags
type AdTagConfig struct {
	// Server configuration
	ServerURL string // Base URL of the ad server (e.g., https://ads.thenexusengine.com)

	// Ad unit configuration
	PublisherID string // Publisher account ID
	PlacementID string // Ad placement/slot ID
	Width       int    // Ad width in pixels
	Height      int    // Ad height in pixels

	// Optional configuration
	PageURL     string            // Page URL for targeting
	Domain      string            // Site domain
	Keywords    []string          // Targeting keywords
	CustomData  map[string]string // Custom key-value pairs
	RefreshRate int               // Auto-refresh interval in seconds (0 = no refresh)
}

// AdTagFormat represents the format of the ad tag
type AdTagFormat string

const (
	// FormatAsync generates asynchronous JavaScript tag
	FormatAsync AdTagFormat = "async"
	// FormatSync generates synchronous JavaScript tag
	FormatSync AdTagFormat = "sync"
	// FormatIframe generates iframe tag
	FormatIframe AdTagFormat = "iframe"
	// FormatGAM generates GAM 3rd party creative script
	FormatGAM AdTagFormat = "gam"
)

// AdTag represents a generated ad tag
type AdTag struct {
	HTML       string // Complete HTML tag code
	JavaScript string // Standalone JavaScript code
	IframeURL  string // Direct iframe URL
	GAMScript  string // GAM 3rd party script
}

// Generator generates ad tags for publisher integration
type Generator struct {
	serverURL string
}

// NewGenerator creates a new ad tag generator
func NewGenerator(serverURL string) *Generator {
	return &Generator{
		serverURL: strings.TrimRight(serverURL, "/"),
	}
}

// Generate generates an ad tag in the specified format
func (g *Generator) Generate(config *AdTagConfig, format AdTagFormat) (*AdTag, error) {
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	tag := &AdTag{}

	switch format {
	case FormatAsync:
		tag.HTML = g.generateAsyncTag(config)
		tag.JavaScript = g.generateAsyncScript(config)
	case FormatSync:
		tag.HTML = g.generateSyncTag(config)
		tag.JavaScript = g.generateSyncScript(config)
	case FormatIframe:
		tag.IframeURL = g.generateIframeURL(config)
		tag.HTML = g.generateIframeTag(config)
	case FormatGAM:
		tag.GAMScript = g.generateGAMScript(config)
		tag.HTML = tag.GAMScript
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	return tag, nil
}

// generateAsyncTag generates asynchronous JavaScript ad tag
func (g *Generator) generateAsyncTag(config *AdTagConfig) string {
	divID := fmt.Sprintf("tne-ad-%s", config.PlacementID)

	tmpl := `<!-- TNE Catalyst Ad Tag - {{ .Width }}x{{ .Height }} -->
<div id="{{ .DivID }}" style="width:{{ .Width }}px;height:{{ .Height }}px;"></div>
<script>
(function() {
  var tne = tne || {};
  tne.cmd = tne.cmd || [];
  tne.cmd.push(function() {
    tne.display({
      publisherId: '{{ .PublisherID }}',
      placementId: '{{ .PlacementID }}',
      divId: '{{ .DivID }}',
      size: [{{ .Width }}, {{ .Height }}],
      serverUrl: '{{ .ServerURL }}'{{ if .PageURL }},
      pageUrl: '{{ .PageURL }}'{{ end }}{{ if .Domain }},
      domain: '{{ .Domain }}'{{ end }}{{ if .Keywords }},
      keywords: [{{ .KeywordsJS }}]{{ end }}{{ if .CustomData }},
      customData: {{ .CustomDataJS }}{{ end }}{{ if .RefreshRate }},
      refreshRate: {{ .RefreshRate }}{{ end }}
    });
  });
})();
</script>
<script async src="{{ .ServerURL }}/assets/tne-ads.js"></script>`

	data := map[string]interface{}{
		"DivID":        divID,
		"PublisherID":  config.PublisherID,
		"PlacementID":  config.PlacementID,
		"Width":        config.Width,
		"Height":       config.Height,
		"ServerURL":    g.serverURL,
		"PageURL":      config.PageURL,
		"Domain":       config.Domain,
		"Keywords":     config.Keywords,
		"KeywordsJS":   formatKeywordsJS(config.Keywords),
		"CustomData":   config.CustomData,
		"CustomDataJS": formatCustomDataJS(config.CustomData),
		"RefreshRate":  config.RefreshRate,
	}

	return renderTemplate(tmpl, data)
}

// generateAsyncScript generates standalone async JavaScript
func (g *Generator) generateAsyncScript(config *AdTagConfig) string {
	divID := fmt.Sprintf("tne-ad-%s", config.PlacementID)

	tmpl := `tne.display({
  publisherId: '{{ .PublisherID }}',
  placementId: '{{ .PlacementID }}',
  divId: '{{ .DivID }}',
  size: [{{ .Width }}, {{ .Height }}],
  serverUrl: '{{ .ServerURL }}'{{ if .PageURL }},
  pageUrl: '{{ .PageURL }}'{{ end }}{{ if .Domain }},
  domain: '{{ .Domain }}'{{ end }}{{ if .Keywords }},
  keywords: [{{ .KeywordsJS }}]{{ end }}{{ if .CustomData }},
  customData: {{ .CustomDataJS }}{{ end }}{{ if .RefreshRate }},
  refreshRate: {{ .RefreshRate }}{{ end }}
});`

	data := map[string]interface{}{
		"DivID":        divID,
		"PublisherID":  config.PublisherID,
		"PlacementID":  config.PlacementID,
		"Width":        config.Width,
		"Height":       config.Height,
		"ServerURL":    g.serverURL,
		"PageURL":      config.PageURL,
		"Domain":       config.Domain,
		"Keywords":     config.Keywords,
		"KeywordsJS":   formatKeywordsJS(config.Keywords),
		"CustomData":   config.CustomData,
		"CustomDataJS": formatCustomDataJS(config.CustomData),
		"RefreshRate":  config.RefreshRate,
	}

	return renderTemplate(tmpl, data)
}

// generateSyncTag generates synchronous JavaScript ad tag
func (g *Generator) generateSyncTag(config *AdTagConfig) string {
	divID := fmt.Sprintf("tne-ad-%s", config.PlacementID)

	tmpl := `<!-- TNE Catalyst Ad Tag - {{ .Width }}x{{ .Height }} -->
<div id="{{ .DivID }}" style="width:{{ .Width }}px;height:{{ .Height }}px;"></div>
<script>
document.write('<scr' + 'ipt src="{{ .AdURL }}"></scr' + 'ipt>');
</script>`

	adURL := g.buildAdURL(config, divID)

	data := map[string]interface{}{
		"DivID":  divID,
		"Width":  config.Width,
		"Height": config.Height,
		"AdURL":  adURL,
	}

	return renderTemplate(tmpl, data)
}

// generateSyncScript generates standalone sync JavaScript
func (g *Generator) generateSyncScript(config *AdTagConfig) string {
	divID := fmt.Sprintf("tne-ad-%s", config.PlacementID)
	adURL := g.buildAdURL(config, divID)
	return fmt.Sprintf("document.write('<scr' + 'ipt src=\"%s\"></scr' + 'ipt>');", adURL)
}

// generateIframeTag generates iframe ad tag
func (g *Generator) generateIframeTag(config *AdTagConfig) string {
	iframeURL := g.generateIframeURL(config)

	tmpl := `<!-- TNE Catalyst Ad Tag - {{ .Width }}x{{ .Height }} -->
<iframe src="{{ .URL }}"
        width="{{ .Width }}"
        height="{{ .Height }}"
        frameborder="0"
        scrolling="no"
        marginheight="0"
        marginwidth="0"
        style="border:0;vertical-align:bottom;"
        sandbox="allow-forms allow-pointer-lock allow-popups allow-popups-to-escape-sandbox allow-same-origin allow-scripts allow-top-navigation-by-user-activation"
        loading="lazy">
</iframe>`

	data := map[string]interface{}{
		"URL":    iframeURL,
		"Width":  config.Width,
		"Height": config.Height,
	}

	return renderTemplate(tmpl, data)
}

// generateIframeURL generates iframe URL
func (g *Generator) generateIframeURL(config *AdTagConfig) string {
	params := url.Values{}
	params.Set("pub", config.PublisherID)
	params.Set("placement", config.PlacementID)
	params.Set("w", fmt.Sprintf("%d", config.Width))
	params.Set("h", fmt.Sprintf("%d", config.Height))

	if config.PageURL != "" {
		params.Set("url", config.PageURL)
	}
	if config.Domain != "" {
		params.Set("domain", config.Domain)
	}
	if len(config.Keywords) > 0 {
		params.Set("kw", strings.Join(config.Keywords, ","))
	}
	for k, v := range config.CustomData {
		params.Set(k, v)
	}

	return fmt.Sprintf("%s/ad/iframe?%s", g.serverURL, params.Encode())
}

// generateGAMScript generates GAM 3rd party creative script
func (g *Generator) generateGAMScript(config *AdTagConfig) string {
	// GAM 3rd party script format
	tmpl := `<script>
(function() {
  // TNE Catalyst - GAM Integration
  var tneConfig = {
    publisherId: '{{ .PublisherID }}',
    placementId: '{{ .PlacementID }}',
    width: {{ .Width }},
    height: {{ .Height }},
    serverUrl: '{{ .ServerURL }}'{{ if .PageURL }},
    pageUrl: '{{ .PageURL }}'{{ end }}{{ if .Domain }},
    domain: '{{ .Domain }}'{{ end }}{{ if .Keywords }},
    keywords: [{{ .KeywordsJS }}]{{ end }}
  };

  // Create container
  var container = document.createElement('div');
  container.id = 'tne-gam-' + tneConfig.placementId;
  container.style.width = tneConfig.width + 'px';
  container.style.height = tneConfig.height + 'px';
  document.write(container.outerHTML);

  // Load ad
  var script = document.createElement('script');
  script.src = tneConfig.serverUrl + '/ad/gam?' +
    'pub=' + encodeURIComponent(tneConfig.publisherId) +
    '&placement=' + encodeURIComponent(tneConfig.placementId) +
    '&w=' + tneConfig.width +
    '&h=' + tneConfig.height +
    '&div=' + encodeURIComponent(container.id){{ if .PageURL }} +
    '&url=' + encodeURIComponent(tneConfig.pageUrl){{ end }}{{ if .Domain }} +
    '&domain=' + encodeURIComponent(tneConfig.domain){{ end }}{{ if .Keywords }} +
    '&kw=' + encodeURIComponent([{{ .KeywordsJS }}].join(',')){{ end }};
  script.async = true;
  document.body.appendChild(script);
})();
</script>`

	data := map[string]interface{}{
		"PublisherID": config.PublisherID,
		"PlacementID": config.PlacementID,
		"Width":       config.Width,
		"Height":      config.Height,
		"ServerURL":   g.serverURL,
		"PageURL":     config.PageURL,
		"Domain":      config.Domain,
		"Keywords":    config.Keywords,
		"KeywordsJS":  formatKeywordsJS(config.Keywords),
	}

	return renderTemplate(tmpl, data)
}

// buildAdURL builds the ad request URL
func (g *Generator) buildAdURL(config *AdTagConfig, divID string) string {
	params := url.Values{}
	params.Set("pub", config.PublisherID)
	params.Set("placement", config.PlacementID)
	params.Set("div", divID)
	params.Set("w", fmt.Sprintf("%d", config.Width))
	params.Set("h", fmt.Sprintf("%d", config.Height))

	if config.PageURL != "" {
		params.Set("url", config.PageURL)
	}
	if config.Domain != "" {
		params.Set("domain", config.Domain)
	}
	if len(config.Keywords) > 0 {
		params.Set("kw", strings.Join(config.Keywords, ","))
	}
	for k, v := range config.CustomData {
		params.Set(k, v)
	}

	return fmt.Sprintf("%s/ad/js?%s", g.serverURL, params.Encode())
}

// validateConfig validates ad tag configuration
func validateConfig(config *AdTagConfig) error {
	if config.PublisherID == "" {
		return fmt.Errorf("publisher ID is required")
	}
	if config.PlacementID == "" {
		return fmt.Errorf("placement ID is required")
	}
	if config.Width <= 0 {
		return fmt.Errorf("width must be positive")
	}
	if config.Height <= 0 {
		return fmt.Errorf("height must be positive")
	}
	return nil
}

// formatKeywordsJS formats keywords for JavaScript output
func formatKeywordsJS(keywords []string) string {
	if len(keywords) == 0 {
		return ""
	}
	quoted := make([]string, len(keywords))
	for i, kw := range keywords {
		quoted[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(kw, "'", "\\'"))
	}
	return strings.Join(quoted, ", ")
}

// formatCustomDataJS formats custom data for JavaScript output
func formatCustomDataJS(data map[string]string) string {
	if len(data) == 0 {
		return "{}"
	}
	parts := make([]string, 0, len(data))
	for k, v := range data {
		parts = append(parts, fmt.Sprintf("'%s': '%s'",
			strings.ReplaceAll(k, "'", "\\'"),
			strings.ReplaceAll(v, "'", "\\'")))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// renderTemplate renders a template string with data
func renderTemplate(tmplStr string, data interface{}) string {
	tmpl, err := template.New("adtag").Parse(tmplStr)
	if err != nil {
		return ""
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return ""
	}

	return buf.String()
}

// GenerateAllFormats generates ad tags in all formats
func (g *Generator) GenerateAllFormats(config *AdTagConfig) (map[AdTagFormat]*AdTag, error) {
	formats := []AdTagFormat{FormatAsync, FormatSync, FormatIframe, FormatGAM}
	result := make(map[AdTagFormat]*AdTag)

	for _, format := range formats {
		tag, err := g.Generate(config, format)
		if err != nil {
			return nil, fmt.Errorf("failed to generate %s format: %w", format, err)
		}
		result[format] = tag
	}

	return result, nil
}
