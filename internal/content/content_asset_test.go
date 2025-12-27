package content

import (
	"io"
	"path/filepath"
	"testing"
)

func TestGetTimeStampsMarkdown(t *testing.T) {
	f := testdataPath("sample.md")
	timestamps, err := GetTimeStamps(f)
	if err != nil {
		t.Fatalf("GetTimeStamps failed: %v", err)
	}

	if len(timestamps) != 1 {
		t.Errorf("GetTimeStamps returned %d timestamps, want 1", len(timestamps))
	}

	if timestamps[0].IsZero() {
		t.Error("timestamp is zero")
	}
}

func TestGetTimeStampsTextBundle(t *testing.T) {
	f := testdataPath("sample.textbundle")
	timestamps, err := GetTimeStamps(f)
	if err != nil {
		t.Fatalf("GetTimeStamps failed: %v", err)
	}

	// should have at least: text.md, info.json, and 2 image assets
	if len(timestamps) < 4 {
		t.Errorf("GetTimeStamps returned %d timestamps, want at least 4", len(timestamps))
	}

	for i, ts := range timestamps {
		if ts.IsZero() {
			t.Errorf("Timestamp %d is zero", i)
		}
	}
}

func TestGetTimeStampsTextPack(t *testing.T) {
	f := testdataPath("sample.textpack")
	timestamps, err := GetTimeStamps(f)
	if err != nil {
		t.Fatalf("GetTimeStamps failed: %v", err)
	}

	// should have at least the zip file modification time
	if len(timestamps) < 1 {
		t.Errorf("GetTimeStamps returned %d timestamps, want at least 1", len(timestamps))
	}

	if timestamps[0].IsZero() {
		t.Error("timestamp is zero")
	}
}

func TestGetTimeStampsUnsupported(t *testing.T) {
	if _, err := GetTimeStamps("file.txt"); err == nil {
		t.Error("GetTimeStamps should return error for unsupported file")
	}
}

func TestListAssetsMarkdown(t *testing.T) {
	f := testdataPath("sample.md")
	assets, err := ListAssets(f)
	if err != nil {
		t.Errorf("ListAssets should not return error for markdown files, got: %v", err)
	}

	if assets != nil {
		t.Errorf("ListAssets returned %v, want nil", assets)
	}
}

func TestListAssetsHTML(t *testing.T) {
	f := testdataPath("sample.html")
	assets, err := ListAssets(f)
	if err != nil {
		t.Errorf("ListAssets should not return error for HTML files, got: %v", err)
	}

	if assets != nil {
		t.Errorf("ListAssets returned %v, want nil", assets)
	}
}

func TestListAssetsTextBundle(t *testing.T) {
	f := testdataPath("sample.textbundle")
	assets, err := ListAssets(f)
	if err != nil {
		t.Fatalf("ListAssets failed: %v", err)
	}

	if len(assets) != 2 {
		t.Errorf("ListAssets returned %d assets, want 2", len(assets))
	}

	hasImage := false
	hasNested := false
	for _, asset := range assets {
		if asset == "image.png" {
			hasImage = true
		}
		if asset == filepath.Join("nested", "nested.png") {
			hasNested = true
		}
	}

	if !hasImage {
		t.Error("ListAssets missing image.png")
	}
	if !hasNested {
		t.Error("ListAssets missing nested/nested.png")
	}
}

func TestListAssetsTextPack(t *testing.T) {
	f := testdataPath("sample.textpack")
	assets, err := ListAssets(f)
	if err != nil {
		t.Fatalf("ListAssets failed: %v", err)
	}

	if len(assets) != 2 {
		t.Errorf("ListAssets returned %d assets, want 2", len(assets))
	}

	hasImage := false
	hasNested := false
	for _, asset := range assets {
		if asset == "image.png" {
			hasImage = true
		}
		if asset == filepath.Join("nested", "nested.png") {
			hasNested = true
		}
	}

	if !hasImage {
		t.Error("ListAssets missing image.png")
	}
	if !hasNested {
		t.Error("ListAssets missing nested/nested.png")
	}
}

func TestListAssetsUnsupported(t *testing.T) {
	if _, err := ListAssets("file.txt"); err == nil {
		t.Error("ListAssets should return error for unsupported file")
	}
}

func TestOpenAssetTextBundle(t *testing.T) {
	f := testdataPath("sample.textbundle")
	path, reader, err := OpenAsset(f, "image.png")
	if err != nil {
		t.Fatalf("OpenAsset failed: %v", err)
	}
	defer reader.Close()

	if path != filepath.Join("assets", "image.png") {
		t.Errorf("asset path=%q, want %q", path, filepath.Join("assets", "image.png"))
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read asset: %v", err)
	}

	if len(data) == 0 {
		t.Error("asset content is empty")
	}

	if len(data) < 4 || data[0] != 0x89 || data[1] != 'P' || data[2] != 'N' || data[3] != 'G' {
		t.Error("asset does not appear to be a valid PNG")
	}
}

