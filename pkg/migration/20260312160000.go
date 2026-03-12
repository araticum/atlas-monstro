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

package migration

import (
	"code.vikunja.io/api/pkg/config"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260312160000",
		Description: "Add is_template flag to tasks",
		Migrate: func(tx *xorm.Engine) error {
			switch config.DatabaseType.GetString() {
			case "postgres":
				_, err := tx.Exec("ALTER TABLE tasks ADD COLUMN IF NOT EXISTS is_template BOOLEAN DEFAULT FALSE")
				return err
			case "mysql":
				_, err := tx.Exec("ALTER TABLE tasks ADD COLUMN is_template BOOLEAN DEFAULT FALSE")
				return err
			case "sqlite":
				exists, err := columnExists(tx, "tasks", "is_template")
				if err != nil {
					return err
				}
				if exists {
					return nil
				}
				_, err = tx.Exec("ALTER TABLE tasks ADD COLUMN is_template BOOLEAN DEFAULT FALSE")
				return err
			default:
				_, err := tx.Exec("ALTER TABLE tasks ADD COLUMN is_template BOOLEAN DEFAULT FALSE")
				return err
			}
		},
		Rollback: func(tx *xorm.Engine) error {
			return dropTableColum(tx, "tasks", "is_template")
		},
	})
}
