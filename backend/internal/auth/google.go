package auth

import (
	"context"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func GoogleOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func BeginGoogleLogin(w http.ResponseWriter, r *http.Request) {
	cfg := GoogleOAuthConfig()
	state := "dev" // TODO: CSRF-safe random state and session storage
	url := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

func HandleGoogleCallback(r *http.Request) (*oauth2.Token, error) {
	cfg := GoogleOAuthConfig()
	code := r.URL.Query().Get("code")
	ctx := context.Background()
	return cfg.Exchange(ctx, code)
}


