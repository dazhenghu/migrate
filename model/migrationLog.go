package model

import "time"

/**
migration执行记录表
 */
type MigrationLog struct {
    Version string `gorm:"primary_key"`
    CreateAt time.Time
}

func (migrationLog MigrationLog)TableName() string {
    return "migration_log"
}
