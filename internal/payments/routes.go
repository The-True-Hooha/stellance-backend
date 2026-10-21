package payments

import (
	"net/http"

	"github.com/The-True-Hooha/stellance-backend/internal/middleware"
	"github.com/The-True-Hooha/stellance-backend/pkg/httpx"
)

func RegisterPaymentRoutes(apiV1 *httpx.RouteGroup, router *http.ServeMux, s *PaymentService) {
	h := NewPaymentHandler(s)
	auth := middleware.NewAuthMiddleware(s.jwt)

	g := apiV1.AddGroup("/payments")
	g.HandleFunc("GET /", auth.Authenticate(http.HandlerFunc(h.ListPaymentsHandler)).ServeHTTP)
	g.HandleFunc("GET /invoice/{id}", auth.Authenticate(http.HandlerFunc(h.GetInvoicePaymentsHandler)).ServeHTTP)
	g.HandleFunc("GET /request/{invoice_url}", auth.IsPublicAccess(http.HandlerFunc(h.GetPublicRequestHandler)).ServeHTTP)
	g.HandleFunc("POST /verify", auth.Authenticate(http.HandlerFunc(h.VerifyTransactionHandler)).ServeHTTP)
}
