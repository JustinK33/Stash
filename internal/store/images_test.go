package store

import (
	"bytes"
	"os"
	"testing"
)

func TestImportImageSavesAndLoads(t *testing.T) {
	stashStore := newTestStore(t)
	file := File{Settings: DefaultSettings()}

	savedImage, err := stashStore.ImportImage(&file, bytes.NewReader(testPNG(t)), "example.png")
	if err != nil {
		t.Fatal(err)
	}

	if savedImage.Name != "example.png" {
		t.Errorf("got name %q, want %q", savedImage.Name, "example.png")
	}
	if savedImage.MediaType != "image/png" {
		t.Errorf("got media type %q, want %q", savedImage.MediaType, "image/png")
	}
	if len(file.Images) != 1 {
		t.Fatalf("got %d images, want 1: %#v", len(file.Images), file.Images)
	}
	if _, err := os.Stat(stashStore.ImagePath(savedImage)); err != nil {
		t.Errorf("managed file missing at %s: %v", stashStore.ImagePath(savedImage), err)
	}

	loaded, err := stashStore.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Images) != 1 || loaded.Images[0].ID != savedImage.ID {
		t.Errorf("got %#v, want the imported image to survive Load", loaded.Images)
	}
}

func TestImportImageDeduplicatesByContent(t *testing.T) {
	stashStore := newTestStore(t)
	file := File{Settings: DefaultSettings()}
	imageData := testPNG(t)

	first, err := stashStore.ImportImage(&file, bytes.NewReader(imageData), "first.png")
	if err != nil {
		t.Fatal(err)
	}
	second, err := stashStore.ImportImage(&file, bytes.NewReader(imageData), "second.png")
	if err != nil {
		t.Fatal(err)
	}

	if first.ID != second.ID {
		t.Errorf("got IDs %q and %q, want identical content to share one ID", first.ID, second.ID)
	}
	if len(file.Images) != 1 {
		t.Errorf("got %d images, want 1: %#v", len(file.Images), file.Images)
	}
}

func TestImportImageRejects(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{name: "not an image", data: []byte("not an image")},
		{name: "oversized", data: make([]byte, MaxImageBytes+1)},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			stashStore := newTestStore(t)
			file := File{Settings: DefaultSettings()}

			if _, err := stashStore.ImportImage(&file, bytes.NewReader(testCase.data), "input.png"); err == nil {
				t.Error("got nil error, want a rejection")
			}
			if len(file.Images) != 0 {
				t.Errorf("got %d images, want the rejected import to leave none", len(file.Images))
			}
		})
	}
}

func TestDeleteImageRemovesMetadataAndFile(t *testing.T) {
	stashStore := newTestStore(t)
	file := File{Settings: DefaultSettings()}
	savedImage, err := stashStore.ImportImage(&file, bytes.NewReader(testPNG(t)), "example.png")
	if err != nil {
		t.Fatal(err)
	}

	if err := stashStore.DeleteImage(&file, savedImage.ID); err != nil {
		t.Fatal(err)
	}

	if len(file.Images) != 0 {
		t.Errorf("got %#v, want no image metadata", file.Images)
	}
	if _, err := os.Stat(stashStore.ImagePath(savedImage)); !os.IsNotExist(err) {
		t.Errorf("got %v, want the managed file to be gone", err)
	}
}

func TestDeleteImageUnknownIDIsNoOp(t *testing.T) {
	stashStore := newTestStore(t)
	file := File{Settings: DefaultSettings()}
	if _, err := stashStore.ImportImage(&file, bytes.NewReader(testPNG(t)), "example.png"); err != nil {
		t.Fatal(err)
	}

	if err := stashStore.DeleteImage(&file, "absent"); err != nil {
		t.Fatal(err)
	}

	if len(file.Images) != 1 {
		t.Errorf("got %d images, want the existing image untouched", len(file.Images))
	}
}

func TestClearImagesKeepsSnippets(t *testing.T) {
	stashStore := newTestStore(t)
	file := File{
		Snippets: Add(nil, "keep me"),
		Settings: DefaultSettings(),
	}
	savedImage, err := stashStore.ImportImage(&file, bytes.NewReader(testPNG(t)), "example.png")
	if err != nil {
		t.Fatal(err)
	}

	if err := stashStore.ClearImages(&file); err != nil {
		t.Fatal(err)
	}

	if len(file.Images) != 0 {
		t.Errorf("got %#v, want no images", file.Images)
	}
	if _, err := os.Stat(stashStore.ImagePath(savedImage)); !os.IsNotExist(err) {
		t.Errorf("got %v, want the managed file to be gone", err)
	}
	if len(file.Snippets) != 1 || file.Snippets[0].Text != "keep me" {
		t.Errorf("got %#v, want the text snippet to remain", file.Snippets)
	}
}
