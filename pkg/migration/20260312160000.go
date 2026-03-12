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
