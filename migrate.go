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
    "strings"
    "errors"
)

var defaultDbIndex string = "db"

type MigrateInterface interface {
    Up() (error)
    Down() (error)
}

type DbConf struct {
    Type string  // 数据库类型，如：mysql
    Dsn  string  // 数据库连接
}


type migrate struct {
    db *gorm.DB  // 数据库连接
    migrationDirPath string // 需执行migrate的目录
    migrateProcess MigrateInterface // 执行操作
    dbConfMap map[string]*DbConf // 数据库连接配置map，key:配置名，DbConf:数据库连接配置
    dbConnBuff map[string]*gorm.DB // 数据库连接
}

func New(migrationDirPath string, dbconfigMap map[string]*DbConf) *migrate  {
    if dbconfigMap == nil {
        dbconfigMap = make(map[string]*DbConf)
    }

    var defaultDb *gorm.DB
    for key, conf := range dbconfigMap {
        if key == defaultDbIndex {
            db, err := gorm.Open(conf.Type, conf.Dsn)
            if err != nil {
                panic(err)
            }
            defaultDb = db
        }
    }

    obj := &migrate{
        db:defaultDb,
        migrationDirPath:migrationDirPath,
        dbConfMap: dbconfigMap,
        dbConnBuff: make(map[string]*gorm.DB),
    }
    if defaultDb != nil {
        // 将默认链接加到缓存中
        obj.dbConnBuff[defaultDbIndex] = defaultDb
    }
    return obj
}

/**
初始化migrate模块自身所需资源
 */
func (self *migrate)InitSelf() {
    has := self.db.HasTable("migration_log")

    if !has {
        sql := "CREATE TABLE `migration_log` (" +
            "`version` varchar(180) NOT NULL, " +
            "`create_at` datetime DEFAULT NULL, " +
            "PRIMARY KEY (`version`)" +
            ") ENGINE=InnoDB DEFAULT CHARSET=utf8;"

        self.db.Exec(sql)
    }
}

/**
添加db配置到confmap中
 */
func (self *migrate)PushDbConf(confName string, conf *DbConf)  {
    self.dbConfMap[confName] = conf
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

        if migrationInfo.DbIndex == "" {
            migrationInfo.DbIndex = defaultDbIndex
        }

        // 先从缓存中获取数据库连接，没有的话再建立链接
        dbConn, ok := self.dbConnBuff[migrationInfo.DbIndex]
        if !ok {
            dbconf, confOk := self.dbConfMap[migrationInfo.DbIndex]
            if !confOk {
                errRet = errors.New(fmt.Sprintf("migration err, undefined dbconf index:%s", migrationInfo.DbIndex))
                return
            }

            dbConn, errRet = gorm.Open("mysql", dbconf.Dsn)
            if errRet != nil {
                return
            }
            self.dbConnBuff[migrationInfo.DbIndex] = dbConn
        }


        // 事务处理
        dbConn.Begin()
        //defer func(conn *gorm.DB) {
        //    rec := recover()
        //    if rec != nil {
        //        errRet = rec.(error)
        //        fmt.Printf("sql err:%+v\n", errRet)
        //        conn.Rollback()
        //        return
        //    }
        //
        //    conn.Commit()
        //}(dbConn)

        // 执行UP语句
        errRet = self.Up(dbConn, migrationInfo)

        if errRet != nil {
            fmt.Printf("sql err:%+v\n", errRet)
            dbConn.Rollback()
            return
        }

        dbConn.Commit()

        // 更新执行记录
        migrationLog := &model.MigrationLog{
            Version: info.Name(),
            CreateAt: time.Now(),
        }
        errRet = self.db.Save(migrationLog).Error
        return
    })

    defer func() {
        for _, dbConn := range self.dbConnBuff  {
            if dbConn != nil {
                dbConn.Close()
            }
        }
    }()

    return err
}

/**
执行migration的up操作
 */
func (self *migrate)Up(conn *gorm.DB, migration *migration) (err error) {
    for _, sql := range migration.UpList {
        fmt.Printf("exec sql:%s\n", sql)
        err = self.ExecSql(conn, sql)
        if err != nil {
            return err
        }
    }

    return nil
}

/**
执行sql
 */
func (self *migrate)ExecSql(conn *gorm.DB, sql string) error {

    return conn.Exec(sql).Error
}


/**
控制台命令生成migration文件
 */
func CreateMigrationFile(migrationDirPath string)  {
    fmt.Println("Please input file name:")
    var fileName string
    fmt.Scanln(&fileName)

    fileName = GenerateMigrationFileName(fileName) + ".yaml" // 增加前缀

    fmt.Println("Create new migration:" + fileName + " (y|n)")

    var yn string
    fmt.Scanln(&yn)

    yn = strings.ToUpper(yn)

    if yn == "Y" {
        // 创建文件
        err := GenerateMigrationFile(migrationDirPath, fileName, []string{}, []string{})
        fmt.Printf("err:%+v", err)
    }
}