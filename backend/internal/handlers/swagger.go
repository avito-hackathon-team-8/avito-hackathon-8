package handlers

import (
	"net/http"
	"os"
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
    window.ui = SwaggerUIBundle({
      url: '/api/openapi.yaml',
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
      layout: 'StandaloneLayout'
    });
  </script>
</body>
</html>`

func swaggerUI(response http.ResponseWriter, _ *http.Request) {
	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = response.Write([]byte(swaggerHTML))
}

func openAPISpec(response http.ResponseWriter, _ *http.Request) {
	spec, err := os.ReadFile("docs/openapi.yaml")
	if err != nil {
		http.Error(response, "OpenAPI specification is unavailable", http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	_, _ = response.Write(spec)
}
