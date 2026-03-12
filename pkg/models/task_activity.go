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
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

type TaskActivity struct {
	ID            int64      `xorm:"bigint pk autoincr" json:"id"`
	TaskID        int64      `xorm:"bigint notnull index" json:"task_id" param:"task"`
	UserID        int64      `xorm:"bigint notnull index" json:"user_id"`
	Action        string     `xorm:"varchar(50) notnull" json:"action"`
	ChangedFields string     `xorm:"longtext" json:"changed_fields"`
	CreatedAt     time.Time  `xorm:"created" json:"created_at"`
	User          *user.User `xorm:"-" json:"user,omitempty"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

func (*TaskActivity) TableName() string {
	return "task_activities"
}

func (ta *TaskActivity) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	return (&Task{ID: ta.TaskID}).CanRead(s, a)
}

func (ta *TaskActivity) CanCreate(*xorm.Session, web.Auth) (bool, error) { return false, nil }
func (ta *TaskActivity) CanUpdate(*xorm.Session, web.Auth) (bool, error) { return false, nil }
func (ta *TaskActivity) CanDelete(*xorm.Session, web.Auth) (bool, error) { return false, nil }
func (ta *TaskActivity) Create(*xorm.Session, web.Auth) error {
	return fmt.Errorf("task activities are read-only")
}
func (ta *TaskActivity) ReadOne(*xorm.Session, web.Auth) error {
	return fmt.Errorf("task activities are read-only")
}
func (ta *TaskActivity) Update(*xorm.Session, web.Auth) error {
	return fmt.Errorf("task activities are read-only")
}
func (ta *TaskActivity) Delete(*xorm.Session, web.Auth) error {
	return fmt.Errorf("task activities are read-only")
}

func (ta *TaskActivity) ReadAll(s *xorm.Session, auth web.Auth, _ string, _ int, _ int) (interface{}, int, int64, error) {
	activities, err := GetTaskActivities(s, ta.TaskID, auth)
	if err != nil {
		return nil, 0, 0, err
	}
	return activities, len(activities), int64(len(activities)), nil
}

func RecordTaskActivity(s *xorm.Session, taskID, userID int64, action string, changes interface{}) error {
	if taskID == 0 || userID == 0 || action == "" {
		return nil
	}

	payload := "{}"
	if changes != nil {
		data, err := json.Marshal(changes)
		if err != nil {
			return err
		}
		payload = string(data)
	}

	_, err := s.Insert(&TaskActivity{
		TaskID:        taskID,
		UserID:        userID,
		Action:        action,
		ChangedFields: payload,
	})
	return err
}

func GetTaskActivities(s *xorm.Session, taskID int64, auth web.Auth) ([]*TaskActivity, error) {
	task := &Task{ID: taskID}
	canRead, _, err := task.CanRead(s, auth)
	if err != nil {
		return nil, err
	}
	if !canRead {
		return nil, ErrGenericForbidden{}
	}

	activities := []*TaskActivity{}
	if err := s.Where("task_id = ?", taskID).Desc("created_at", "id").Find(&activities); err != nil {
		return nil, err
	}

	userIDsMap := map[int64]struct{}{}
	for _, activity := range activities {
		userIDsMap[activity.UserID] = struct{}{}
	}
	ids := make([]int64, 0, len(userIDsMap))
	for id := range userIDsMap {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	users, err := getUsersOrLinkSharesFromIDs(s, ids)
	if err != nil {
		return nil, err
	}
	for _, activity := range activities {
		activity.User = users[activity.UserID]
	}

	return activities, nil
}

type TaskActivityFieldChange struct {
	Old interface{} `json:"old"`
	New interface{} `json:"new"`
}

func BuildTaskAuditChanges(oldTask, newTask *Task) map[string]TaskActivityFieldChange {
	if oldTask == nil || newTask == nil {
		return map[string]TaskActivityFieldChange{}
	}

	changes := map[string]TaskActivityFieldChange{}
	if oldTask.Title != newTask.Title {
		changes["title"] = TaskActivityFieldChange{Old: oldTask.Title, New: newTask.Title}
	}
	if oldTask.Description != newTask.Description {
		changes["description"] = TaskActivityFieldChange{Old: oldTask.Description, New: newTask.Description}
	}
	if oldTask.Done != newTask.Done {
		changes["status"] = TaskActivityFieldChange{Old: taskStatusLabel(oldTask), New: taskStatusLabel(newTask)}
	}
	if !oldTask.DueDate.Equal(newTask.DueDate) {
		changes["due_date"] = TaskActivityFieldChange{Old: formatTaskActivityTime(oldTask.DueDate), New: formatTaskActivityTime(newTask.DueDate)}
	}
	if oldTask.Priority != newTask.Priority {
		changes["priority"] = TaskActivityFieldChange{Old: oldTask.Priority, New: newTask.Priority}
	}

	oldAssignees := normalizeTaskActivityAssignees(oldTask.Assignees)
	newAssignees := normalizeTaskActivityAssignees(newTask.Assignees)
	if !stringSlicesEqual(oldAssignees, newAssignees) {
		changes["assignees"] = TaskActivityFieldChange{Old: oldAssignees, New: newAssignees}
	}

	return changes
}

func taskStatusLabel(t *Task) string {
	if t != nil && t.Done {
		return "done"
	}
	return "open"
}

func formatTaskActivityTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}

func normalizeTaskActivityAssignees(assignees []*user.User) []string {
	if len(assignees) == 0 {
		return []string{}
	}
	names := make([]string, 0, len(assignees))
	for _, assignee := range assignees {
		if assignee == nil {
			continue
		}
		if assignee.Name != "" {
			names = append(names, assignee.Name)
			continue
		}
		names = append(names, assignee.Username)
	}
	sort.Strings(names)
	return names
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
