package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	authports "ctf-platform/internal/module/auth/ports"
)

func TestCASTicketValidatorReturnsPrincipal(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("ticket"); got != "ST-1" {
			t.Fatalf("unexpected ticket: %s", got)
		}
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<cas:serviceResponse xmlns:cas="http://www.yale.edu/tp/cas">
  <cas:authenticationSuccess>
    <cas:user>cas_user_1</cas:user>
    <cas:attributes>
      <cas:displayName>CAS User</cas:displayName>
      <cas:mail>cas_user_1@example.edu</cas:mail>
      <cas:className>CTF-1</cas:className>
      <cas:studentNo>20260001</cas:studentNo>
    </cas:attributes>
  </cas:authenticationSuccess>
</cas:serviceResponse>`)
	}))
	defer server.Close()

	validator := NewCASTicketValidator(zap.NewNop(), server.Client())
	principal, err := validator.ValidateTicket(
		context.Background(),
		server.URL+"/serviceValidate?service=https%3A%2F%2Fctf.example.edu%2Fapi%2Fv1%2Fauth%2Fcas%2Fcallback&ticket=ST-1",
	)
	if err != nil {
		t.Fatalf("ValidateTicket() error = %v", err)
	}
	if principal.Username != "cas_user_1" || principal.Name != "CAS User" {
		t.Fatalf("unexpected principal: %+v", principal)
	}
	if principal.Email != "cas_user_1@example.edu" || principal.ClassName != "CTF-1" {
		t.Fatalf("unexpected principal attributes: %+v", principal)
	}
}

func TestCASTicketValidatorReturnsInvalidTicketSentinel(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<serviceResponse>
  <authenticationFailure code="INVALID_TICKET">ticket not recognized</authenticationFailure>
</serviceResponse>`)
	}))
	defer server.Close()

	validator := NewCASTicketValidator(zap.NewNop(), server.Client())
	_, err := validator.ValidateTicket(context.Background(), server.URL+"/serviceValidate?ticket=ST-invalid")
	if !errors.Is(err, authports.ErrCASTicketInvalid) {
		t.Fatalf("expected ErrCASTicketInvalid, got %v", err)
	}
}

func TestCASTicketValidatorRejectsInvalidUsername(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<serviceResponse>
  <authenticationSuccess>
    <user>bad user</user>
  </authenticationSuccess>
</serviceResponse>`)
	}))
	defer server.Close()

	validator := NewCASTicketValidator(zap.NewNop(), server.Client())
	_, err := validator.ValidateTicket(context.Background(), server.URL+"/serviceValidate?ticket=ST-invalid-user")
	if !errors.Is(err, authports.ErrCASTicketInvalid) {
		t.Fatalf("expected ErrCASTicketInvalid, got %v", err)
	}
}
