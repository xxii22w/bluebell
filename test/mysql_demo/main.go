package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // init()
)

var db *sql.DB

func initMysql() (err error) {
	// DSN: Data Source Name
	dsn := "root:123@tcp(127.0.0.1:3307)/mysql_demo"
	// 初始化全局的db变量
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	// 尝试与数据库连接 (校验dsn是否正确)
	err = db.Ping()
	if err != nil {
		fmt.Printf("connect to db faild,err:%v\n", err)
		return err
	}
	// 数值需要根据业务情况来确定
	db.SetConnMaxLifetime(time.Second * 10)
	db.SetMaxOpenConns(200) // 最大连接数
	db.SetMaxIdleConns(10)  // 最大空闲连接数
	return
}

type user struct {
	id   int
	age  int
	name string
}

// 查找单条数据示例
func queryRowDemo() {
	sqlStr := "select id , name , age from user where id=?"
	var u user
	// 非常重要：确保QueryRow之后调用Scan方法，否则持有数据库链接不会被释放
	err := db.QueryRow(sqlStr, 1).Scan(&u.id, &u.name, &u.age)
	if err != nil {
		fmt.Printf("scan failed,err:&v\n", err)
		return
	}
	fmt.Printf("id:%d name:%s age:%d\n", u.id, u.name, u.age)
}

// 查询多条数据
func queryMultiRowDemo() {
	sqlStr := "select id , name , age from user where id > ?"
	rows, err := db.Query(sqlStr, 0)
	if err != nil {
		fmt.Printf("query failed,err:%v\n", err)
		return
	}
	// 非常重要，关闭rows释放持有的数据库连接
	defer rows.Close()

	// 循环读取结果集中的数据
	for rows.Next() {
		var u user
		err := rows.Scan(&u.id, &u.name, &u.age)
		if err != nil {
			fmt.Printf("scan failed, err:&v\n", err)
			return
		}
		fmt.Printf("id:%d name:%s age:%d\n", u.id, u.name, u.age)
	}
}

// 插入数据
func insertRowDemo() {
	sqlStr := "insert into user(name,age) values (?,?)"
	ret, err := db.Exec(sqlStr, "王五", 38)
	if err != nil {
		fmt.Printf("insert failed, err:%v\n", err)
		return
	}
	var theID int64
	theID, err = ret.LastInsertId() // 新插入数据的id
	if err != nil {
		fmt.Printf("get lastinsert ID failed, err:%v\n", err)
		return
	}
	fmt.Printf("insert success, the id is %d.\n", theID)
}

// 更新数据
func updateRowDemo() {
	sqlStr := "update user set age=? where id = ?"
	ret, err := db.Exec(sqlStr, 16, 3)
	if err != nil {
		fmt.Printf("update failed, err:%v\n", err)
		return
	}
	var n int64
	n, err = ret.RowsAffected() // 操作影响的行数
	if err != nil {
		fmt.Printf("get RowsAffected ID failed, err:%v\n", err)
		return
	}
	fmt.Printf("update success, affected rows:%d\n", n)
}

// 删除数据
func deleteRowDemo() {
	sqlStr := "delete from user where id = ?"
	ret, err := db.Exec(sqlStr, 3)
	if err != nil {
		fmt.Printf("delete failed, err:%v\n", err)
		return
	}
	var n int64
	n, err = ret.RowsAffected() // 操作影响的行数
	if err != nil {
		fmt.Printf("get RowsAffected ID failed, err:%v\n", err)
		return
	}
	fmt.Printf("delete success, affected rows:%d\n", n)
}

// 预处理的作用是:
// 1.优化重复插入查询删除等速度 先把sql发送给服务端，再把数据发送给服务端
// 2. 防止sql注入
// 预处理查询示例
func prepareQueryDemo() {
	sqlStr := "select id, name, age from user where id > ?"
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		fmt.Printf("prepare failed ,err:%v\n", err)
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query(0)
	if err != nil {
		fmt.Printf("query failed, err:%v\n", err)
		return
	}
	defer rows.Close()
	// 循环读取结果集中的数据
	for rows.Next() {
		var u user
		err := rows.Scan(&u.id, &u.name, &u.age)
		if err != nil {
			fmt.Printf("scan failed, err:%v\n", err)
			return
		}
		fmt.Printf("id:%d name:%s age:%d\n", u.id, u.name, u.age)
	}
}

// 预处理插入示例
func prepareInsertDemo() {
	sqlStr := "insert into user(name,age) values(?,?)"
	stmt, err := db.Prepare(sqlStr)
	if err != nil {
		fmt.Printf("prepare failed ,err:%v\n", err)
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec("huhu", 43)
	if err != nil {
		fmt.Printf("insert failed, err:%v\n", err)
		return
	}
	_, err = stmt.Exec("xixi", 24)
	if err != nil {
		fmt.Printf("insert failed, err:%v\n", err)
		return
	}
	fmt.Printf("insert success.")
}

// 永远不要自己拼接sql语句
func sqlInjectDemo(name string) {
	sqlStr := fmt.Sprintf("select id,name,age from user where name='%s'", name)
	fmt.Printf("SQL:%s\n", sqlStr)
	var u user
	err := db.QueryRow(sqlStr).Scan(&u.id, &u.name, &u.age)
	if err != nil {
		fmt.Printf("exec failed, err:%v\n", err)
		return
	}
	fmt.Printf("user:%#v\n", u)
}

// 事务操作示例
func transactionDemo() {
	tx, err := db.Begin() // 开启事务
	if err != nil {
		if tx != nil {
			tx.Rollback() // 回滚
		}
		fmt.Printf("begin trans failed, err:%v\n", err)
		return
	}
	sqlStr := "update user set age = 30 where id = ?"
	ret1, err := tx.Exec(sqlStr, 2)
	if err != nil {
		tx.Rollback() // 回滚
		fmt.Printf("exec sql1 failed, err:%v\n", err)
		return
	}
	affRow1, err := ret1.RowsAffected()
	if err != nil {
		tx.Rollback() // 回滚
		fmt.Printf("exec ret1.RowsAffected() failed, err:%v\n", err)
		return
	}

	sqlStr2 := "update user set age = 40 where id = ?"
	ret2, err := tx.Exec(sqlStr2, 3)
	if err != nil {
		tx.Rollback() // 回滚
		fmt.Printf("exec sql2 failed, err:%v\n", err)
		return
	}
	affRow2, err := ret2.RowsAffected()
	if err != nil {
		tx.Rollback() // 回滚
		fmt.Printf("exec ret2.RowsAffected() failed, err:%v\n", err)
		return
	}

	// 当affRow1 == 1 && affRow2 == 1
	if affRow1 == 1 && affRow2 == 1 {
		fmt.Println("事务提交...")

		tx.Commit() // 提交事务
	} else {
		tx.Rollback()
		fmt.Println("事务回滚...")
	}

	fmt.Printf("ecex trans success!")
}

func main() {
	if err := initMysql(); err != nil {
		fmt.Printf("connect to db faild,err:%v\n", err)
	}
	// 做完错误检查之后，确保db不为nil
	defer db.Close()
	fmt.Println("connect to db success")
	// db.xx()  // 去使用数据库操作
	// queryRowDemo()
	// queryMultiRowDemo()
	// insertRowDemo()
	// updateRowDemo()
	// deleteRowDemo()
	// prepareQueryDemo()
	// prepareInsertDemo()
	// sqlInjectDemo("xxx' or 1=1#")	// 这样是可以搜索出来的，千万不要通过拼接来使用sql
	transactionDemo()
}
