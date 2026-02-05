package data

import (
	"testing"
	"time"
)

func TestCitext(t *testing.T) {
	val := "TestValue"
	result := Citext(val)

	if result == nil {
		t.Errorf("Expected Citext to return non-nil expression")
	}
}

func TestSTPoint(t *testing.T) {
	longitude := -74.0060
	latitude := 40.7128

	result := STPoint(longitude, latitude)

	if result == nil {
		t.Errorf("Expected STPoint to return non-nil expression")
	}
}

func TestSTLongitude(t *testing.T) {
	location := STPoint(-74.0060, 40.7128)
	result := STLongitude(location)

	if result == nil {
		t.Errorf("Expected STLongitude to return non-nil expression")
	}
}

func TestSTLatitude(t *testing.T) {
	location := STPoint(-74.0060, 40.7128)
	result := STLatitude(location)

	if result == nil {
		t.Errorf("Expected STLatitude to return non-nil expression")
	}
}

func TestNow(t *testing.T) {
	result := Now()

	if result == nil {
		t.Errorf("Expected Now to return non-nil expression")
	}
}

func TestStandardTableFields(t *testing.T) {
	now := time.Now()
	st := StandardTable{
		ID:        1,
		UUID:      "550e8400-e29b-41d4-a716-446655440000",
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}

	if st.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", st.ID)
	}

	if st.UUID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("Expected UUID to match")
	}

	if st.CreatedAt != now {
		t.Errorf("Expected CreatedAt to match")
	}

	if st.UpdatedAt != now {
		t.Errorf("Expected UpdatedAt to match")
	}

	if st.Version != 1 {
		t.Errorf("Expected Version to be 1, got %d", st.Version)
	}
}
