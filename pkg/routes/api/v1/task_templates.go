package v1

import (
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	auth2 "code.vikunja.io/api/pkg/modules/auth"
	"github.com/labstack/echo/v5"
)

type taskTemplateUpdatePayload struct {
	IsTemplate bool `json:"is_template"`
}

func ListProjectTaskTemplates(c *echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID.").Wrap(err)
	}

	s := db.NewSession()
	defer s.Close()

	a, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	project := &models.Project{ID: projectID}
	canRead, _, err := project.CanRead(s, a)
	if err != nil {
		return err
	}
	if !canRead {
		return echo.ErrForbidden
	}

	var rawTemplates []*models.Task
	err = s.Where("project_id = ? AND is_template = ?", projectID, true).Asc("`index`").Find(&rawTemplates)
	if err != nil {
		return err
	}

	templates := make([]*models.Task, 0, len(rawTemplates))
	for _, raw := range rawTemplates {
		template := &models.Task{ID: raw.ID}
		if err := template.ReadOne(s, a); err != nil {
			return err
		}
		templates = append(templates, template)
	}

	return c.JSON(http.StatusOK, templates)
}

func DuplicateTaskTemplate(c *echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID.").Wrap(err)
	}
	templateID, err := strconv.ParseInt(c.Param("template"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid template ID.").Wrap(err)
	}

	s := db.NewSession()
	defer s.Close()
	if err := s.Begin(); err != nil {
		return err
	}

	a, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	project := &models.Project{ID: projectID}
	canWrite, err := project.CanWrite(s, a)
	if err != nil {
		return err
	}
	if !canWrite {
		return echo.ErrForbidden
	}

	originalTask := &models.Task{ID: templateID}
	if err := originalTask.ReadOne(s, a); err != nil {
		return err
	}
	if originalTask.ProjectID != projectID || !originalTask.IsTemplate {
		return models.ErrTaskDoesNotExist{ID: templateID}
	}

	newTask := &models.Task{
		Title:       originalTask.Title,
		Description: originalTask.Description,
		Done:        false,
		DueDate:     originalTask.DueDate,
		ProjectID:   originalTask.ProjectID,
		RepeatAfter: originalTask.RepeatAfter,
		RepeatMode:  originalTask.RepeatMode,
		Priority:    originalTask.Priority,
		StartDate:   originalTask.StartDate,
		EndDate:     originalTask.EndDate,
		Assignees:   originalTask.Assignees,
		HexColor:    originalTask.HexColor,
		PercentDone: originalTask.PercentDone,
		Reminders:   originalTask.Reminders,
		IsTemplate:  false,
	}
	if err := newTask.Create(s, a); err != nil {
		_ = s.Rollback()
		return err
	}

	labelTasks := []*models.LabelTask{}
	if err := s.Where("task_id = ?", templateID).Find(&labelTasks); err != nil {
		_ = s.Rollback()
		return err
	}
	for _, lt := range labelTasks {
		lt.ID = 0
		lt.TaskID = newTask.ID
	}
	if len(labelTasks) > 0 {
		if _, err := s.Insert(&labelTasks); err != nil {
			_ = s.Rollback()
			return err
		}
	}

	customFieldValues, err := models.GetCustomFieldValuesForTask(s, templateID)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	newValues := make([]*models.TaskCustomFieldValue, 0, len(customFieldValues))
	for _, value := range customFieldValues {
		newValues = append(newValues, &models.TaskCustomFieldValue{
			TaskID:  newTask.ID,
			FieldID: value.FieldID,
			Value:   value.Value,
		})
	}
	if len(newValues) > 0 {
		if _, err := s.Insert(&newValues); err != nil {
			_ = s.Rollback()
			return err
		}
	}

	if err := newTask.ReadOne(s, a); err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusCreated, newTask)
}

func UpdateTaskTemplate(c *echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID.").Wrap(err)
	}

	payload := &taskTemplateUpdatePayload{}
	if err := c.Bind(payload); err != nil {
		return models.ErrInvalidModel{Message: "Could not parse task template payload.", Err: err}
	}

	s := db.NewSession()
	defer s.Close()
	if err := s.Begin(); err != nil {
		return err
	}

	a, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	task := &models.Task{ID: taskID}
	canWrite, err := task.CanWrite(s, a)
	if err != nil {
		return err
	}
	if !canWrite {
		return echo.ErrForbidden
	}

	originalTask, err := models.GetTaskByIDSimple(s, taskID)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	originalTask.IsTemplate = payload.IsTemplate
	if _, err := s.ID(taskID).Cols("is_template").Update(&originalTask); err != nil {
		_ = s.Rollback()
		return err
	}
	task = &originalTask

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, task)
}
