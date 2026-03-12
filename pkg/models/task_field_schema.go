package models

import (
	"fmt"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

const (
	TaskFieldTypeText     = "text"
	TaskFieldTypeTextarea = "textarea"
	TaskFieldTypeNumber   = "number"
	TaskFieldTypeDate     = "date"
	TaskFieldTypeSelect   = "select"
	TaskFieldTypeCheckbox = "checkbox"
)

type TaskFieldSchema struct {
	ID           int64     `xorm:"bigint autoincr not null unique pk" json:"id" param:"schema"`
	ProjectID    int64     `xorm:"bigint not null index" json:"project_id" param:"project"`
	Name         string    `xorm:"varchar(255) not null" json:"name" valid:"minstringlength(1)"`
	FieldType    string    `xorm:"varchar(50) not null" json:"field_type"`
	Description  string    `xorm:"varchar(500) null" json:"description"`
	OptionsJSON  string    `xorm:"TEXT null" json:"options"`
	Required     bool      `xorm:"not null default false" json:"required"`
	DefaultValue string    `xorm:"TEXT null" json:"default_value"`
	Created      time.Time `xorm:"created not null" json:"created"`
	Updated      time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (*TaskFieldSchema) TableName() string {
	return "task_field_schema"
}

func (s *TaskFieldSchema) Validate() error {
	s.Name = strings.TrimSpace(s.Name)
	s.FieldType = strings.TrimSpace(s.FieldType)

	if s.ProjectID < 1 {
		return ErrProjectDoesNotExist{ID: s.ProjectID}
	}
	if s.Name == "" {
		return ErrInvalidData{Message: "The field name must not be empty."}
	}

	switch s.FieldType {
	case TaskFieldTypeText, TaskFieldTypeTextarea, TaskFieldTypeNumber, TaskFieldTypeDate, TaskFieldTypeSelect, TaskFieldTypeCheckbox:
		return nil
	default:
		return ErrInvalidData{Message: fmt.Sprintf("Invalid field type '%s'.", s.FieldType)}
	}
}

func (s *TaskFieldSchema) CanRead(sess *xorm.Session, a web.Auth) (bool, int, error) {
	return (&Project{ID: s.ProjectID}).CanRead(sess, a)
}

func (s *TaskFieldSchema) CanCreate(sess *xorm.Session, a web.Auth) (bool, error) {
	return (&Project{ID: s.ProjectID}).CanWrite(sess, a)
}

func (s *TaskFieldSchema) CanUpdate(sess *xorm.Session, a web.Auth) (bool, error) {
	return (&Project{ID: s.ProjectID}).CanWrite(sess, a)
}

func (s *TaskFieldSchema) CanDelete(sess *xorm.Session, a web.Auth) (bool, error) {
	return (&Project{ID: s.ProjectID}).CanWrite(sess, a)
}

func GetFieldSchemasForProject(sess *xorm.Session, projectID int64) ([]*TaskFieldSchema, error) {
	var schemas []*TaskFieldSchema
	err := sess.Where("project_id = ?", projectID).Asc("id").Find(&schemas)
	return schemas, err
}

func GetFieldSchemaByID(sess *xorm.Session, id int64) (*TaskFieldSchema, error) {
	schema := &TaskFieldSchema{}
	exists, err := sess.ID(id).Get(schema)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrTaskFieldSchemaNotFound{ID: id}
	}
	return schema, nil
}

type ErrTaskFieldSchemaNotFound struct{ ID int64 }

func (err ErrTaskFieldSchemaNotFound) Error() string {
	return fmt.Sprintf("task field schema not found [id: %d]", err.ID)
}

func (err ErrTaskFieldSchemaNotFound) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: 404, Code: 9701, Message: "This custom field schema does not exist."}
}
