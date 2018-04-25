package migrate

import (
    "github.com/dazhenghu/util/fileutil"
    "github.com/go-yaml/yaml"
    "io/ioutil"
    "github.com/dazhenghu/util/dhutil"
    "path/filepath"
)

/**
用于构建执行migrate操作的数据结构
 */
type migration struct {
    DbIndex string `yaml:"dbindex"` // db配置key值
    UpList []string `yaml:"up,flow"`     // up操作执行的sql列表
    DownList []string `yaml:"down,flow"` // down操作执行的sql列表
}

func NewMigration() *migration {
    return &migration{
        UpList: make([]string, 0, 10),
        DownList: make([]string, 0, 10),
    }
}

/**
生成migration的文件名，时间+输入的文件名
 */
func GenerateMigrationFileName(fileName string) string {
    prefix := dhutil.CurrTimeFormat(dhutil.TIME_FORMAT_NO_SPLIT)

    // 文件名为时间戳+原文件名
    fileName = prefix + "_" + fileName
    return fileName
}

/**
生成migration的文件名，时间+输入的文件名
 */
func GenerateMigrationFile(migrationDirPath, migrationName string, upList,downList []string) error {
    isExists, err := fileutil.PathExists(migrationDirPath)
    if !isExists {
        return err
    }

    migrationInfo := &migration{
        UpList: upList,
        DownList: downList,
    }

    migrationBytes, err := yaml.Marshal(migrationInfo)
    if err != nil {
        return err
    }

    destFilePath := filepath.Join(migrationDirPath, migrationName)

    return ioutil.WriteFile(destFilePath, migrationBytes, 0666)
}