func TestOpenAssetTextBundleNested(t *testing.T) {
	f := testdataPath("sample.textbundle")
	nestedPath := filepath.Join("nested", "nested.png")
	path, reader, err := OpenAsset(f, nestedPath)
	if err != nil {
		t.Fatalf("OpenAsset failed: %v", err)
	}
	defer reader.Close()

	if path != filepath.Join("assets", nestedPath) {
		t.Errorf("asset path=%q, want %q", path, filepath.Join("assets", nestedPath))
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read asset: %v", err)
	}

	if len(data) == 0 {
		t.Error("asset content is empty")
	}

	if len(data) < 4 || data[0] != 0x89 || data[1] != 'P' || data[2] != 'N' || data[3] != 'G' {
		t.Error("asset does not appear to be a valid PNG")
	}
}

func TestOpenAssetTextBundleMissing(t *testing.T) {
	f := testdataPath("sample.textbundle")
	if _, _, err := OpenAsset(f, "nonexistent.png"); err == nil {
		t.Error("OpenAsset should return error for missing asset")
	}
}

func TestOpenAssetTextPack(t *testing.T) {
	f := testdataPath("sample.textpack")
	path, reader, err := OpenAsset(f, "image.png")
	if err != nil {
		t.Fatalf("OpenAsset failed: %v", err)
	}
	defer reader.Close()

	if path != filepath.Join("assets", "image.png") {
		t.Errorf("asset path=%q, want %q", path, filepath.Join("assets", "image.png"))
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read asset: %v", err)
	}

	if len(data) == 0 {
		t.Error("asset content is empty")
	}

	if len(data) < 4 || data[0] != 0x89 || data[1] != 'P' || data[2] != 'N' || data[3] != 'G' {
		t.Error("asset does not appear to be a valid PNG")
	}
}

func TestOpenAssetTextPackNested(t *testing.T) {
	f := testdataPath("sample.textpack")
	nestedPath := filepath.Join("nested", "nested.png")
	path, reader, err := OpenAsset(f, nestedPath)
	if err != nil {
		t.Fatalf("OpenAsset failed: %v", err)
	}
	defer reader.Close()

	if path != filepath.Join("assets", nestedPath) {
		t.Errorf("asset path=%q, want %q", path, filepath.Join("assets", nestedPath))
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read asset: %v", err)
	}

	if len(data) == 0 {
		t.Error("asset content is empty")
	}

	if len(data) < 4 || data[0] != 0x89 || data[1] != 'P' || data[2] != 'N' || data[3] != 'G' {
		t.Error("asset does not appear to be a valid PNG")
	}
}

func TestOpenAssetTextPackMissing(t *testing.T) {
	f := testdataPath("sample.textpack")
	if _, _, err := OpenAsset(f, "nonexistent.png"); err == nil {
		t.Error("OpenAsset should return error for missing asset")
	}
}

func TestOpenAssetMarkdown(t *testing.T) {
	f := testdataPath("sample.md")
	if _, _, err := OpenAsset(f, "image.png"); err == nil {
		t.Error("OpenAsset should return error for markdown files")
	}
}

func TestOpenAssetHTML(t *testing.T) {
	f := testdataPath("sample.html")
	if _, _, err := OpenAsset(f, "image.png"); err == nil {
		t.Error("OpenAsset should return error for HTML files")
	}
}

func TestOpenAssetUnsupported(t *testing.T) {
	if _, _, err := OpenAsset("file.txt", "image.png"); err == nil {
		t.Error("OpenAsset should return error for unsupported file")
	}
}

func TestGetMetadata(t *testing.T) {
	tests := []struct {
		name          string
		file          string
		expectedTitle string
	}{
		{"Markdown", testdataPath("sample.md"), "Test Markdown"},
		{"HTML", testdataPath("sample.html"), "Test HTML"},
		{"TextBundle", testdataPath("sample.textbundle"), "Main Heading"},
		{"TextPack", testdataPath("sample.textpack"), "Main Heading"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta, err := GetMetadata(tt.file)
			if err != nil {
				t.Fatalf("GetMetadata failed: %v", err)
			}

			if meta == nil {
				t.Fatal("GetMetadata returned nil metadata")
			}

			if meta.Title != tt.expectedTitle {
				t.Errorf("title=%q, want %q", meta.Title, tt.expectedTitle)
			}
		})
	}
}

func TestGetMetadataUnsupported(t *testing.T) {
	if _, err := GetMetadata("file.txt"); err == nil {
		t.Error("GetMetadata should return error for unsupported file")
	}
}
