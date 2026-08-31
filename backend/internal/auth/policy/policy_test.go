package policy

import "testing"

func TestCanDeleteFileRequiresUploader(t *testing.T) {
	owner := &Principal{UserID: "owner", Role: RoleUser}
	if err := CanDeleteFile(owner, "owner"); err != nil {
		t.Fatalf("owner should be allowed to delete: %v", err)
	}
	if err := CanDeleteFile(owner, "other"); err != ErrForbidden {
		t.Fatalf("other uploader should be forbidden, got %v", err)
	}
}

func TestRequireAdmin(t *testing.T) {
	if err := RequireAdmin(&Principal{UserID: "user", Role: RoleUser}); err != ErrForbidden {
		t.Fatalf("regular user should not be admin, got %v", err)
	}
	if err := RequireAdmin(&Principal{UserID: "admin", Role: RoleAdmin}); err != nil {
		t.Fatalf("admin should be allowed: %v", err)
	}
}
