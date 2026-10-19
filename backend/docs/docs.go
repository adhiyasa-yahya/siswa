// Package docs embeds the hand written OpenAPI 3.0 specification and serves it
// together with a Swagger UI page.
//
// The spec is the source of truth (docs/openapi.yaml) - no code generation step
// is needed, so the documentation cannot drift out of the repository.
package docs

import (
	"embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed openapi.yaml
var FS embed.FS

// SpecPath is the route serving the raw OpenAPI document.
const SpecPath = "/swagger/openapi.yaml"

const swaggerUIVersion = "5.17.14"

const indexHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>SISWA API - Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@` + swaggerUIVersion + `/swagger-ui.css">
  <style>
    body { margin: 0; background: #fafafa; }
    .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@` + swaggerUIVersion + `/swagger-ui-bundle.js" crossorigin></script>
  <script src="https://unpkg.com/swagger-ui-dist@` + swaggerUIVersion + `/swagger-ui-standalone-preset.js" crossorigin></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: "` + SpecPath + `",
        dom_id: "#swagger-ui",
        deepLinking: true,
        displayRequestDuration: true,
        defaultModelsExpandDepth: 1,
        docExpansion: "list",
        tryItOutEnabled: true,
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
        layout: "StandaloneLayout"
      });
    };
  </script>
</body>
</html>`

// Register mounts the Swagger UI and the OpenAPI document on the router.
func Register(r gin.IRouter) {
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	r.GET("/swagger/index.html", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(indexHTML))
	})

	r.GET(SpecPath, func(c *gin.Context) {
		spec, err := FS.ReadFile("openapi.yaml")
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", spec)
	})
}
