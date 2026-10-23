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
// @Success      200  {object}  utils.ApiResponse{data=PaymentRequest}
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
// @Success      200  {object}  utils.ApiResponse{data=PaymentRequest}
// @Router       /payments/request/{invoice_url} [get]
func (h *PaymentHandler) GetPublicRequestHandler(w http.ResponseWriter, r *http.Request) {
	write(w, h.service.GetPublicRequest(r.Context(), r.PathValue("invoice_url")))
}

// ListPaymentsHandler godoc
// @Summary      Payments received on the user's wallets (verified, invalid, unmatched, failed) (v2)
// @Tags         payments-v2
// @Produce      json
// @Param        status  query  string  false  "verification status"
// @Param        page    query  int     false  "page number, from 1"
// @Param        limit   query  int     false  "page size, default 20, max 100"
// @Success      200  {object}  utils.ApiResponse{data=PaymentList}
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
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	write(w, h.service.ListPayments(r.Context(), uid, status, page, limit))
}

// VerifyTransactionHandler godoc
// @Summary      Re-check a Stellar transaction against invoices
// @Description  Returns 404 when the hash is unknown to Stellar. Status values and transitions: docs/payment-status-model.md (v2)
// @Tags         payments-v2
// @Accept       json
// @Produce      json
// @Param        body  body  VerifyTransactionDTO  true  "Transaction hash"
// @Success      200  {object}  utils.ApiResponse{data=[]PaymentRecord}
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

// AdminListPaymentsHandler godoc
// @Summary      All platform payments, with merchant email (v2)
// @Description  Paginated list of every detected payment, including unmatched, invalid and failed ones
// @Tags         admin
// @Produce      json
// @Param        page    query  int     false  "Page number"
// @Param        limit   query  int     false  "Items per page, max 100"
// @Param        status  query  string  false  "verification status"
// @Param        search  query  string  false  "transaction hash, payment reference, address or invoice number"
// @Success      200  {object}  utils.ApiResponse{data=PaymentList}
// @Failure      403  {object}  utils.ApiResponse
// @Security     BearerAuth
// @Router       /admin/payments [get]
func (h *PaymentHandler) AdminListPaymentsHandler(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	switch VerificationStatus(status) {
	case "", StatusPending, StatusDetected, StatusVerified, StatusInvalid, StatusUnmatched, StatusFailed:
	default:
		write(w, &utils.ApiResponse{StatusCode: http.StatusBadRequest, Message: "invalid status filter"})
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	write(w, h.service.AdminListPayments(r.Context(), status, r.URL.Query().Get("search"), page, limit))
}

// GetPaymentHandler godoc
// @Summary      One payment received on the user's wallets (v2)
// @Tags         payments-v2
// @Produce      json
// @Param        id   path  string  true  "Payment ID"
// @Success      200  {object}  utils.ApiResponse{data=PaymentRecord}
// @Failure      404  {object}  utils.ApiResponse
// @Security     BearerAuth
// @Router       /payments/{id} [get]
func (h *PaymentHandler) GetPaymentHandler(w http.ResponseWriter, r *http.Request) {
	uid, ok := userID(w, r)
	if !ok {
		return
	}
	write(w, h.service.GetPayment(r.Context(), uid, r.PathValue("id")))
}
