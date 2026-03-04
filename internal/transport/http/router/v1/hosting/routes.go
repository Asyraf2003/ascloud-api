package hosting

import (
	"time"

	"github.com/labstack/echo/v4"

	hostingHTTP "example.com/your-api/internal/modules/hosting/transport/http"
	"example.com/your-api/internal/transport/http/middleware/trust"
)

func Register(g *echo.Group) {
	// Balanced limits (per acc, fallback IP):
	// - read:    120/min
	// - create:   30/min
	// - initiate: 30/min
	// - complete: 20/min
	// - rollback: 10/min

	rlRead := trust.RateLimit("hosting_read", 120, time.Minute)
	rlCreate := trust.RateLimit("hosting_site_create", 30, time.Minute)
	rlInitiate := trust.RateLimit("hosting_upload_initiate", 30, time.Minute)
	rlComplete := trust.RateLimit("hosting_upload_complete", 20, time.Minute)
	rlRollback := trust.RateLimit("hosting_rollback", 10, time.Minute)

	// Sites
	g.GET("/hosting/sites", hostingHTTP.ListSites, rlRead)
	g.POST("/hosting/sites", hostingHTTP.CreateSite, rlCreate)
	g.GET("/hosting/sites/:site_id", hostingHTTP.GetSite, rlRead)

	// Upload flow
	g.POST("/hosting/sites/:site_id/uploads", hostingHTTP.InitiateUpload, rlInitiate)
	g.POST("/hosting/sites/:site_id/uploads/:upload_id/complete", hostingHTTP.CompleteUpload, rlComplete)

	// Releases
	g.GET("/hosting/sites/:site_id/releases", hostingHTTP.ListReleases, rlRead)
	g.GET("/hosting/sites/:site_id/releases/:release_id", hostingHTTP.GetRelease, rlRead)

	// Rollback (activate/pointer update)
	g.POST("/hosting/sites/:site_id/rollback", hostingHTTP.RollbackSite, rlRollback)
}
