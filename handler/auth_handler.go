package handler

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"phase3-api-architecture/models"
	"phase3-api-architecture/repository"
	"phase3-api-architecture/utils"
)

type AuthHandler struct {
	// Repo *repository.UserRepository
	Repo repository.UserRepoInterface
}

// Register godoc
// @Summary      Mendaftarkan user baru
// @Description  Input data user untuk disimpan ke database
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body models.User true "Data User"
// @Success      201  {object}  utils.APIResponse
// @Failure      400  {object}  utils.APIResponse
// @Failure      500  {object}  utils.APIResponse
// @Router       /register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var u models.User

	// 1. Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		// Log Warning (Client Error)
		slog.Warn("register decode error", "error", err)
		utils.ResponseError(w, http.StatusBadRequest, "Invalid Input Format")
		return
	}

	// 2. Set Default Role (Penting untuk RBAC)
	// Kita force role disini, supaya user gak bisa set role sendiri
	u.Role = "user"

	// 3. Hash Password
	hashedPassword, err := utils.HashPassword(u.Password)
	if err != nil {
		// Log Error (Server Error)
		slog.Error("hashing failed", "error", err)
		utils.ResponseError(w, http.StatusInternalServerError, "Gagal memproses password")
		return
	}
	u.Password = hashedPassword

	// 4. Simpan ke DB
	if err := h.Repo.Register(u); err != nil {
		// Log Error dengan Context Email
		slog.Error("register db error", "error", err, "email", u.Email)

		// Kita asumsikan errornya karena Duplicate Key (Email sudah ada)
		utils.ResponseError(w, http.StatusConflict, "Email mungkin sudah terdaftar")
		return
	}

	// 5. Success Log & Response
	slog.Info("user registered successfully", "email", u.Email, "role", u.Role)
	utils.ResponseJSON(w, http.StatusCreated, "User berhasil didaftarkan", nil)
}

// Login godoc
// @Summary      Masuk ke dalam sistem
// @Description  Mengecek email & password, lalu mengembalikan token JWT
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body models.User true "Email dan Password"
// @Success      200  {object}  models.LoginResponse
// @Failure      400  {object}  utils.APIResponse
// @Failure      401  {object}  utils.APIResponse
// @Router       /login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input models.User

	// 1. Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		slog.Warn("login decode error", "error", err)
		utils.ResponseError(w, http.StatusBadRequest, "Invalid Input Format")
		return
	}

	// 2. Cek User di DB
	userInDB, err := h.Repo.GetByEmail(input.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// LOG: Spesifik (User not found)
			slog.Warn("login failed: user not found", "email", input.Email)
			// RESPON: Generik (Demi keamanan, jangan bilang "User gak ada")
			utils.ResponseError(w, http.StatusUnauthorized, "Email atau password salah")
			return
		}
		// Error lain (DB mati, koneksi putus, dll)
		slog.Error("login db error", "error", err, "email", input.Email)
		utils.ResponseError(w, http.StatusInternalServerError, "Terjadi kesalahan pada server")
		return
	}

	// 3. Cek Password
	isValid := utils.CheckPasswordHash(input.Password, userInDB.Password)
	if !isValid {
		// LOG: Spesifik (Wrong Password)
		slog.Warn("login failed: wrong password", "email", input.Email)
		// RESPON: Generik
		utils.ResponseError(w, http.StatusUnauthorized, "Email atau password salah")
		return
	}

	// 4. Generate Token
	token, err := utils.GenerateToken(userInDB.ID, userInDB.Email, userInDB.Role)
	if err != nil {
		slog.Error("token generation failed", "error", err, "user_id", userInDB.ID)
		utils.ResponseError(w, http.StatusInternalServerError, "Gagal membuat token")
		return
	}

	// 5. Success Log & Response
	slog.Info("user logged in", "email", userInDB.Email, "role", userInDB.Role)

	utils.ResponseJSON(w, http.StatusOK, "Login berhasil", models.LoginResponse{
		AccessToken: token,
	})
}
