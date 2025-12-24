package handler

import (
	"encoding/json"
	"net/http"
	"phase3-api-architecture/models"
	"phase3-api-architecture/repository"
	"phase3-api-architecture/utils"
	"strconv"

	"github.com/go-playground/validator/v10"
)

type ProductHandler struct {
	Repo *repository.ProductRepository
}

var validate = validator.New()

func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid Product ID")
		return 0, false
	}
	return id, true
}

func (h *ProductHandler) HandleCreateProduct(w http.ResponseWriter, r *http.Request) {
	h.CreateProduct(w, r)
}

func (h *ProductHandler) HandleUpdateProduct(w http.ResponseWriter, r *http.Request) {
	// Ambil ID dari Path Value
	if id, ok := parseID(w, r); ok {
		h.UpdateProduct(w, r, id)
	}
}

func (h *ProductHandler) HandleDeleteProduct(w http.ResponseWriter, r *http.Request) {
	if id, ok := parseID(w, r); ok {
		h.DeleteProduct(w, r, id)
	}
}

func (h *ProductHandler) HandleGetProductByID(w http.ResponseWriter, r *http.Request) {
	if id, ok := parseID(w, r); ok {
		h.GetProductByID(w, r, id)
	}
}

// GetAllProducts godoc
// @Summary      Ambil Semua Produk
// @Description  Mengambil list produk dengan pagination & search
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        page   query    int     false  "Halaman ke- (Default 1)"
// @Param        limit  query    int     false  "Jumlah data (Default 10)"
// @Param        search query    string  false  "Cari nama produk"
// @Success      200  {object}  utils.APIResponse
// @Failure      500  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /products [get]
func (h *ProductHandler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	pageStr := query.Get("page")
	limitStr := query.Get("limit")
	search := query.Get("search")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 10
	}

	filter := models.ProductFilter{
		Page:   page,
		Limit:  limit,
		Search: search,
	}

	products, err := h.Repo.GetAll(filter)
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "Gagal mengambil data produk")
		return
	}
	utils.ResponseJSON(w, http.StatusOK, "List semua produk", products)
}

// CreateProduct godoc
// @Summary      Tambah Produk Baru (Admin Only)
// @Description  Menambahkan data produk ke database
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        request body models.Product true "Data Produk"
// @Success      201  {object}  utils.APIResponse
// @Failure      400  {object}  utils.APIResponse
// @Failure      401  {object}  utils.APIResponse "Unauthorized"
// @Failure      403  {object}  utils.APIResponse "Forbidden (Bukan Admin)"
// @Security     BearerAuth
// @Router       /products [post]
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if err := validate.Struct(p); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Validation error: "+err.Error())
		return
	}

	if err := h.Repo.Create(p); err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "Gagal membuat produk")
		return
	}

	utils.ResponseJSON(w, http.StatusCreated, "Produk berhasil ditambahkan", p)
}

// HandleGetProductByID godoc
// @Summary      Ambil Detail Produk
// @Description  Mencari produk berdasarkan ID
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Success      200  {object}  utils.APIResponse
// @Failure      404  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /products/{id} [get]
func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request, id int) {
	product, err := h.Repo.GetByID(id)
	if err != nil {
		utils.ResponseError(w, http.StatusNotFound, "Produk tidak ditemukan")
		return
	}

	utils.ResponseJSON(w, http.StatusOK, "Detail produk", product)
}

// HandleUpdateProduct godoc
// @Summary      Update Produk (Admin Only)
// @Description  Mengubah data produk berdasarkan ID
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        id      path    int             true  "Product ID"
// @Param        request body    models.Product  true  "Data Update"
// @Success      200     {object}  utils.APIResponse
// @Failure      400     {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /products/{id} [put]
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request, id int) {
	var p models.Product

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	p.ID = id

	if err := validate.Struct(p); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Validation error: "+err.Error())
		return
	}

	_, err := h.Repo.GetByID(id)
	if err != nil {
		utils.ResponseError(w, http.StatusNotFound, "Produk tidak ditemukan")
		return
	}

	if err := h.Repo.Update(p); err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "Gagal mengupdate produk")
		return
	}

	utils.ResponseJSON(w, http.StatusOK, "Produk berhasil diupdate", p)
}

// HandleDeleteProduct godoc
// @Summary      Hapus Produk (Admin Only)
// @Description  Menghapus produk dari database
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Success      200  {object}  utils.APIResponse
// @Failure      404  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /products/{id} [delete]
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request, id int) {
	_, err := h.Repo.GetByID(id)
	if err != nil {
		utils.ResponseError(w, http.StatusNotFound, "Produk tidak ditemukan")
		return
	}

	if err := h.Repo.Delete(id); err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "Gagal menghapus produk")
		return
	}

	utils.ResponseJSON(w, http.StatusOK, "Produk berhasil dihapus", nil)

}

// / HandleCheckout godoc
// @Summary      Beli Produk
// @Description  User membeli produk (mengurangi stok dan catat transaksi)
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Param        request body models.CheckoutRequest true "Data Pembelian"  <-- Pastikan model ini ada
// @Success      200  {object}  utils.APIResponse
// @Failure      400  {object}  utils.APIResponse
// @Security     BearerAuth
// @Router       /checkout [post]
func (h *ProductHandler) HandleCheckout(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		utils.ResponseError(w, http.StatusUnauthorized, "User ID tidak valid!")
		return
	}

	userEmail, ok := r.Context().Value("email").(string)
	if !ok {
		userEmail = ""
	}

	var req models.CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Format input salah!")
		return
	}

	if err := validate.Struct(req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.Repo.Checkout(r.Context(), userID, userEmail, req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.ResponseJSON(w, http.StatusOK, "Pembelian berhasil, invoice akan dikirim via email", nil)
}
