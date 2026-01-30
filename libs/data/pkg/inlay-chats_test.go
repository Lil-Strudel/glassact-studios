package data

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestInlayChat(t *testing.T, models Models, inlayID int, userID int) *InlayChat {
	t.Helper()

	chat := &InlayChat{
		InlayID:    inlayID,
		UserID:     userID,
		SenderType: SenderTypes.Customer,
		Message:    "Test chat message",
	}

	err := models.InlayChats.Insert(chat)
	if err != nil {
		t.Fatalf("Failed to create test inlay chat: %v", err)
	}

	return chat
}

func TestInlayChatModel_Insert(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	user := createTestUser(t, models, dealership.ID)
	inlay := createTestCustomInlay(t, models, project.ID)

	t.Run("successful insert customer message", func(t *testing.T) {
		chat := &InlayChat{
			InlayID:    inlay.ID,
			UserID:     user.ID,
			SenderType: SenderTypes.Customer,
			Message:    "Hello, I have a question about my inlay.",
		}

		err := models.InlayChats.Insert(chat)
		require.NoError(t, err)

		assert.NotZero(t, chat.ID)
		assert.NotEmpty(t, chat.UUID)
		assert.NotZero(t, chat.CreatedAt)
		assert.NotZero(t, chat.UpdatedAt)
		assert.Equal(t, 1, chat.Version)

		_, err = uuid.Parse(chat.UUID)
		assert.NoError(t, err)
	})

	t.Run("successful insert glassact message", func(t *testing.T) {
		chat := &InlayChat{
			InlayID:    inlay.ID,
			UserID:     user.ID,
			SenderType: SenderTypes.GlassAct,
			Message:    "Thank you for reaching out. Let me help you.",
		}

		err := models.InlayChats.Insert(chat)
		require.NoError(t, err)

		assert.Equal(t, SenderTypes.GlassAct, chat.SenderType)
	})

	t.Run("insert with all sender types", func(t *testing.T) {
		senderTypes := []SenderType{SenderTypes.GlassAct, SenderTypes.Customer}

		for _, senderType := range senderTypes {
			chat := &InlayChat{
				InlayID:    inlay.ID,
				UserID:     user.ID,
				SenderType: senderType,
				Message:    fmt.Sprintf("Message from %s", senderType),
			}

			err := models.InlayChats.Insert(chat)
			require.NoError(t, err, "failed to insert chat with sender type %s", senderType)
			assert.Equal(t, senderType, chat.SenderType)
		}
	})

	t.Run("insert with invalid inlay fails", func(t *testing.T) {
		chat := &InlayChat{
			InlayID:    99999,
			UserID:     user.ID,
			SenderType: SenderTypes.Customer,
			Message:    "This should fail",
		}

		err := models.InlayChats.Insert(chat)
		assert.Error(t, err)
	})

	t.Run("insert with invalid user fails", func(t *testing.T) {
		chat := &InlayChat{
			InlayID:    inlay.ID,
			UserID:     99999,
			SenderType: SenderTypes.Customer,
			Message:    "This should fail",
		}

		err := models.InlayChats.Insert(chat)
		assert.Error(t, err)
	})

	t.Run("insert with invalid sender type fails", func(t *testing.T) {
		chat := &InlayChat{
			InlayID:    inlay.ID,
			UserID:     user.ID,
			SenderType: SenderType("invalid-type"),
			Message:    "This should fail",
		}

		err := models.InlayChats.Insert(chat)
		assert.Error(t, err)
	})

	t.Run("insert with empty message succeeds", func(t *testing.T) {
		chat := &InlayChat{
			InlayID:    inlay.ID,
			UserID:     user.ID,
			SenderType: SenderTypes.Customer,
			Message:    "",
		}

		err := models.InlayChats.Insert(chat)
		require.NoError(t, err)
	})

	t.Run("insert with long message succeeds", func(t *testing.T) {
		longMessage := ""
		for i := 0; i < 10000; i++ {
			longMessage += "a"
		}

		chat := &InlayChat{
			InlayID:    inlay.ID,
			UserID:     user.ID,
			SenderType: SenderTypes.Customer,
			Message:    longMessage,
		}

		err := models.InlayChats.Insert(chat)
		require.NoError(t, err)
		assert.Equal(t, longMessage, chat.Message)
	})

	t.Run("insert multiple messages for same inlay", func(t *testing.T) {
		chats := make([]*InlayChat, 10)
		for i := 0; i < 10; i++ {
			chat := &InlayChat{
				InlayID:    inlay.ID,
				UserID:     user.ID,
				SenderType: SenderTypes.Customer,
				Message:    fmt.Sprintf("Message %d", i),
			}
			err := models.InlayChats.Insert(chat)
			require.NoError(t, err)
			chats[i] = chat
		}

		ids := make(map[int]bool)
		for _, chat := range chats {
			assert.False(t, ids[chat.ID], "duplicate ID found")
			ids[chat.ID] = true
		}
	})

	t.Run("insert with unicode content", func(t *testing.T) {
		chat := &InlayChat{
			InlayID:    inlay.ID,
			UserID:     user.ID,
			SenderType: SenderTypes.Customer,
			Message:    "Hello! I'd like to know about the price. This is in Japanese: \u65e5\u672c\u8a9e. And emojis: \U0001F600\U0001F60D",
		}

		err := models.InlayChats.Insert(chat)
		require.NoError(t, err)

		retrieved, found, err := models.InlayChats.GetByID(chat.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, chat.Message, retrieved.Message)
	})
}

