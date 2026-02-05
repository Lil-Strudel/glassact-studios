package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInlayProof_Struct(t *testing.T) {
	t.Run("struct fields exist", func(t *testing.T) {
		proof := &InlayProof{
			StandardTable: StandardTable{
				ID:      1,
				UUID:    "test-uuid",
				Version: 1,
			},
			InlayID: 123,
		}

		assert.Equal(t, 1, proof.ID)
		assert.Equal(t, "test-uuid", proof.UUID)
		assert.Equal(t, 123, proof.InlayID)
		assert.Equal(t, 1, proof.Version)
	})
}

func TestInlayProofModel_Struct(t *testing.T) {
	t.Run("model has database connections", func(t *testing.T) {
		models := getTestModels(t)

		assert.NotNil(t, models.InlayProofs.DB)
		assert.NotNil(t, models.InlayProofs.STDB)
	})
}
