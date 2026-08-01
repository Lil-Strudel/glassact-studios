package modules

import (
	"fmt"
	"net/http"
	"testing"

	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// patchDealershipBody is the full PATCH payload. Name and address are required
// by the handler, so every case has to send them even when it is only exercising
// one of the newer fields.
func patchDealershipBody(dealership *data.Dealership) map[string]any {
	return map[string]any{
		"name":                  dealership.Name,
		"phone":                 dealership.Phone,
		"sandblast_file_format": string(dealership.SandblastFileFormat),
		"address": map[string]any{
			"street":      dealership.Address.Street,
			"street_ext":  dealership.Address.StreetExt,
			"city":        dealership.Address.City,
			"state":       dealership.Address.State,
			"postal_code": dealership.Address.PostalCode,
			"country":     dealership.Address.Country,
			"latitude":    dealership.Address.Latitude,
			"longitude":   dealership.Address.Longitude,
		},
	}
}

func TestPatchDealership_DealershipUserCannotChangePaymentTiming(t *testing.T) {
	testCtx, cleanup := setupTestApp(t)
	defer cleanup()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, testCtx)

	dealership, found, err := testCtx.db.Dealerships.GetByID(dealershipUser.DealershipID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, data.PaymentTimings.PostShipping, dealership.PaymentTiming)

	body := patchDealershipBody(dealership)
	body["payment_timing"] = string(data.PaymentTimings.PreManufacturing)

	res := testCtx.request(testRequest{
		method: http.MethodPatch,
		path:   fmt.Sprintf("/api/dealership/%s", dealership.UUID),
		body:   body,
		token:  dealershipToken,
	})
	require.Equal(t, http.StatusOK, res.statusCode, string(res.body))

	reloaded, _, err := testCtx.db.Dealerships.GetByID(dealership.ID)
	require.NoError(t, err)
	assert.Equal(t, data.PaymentTimings.PostShipping, reloaded.PaymentTiming)
}

func TestPatchDealership_InternalUserChangesPaymentTiming(t *testing.T) {
	testCtx, cleanup := setupTestApp(t)
	defer cleanup()

	dealershipUser, _, _, internalToken := seedTestData(t, testCtx)

	dealership, found, err := testCtx.db.Dealerships.GetByID(dealershipUser.DealershipID)
	require.NoError(t, err)
	require.True(t, found)

	body := patchDealershipBody(dealership)
	body["payment_timing"] = string(data.PaymentTimings.PreManufacturing)

	res := testCtx.request(testRequest{
		method: http.MethodPatch,
		path:   fmt.Sprintf("/api/dealership/%s", dealership.UUID),
		body:   body,
		token:  internalToken,
	})
	require.Equal(t, http.StatusOK, res.statusCode, string(res.body))

	reloaded, _, err := testCtx.db.Dealerships.GetByID(dealership.ID)
	require.NoError(t, err)
	assert.Equal(t, data.PaymentTimings.PreManufacturing, reloaded.PaymentTiming)
}

func TestPatchDealership_DealershipUserChangesPhoneAndSandblastFormat(t *testing.T) {
	testCtx, cleanup := setupTestApp(t)
	defer cleanup()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, testCtx)

	dealership, found, err := testCtx.db.Dealerships.GetByID(dealershipUser.DealershipID)
	require.NoError(t, err)
	require.True(t, found)

	body := patchDealershipBody(dealership)
	body["phone"] = "5559876543"
	body["sandblast_file_format"] = string(data.SandblastFileFormats.DXF)

	res := testCtx.request(testRequest{
		method: http.MethodPatch,
		path:   fmt.Sprintf("/api/dealership/%s", dealership.UUID),
		body:   body,
		token:  dealershipToken,
	})
	require.Equal(t, http.StatusOK, res.statusCode, string(res.body))

	reloaded, _, err := testCtx.db.Dealerships.GetByID(dealership.ID)
	require.NoError(t, err)
	assert.Equal(t, "5559876543", reloaded.Phone)
	assert.Equal(t, data.SandblastFileFormats.DXF, reloaded.SandblastFileFormat)
}

func TestPatchDealership_RejectsBadPhoneAndFormat(t *testing.T) {
	testCtx, cleanup := setupTestApp(t)
	defer cleanup()

	dealershipUser, dealershipToken, _, _ := seedTestData(t, testCtx)

	dealership, found, err := testCtx.db.Dealerships.GetByID(dealershipUser.DealershipID)
	require.NoError(t, err)
	require.True(t, found)

	tests := []struct {
		name  string
		field string
		value any
	}{
		{"phone with too few digits", "phone", "555123"},
		{"phone with punctuation", "phone", "(555) 123-4567"},
		{"unsupported sandblast format", "sandblast_file_format", "jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := patchDealershipBody(dealership)
			body[tt.field] = tt.value

			res := testCtx.request(testRequest{
				method: http.MethodPatch,
				path:   fmt.Sprintf("/api/dealership/%s", dealership.UUID),
				body:   body,
				token:  dealershipToken,
			})
			assert.Equal(t, http.StatusBadRequest, res.statusCode, string(res.body))
		})
	}
}
