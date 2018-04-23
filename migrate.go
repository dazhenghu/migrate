package migrate

import (
    "github.com/jinzhu/gorm"
    "github.com/dazhenghu/util/fileutil"
    "path/filepath"
    "os"
    "io/ioutil"
    "github.com/go-yaml/yaml"
    "github.com/dazhenghu/migrate/model"
    "fmt"
    "time"
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
    for _, item := range versions {
        execedVersions[item] = ""
    }

    err = filepath.Walk(self.migrationDirPath, func(path string, info os.FileInfo, err error) (errRet error) {
        if info == nil || os.IsNotExist(err) {
            errRet = err
            return
        }

        if info.IsDir() {
           return nil
        }

        _, ok := execedVersions[info.Name()]

        if ok {
            // 已执行过
            return nil
        }

        migationBytes, errRet := ioutil.ReadFile(path)
        if errRet != nil {
            return
        }

        migrationInfo := &migration{}
        errRet = yaml.Unmarshal(migationBytes, migrationInfo)

        if errRet != nil {
            return
        }

        // 事务处理
        self.db.Begin()
        defer func() {
            rec := recover()
            if rec != nil {
                errRet = rec.(error)
                self.db.Rollback()
                return
            }

            self.db.Commit()
        }()

        // 执行UP语句
        self.Up(migrationInfo)

        // 更新执行记录
        migrationLog := &model.MigrationLog{
            Version: info.Name(),
            CreateAt: time.Now(),
        }
        errRet = self.db.Save(migrationLog).Error
        return
    })



    return err
}

/**
执行migration的up操作
 */
func (self *migrate)Up(migration *migration) (err error) {
    for _, sql := range migration.UpList {
        fmt.Printf("exec sql:%s\n", sql)
        err = self.ExecSql(sql)
        if err != nil {
            panic(err)
            return
        }
    }

    return nil
}

/**
执行sql
 */
func (self *migrate)ExecSql(sql string) error {

    return self.db.Exec(sql).Error
}