package background

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/upload"
	"github.com/Lil-Strudel/glassact-studios/apps/api/svg"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

// relatedItemsFromIDs keeps the admin's order. A newly added item must be one
// a dealership could actually add; one retired since it was linked may stay,
// so retiring a catalog item never blocks saving the backgrounds it is on.
func (m BackgroundModule) relatedItemsFromIDs(ids []int, current []data.RelatedCatalogItem) ([]data.RelatedCatalogItem, error) {
	alreadyRelated := make(map[int]bool, len(current))
	for _, item := range current {
		alreadyRelated[item.ID] = true
	}

	seen := make(map[int]bool, len(ids))
	items := make([]data.RelatedCatalogItem, len(ids))
	for i, id := range ids {
		if seen[id] {
			return nil, fmt.Errorf("catalog item %d is listed more than once", id)
		}
		seen[id] = true

		item, found, err := m.Db.CatalogItems.GetByID(id)
		if err != nil {
			return nil, err
		}
		if !found {
			return nil, fmt.Errorf("catalog item %d not found", id)
		}
		if !item.IsActive && !alreadyRelated[id] {
			return nil, fmt.Errorf("%s is disabled in the catalog", item.CatalogCode)
		}
		items[i] = data.RelatedCatalogItem{ID: item.ID}
	}
	return items, nil
}

func (m BackgroundModule) ensureCodeIsFree(code string, ownID int) error {
	backgrounds, err := m.Db.SandblastBackgrounds.GetAll()
	if err != nil {
		return err
	}
	for _, b := range backgrounds {
		if b.ID != ownID && strings.EqualFold(b.Code, code) {
			return fmt.Errorf("another background already uses code %s", b.Code)
		}
	}
	return nil
}

func normalizeTags(tags []string) []string {
	seen := map[string]bool{}
	normalized := []string{}
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" || seen[tag] {
			continue
		}
		seen[tag] = true
		normalized = append(normalized, tag)
	}
	return normalized
}

// storeMask rebuilds the mask from the raw upload, the same way the preview
// was prepared, trims it to the measured content and stores it under the
// background's code. Without S3 (tests) the upload's URL is kept as-is.
func (m BackgroundModule) storeMask(ctx context.Context, code, rawURL string, bbox svg.ContentBBox) (string, error) {
	if m.S3 == nil {
		return rawURL, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	raw, err := upload.GetFileFromS3(ctx, m.S3, m.Cfg, strings.TrimPrefix(rawURL, "/"))
	if err != nil {
		return "", fmt.Errorf("failed to fetch uploaded svg: %w", err)
	}

	mask, err := svg.FitMask(raw, bbox)
	if err != nil {
		return "", fmt.Errorf("failed to prepare sandblast mask: %w", err)
	}

	result, err := upload.UploadFileToS3(
		ctx, m.S3, m.Cfg,
		bytes.NewReader(mask),
		upload.SanitizeFilenamePart(code, "background")+".svg",
		int64(len(mask)),
		"image/svg+xml",
		"sandblast-backgrounds",
	)
	if err != nil {
		return "", fmt.Errorf("failed to store sandblast mask: %w", err)
	}
	return result.URL, nil
}
