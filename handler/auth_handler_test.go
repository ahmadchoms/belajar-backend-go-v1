package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"phase3-api-architecture/mocks"
	"phase3-api-architecture/models"
	"phase3-api-architecture/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin_Success(t *testing.T) {
	// 1. SETUP STUNTMAN (MOCK)
	mockRepo := new(mocks.UserRepoMock)

	// Skenario: Kalau Mock ditanya GetByEmail("test@example.com")...
	// ...dia akan jawab: Ini user datanya, dan errornya nil.
	hashedPassword, _ := utils.HashPassword("rahasia123")
	mockUser := models.User{
		ID:       1,
		Email:    "test@example.com",
		Password: hashedPassword,
		Role:     "user",
	}

	mockRepo.On("GetByEmail", "test@example.com").Return(mockUser, nil)

	// 2. SETUP HANDLER (Pakai Mock Repo)
	authHandler := AuthHandler{Repo: mockRepo}

	// 3. SETUP REQUEST (Pura-pura request HTTP)
	requestBody := map[string]string{
		"email":    "test@example.com",
		"password": "rahasia123",
	}
	jsonValue, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder() // Perekam respon

	// 4. EKSEKUSI
	authHandler.Login(w, req)

	// 5. ASSERTION (Cek Hasil)
	res := w.Result()

	// Harusnya 200 OK
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Cek apakah method GetByEmail tadi beneran dipanggil?
	mockRepo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	// 1. SETUP
	mockRepo := new(mocks.UserRepoMock)

	// Password asli di DB "rahasia123"
	hashedPassword, _ := utils.HashPassword("rahasia123")
	mockUser := models.User{ID: 1, Email: "test@example.com", Password: hashedPassword, Role: "user"}

	mockRepo.On("GetByEmail", "test@example.com").Return(mockUser, nil)

	authHandler := AuthHandler{Repo: mockRepo}

	// 2. REQUEST (Password SALAH "salah123")
	requestBody := map[string]string{
		"email":    "test@example.com",
		"password": "salah123",
	}
	jsonValue, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()

	// 3. EKSEKUSI
	authHandler.Login(w, req)

	// 4. ASSERT
	// Harusnya 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
}
