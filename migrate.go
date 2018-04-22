package migrate

import (
    "github.com/jinzhu/gorm"
    "github.com/dazhenghu/util/fileutil"
    "path/filepath"
    "os"
    "io/ioutil"
    "github.com/go-yaml/yaml"
    "github.com/dazhenghu/migrate/model"
)

type MigrateInterface interface {
    Up() (error)
    Down() (error)
}

type migrate struct {
    db *gorm.DB  // 数据库连接
    migrationDirPath string // 需执行migrate的目录
    migrateProcess MigrateInterface // 执行操作
}

func New(db *gorm.DB, migrationDirPath string) *migrate  {
    return &migrate{
        db:db,
        migrationDirPath:migrationDirPath,
    }
}

/**
初始化migrate模块自身所需资源
 */
func InitSelf(db *gorm.DB) {
    has := db.HasTable("migration_log")

    if !has {
        sql := "CREATE TABLE `migration_log` (" +
            "`version` varchar(180) NOT NULL, " +
            "`create_at` datetime DEFAULT NULL, " +
            "PRIMARY KEY (`version`)" +
            ") ENGINE=InnoDB DEFAULT CHARSET=utf8;"

        db.Exec(sql)
    }
}

/**
批量执行UP操作
 */
func (self *migrate)ExecUp() error {
    fileExists, err := fileutil.PathExists(self.migrationDirPath)
    if !fileExists {
        return err
    }

    var migrationLogs []*model.MigrationLog
    var versions []string

    self.db.Find(&migrationLogs).Pluck("version", &versions)

    execedVersions := make(map[string]string) // 主要用来借住hash判断version是否已存在，即已执行过
    for _, item := range migrationLogs {
        execedVersions[(*item).Version] = ""
    }

    err = filepath.Walk(self.migrationDirPath, func(path string, info os.FileInfo, err error) error {
        if info == nil || os.IsNotExist(err) {
            return err
        }

        if info.IsDir() {
           return nil
        }

        _, ok := execedVersions[info.Name()]

        if ok {
            // 已执行过
            return nil
        }

        migationBytes, err := ioutil.ReadFile(path)
        if err != nil {
            return err
        }

        migrationInfo := &migration{}
        err = yaml.Unmarshal(migationBytes, migrationInfo)
        if err != nil {
            return err
        }

        self.Up(migrationInfo)

        return nil
    })



    return err
}

/**
执行migration的up操作
 */
func (self *migrate)Up(migration *migration) error {
    for _, sql := range migration.UpList {
        err := self.ExecSql(sql)
        if err != nil {
            return err
        }
    }

    return nil
}

/**
执行sql
 */
func (self *migrate)ExecSql(sql string) error {

    self.db.Exec(sql)

    return nil
}