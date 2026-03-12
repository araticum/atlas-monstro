package migration

import (
    "code.vikunja.io/api/pkg/models"
    "src.techknowlogick.com/xormigrate"
    "xorm.io/xorm"
)

func init() {
    migrations = append(migrations, &xormigrate.Migration{
        ID:          "20260312151000",
        Description: "Add task activities audit log table",
        Migrate: func(tx *xorm.Engine) error {
            return tx.Sync2(new(models.TaskActivity))
        },
        Rollback: func(tx *xorm.Engine) error {
            return nil
        },
    })
}
