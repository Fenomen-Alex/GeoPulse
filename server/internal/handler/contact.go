package handler

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"

	"github.com/alex/geopulse/server/internal/contact"
)

// handleContact validates and dispatches a contact request. When testMode is
// true no email is sent — the request succeeds silently so dev/demo flows work
// without ever firing a real Resend email.
func handleContact(w http.ResponseWriter, r *http.Request, testMode bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	decoder := json.NewDecoder(r.Body)
	var req contact.ContactRequest
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	// Honeypot: silently accept bot submissions without dispatching an email so
	// bots believe they succeeded (and stop retry loops).
	if strings.TrimSpace(req.Website) != "" {
		writeContactSuccess(w)
		return
	}

	if req.Name == "" || req.Email == "" || req.Subject == "" || req.Message == "" {
		writeContactError(w, http.StatusBadRequest, "All fields are required")
		return
	}

	if len(req.Name) > 100 || len(req.Subject) > 100 || len(req.Email) > 254 || len(req.Message) > 5000 {
		writeContactError(w, http.StatusBadRequest, "One or more fields are too long")
		return
	}

	// Reject obvious non-addresses without relying on the client's type="email".
	if addr, err := mail.ParseAddress(req.Email); err != nil || !strings.Contains(addr.Address, ".") || addr.Address == "" {
		writeContactError(w, http.StatusBadRequest, "Invalid email address")
		return
	}

	if !testMode {
		err = contact.SendContactEmail(req)
		if err != nil {
			writeContactError(w, http.StatusInternalServerError, "Failed to send message")
			return
		}
	}

	writeContactSuccess(w)
}

func writeContactError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeContactSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Your request has been dispatched. Our team will respond shortly."})
}