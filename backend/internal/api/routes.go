package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"mcp-backend/internal/auth"
	"mcp-backend/internal/helm"
	"mcp-backend/internal/storage"
)

type ServerCreateRequest struct {
	OwnerID    string                 `json:"owner_id"`
	Name       string                 `json:"name"`
	ConfigJSON map[string]interface{} `json:"config_json"`
}

// TODO: add middleware for JWT verification and tenant/workspace claims

func AttachRoutes(r *chi.Mux, log *logrus.Logger, db *storage.MongoStore, helmSvc *helm.Service) {
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK); w.Write([]byte("ok")) })

	// Google OAuth (dev-simple version)
	r.Get("/auth/google/login", func(w http.ResponseWriter, r *http.Request) { auth.BeginGoogleLogin(w, r) })
	r.Get("/auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		_, err := auth.HandleGoogleCallback(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// TODO: fetch userinfo, create/update user, issue our JWT
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("google auth ok (complete user linking in next step)"))
	})

	r.Route("/servers", func(sr chi.Router) {
		sr.Post("/", func(w http.ResponseWriter, r *http.Request) {
			var req ServerCreateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if req.Name == "" {
				http.Error(w, "name required", http.StatusBadRequest)
				return
			}
			id := uuid.NewString()
			s := storage.ServerDef{ID: id, OwnerID: req.OwnerID, Name: req.Name, ConfigJSON: req.ConfigJSON, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
			res, err := db.Servers().InsertOne(r.Context(), s)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": res.InsertedID})
		})

		sr.Get("/", func(w http.ResponseWriter, r *http.Request) {
			cur, err := db.Servers().Find(r.Context(), map[string]interface{}{})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer cur.Close(r.Context())
			var out []storage.ServerDef
			for cur.Next(r.Context()) {
				var s storage.ServerDef
				_ = cur.Decode(&s)
				out = append(out, s)
			}
			_ = json.NewEncoder(w).Encode(out)
		})

		sr.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			var s storage.ServerDef
			if err := db.Servers().FindOne(r.Context(), map[string]interface{}{"_id": id}).Decode(&s); err != nil {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(s)
		})

		sr.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			if _, err := db.Servers().DeleteOne(r.Context(), map[string]interface{}{"_id": id}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})

		// Deploy/Upgrade/Uninstall
		sr.Post("/{id}/deploy", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			var s storage.ServerDef
			if err := db.Servers().FindOne(r.Context(), map[string]interface{}{"_id": id}).Decode(&s); err != nil {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			// Serialize config JSON as Helm values directly
			values, _ := json.Marshal(s.ConfigJSON)
			if err := helmSvc.UpsertRelease("mcp-"+s.Name, string(values), ""); err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "deployed"})
		})

		sr.Post("/{id}/upgrade", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			var s storage.ServerDef
			if err := db.Servers().FindOne(r.Context(), map[string]interface{}{"_id": id}).Decode(&s); err != nil {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			var overrides map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&overrides)
			if overrides != nil {
				for k, v := range overrides {
					s.ConfigJSON[k] = v
				}
			}
			values, _ := json.Marshal(s.ConfigJSON)
			if err := helmSvc.UpsertRelease("mcp-"+s.Name, string(values), ""); err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "upgraded"})
		})

		sr.Post("/{id}/uninstall", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			var s storage.ServerDef
			if err := db.Servers().FindOne(r.Context(), map[string]interface{}{"_id": id}).Decode(&s); err != nil {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			if err := helmSvc.UninstallRelease("mcp-"+s.Name, ""); err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "uninstalled"})
		})
	})
}
