package config

import (
	"strings"
	"testing"
)

func TestGetMigrationList(t *testing.T) {
	migrations := getMigrationList()

	if len(migrations) == 0 {
		t.Fatal("Daftar migrasi tidak boleh kosong")
	}

	seenVersions := make(map[string]bool)

	for i, m := range migrations {
		if strings.TrimSpace(m.version) == "" {
			t.Errorf("Migrasi index %d memiliki versi kosong", i)
		}

		if seenVersions[m.version] {
			t.Errorf("Versi migrasi duplikat ditemukan: %s", m.version)
		}
		seenVersions[m.version] = true

		if strings.TrimSpace(m.description) == "" {
			t.Errorf("Migrasi %s tidak memiliki deskripsi", m.version)
		}

		if len(m.queries) == 0 {
			t.Errorf("Migrasi %s tidak memiliki daftar query", m.version)
		}

		for qIdx, q := range m.queries {
			trimmed := strings.TrimSpace(q)
			if trimmed == "" {
				t.Errorf("Migrasi %s memiliki query kosong pada query index %d", m.version, qIdx)
			}
			// Pastikan tidak ada perintah DROP DATABASE atau DROP TABLE sembarangan
			upper := strings.ToUpper(trimmed)
			if strings.Contains(upper, "DROP DATABASE") || strings.Contains(upper, "DROP TABLE") {
				t.Errorf("Migrasi %s mengandung instruksi destruktif (DROP): %s", m.version, trimmed)
			}
		}
	}
}
