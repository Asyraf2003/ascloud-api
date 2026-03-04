package usecase

import (
	"context"
	"testing"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
)

func TestDeployer_SuccessPublishesAndUpdatesPointers(t *testing.T) {
	up := &fakeUploadStore{u: domain.Upload{
		ID:        "u1",
		SiteID:    "s1",
		ReleaseID: "r1",
		ObjectKey: "sites/s1/uploads/u1.zip",
		SizeBytes: 10,
		Status:    domain.UploadStatusQueued,
	}}
	sites := &fakeSiteStore{}
	rel := &fakeReleaseStore{}
	obj := &fakeObjectStore{zipBytes: makeZipBytes(t, map[string][]byte{
		"index.html":     []byte("<h1>x</h1>"),
		"assets/app.css": []byte("body{}"),
	})}

	cfg := DefaultDeployConfig()
	cfg.TmpDir = tempTmpDir(t)

	d := NewDeployer(cfg, up, sites, rel, obj, nil)
	err := d.Deploy(context.Background(), ports.DeployMessage{
		SiteID:    "s1",
		UploadID:  "u1",
		ReleaseID: "r1",
		ObjectKey: "sites/s1/uploads/u1.zip",
		SizeBytes: 10,
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if up.lastStatus != domain.UploadStatusDeployed {
		t.Fatalf("upload status = %s", up.lastStatus)
	}
	if sites.lastRelease != "r1" {
		t.Fatalf("site current release = %s", sites.lastRelease)
	}
	if rel.lastStatus != domain.ReleaseStatusSuccess {
		t.Fatalf("release status = %s", rel.lastStatus)
	}
	if len(obj.putKeys) != 2 {
		t.Fatalf("put keys = %v", obj.putKeys)
	}
	if obj.deleteKey == "" {
		t.Fatalf("expected delete to be called")
	}
}
