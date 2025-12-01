package handlers

import (
	"agent-workspace-manager/internal/database"
	"agent-workspace-manager/internal/models"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
)

// validateProjectName 驗證專案名稱是否只包含英文、數字、底線
func validateProjectName(name string) bool {
	match, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", name)
	return match
}

// validateDirectory 驗證目錄是否存在且為目錄
func validateDirectory(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return "", os.ErrNotExist
	}
	if !info.IsDir() {
		return "", os.ErrInvalid
	}
	return absPath, nil
}

// CreateProject 處理建立新專案的請求
func CreateProject(c *gin.Context) {
	var input struct {
		Name          string `json:"name" binding:"required"`
		Description   string `json:"description"`
		AICliCommand  string `json:"ai_cli_command"`
		DirectoryPath string `json:"directory_path" binding:"required"`
	}

	// 綁定並驗證 JSON 輸入
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 驗證專案名稱格式
	if !validateProjectName(input.Name) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project name. Only alphanumeric characters and underscores are allowed."})
		return
	}

	// 驗證目錄路徑
	absPath, err := validateDirectory(input.DirectoryPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Directory does not exist"})
		} else if err == os.ErrInvalid {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Path is not a directory"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid directory path"})
		}
		return
	}

	// 建立專案模型
	project := models.Project{
		Name:          input.Name,
		Description:   input.Description,
		AICliCommand:  input.AICliCommand,
		DirectoryPath: absPath,
	}

	// 儲存至資料庫
	if err := database.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// GetProjects 取得所有專案列表
func GetProjects(c *gin.Context) {
	var projects []models.Project
	database.DB.Find(&projects)
	c.JSON(http.StatusOK, projects)
}

// GetProject 根據 ID 取得單一專案
func GetProject(c *gin.Context) {
	id := c.Param("id")
	var project models.Project
	if err := database.DB.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	c.JSON(http.StatusOK, project)
}

// UpdateProject 更新專案資訊
func UpdateProject(c *gin.Context) {
	id := c.Param("id")
	var project models.Project
	if err := database.DB.First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var input struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		AICliCommand  string `json:"ai_cli_command"`
		DirectoryPath string `json:"directory_path"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 驗證專案名稱格式 (如果有更新)
	if input.Name != "" {
		if !validateProjectName(input.Name) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project name. Only alphanumeric characters and underscores are allowed."})
			return
		}
		project.Name = input.Name
	}

	// 驗證目錄路徑 (如果有更新)
	if input.DirectoryPath != "" {
		absPath, err := validateDirectory(input.DirectoryPath)
		if err != nil {
			if os.IsNotExist(err) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Directory does not exist"})
			} else if err == os.ErrInvalid {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Path is not a directory"})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid directory path"})
			}
			return
		}
		project.DirectoryPath = absPath
	}

	if input.Description != "" {
		project.Description = input.Description
	}
	if input.AICliCommand != "" {
		project.AICliCommand = input.AICliCommand
	}

	database.DB.Save(&project)
	c.JSON(http.StatusOK, project)
}

// DeleteProject 刪除專案
func DeleteProject(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.Project{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Project deleted"})
}

// GetProjectExecutions 取得特定專案的執行記錄
func GetProjectExecutions(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var executions []models.Execution
	// 依照開始時間倒序排列
	if err := database.DB.Where("project_id = ?", projectID).Order("start_time desc").Find(&executions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch executions"})
		return
	}

	c.JSON(http.StatusOK, executions)
}