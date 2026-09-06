// Package googleauth completes a desktop OAuth connection to Google Drive.
package googleauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// drive.file is non-sensitive and lets Nuvem work only with the folder the
// person explicitly selects through Google's Picker flow.
const driveScope = "https://www.googleapis.com/auth/drive.file"

// Result is the authorization together with the folder selected in Google's
// official Picker flow.
type Result struct {
	Token          *oauth2.Token
	PickedFolderID string
}

type callbackResult struct {
	code     string
	folderID string
	err      error
}

// Connect opens the system browser and receives the browser callback on an
// ephemeral localhost port. The OAuth client must be a Google Desktop client.
func Connect(ctx context.Context, clientID, clientSecret string) (Result, error) {
	if clientID == "" {
		return Result{}, fmt.Errorf("o ID do cliente OAuth do Google é obrigatório")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return Result{}, fmt.Errorf("abrir retorno local: %w", err)
	}
	defer listener.Close()
	state, err := randomState()
	if err != nil {
		return Result{}, err
	}
	redirectURL := "http://" + listener.Addr().String() + "/oauth/callback"
	configuration := oauth2.Config{ClientID: clientID, ClientSecret: clientSecret, RedirectURL: redirectURL, Scopes: []string{driveScope}, Endpoint: oauth2.Endpoint{AuthURL: "https://accounts.google.com/o/oauth2/auth", TokenURL: "https://oauth2.googleapis.com/token"}}
	result := make(chan callbackResult, 1)
	server := &http.Server{Handler: callbackHandler(state, result)}
	go server.Serve(listener)
	defer server.Shutdown(context.Background())

	authURL := configuration.AuthCodeURL(state, oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
		oauth2.SetAuthURLParam("trigger_onepick", "true"),
		oauth2.SetAuthURLParam("allow_folder_selection", "true"),
		// Do not carry a previously granted broad Drive scope into this token.
		oauth2.SetAuthURLParam("include_granted_scopes", "false"),
	)
	if err := exec.CommandContext(ctx, "xdg-open", authURL).Start(); err != nil {
		return Result{}, fmt.Errorf("abrir o navegador para conectar o Google Drive: %w", err)
	}
	select {
	case callback := <-result:
		if callback.err != nil {
			return Result{}, callback.err
		}
		if callback.code == "" {
			return Result{}, fmt.Errorf("o Google Drive recusou ou interrompeu a conexão")
		}
		token, err := configuration.Exchange(ctx, callback.code)
		if err != nil {
			return Result{}, fmt.Errorf("trocar autorização do Google Drive: %w", err)
		}
		return Result{Token: token, PickedFolderID: callback.folderID}, nil
	case <-ctx.Done():
		return Result{}, ctx.Err()
	case <-time.After(3 * time.Minute):
		return Result{}, fmt.Errorf("a conexão com o Google Drive expirou; tente novamente")
	}
}

func callbackHandler(expectedState string, result chan<- callbackResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/callback" || r.URL.Query().Get("state") != expectedState {
			http.Error(w, "Invalid Nuvem authorization callback.", http.StatusBadRequest)
			return
		}
		if r.URL.Query().Get("error") != "" {
			select {
			case result <- callbackResult{}:
			default:
			}
			fmt.Fprint(w, "Nuvem was not connected. You may close this page.")
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Authorization code is missing.", http.StatusBadRequest)
			return
		}
		folderIDs := strings.Split(strings.TrimSpace(r.URL.Query().Get("picked_file_ids")), ",")
		if len(folderIDs) != 1 || strings.TrimSpace(folderIDs[0]) == "" {
			select {
			case result <- callbackResult{err: fmt.Errorf("selecione uma única pasta do Google Drive para continuar")}:
			default:
			}
			http.Error(w, "Select one Google Drive folder and try again.", http.StatusBadRequest)
			return
		}
		select {
		case result <- callbackResult{code: code, folderID: strings.TrimSpace(folderIDs[0])}:
		default:
		}
		fmt.Fprint(w, "Google Drive is connected to Nuvem. You may close this page.")
	}
}

func randomState() (string, error) {
	value := make([]byte, 24)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}
