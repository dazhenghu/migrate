package migrate

import (
    "fmt"
    "strings"
)

/**
控制台命令生成migration文件
 */
func CreateMigrationFile()  {
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
        err := GenerateMigrationFile("./migration", fileName, []string{}, []string{})
        fmt.Printf("err:%+v", err)
    }
}