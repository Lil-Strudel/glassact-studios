package data

import (
	"testing"
)

func newTestArticle(category SupportCategory, title string) *SupportArticle {
	return &SupportArticle{
		Category:    category,
		Title:       title,
		Body:        "Some **markdown** body.",
		SortOrder:   0,
		IsPublished: true,
	}
}

func TestSupportArticle_Insert(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	article := newTestArticle(SupportCategories.Installation, "How to install")

	err := models.SupportArticles.Insert(article)
	if err != nil {
		t.Fatalf("Failed to insert support article: %v", err)
	}

	if article.ID == 0 {
		t.Errorf("Expected non-zero ID, got %d", article.ID)
	}
	if article.UUID == "" {
		t.Errorf("Expected UUID, got empty string")
	}
	if article.Version == 0 {
		t.Errorf("Expected non-zero version")
	}
	if article.CreatedAt.IsZero() {
		t.Errorf("Expected non-zero CreatedAt")
	}
}

func TestSupportArticle_InsertWithYoutubeURL(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	url := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	article := newTestArticle(SupportCategories.Installation, "Video guide")
	article.YoutubeURL = &url

	err := models.SupportArticles.Insert(article)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.SupportArticles.GetByUUID(article.UUID)
	if err != nil {
		t.Fatalf("Failed to get by UUID: %v", err)
	}
	if !found {
		t.Fatalf("Article not found")
	}
	if retrieved.YoutubeURL == nil {
		t.Fatalf("Expected youtube_url to be set")
	}
	if *retrieved.YoutubeURL != url {
		t.Errorf("Expected youtube_url %s, got %s", url, *retrieved.YoutubeURL)
	}
}

func TestSupportArticle_GetByID(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	original := newTestArticle(SupportCategories.Ordering, "Placing orders")

	err := models.SupportArticles.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	retrieved, found, err := models.SupportArticles.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get by ID: %v", err)
	}
	if !found {
		t.Fatalf("Article not found")
	}
	if retrieved.Title != original.Title {
		t.Errorf("Expected title %s, got %s", original.Title, retrieved.Title)
	}
	if retrieved.Category != SupportCategories.Ordering {
		t.Errorf("Expected category ordering, got %s", retrieved.Category)
	}
}

func TestSupportArticle_Update(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	original := newTestArticle(SupportCategories.General, "Original title")

	err := models.SupportArticles.Insert(original)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	original.Title = "Updated title"
	original.Category = SupportCategories.Pricing
	original.IsPublished = false

	err = models.SupportArticles.Update(original)
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	retrieved, found, err := models.SupportArticles.GetByID(original.ID)
	if err != nil {
		t.Fatalf("Failed to get after update: %v", err)
	}
	if !found {
		t.Fatalf("Article not found after update")
	}
	if retrieved.Title != "Updated title" {
		t.Errorf("Expected updated title, got %s", retrieved.Title)
	}
	if retrieved.Category != SupportCategories.Pricing {
		t.Errorf("Expected category pricing, got %s", retrieved.Category)
	}
	if retrieved.IsPublished != false {
		t.Errorf("Expected IsPublished to be false")
	}
}

func TestSupportArticle_VersionIncrement(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	article := newTestArticle(SupportCategories.Contact, "Contact us")

	err := models.SupportArticles.Insert(article)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	initialVersion := article.Version

	article.Title = "Contact us (updated)"
	err = models.SupportArticles.Update(article)
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	if article.Version <= initialVersion {
		t.Errorf("Expected version to increment, was %d, now %d", initialVersion, article.Version)
	}
}

func TestSupportArticle_Delete(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	article := newTestArticle(SupportCategories.General, "Temporary")

	err := models.SupportArticles.Insert(article)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	err = models.SupportArticles.Delete(article.ID)
	if err != nil {
		t.Fatalf("Failed to delete: %v", err)
	}

	_, found, err := models.SupportArticles.GetByID(article.ID)
	if err != nil {
		t.Fatalf("Failed to query after delete: %v", err)
	}
	if found {
		t.Errorf("Expected article to be deleted")
	}
}

func TestSupportArticle_GetAllPublished_ExcludesUnpublished(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	published := newTestArticle(SupportCategories.Installation, "Published")
	if err := models.SupportArticles.Insert(published); err != nil {
		t.Fatalf("Failed to insert published: %v", err)
	}

	hidden := newTestArticle(SupportCategories.Installation, "Hidden")
	hidden.IsPublished = false
	if err := models.SupportArticles.Insert(hidden); err != nil {
		t.Fatalf("Failed to insert hidden: %v", err)
	}

	articles, err := models.SupportArticles.GetAllPublished()
	if err != nil {
		t.Fatalf("Failed to get all published: %v", err)
	}

	for _, a := range articles {
		if a.ID == hidden.ID {
			t.Errorf("Unpublished article should not be returned by GetAllPublished")
		}
	}

	found := false
	for _, a := range articles {
		if a.ID == published.ID {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected published article to be returned")
	}
}

func TestSupportArticle_GetAll_OrdersByCategoryThenSortOrder(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)

	second := newTestArticle(SupportCategories.Installation, "Second")
	second.SortOrder = 20
	if err := models.SupportArticles.Insert(second); err != nil {
		t.Fatalf("Failed to insert second: %v", err)
	}

	first := newTestArticle(SupportCategories.Installation, "First")
	first.SortOrder = 10
	if err := models.SupportArticles.Insert(first); err != nil {
		t.Fatalf("Failed to insert first: %v", err)
	}

	articles, err := models.SupportArticles.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all: %v", err)
	}

	var installIdxFirst, installIdxSecond = -1, -1
	for i, a := range articles {
		if a.ID == first.ID {
			installIdxFirst = i
		}
		if a.ID == second.ID {
			installIdxSecond = i
		}
	}

	if installIdxFirst == -1 || installIdxSecond == -1 {
		t.Fatalf("Expected both articles to be returned")
	}
	if installIdxFirst > installIdxSecond {
		t.Errorf("Expected lower sort_order to come first, got first at %d, second at %d", installIdxFirst, installIdxSecond)
	}
}