func TestInlayChatModel_GetByID(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	user := createTestUser(t, models, dealership.ID)
	inlay := createTestCustomInlay(t, models, project.ID)
	chat := createTestInlayChat(t, models, inlay.ID, user.ID)

	t.Run("existing chat", func(t *testing.T) {
		retrieved, found, err := models.InlayChats.GetByID(chat.ID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, chat.ID, retrieved.ID)
		assert.Equal(t, chat.UUID, retrieved.UUID)
		assert.Equal(t, chat.InlayID, retrieved.InlayID)
		assert.Equal(t, chat.UserID, retrieved.UserID)
		assert.Equal(t, chat.SenderType, retrieved.SenderType)
		assert.Equal(t, chat.Message, retrieved.Message)
	})

	t.Run("non-existing chat", func(t *testing.T) {
		retrieved, found, err := models.InlayChats.GetByID(99999)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("negative ID", func(t *testing.T) {
		retrieved, found, err := models.InlayChats.GetByID(-1)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("zero ID", func(t *testing.T) {
		retrieved, found, err := models.InlayChats.GetByID(0)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})
}

func TestInlayChatModel_GetByUUID(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	user := createTestUser(t, models, dealership.ID)
	inlay := createTestCustomInlay(t, models, project.ID)
	chat := createTestInlayChat(t, models, inlay.ID, user.ID)

	t.Run("existing chat", func(t *testing.T) {
		retrieved, found, err := models.InlayChats.GetByUUID(chat.UUID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, chat.ID, retrieved.ID)
		assert.Equal(t, chat.UUID, retrieved.UUID)
	})

	t.Run("non-existing UUID", func(t *testing.T) {
		nonExistentUUID := uuid.New().String()
		retrieved, found, err := models.InlayChats.GetByUUID(nonExistentUUID)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("invalid UUID format", func(t *testing.T) {
		_, _, err := models.InlayChats.GetByUUID("not-a-valid-uuid")
		assert.Error(t, err)
	})

	t.Run("empty UUID", func(t *testing.T) {
		_, _, err := models.InlayChats.GetByUUID("")
		assert.Error(t, err)
	})
}

func TestInlayChatModel_GetAll(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	t.Run("empty table", func(t *testing.T) {
		chats, err := models.InlayChats.GetAll()
		require.NoError(t, err)
		assert.Empty(t, chats)
	})

	t.Run("multiple chats", func(t *testing.T) {
		dealership := createTestDealership(t, models)
		project := createTestProject(t, models, dealership.ID)
		user := createTestUser(t, models, dealership.ID)
		inlay := createTestCustomInlay(t, models, project.ID)

		for i := 0; i < 5; i++ {
			createTestInlayChat(t, models, inlay.ID, user.ID)
		}

		chats, err := models.InlayChats.GetAll()
		require.NoError(t, err)
		assert.Len(t, chats, 5)

		for _, chat := range chats {
			assert.NotZero(t, chat.ID)
			assert.NotEmpty(t, chat.UUID)
			assert.NotEmpty(t, chat.Message)
		}
	})

	t.Run("chats from multiple inlays", func(t *testing.T) {
		cleanupTables(t)

		dealership := createTestDealership(t, models)
		project := createTestProject(t, models, dealership.ID)
		user := createTestUser(t, models, dealership.ID)
		inlay1 := createTestCustomInlay(t, models, project.ID)
		inlay2 := createTestCustomInlay(t, models, project.ID)

		for i := 0; i < 3; i++ {
			createTestInlayChat(t, models, inlay1.ID, user.ID)
		}
		for i := 0; i < 2; i++ {
			createTestInlayChat(t, models, inlay2.ID, user.ID)
		}

		chats, err := models.InlayChats.GetAll()
		require.NoError(t, err)
		assert.Len(t, chats, 5)
	})

	t.Run("mixed sender types", func(t *testing.T) {
		cleanupTables(t)

		dealership := createTestDealership(t, models)
		project := createTestProject(t, models, dealership.ID)
		user := createTestUser(t, models, dealership.ID)
		inlay := createTestCustomInlay(t, models, project.ID)

		chatCustomer := &InlayChat{
			InlayID:    inlay.ID,
			UserID:     user.ID,
			SenderType: SenderTypes.Customer,
			Message:    "Customer message",
		}
		err := models.InlayChats.Insert(chatCustomer)
		require.NoError(t, err)

		chatGlassAct := &InlayChat{
			InlayID:    inlay.ID,
			UserID:     user.ID,
			SenderType: SenderTypes.GlassAct,
			Message:    "GlassAct response",
		}
		err = models.InlayChats.Insert(chatGlassAct)
		require.NoError(t, err)

		chats, err := models.InlayChats.GetAll()
		require.NoError(t, err)
		assert.Len(t, chats, 2)

		senderTypeCount := map[SenderType]int{}
		for _, chat := range chats {
			senderTypeCount[chat.SenderType]++
		}
		assert.Equal(t, 1, senderTypeCount[SenderTypes.Customer])
		assert.Equal(t, 1, senderTypeCount[SenderTypes.GlassAct])
	})
}

func TestInlayChatModel_GetAllByInlayID(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	user := createTestUser(t, models, dealership.ID)
	inlay1 := createTestCustomInlay(t, models, project.ID)
	inlay2 := createTestCustomInlay(t, models, project.ID)

	for i := 0; i < 3; i++ {
		createTestInlayChat(t, models, inlay1.ID, user.ID)
	}

	for i := 0; i < 2; i++ {
		createTestInlayChat(t, models, inlay2.ID, user.ID)
	}

	t.Run("get chats for specific inlay", func(t *testing.T) {
		chats, err := models.InlayChats.GetAllByInlayID(inlay1.ID)
		require.NoError(t, err)
		assert.Len(t, chats, 3)

		for _, chat := range chats {
			assert.Equal(t, inlay1.ID, chat.InlayID)
		}
	})

	t.Run("get chats for different inlay", func(t *testing.T) {
		chats, err := models.InlayChats.GetAllByInlayID(inlay2.ID)
		require.NoError(t, err)
		assert.Len(t, chats, 2)

		for _, chat := range chats {
			assert.Equal(t, inlay2.ID, chat.InlayID)
		}
	})

	t.Run("non-existent inlay returns empty", func(t *testing.T) {
		chats, err := models.InlayChats.GetAllByInlayID(99999)
		require.NoError(t, err)
		assert.Empty(t, chats)
	})

	t.Run("negative inlay ID returns empty", func(t *testing.T) {
		chats, err := models.InlayChats.GetAllByInlayID(-1)
		require.NoError(t, err)
		assert.Empty(t, chats)
	})

	t.Run("zero inlay ID returns empty", func(t *testing.T) {
		chats, err := models.InlayChats.GetAllByInlayID(0)
		require.NoError(t, err)
		assert.Empty(t, chats)
	})
}

func TestInlayChatModel_Update(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	user := createTestUser(t, models, dealership.ID)
	inlay := createTestCustomInlay(t, models, project.ID)

	t.Run("successful update", func(t *testing.T) {
		chat := createTestInlayChat(t, models, inlay.ID, user.ID)
		originalVersion := chat.Version
		originalUpdatedAt := chat.UpdatedAt

		time.Sleep(10 * time.Millisecond)

		chat.Message = "Updated message content"
		chat.SenderType = SenderTypes.GlassAct

		err := models.InlayChats.Update(chat)
		require.NoError(t, err)

		assert.Equal(t, originalVersion+1, chat.Version)
		assert.True(t, chat.UpdatedAt.After(originalUpdatedAt) || chat.UpdatedAt.Equal(originalUpdatedAt))

		retrieved, found, err := models.InlayChats.GetByID(chat.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, "Updated message content", retrieved.Message)
		assert.Equal(t, SenderTypes.GlassAct, retrieved.SenderType)
	})

	t.Run("update with wrong version fails", func(t *testing.T) {
		chat := createTestInlayChat(t, models, inlay.ID, user.ID)

		chat.Version = 999

		chat.Message = "Should Fail"
		err := models.InlayChats.Update(chat)
		assert.Error(t, err)
	})

	t.Run("update preserves ID and UUID", func(t *testing.T) {
		chat := createTestInlayChat(t, models, inlay.ID, user.ID)
		originalID := chat.ID
		originalUUID := chat.UUID

		chat.Message = "New Message"
		err := models.InlayChats.Update(chat)
		require.NoError(t, err)

		assert.Equal(t, originalID, chat.ID)
		assert.Equal(t, originalUUID, chat.UUID)
	})

	t.Run("concurrent update detection", func(t *testing.T) {
		chat := createTestInlayChat(t, models, inlay.ID, user.ID)

		chat1, found, err := models.InlayChats.GetByID(chat.ID)
		require.NoError(t, err)
		require.True(t, found)

		chat2, found, err := models.InlayChats.GetByID(chat.ID)
		require.NoError(t, err)
		require.True(t, found)

		chat1.Message = "First Update"
		err = models.InlayChats.Update(chat1)
		require.NoError(t, err)

		chat2.Message = "Second Update"
		err = models.InlayChats.Update(chat2)
		assert.Error(t, err, "should fail due to optimistic locking")
	})

	t.Run("update sender type", func(t *testing.T) {
		chat := &InlayChat{
			InlayID:    inlay.ID,
			UserID:     user.ID,
			SenderType: SenderTypes.Customer,
			Message:    "Original message",
		}
		err := models.InlayChats.Insert(chat)
		require.NoError(t, err)

		chat.SenderType = SenderTypes.GlassAct
		err = models.InlayChats.Update(chat)
		require.NoError(t, err)

		retrieved, found, err := models.InlayChats.GetByID(chat.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, SenderTypes.GlassAct, retrieved.SenderType)
	})

	t.Run("update to invalid sender type fails", func(t *testing.T) {
		chat := createTestInlayChat(t, models, inlay.ID, user.ID)

		chat.SenderType = SenderType("invalid-type")
		err := models.InlayChats.Update(chat)
		assert.Error(t, err)
	})
}

func TestInlayChatModel_Delete(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	user := createTestUser(t, models, dealership.ID)
	inlay := createTestCustomInlay(t, models, project.ID)

	t.Run("successful delete", func(t *testing.T) {
		chat := createTestInlayChat(t, models, inlay.ID, user.ID)

		err := models.InlayChats.Delete(chat.ID)
		require.NoError(t, err)

		retrieved, found, err := models.InlayChats.GetByID(chat.ID)
		require.NoError(t, err)
		assert.False(t, found)
		assert.Nil(t, retrieved)
	})

	t.Run("delete non-existent chat", func(t *testing.T) {
		err := models.InlayChats.Delete(99999)
		require.NoError(t, err)
	})

	t.Run("delete one chat leaves others", func(t *testing.T) {
		chat1 := createTestInlayChat(t, models, inlay.ID, user.ID)
		chat2 := createTestInlayChat(t, models, inlay.ID, user.ID)

		err := models.InlayChats.Delete(chat1.ID)
		require.NoError(t, err)

		retrieved, found, err := models.InlayChats.GetByID(chat2.ID)
		require.NoError(t, err)
		assert.True(t, found)
		assert.NotNil(t, retrieved)
	})

	t.Run("delete with negative ID", func(t *testing.T) {
		err := models.InlayChats.Delete(-1)
		require.NoError(t, err)
	})

	t.Run("delete with zero ID", func(t *testing.T) {
		err := models.InlayChats.Delete(0)
		require.NoError(t, err)
	})
}

func TestSenderType_Constants(t *testing.T) {
	t.Run("sender type values", func(t *testing.T) {
		assert.Equal(t, SenderType("glassact"), SenderTypes.GlassAct)
		assert.Equal(t, SenderType("customer"), SenderTypes.Customer)
	})

	t.Run("sender type string conversion", func(t *testing.T) {
		assert.Equal(t, "glassact", string(SenderTypes.GlassAct))
		assert.Equal(t, "customer", string(SenderTypes.Customer))
	})

	t.Run("sender types are distinct", func(t *testing.T) {
		assert.NotEqual(t, SenderTypes.GlassAct, SenderTypes.Customer)
	})
}

func TestInlayChat_StandardTable(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	user := createTestUser(t, models, dealership.ID)
	inlay := createTestCustomInlay(t, models, project.ID)
	chat := createTestInlayChat(t, models, inlay.ID, user.ID)

	t.Run("standard fields populated", func(t *testing.T) {
		assert.NotZero(t, chat.StandardTable.ID)
		assert.NotEmpty(t, chat.StandardTable.UUID)
		assert.NotZero(t, chat.StandardTable.CreatedAt)
		assert.NotZero(t, chat.StandardTable.UpdatedAt)
		assert.Equal(t, 1, chat.StandardTable.Version)
	})

	t.Run("created_at immutable", func(t *testing.T) {
		originalCreatedAt := chat.CreatedAt

		chat.Message = "New Message"
		err := models.InlayChats.Update(chat)
		require.NoError(t, err)

		retrieved, found, err := models.InlayChats.GetByID(chat.ID)
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, originalCreatedAt.Unix(), retrieved.CreatedAt.Unix())
	})
}

func TestInlayChat_Cascade(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	project := createTestProject(t, models, dealership.ID)
	user := createTestUser(t, models, dealership.ID)
	inlay := createTestCustomInlay(t, models, project.ID)

	t.Run("delete inlay cascades to chats", func(t *testing.T) {
		chat1 := createTestInlayChat(t, models, inlay.ID, user.ID)
		chat2 := createTestInlayChat(t, models, inlay.ID, user.ID)

		err := models.Inlays.Delete(inlay.ID)
		require.NoError(t, err)

		_, found, err := models.InlayChats.GetByID(chat1.ID)
		require.NoError(t, err)
		assert.False(t, found)

		_, found, err = models.InlayChats.GetByID(chat2.ID)
		require.NoError(t, err)
		assert.False(t, found)
	})
}

func BenchmarkInlayChatModel_Insert(b *testing.B) {
	models := NewModels(testDB.Pool, testDB.STDB)

	dealership := &Dealership{
		Name: "Benchmark Dealership",
		Address: Address{
			Street:     "123 Main St",
			StreetExt:  "",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "USA",
			Latitude:   40.7128,
			Longitude:  -74.0060,
		},
	}
	_ = models.Dealerships.Insert(dealership)

	project := &Project{
		Name:         "Benchmark Project",
		Status:       ProjectStatusi.AwaitingProof,
		Approved:     false,
		DealershipID: dealership.ID,
	}
	_ = models.Projects.Insert(project)

	user := &User{
		Name:         "Benchmark User",
		Email:        fmt.Sprintf("benchmarkchat%d@example.com", time.Now().UnixNano()),
		Avatar:       "https://example.com/avatar.png",
		DealershipID: dealership.ID,
		Role:         UserRoles.User,
	}
	_ = models.Users.Insert(user)

	inlay := &Inlay{
		ProjectID:  project.ID,
		Name:       "Benchmark Inlay",
		PreviewURL: "https://example.com/preview.png",
		PriceGroup: 1,
		Type:       InlayTypes.Custom,
		CustomInfo: &InlayCustomInfo{
			Description: "Benchmark inlay",
			Width:       24.5,
			Height:      36.75,
		},
	}
	_ = models.Inlays.Insert(inlay)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chat := &InlayChat{
			InlayID:    inlay.ID,
			UserID:     user.ID,
			SenderType: SenderTypes.Customer,
			Message:    fmt.Sprintf("Benchmark message %d", i),
		}
		_ = models.InlayChats.Insert(chat)
	}
}

func BenchmarkInlayChatModel_GetByID(b *testing.B) {
	models := NewModels(testDB.Pool, testDB.STDB)

	dealership := &Dealership{
		Name: "Benchmark Dealership",
		Address: Address{
			Street:     "123 Main St",
			StreetExt:  "",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "USA",
			Latitude:   40.7128,
			Longitude:  -74.0060,
		},
	}
	_ = models.Dealerships.Insert(dealership)

	project := &Project{
		Name:         "Benchmark Project",
		Status:       ProjectStatusi.AwaitingProof,
		Approved:     false,
		DealershipID: dealership.ID,
	}
	_ = models.Projects.Insert(project)

	user := &User{
		Name:         "Benchmark User",
		Email:        fmt.Sprintf("benchmarkchatget%d@example.com", time.Now().UnixNano()),
		Avatar:       "https://example.com/avatar.png",
		DealershipID: dealership.ID,
		Role:         UserRoles.User,
	}
	_ = models.Users.Insert(user)

	inlay := &Inlay{
		ProjectID:  project.ID,
		Name:       "Benchmark Inlay",
		PreviewURL: "https://example.com/preview.png",
		PriceGroup: 1,
		Type:       InlayTypes.Custom,
		CustomInfo: &InlayCustomInfo{
			Description: "Benchmark inlay",
			Width:       24.5,
			Height:      36.75,
		},
	}
	_ = models.Inlays.Insert(inlay)

	chat := &InlayChat{
		InlayID:    inlay.ID,
		UserID:     user.ID,
		SenderType: SenderTypes.Customer,
		Message:    "Benchmark message",
	}
	_ = models.InlayChats.Insert(chat)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = models.InlayChats.GetByID(chat.ID)
	}
}

func BenchmarkInlayChatModel_GetAllByInlayID(b *testing.B) {
	models := NewModels(testDB.Pool, testDB.STDB)

	dealership := &Dealership{
		Name: "Benchmark Dealership",
		Address: Address{
			Street:     "123 Main St",
			StreetExt:  "",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "USA",
			Latitude:   40.7128,
			Longitude:  -74.0060,
		},
	}
	_ = models.Dealerships.Insert(dealership)

	project := &Project{
		Name:         "Benchmark Project",
		Status:       ProjectStatusi.AwaitingProof,
		Approved:     false,
		DealershipID: dealership.ID,
	}
	_ = models.Projects.Insert(project)

	user := &User{
		Name:         "Benchmark User",
		Email:        fmt.Sprintf("benchmarkchatall%d@example.com", time.Now().UnixNano()),
		Avatar:       "https://example.com/avatar.png",
		DealershipID: dealership.ID,
		Role:         UserRoles.User,
	}
	_ = models.Users.Insert(user)

	inlay := &Inlay{
		ProjectID:  project.ID,
		Name:       "Benchmark Inlay",
		PreviewURL: "https://example.com/preview.png",
		PriceGroup: 1,
		Type:       InlayTypes.Custom,
		CustomInfo: &InlayCustomInfo{
			Description: "Benchmark inlay",
			Width:       24.5,
			Height:      36.75,
		},
	}
	_ = models.Inlays.Insert(inlay)

	for i := 0; i < 50; i++ {
		chat := &InlayChat{
			InlayID:    inlay.ID,
			UserID:     user.ID,
			SenderType: SenderTypes.Customer,
			Message:    fmt.Sprintf("Benchmark message %d", i),
		}
		_ = models.InlayChats.Insert(chat)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = models.InlayChats.GetAllByInlayID(inlay.ID)
	}
}

func TestInlayChatToGenErrorHandling(t *testing.T) {
	t.Run("invalid UUID in inlayChatToGen", func(t *testing.T) {
		chat := &InlayChat{
			StandardTable: StandardTable{
				ID:   1,
				UUID: "not-a-valid-uuid",
			},
			InlayID:    1,
			UserID:     1,
			SenderType: SenderTypes.Customer,
			Message:    "Test message",
		}

		_, err := inlayChatToGen(chat)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid UUID")
	})
}

func TestInlayChatModel_Update_InvalidUUID(t *testing.T) {
	cleanupTables(t)
	models := getTestModels(t)

	dealership := createTestDealership(t, models)
	user := createTestUser(t, models, dealership.ID)
	project := createTestProject(t, models, dealership.ID)
	inlay := createTestCatalogInlay(t, models, project.ID, createTestCatalogItem(t, models))

	t.Run("update with invalid UUID fails", func(t *testing.T) {
		chat := &InlayChat{
			StandardTable: StandardTable{
				ID:      1,
				UUID:    "invalid-uuid-format",
				Version: 1,
			},
			InlayID:    inlay.ID,
			UserID:     user.ID,
			SenderType: SenderTypes.Customer,
			Message:    "Test message",
		}

		err := models.InlayChats.Update(chat)
		assert.Error(t, err)
	})
}
