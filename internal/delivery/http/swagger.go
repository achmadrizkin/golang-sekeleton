package http

import (
	"net/http"
)

const swaggerIndexHTML = `<!DOCTYPE html>
<html>
<head>
  <title>golang-sekeleton API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        urls: [
          { url: "/swagger/user.swagger.json", name: "User" },
        ],
        dom_id: "#swagger-ui",
      });
    };
  </script>
</body>
</html>`

// SwaggerHandler serves the Swagger UI shell plus the generated
// *.swagger.json files from the proto/ directory. Mount it only when
// cfg.Server.SwaggerEnabled is true — it should stay off in production.
func SwaggerHandler(protoDir string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(swaggerIndexHTML))
	})
	mux.Handle("/swagger/user.swagger.json", http.StripPrefix("/swagger/", http.FileServer(http.Dir(protoDir))))
	mux.Handle("/swagger/common.swagger.json", http.StripPrefix("/swagger/", http.FileServer(http.Dir(protoDir))))
	return mux
}
