package command

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
)

type DynamicAutocompleteHandler struct {
	root *mux.Router
	API  *pluginapi.Client
}

func NewDynamicAutocompleteHandler(api *pluginapi.Client) *DynamicAutocompleteHandler {
	h := &DynamicAutocompleteHandler{}

	root := mux.NewRouter()
	root.Use(MattermostAuthorizationRequired)

	root.HandleFunc("/dynamic_users", h.handleDynamicUsers)

	root.Handle("{anything:.*}", http.NotFoundHandler())
	root.NotFoundHandler = http.NotFoundHandler()

	h.root = root
	h.API = api

	return h
}

// PluginIDContextKey Key used to store the sourcePluginID for http requests.
const PluginIDContextKey = "plugin_id"

func MattermostAuthorizationRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		if userID != "" {
			next.ServeHTTP(w, r)
			return
		}

		pluginID, ok := r.Context().Value(PluginIDContextKey).(string)
		if ok && pluginID != "" {
			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, "Not authorized", http.StatusUnauthorized)
	})
}

func (h *DynamicAutocompleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, sourcePluginID string) {
	h.root.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), PluginIDContextKey, sourcePluginID)))
}

func (h *DynamicAutocompleteHandler) handleDynamicUsers(w http.ResponseWriter, r *http.Request) {
	a := make([]model.AutocompleteListItem, 0)
	a = append(a, model.AutocompleteListItem{HelpText: "Tester Two", Item: "@tester2"})
	a = append(a, model.AutocompleteListItem{HelpText: "Some Guy", Item: "@someguy"})
	a = append(a, model.AutocompleteListItem{HelpText: "Some Other Guy", Item: "@someotherguy"})
	b, _ := json.Marshal(a)
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(b); err != nil {
		h.API.Log.Error("failed to write status", "err", err.Error())
	}
}
