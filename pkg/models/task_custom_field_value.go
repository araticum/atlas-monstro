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

package models

import (
	"fmt"
	"time"

	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

type TaskCustomFieldValue struct {
	ID      int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
	TaskID  int64     `xorm:"bigint not null index unique(task_field)" json:"task_id" param:"task"`
	FieldID int64     `xorm:"bigint not null index unique(task_field)" json:"field_id" param:"schema"`
	Value   string    `xorm:"TEXT null" json:"value"`
	Created time.Time `xorm:"created not null" json:"created"`
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (*TaskCustomFieldValue) TableName() string {
	return "task_custom_field_value"
}

func (v *TaskCustomFieldValue) CanRead(sess *xorm.Session, a web.Auth) (bool, int, error) {
	return (&Task{ID: v.TaskID}).CanRead(sess, a)
}

func (v *TaskCustomFieldValue) CanCreate(sess *xorm.Session, a web.Auth) (bool, error) {
	return (&Task{ID: v.TaskID}).CanWrite(sess, a)
}

func (v *TaskCustomFieldValue) CanUpdate(sess *xorm.Session, a web.Auth) (bool, error) {
	return (&Task{ID: v.TaskID}).CanWrite(sess, a)
}

func (v *TaskCustomFieldValue) CanDelete(sess *xorm.Session, a web.Auth) (bool, error) {
	return (&Task{ID: v.TaskID}).CanWrite(sess, a)
}

func GetCustomFieldValuesForTask(sess *xorm.Session, taskID int64) ([]*TaskCustomFieldValue, error) {
	var values []*TaskCustomFieldValue
	err := sess.Where("task_id = ?", taskID).Asc("field_id").Find(&values)
	return values, err
}

func GetCustomFieldValue(sess *xorm.Session, taskID, fieldID int64) (*TaskCustomFieldValue, error) {
	value := &TaskCustomFieldValue{}
	exists, err := sess.Where("task_id = ? AND field_id = ?", taskID, fieldID).Get(value)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrTaskCustomFieldValueNotFound{TaskID: taskID, FieldID: fieldID}
	}
	return value, nil
}

type ErrTaskCustomFieldValueNotFound struct {
	TaskID  int64
	FieldID int64
}

func (err ErrTaskCustomFieldValueNotFound) Error() string {
	return fmt.Sprintf("custom field value not found [task_id: %d, field_id: %d]", err.TaskID, err.FieldID)
}

func (err ErrTaskCustomFieldValueNotFound) HTTPError() web.HTTPError {
	return web.HTTPError{HTTPCode: 404, Code: 9702, Message: "This custom field value does not exist."}
}
