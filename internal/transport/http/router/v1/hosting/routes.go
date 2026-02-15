package hosting

import (
	"github.com/labstack/echo/v4"

	hostingHTTP "example.com/your-api/internal/modules/hosting/transport/http"
)

func Register(g *echo.Group) {
	// Initiate: create upload + return presigned PUT URL
	g.POST("/hosting/sites/:site_id/uploads", hostingHTTP.InitiateUpload)

	// Complete: verify uploaded object (Head), enforce zip size limit, enqueue deploy
	g.POST("/hosting/sites/:site_id/uploads/:upload_id/complete", hostingHTTP.CompleteUpload)
}
