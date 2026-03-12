package v1

import (
	"net/http"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	auth2 "code.vikunja.io/api/pkg/modules/auth"
	"github.com/labstack/echo/v5"
)

func ListTaskFieldSchemas(c *echo.Context) error {
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

	schemas, err := models.GetFieldSchemasForProject(s, projectID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, schemas)
}

func CreateTaskFieldSchema(c *echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID.").Wrap(err)
	}

	payload := &models.TaskFieldSchema{ProjectID: projectID}
	if err := c.Bind(payload); err != nil {
		return models.ErrInvalidModel{Message: "Could not parse custom field schema payload.", Err: err}
	}
	payload.ProjectID = projectID
	payload.Name = strings.TrimSpace(payload.Name)
	payload.FieldType = strings.TrimSpace(payload.FieldType)
	if err := payload.Validate(); err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	a, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	canWrite, err := payload.CanCreate(s, a)
	if err != nil {
		return err
	}
	if !canWrite {
		return echo.ErrForbidden
	}

	exists, err := s.Where("project_id = ? AND name = ?", payload.ProjectID, payload.Name).Exist(&models.TaskFieldSchema{})
	if err != nil {
		return err
	}
	if exists {
		return models.ErrInvalidData{Message: "A custom field with this name already exists in the project."}
	}

	_, err = s.Insert(payload)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusCreated, payload)
}

func DeleteTaskFieldSchema(c *echo.Context) error {
	projectID, err := strconv.ParseInt(c.Param("project"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid project ID.").Wrap(err)
	}
	schemaID, err := strconv.ParseInt(c.Param("schema"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid schema ID.").Wrap(err)
	}

	s := db.NewSession()
	defer s.Close()

	a, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	schema, err := models.GetFieldSchemaByID(s, schemaID)
	if err != nil {
		return err
	}
	if schema.ProjectID != projectID {
		return models.ErrTaskFieldSchemaNotFound{ID: schemaID}
	}

	canDelete, err := schema.CanDelete(s, a)
	if err != nil {
		return err
	}
	if !canDelete {
		return echo.ErrForbidden
	}

	if _, err = s.Where("field_id = ?", schemaID).Delete(&models.TaskCustomFieldValue{}); err != nil {
		_ = s.Rollback()
		return err
	}
	if _, err = s.ID(schemaID).Delete(&models.TaskFieldSchema{}); err != nil {
		_ = s.Rollback()
		return err
	}
	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "The custom field schema was deleted successfully."})
}
