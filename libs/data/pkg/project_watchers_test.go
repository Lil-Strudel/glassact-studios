package data

import (
	"testing"
)

func TestProjectWatcher_AutoSubscribeIsIdempotent(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	user := createTestDealershipUser(t, models, dealership.ID)

	for i := 0; i < 3; i++ {
		if err := models.ProjectWatchers.AutoSubscribe(project.ID, user); err != nil {
			t.Fatalf("AutoSubscribe call %d failed: %v", i+1, err)
		}
	}

	isWatching, err := models.ProjectWatchers.IsWatching(project.ID, user)
	if err != nil {
		t.Fatalf("IsWatching failed: %v", err)
	}
	if !isWatching {
		t.Errorf("Expected user to be watching after auto-subscribe")
	}

	count, err := models.ProjectWatchers.CountForProject(project.ID)
	if err != nil {
		t.Fatalf("CountForProject failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 watcher after repeated auto-subscribe, got %d", count)
	}
}

// An explicit unwatch is a decision the user made. Later activity on the project
// must not quietly resubscribe them.
func TestProjectWatcher_AutoSubscribeDoesNotReviveExplicitUnwatch(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	user := createTestDealershipUser(t, models, dealership.ID)

	if err := models.ProjectWatchers.SetWatching(project.ID, user, false); err != nil {
		t.Fatalf("SetWatching(false) failed: %v", err)
	}

	if err := models.ProjectWatchers.AutoSubscribe(project.ID, user); err != nil {
		t.Fatalf("AutoSubscribe failed: %v", err)
	}

	isWatching, err := models.ProjectWatchers.IsWatching(project.ID, user)
	if err != nil {
		t.Fatalf("IsWatching failed: %v", err)
	}
	if isWatching {
		t.Errorf("Expected an explicit unwatch to survive auto-subscribe")
	}
}

func TestProjectWatcher_SetWatchingRoundTrips(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	user := createTestDealershipUser(t, models, dealership.ID)

	// No row at all means not watching.
	isWatching, err := models.ProjectWatchers.IsWatching(project.ID, user)
	if err != nil {
		t.Fatalf("IsWatching failed: %v", err)
	}
	if isWatching {
		t.Errorf("Expected a user with no watcher row to not be watching")
	}

	for _, want := range []bool{true, false, true} {
		if err := models.ProjectWatchers.SetWatching(project.ID, user, want); err != nil {
			t.Fatalf("SetWatching(%v) failed: %v", want, err)
		}

		got, err := models.ProjectWatchers.IsWatching(project.ID, user)
		if err != nil {
			t.Fatalf("IsWatching failed: %v", err)
		}
		if got != want {
			t.Errorf("Expected is_watching %v, got %v", want, got)
		}
	}
}

func TestProjectWatcher_GetWatchersExcludesInactiveAndUnwatched(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	watching := createTestDealershipUser(t, models, dealership.ID)
	unwatched := createTestDealershipUser(t, models, dealership.ID)
	inactive := createTestDealershipUser(t, models, dealership.ID)
	createTestDealershipUser(t, models, dealership.ID) // never subscribed at all

	if err := models.ProjectWatchers.SetWatching(project.ID, watching, true); err != nil {
		t.Fatalf("SetWatching failed: %v", err)
	}
	if err := models.ProjectWatchers.SetWatching(project.ID, unwatched, false); err != nil {
		t.Fatalf("SetWatching failed: %v", err)
	}
	if err := models.ProjectWatchers.SetWatching(project.ID, inactive, true); err != nil {
		t.Fatalf("SetWatching failed: %v", err)
	}

	inactive.IsActive = false
	if err := models.DealershipUsers.Update(inactive); err != nil {
		t.Fatalf("Failed to deactivate user: %v", err)
	}

	watchers, err := models.ProjectWatchers.GetDealershipWatchers(project.ID)
	if err != nil {
		t.Fatalf("GetDealershipWatchers failed: %v", err)
	}

	if len(watchers) != 1 {
		t.Fatalf("Expected exactly 1 dealership watcher, got %d", len(watchers))
	}
	if watchers[0].ID != watching.ID {
		t.Errorf("Expected watcher %d, got %d", watching.ID, watchers[0].ID)
	}
}

// The two nullable user columns are mutually exclusive, so each side's query
// must never pick up the other's rows.
func TestProjectWatcher_SidesAreIsolated(t *testing.T) {
	t.Cleanup(func() { cleanupTables(t) })

	models := getTestModels(t)
	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)

	dealershipUser := createTestDealershipUser(t, models, dealership.ID)
	internalUser := createTestInternalUser(t, models)

	if err := models.ProjectWatchers.AutoSubscribe(project.ID, dealershipUser); err != nil {
		t.Fatalf("AutoSubscribe(dealership) failed: %v", err)
	}
	if err := models.ProjectWatchers.AutoSubscribe(project.ID, internalUser); err != nil {
		t.Fatalf("AutoSubscribe(internal) failed: %v", err)
	}

	dealershipWatchers, err := models.ProjectWatchers.GetDealershipWatchers(project.ID)
	if err != nil {
		t.Fatalf("GetDealershipWatchers failed: %v", err)
	}
	if len(dealershipWatchers) != 1 || dealershipWatchers[0].ID != dealershipUser.ID {
		t.Errorf("Expected only the dealership user, got %d watchers", len(dealershipWatchers))
	}

	internalWatchers, err := models.ProjectWatchers.GetInternalWatchers(project.ID)
	if err != nil {
		t.Fatalf("GetInternalWatchers failed: %v", err)
	}
	if len(internalWatchers) != 1 || internalWatchers[0].ID != internalUser.ID {
		t.Errorf("Expected only the internal user, got %d watchers", len(internalWatchers))
	}

	count, err := models.ProjectWatchers.CountForProject(project.ID)
	if err != nil {
		t.Fatalf("CountForProject failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 watchers across both sides, got %d", count)
	}
}
