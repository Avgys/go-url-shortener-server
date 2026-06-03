package audit_test

import (
	"testing"

	"github.com/rs/zerolog"

	"go-url-shortener/internal/service/audit"
)

func TestAuditFile_Close_withoutUpdate(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	log := zerolog.Nop()
	file := audit.NewAuditFile(ctx, t.TempDir()+"/audit.csv", &log)

	if err := file.Close(); err != nil {
		t.Fatalf("Close before Update: %v", err)
	}
}
