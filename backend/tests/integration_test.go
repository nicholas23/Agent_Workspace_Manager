package tests

import (
	"agent-workspace-manager/internal/api"
	"agent-workspace-manager/internal/config"
	"agent-workspace-manager/internal/database"
	"agent-workspace-manager/internal/services/scheduler"
	"agent-workspace-manager/internal/services/telegram"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	// Setup Test DB
	os.Setenv("DATABASE_URL", ":memory:")
	cfg := config.LoadConfig()
	database.Connect(cfg.DatabaseURL)
	
	// Init Services (Mock or Real)
	// For integration test, we might want to mock Telegram/Executor if possible, 
	// but here we test the API flow.
	// We can skip Telegram init or let it fail gracefully (it logs and skips).
	telegram.InitBot(cfg)
	scheduler.InitScheduler()

	r := gin.Default()
	api.SetupRoutes(r)
	return r
}

func TestProjectFlow(t *testing.T) {
	r := setupRouter()

	// 1. Create Project
	// Use absolute path for mock AI CLI
	mockScript, _ := os.Getwd()
	mockScript = mockScript + "/mock_ai_cli.sh"
	
	projectPayload := map[string]string{
		"name":           "integration_test_project",
		"description":    "Test Project",
		"ai_cli_command": mockScript + " {prompt}", // Need {prompt} placeholder
		"directory_path": ".",
	}
	body, _ := json.Marshal(projectPayload)
	req, _ := http.NewRequest("POST", "/api/projects", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	
	var project map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &project)
	assert.NoError(t, err)
	
	// Try both lowercase "id" and uppercase "ID" (gorm.Model uses ID)
	var projectID int
	if id, ok := project["ID"].(float64); ok {
		projectID = int(id)
	} else if id, ok := project["id"].(float64); ok {
		projectID = int(id)
	} else {
		t.Fatalf("Failed to get project ID from response: %+v", project)
	}
	assert.Greater(t, projectID, 0)

	// 2. Schedule Task
	futureTime := time.Now().Add(2 * time.Second)
	schedulePayload := map[string]interface{}{
		"command":        "Scheduled Test",
		"scheduled_time": futureTime,
	}
	body, _ = json.Marshal(schedulePayload)
	req, _ = http.NewRequest("POST", "/api/projects/1/schedules", bytes.NewBuffer(body))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// 3. Wait for execution
	time.Sleep(3 * time.Second)

	// 4. Check Executions
	req, _ = http.NewRequest("GET", "/api/projects/1/executions", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var executions []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &executions)
	
	// Since we are running in memory DB and "echo" command might run fast, 
	// we expect at least one execution (from schedule).
	// However, the executor runs `os/exec` which depends on the system.
	// `echo` should work.
	
	assert.NotEmpty(t, executions)
	assert.Equal(t, "Scheduled Test", executions[0]["command"])
	assert.Equal(t, "completed", executions[0]["status"])
}
