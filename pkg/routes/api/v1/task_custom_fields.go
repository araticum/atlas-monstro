// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	auth2 "code.vikunja.io/api/pkg/modules/auth"
	"github.com/labstack/echo/v5"
)

type taskCustomFieldValuePayload struct {
	Value interface{} `json:"value"`
}

type taskCustomFieldsResponse struct {
	Schemas []*models.TaskFieldSchema      `json:"schemas"`
	Values  []*models.TaskCustomFieldValue `json:"values"`
}

func GetTaskCustomFields(c *echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID.").Wrap(err)
	}

	s := db.NewSession()
	defer s.Close()

	a, err := auth2.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	task := &models.Task{ID: taskID}
	canRead, _, err := task.CanRead(s, a)
	if err != nil {
		return err
	}
	if !canRead {
		return echo.ErrForbidden
	}

	rawTask, err := models.GetTaskByIDSimple(s, taskID)
	if err != nil {
		return err
	}
	schemas, err := models.GetFieldSchemasForProject(s, rawTask.ProjectID)
	if err != nil {
		return err
	}
	values, err := models.GetCustomFieldValuesForTask(s, taskID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, taskCustomFieldsResponse{Schemas: schemas, Values: values})
}

func PutTaskCustomFieldValue(c *echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID.").Wrap(err)
	}
	schemaID, err := strconv.ParseInt(c.Param("schema"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid schema ID.").Wrap(err)
	}

	payload := &taskCustomFieldValuePayload{}
	if err := c.Bind(payload); err != nil {
		return models.ErrInvalidModel{Message: "Could not parse custom field value payload.", Err: err}
	}

	s := db.NewSession()
	defer s.Close()

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

	rawTask, err := models.GetTaskByIDSimple(s, taskID)
	if err != nil {
		return err
	}
	schema, err := models.GetFieldSchemaByID(s, schemaID)
	if err != nil {
		return err
	}
	if schema.ProjectID != rawTask.ProjectID {
		return models.ErrTaskFieldSchemaNotFound{ID: schemaID}
	}

	valueString, err := normalizeCustomFieldValue(schema.FieldType, payload.Value)
	if err != nil {
		return err
	}

	existing := &models.TaskCustomFieldValue{}
	has, err := s.Where("task_id = ? AND field_id = ?", taskID, schemaID).Get(existing)
	if err != nil {
		return err
	}
	if has {
		oldValue := existing.Value
		existing.Value = valueString
		_, err = s.ID(existing.ID).Cols("value").Update(existing)
		if err != nil {
			_ = s.Rollback()
			return err
		}
		if authUser, err := models.GetUserOrLinkShareUser(s, a); err == nil && authUser != nil {
			if err := models.RecordTaskActivity(s, taskID, authUser.ID, "custom_field_changed", map[string]models.TaskActivityFieldChange{
				fmt.Sprintf("custom_field_%d", schemaID): {Old: oldValue, New: valueString},
			}); err != nil {
				_ = s.Rollback()
				return err
			}
		}
		if err := s.Commit(); err != nil {
			_ = s.Rollback()
			return err
		}
		return c.JSON(http.StatusOK, existing)
	}

	value := &models.TaskCustomFieldValue{TaskID: taskID, FieldID: schemaID, Value: valueString}
	_, err = s.Insert(value)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	if authUser, err := models.GetUserOrLinkShareUser(s, a); err == nil && authUser != nil {
		if err := models.RecordTaskActivity(s, taskID, authUser.ID, "custom_field_changed", map[string]models.TaskActivityFieldChange{
			fmt.Sprintf("custom_field_%d", schemaID): {Old: nil, New: valueString},
		}); err != nil {
			_ = s.Rollback()
			return err
		}
	}
	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}
	return c.JSON(http.StatusCreated, value)
}

func DeleteTaskCustomFieldValue(c *echo.Context) error {
	taskID, err := strconv.ParseInt(c.Param("task"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid task ID.").Wrap(err)
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

	task := &models.Task{ID: taskID}
	canDelete, err := task.CanWrite(s, a)
	if err != nil {
		return err
	}
	if !canDelete {
		return echo.ErrForbidden
	}

	_, err = models.GetFieldSchemaByID(s, schemaID)
	if err != nil {
		return err
	}

	value, err := models.GetCustomFieldValue(s, taskID, schemaID)
	if err != nil {
		return err
	}
	if _, err = s.ID(value.ID).Delete(&models.TaskCustomFieldValue{}); err != nil {
		_ = s.Rollback()
		return err
	}
	if authUser, err := models.GetUserOrLinkShareUser(s, a); err == nil && authUser != nil {
		if err := models.RecordTaskActivity(s, taskID, authUser.ID, "custom_field_changed", map[string]models.TaskActivityFieldChange{
			fmt.Sprintf("custom_field_%d", schemaID): {Old: value.Value, New: nil},
		}); err != nil {
			_ = s.Rollback()
			return err
		}
	}
	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "The custom field value was deleted successfully."})
}

func normalizeCustomFieldValue(fieldType string, value interface{}) (string, error) {
	if value == nil {
		return "", nil
	}

	switch fieldType {
	case models.TaskFieldTypeCheckbox:
		b, ok := value.(bool)
		if !ok {
			return "", models.ErrInvalidData{Message: "Checkbox custom field values must be boolean."}
		}
		if b {
			return "true", nil
		}
		return "false", nil
	case models.TaskFieldTypeNumber:
		switch n := value.(type) {
		case float64:
			return strconv.FormatFloat(n, 'f', -1, 64), nil
		case string:
			if _, err := strconv.ParseFloat(strings.TrimSpace(n), 64); err != nil {
				return "", models.ErrInvalidData{Message: "Number custom field values must be numeric."}
			}
			return strings.TrimSpace(n), nil
		default:
			return "", models.ErrInvalidData{Message: "Number custom field values must be numeric."}
		}
	case models.TaskFieldTypeText, models.TaskFieldTypeTextarea, models.TaskFieldTypeDate, models.TaskFieldTypeSelect:
		s, ok := value.(string)
		if ok {
			return s, nil
		}
		buf, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		return string(buf), nil
	default:
		return "", models.ErrInvalidData{Message: "Invalid custom field type."}
	}
}
