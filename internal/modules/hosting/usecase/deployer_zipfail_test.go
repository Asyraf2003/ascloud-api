package usecase

import (
	"context"
	"testing"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
	"example.com/your-api/internal/shared/apperr"
)

func TestDeployer_ZipSlipIsPermanentFail(t *testing.T) {
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
		"../evil.txt": []byte("x"),
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
	ae, ok := apperr.As(err)
	if !ok || ae.Code != "hosting.zip_slip" {
		t.Fatalf("want hosting.zip_slip, got %v", err)
	}
	if up.lastStatus != domain.UploadStatusFailed {
		t.Fatalf("upload status = %s", up.lastStatus)
	}
	if rel.lastStatus != domain.ReleaseStatusFailed || rel.lastCode != "hosting.zip_slip" {
		t.Fatalf("release status/code = %s/%s", rel.lastStatus, rel.lastCode)
	}
	if len(obj.putKeys) != 0 {
		t.Fatalf("unexpected put keys: %v", obj.putKeys)
	}
	if obj.deleteKey != "" {
		t.Fatalf("should not delete on zip slip")
	}
	_ = sites
}
