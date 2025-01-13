package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/layer5io/meshery/server/models"
	"github.com/sirupsen/logrus"
)

type LogLevelResponse struct {
	Status    string   `json:"status,omitempty"`
	LogLevel  string   `json:"log_level"`
	Available []string `json:"available_levels,omitempty"`
}

// getAvailableLogLevels returns all valid logging levels
func getAvailableLogLevels() []string {
	levels := make([]string, 0)
	for _, level := range logrus.AllLevels {
		levels = append(levels, level.String())
	}
	return levels
}

func (h *Handler) ServerEventConfigurationHandler(w http.ResponseWriter, req *http.Request,
	prefObj *models.Preference, user *models.User, provider models.Provider) {

	switch req.Method {
	case http.MethodPost:
		h.ServerEventConfigurationSet(w, req, prefObj, user, provider)
	case http.MethodGet:
		h.ServerEventConfigurationGet(w, req, prefObj, user, provider)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ServerEventConfigurationSet handles setting the log level
// swagger:route POST /api/system/events
func (h *Handler) ServerEventConfigurationSet(w http.ResponseWriter, req *http.Request,
	prefObj *models.Preference, user *models.User, provider models.Provider) {

	var request struct {
		LogLevel string `json:"log_level"`
	}

	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Normalize input to lowercase for case-insensitive comparison
	requestedLevel := strings.ToLower(strings.TrimSpace(request.LogLevel))

	// Validate the requested level
	level, err := logrus.ParseLevel(requestedLevel)
	if err != nil {
		response := LogLevelResponse{
			Status:    "error",
			LogLevel:  h.log.GetLevel().String(),
			Available: getAvailableLogLevels(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Set the new log level
	h.log.SetLevel(level)

	// Prepare success response
	response := LogLevelResponse{
		Status:    "success",
		LogLevel:  level.String(),
		Available: getAvailableLogLevels(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ServerEventConfigurationGet retrieves the current log level
// swagger:route GET /api/system/events
func (h *Handler) ServerEventConfigurationGet(w http.ResponseWriter, req *http.Request,
	prefObj *models.Preference, user *models.User, provider models.Provider) {

	currentLevel := h.log.GetLevel()

	response := LogLevelResponse{
		LogLevel:  currentLevel.String(),
		Available: getAvailableLogLevels(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
