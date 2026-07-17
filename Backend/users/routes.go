package users

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *Handler, authenticate func(http.Handler) http.Handler) {
	// Profile endpoints
	mux.Handle("/profile", authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetCurrentProfile(w, r)
		} else if r.Method == http.MethodPut || r.Method == http.MethodPatch {
			handler.UpdateCurrentProfile(w, r)
		} else {
			w.Header().Set("Allow", "GET, PUT, PATCH")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Verification endpoints
	mux.Handle("/profile/verification", authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetCurrentVerification(w, r)
		} else if r.Method == http.MethodPost {
			handler.SubmitCurrentVerification(w, r)
		} else {
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Admin endpoints - list all verifications
	mux.Handle("/admin/verifications", authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.ListVerifications(w, r)
		} else {
			w.Header().Set("Allow", "GET")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Admin endpoints - review a verification
	mux.Handle("/admin/verifications/review", authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			handler.ReviewVerification(w, r)
		} else {
			w.Header().Set("Allow", "POST, PUT, PATCH")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// User endpoints (optional / standard listing & view user by ID)
	mux.Handle("/users", authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.ListUsers(w, r)
		} else {
			w.Header().Set("Allow", "GET")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/users/", authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.GetUserByID(w, r)
		} else if r.Method == http.MethodPut || r.Method == http.MethodPatch {
			handler.UpdateProfile(w, r)
		} else {
			w.Header().Set("Allow", "GET, PUT, PATCH")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))
}
