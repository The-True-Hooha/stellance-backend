package payments

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/The-True-Hooha/stellance-backend/pkg/utils"
	"github.com/go-playground/validator/v10"
)

type PaymentHandler struct {
	service   *PaymentService
	validator *validator.Validate
}

func NewPaymentHandler(s *PaymentService) *PaymentHandler {
	return &PaymentHandler{service: s, validator: validator.New()}
}

func write(w http.ResponseWriter, resp *utils.ApiResponse) {
	utils.WriteToJson(w, resp.StatusCode, resp)
}

func userID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id, ok := utils.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}
	return id, ok
}

// GetInvoicePaymentsHandler godoc
// @Summary      Payment request, status and payment records of an invoice (v2)
// @Tags         payments-v2
// @Produce      json
// @Param        id   path  string  true  "Invoice ID"
// @Success      200  {object}  utils.ApiResponse
// @Security     BearerAuth
// @Router       /payments/invoice/{id} [get]
func (h *PaymentHandler) GetInvoicePaymentsHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(w, r)
	if !ok {
		return
	}
	write(w, h.service.GetInvoicePayments(r.Context(), uid, r.PathValue("id")))
}

// GetPublicRequestHandler godoc
// @Summary      Payment instructions for an invoice, by invoice url (v2)
// @Tags         payments-v2
// @Produce      json
// @Param        invoice_url  path  string  true  "Invoice URL"
// @Success      200  {object}  utils.ApiResponse
// @Router       /payments/request/{invoice_url} [get]
func (h *PaymentHandler) GetPublicRequestHandler(w http.ResponseWriter, r *http.Request) {
	write(w, h.service.GetPublicRequest(r.Context(), r.PathValue("invoice_url")))
}

// ListPaymentsHandler godoc
// @Summary      Payments received on the user's wallets (verified, invalid, unmatched, failed) (v2)
// @Tags         payments-v2
// @Produce      json
// @Param        status  query  string  false  "verification status"
// @Param        limit   query  int     false  "page size, max 100"
// @Param        offset  query  int     false  "offset"
// @Success      200  {object}  utils.ApiResponse
// @Security     BearerAuth
// @Router       /payments [get]
func (h *PaymentHandler) ListPaymentsHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(w, r)
	if !ok {
		return
	}
	status := r.URL.Query().Get("status")
	switch VerificationStatus(status) {
	case "", StatusPending, StatusDetected, StatusVerified, StatusInvalid, StatusUnmatched, StatusFailed:
	default:
		write(w, &utils.ApiResponse{StatusCode: http.StatusBadRequest, Message: "invalid status filter"})
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}
	write(w, h.service.ListPayments(r.Context(), uid, status, limit, offset))
}

// VerifyTransactionHandler godoc
// @Summary      Re-check a Stellar transaction against invoices (v2)
// @Tags         payments-v2
// @Accept       json
// @Produce      json
// @Param        body  body  VerifyTransactionDTO  true  "Transaction hash"
// @Success      200  {object}  utils.ApiResponse
// @Security     BearerAuth
// @Router       /payments/verify [post]
func (h *PaymentHandler) VerifyTransactionHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(w, r)
	if !ok {
		return
	}
	var dto VerifyTransactionDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		write(w, &utils.ApiResponse{StatusCode: http.StatusBadRequest, Message: "invalid request body"})
		return
	}
	if err := h.validator.Struct(dto); err != nil {
		utils.HandleValidationError(w, err)
		return
	}
	write(w, h.service.VerifyTransactionForUser(r.Context(), uid, dto.TransactionHash))
}
