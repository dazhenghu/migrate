基于gorm、yaml的数据库迁移组件
===

#说明：

>基于gorm、yaml的数据库迁移组件，目前还在不断完善中，目前只支持mysql，后续会按需进行完善。之所以编写这个项目，主要是因为gorm提供的数据迁移
不太符合我的项目需求，我们想要的是可以显示记录数据库修改记录的流水信息。


#示例：

```
// 初始化数据库链接，执行sql的时候会用到
db, err := gorm.Open("mysql", "root:qsqfrms@tcp(127.0.0.1:3306)/test?charset=utf8&parseTime=True&loc=Local")
// 释放链接
defer db.Close()
// 初始化，如果没有migration_log表，将会创建
migrate.InitSelf(db)
// 创建执行migrate操作的对象，第二个参数显示指定migration文件所在路径
migrateObj := migrate.New(db, "./migration")
// 此句是执行migration文件中up所指定的sql
err = migrateObj.ExecUp()
```





