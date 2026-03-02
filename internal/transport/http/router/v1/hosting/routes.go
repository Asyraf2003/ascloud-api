package hosting

import (
	"github.com/labstack/echo/v4"

	hostingHTTP "example.com/your-api/internal/modules/hosting/transport/http"
)

func Register(g *echo.Group) {
	// Sites
	g.GET("/hosting/sites", hostingHTTP.ListSites)
	g.POST("/hosting/sites", hostingHTTP.CreateSite)
	g.GET("/hosting/sites/:site_id", hostingHTTP.GetSite)

	// Upload flow (existing)
	g.POST("/hosting/sites/:site_id/uploads", hostingHTTP.InitiateUpload)
	g.POST("/hosting/sites/:site_id/uploads/:upload_id/complete", hostingHTTP.CompleteUpload)

	// Releases
	g.GET("/hosting/sites/:site_id/releases", hostingHTTP.ListReleases)
	g.GET("/hosting/sites/:site_id/releases/:release_id", hostingHTTP.GetRelease)

	// Rollback (activate/pointer update)
	g.POST("/hosting/sites/:site_id/rollback", hostingHTTP.RollbackSite)
}
