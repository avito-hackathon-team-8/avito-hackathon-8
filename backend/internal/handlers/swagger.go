package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

const swaggerHTML = `<!doctype html>
<html lang="ru">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Коробыш API — Swagger</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
  <script>
    const openAPISpec = __OPENAPI_SPEC__;

    window.ui = SwaggerUIBundle({
      url: 'data:text/yaml;charset=utf-8,' + encodeURIComponent(openAPISpec),
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
      layout: 'StandaloneLayout'
    });
  </script>
</body>
</html>`

func swaggerUI(response http.ResponseWriter, _ *http.Request) {
	spec, err := os.ReadFile("docs/openapi.yaml")
	if err != nil {
		http.Error(response, "OpenAPI specification is unavailable", http.StatusInternalServerError)

		return
	}

	specJSON, err := json.Marshal(string(spec))
	if err != nil {
		http.Error(response, "OpenAPI specification is unavailable", http.StatusInternalServerError)

		return
	}

	html := strings.Replace(swaggerHTML, "__OPENAPI_SPEC__", string(specJSON), 1)

	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = response.Write([]byte(html))
}
