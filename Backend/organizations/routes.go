package organizations

import (
	"context"
	"net/http"
	"strings"

	"garuda-hacks/backend/auth"
)

type contextKey string

const OrgMembershipKey contextKey = "org_membership"

func OrganizationContext(repo Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := auth.ClaimsFromContext(r.Context())
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"success":false,"error":"unauthorized"}`))
				return
			}

			orgID := extractIDFromPath(r.URL.Path)
			if orgID == "" {
				orgID = r.URL.Query().Get("org_id")
			}

			// Some actions like Join don't require pre-existing membership
			path := r.URL.Path
			segments := strings.Split(strings.Trim(path, "/"), "/")
			isJoinRequest := len(segments) == 4 && segments[2] == "members" && segments[3] == "join"

			if orgID == "" || isJoinRequest {
				next.ServeHTTP(w, r)
				return
			}

			membership, err := repo.GetMembership(r.Context(), orgID, claims.UserID)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"success":false,"error":"forbidden: you are not a member of this organization"}`))
				return
			}

			ctx := context.WithValue(r.Context(), OrgMembershipKey, membership)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func MembershipFromContext(ctx context.Context) (*OrganizationMember, bool) {
	membership, ok := ctx.Value(OrgMembershipKey).(*OrganizationMember)
	return membership, ok
}

func RegisterRoutes(mux *http.ServeMux, handler *Handler, authenticate func(http.Handler) http.Handler, repo Repository) {
	orgCtx := OrganizationContext(repo)

	// Create organization
	mux.Handle("/organizations", authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handler.Create(w, r)
		} else {
			w.Header().Set("Allow", "POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Find by ID and Update organization
	mux.Handle("/organizations/", authenticate(orgCtx(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		segments := strings.Split(strings.Trim(path, "/"), "/")

		if len(segments) == 2 {
			// e.g. /organizations/:id
			if r.Method == http.MethodGet {
				handler.FindByID(w, r)
			} else if r.Method == http.MethodPut || r.Method == http.MethodPatch {
				handler.Update(w, r)
			} else {
				w.Header().Set("Allow", "GET, PUT, PATCH")
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if len(segments) >= 3 && segments[2] == "members" {
			if len(segments) == 3 {
				// e.g. /organizations/:id/members
				if r.Method == http.MethodGet {
					handler.ListMembers(w, r)
				} else {
					w.Header().Set("Allow", "GET")
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
				return
			}
			if len(segments) == 4 {
				action := segments[3]
				switch action {
				case "join":
					if r.Method == http.MethodPost {
						handler.Join(w, r)
					} else {
						w.Header().Set("Allow", "POST")
						http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
					}
				case "leave":
					if r.Method == http.MethodPost {
						handler.Leave(w, r)
					} else {
						w.Header().Set("Allow", "POST")
						http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
					}
				case "transfer":
					if r.Method == http.MethodPost {
						handler.TransferOwnership(w, r)
					} else {
						w.Header().Set("Allow", "POST")
						http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
					}
				default:
					http.NotFound(w, r)
				}
				return
			}
		}

		http.NotFound(w, r)
	}))))
}
