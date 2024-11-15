package main

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql" // init()
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func initDB() (err error) {
	// DSN: Data Source Name
	dsn := "root:123@tcp(127.0.0.1:3307)/mysql_demo?charset=utf8mb4&parseTime=True"
	// 也可以使用MustConnection连接不成功就panic
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		fmt.Printf("connect DB failed, err:%v\n", err)
		return
	}
	// 数值需要根据业务情况来确定
	db.SetMaxOpenConns(10) // 最大连接数
	db.SetMaxIdleConns(10) // 最大空闲连接数
	return
}

type user struct {
	Id   int    `db:"id"`
	Age  int    `db:"age"`
	Name string `db:"name"`
}

func (u user) Value() (driver.Value, error) {
	return []interface{}{u.Name, u.Age}, nil
}

// 查询单条数据示例
func queryRowDemo() {
	sqlStr := "select id , name , age from user where id=?"
	var u user
	err := db.Get(&u, sqlStr, 1)
	if err != nil {
		fmt.Printf("get failed, err:%v\n", err)
		return
	}
	fmt.Printf("id:%d name:%s age:%d\n", u.Id, u.Name, u.Age)
}

// 查询多条
func queryMultiRowDemo() {
	sqlStr := "select id,name,age from user where id > ?"
	var Users []user
	err := db.Select(&Users, sqlStr, 0)
	if err != nil {
		fmt.Printf("query failed, err:%v\n", err)
		return
	}
	fmt.Printf("user:%#v\n", Users)
}

// 插入数据
func insertRowDemo() {
	sqlStr := "insert into user(name,age) values (?,?)"
	ret, err := db.Exec(sqlStr, "王五", 38)
	if err != nil {
		fmt.Printf("insert failed, err:%v\n", err)
		return
	}

	theID, err := ret.LastInsertId() // 新插入数据的id
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

	n, err := ret.RowsAffected() // 操作影响的行数
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

	n, err := ret.RowsAffected() // 操作影响的行数
	if err != nil {
		fmt.Printf("get RowsAffected ID failed, err:%v\n", err)
		return
	}
	fmt.Printf("delete success, affected rows:%d\n", n)
}

func insertUserDemo() (err error) {
	_, err = db.NamedExec(`insert into user(name,age) values(:name,:age)`,
		map[string]interface{}{
			"name": "吃大亏",
			"age":  56,
		})
	return
}

func namedQuery() {
	sqlStr := `select * from user where name=:name`
	// 使用map做命名查询
	rows, err := db.NamedQuery(sqlStr, map[string]interface{}{"name": "hhh"})
	if err != nil {
		fmt.Printf("db.NamedQuery failed,err:%v\n", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var u user
		rows.StructScan(&u)
		fmt.Printf("user:%#v\n", u)
	}

	u := user{
		Name: "hhh",
	}
	// 使用结构体命名查询，根据结构体字段的db tag进行映射
	rows, err = db.NamedQuery(sqlStr, u)
	if err != nil {
		fmt.Printf("db.NamedQuery failed,err:%v\n", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var u user
		rows.StructScan(&u)
		fmt.Printf("user:%#v\n", u)
	}
}

func transactionDemo2()(err error) {
	tx, err := db.Beginx() // 开启事务
	if err != nil {
		fmt.Printf("begin trans failed, err:%v\n", err)
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			fmt.Println("rollback")
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err
			fmt.Println("commit")
		}
	}()

	sqlStr1 := "Update user set age=20 where id=?"

	rs, err := tx.Exec(sqlStr1, 1)
	if err!= nil{
		return err
	}
	n, err := rs.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("exec sqlStr1 failed")
	}
	sqlStr2 := "Update user set age=50 where i=?"
	rs, err = tx.Exec(sqlStr2, 5)
	if err!=nil{
		return err
	}
	n, err = rs.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("exec sqlStr1 failed")
	}
	return err
}

// BatchInsertUsers2 使用sqlx.In帮我们拼接语句和参数, 注意传入的参数是[]interface{}
func BatchInsertUsers2(users []interface{}) error {
	query, args, _ := sqlx.In(
		"INSERT INTO user (name, age) VALUES (?), (?), (?)",
		users..., // 如果arg实现了 driver.Valuer, sqlx.In 会通过调用 Value()来展开它
	)
	fmt.Println(query) // 查看生成的querystring
	fmt.Println(args)  // 查看生成的args
	_, err := db.Exec(query, args...) 
	return err
}

// BatchInsertUsers3 使用NamedExec实现批量插入
func BatchInsertUsers3(users []*user) error {
	_, err := db.NamedExec("INSERT INTO user (name, age) VALUES (:name, :age)", users)
	return err
}

// QueryByIDs 根据给定ID查询
func QueryByIDs(ids []int)(users []user, err error){
	// 动态填充id
	query, args, err := sqlx.In("SELECT name, age FROM user WHERE id IN (?)", ids)
	if err != nil {
		return
	}
	// sqlx.In 返回带 `?` bindvar的查询语句, 我们使用Rebind()重新绑定。
	// 重新生成对应数据库的查询语句（如PostgreSQL 用 `$1`, `$2` bindvar）
	query = db.Rebind(query)

	err = db.Select(&users, query, args...)
	return
}

// QueryAndOrderByIDs 按照指定id查询并维护顺序
func QueryAndOrderByIDs(ids []int)(users []user, err error){
	// 动态填充id
	strIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		strIDs = append(strIDs, fmt.Sprintf("%d", id))
	}
	query, args, err := sqlx.In("SELECT name, age FROM user WHERE id IN (?) ORDER BY FIND_IN_SET(id, ?)", ids, strings.Join(strIDs, ","))
	if err != nil {
		return
	}

	// sqlx.In 返回带 `?` bindvar的查询语句, 我们使用Rebind()重新绑定它
	query = db.Rebind(query)

	err = db.Select(&users, query, args...)
	return
}

func main() {
	if err := initDB(); err != nil {
		fmt.Printf("init DB failed, err:%v\n", err)
		return
	}
	fmt.Println("init DB success...")

	// queryRowDemo()
	// queryMultiRowDemo()
	// insertUserDemo()
	// namedQuery()
	// transactionDemo2()

	// u1 := user{Name: "xx", Age: 18}
	// u2 := user{Name: "xxx", Age: 28}
	// u3 := user{Name: "xxxx", Age: 38}

	// users := []interface{}{u1, u2, u3}
	// BatchInsertUsers2(users)

	// users3 := []*user{&u1, &u2, &u3}
	// err := BatchInsertUsers3(users3)
	// if err != nil {
	// 	fmt.Printf("BatchInsertUsers3 failed, err:%v\n", err)
	// }

	users,err := QueryByIDs([]int{1,6,5,4})
	if err != nil {
		fmt.Printf("QueryByIDs failed, err:%v\n",err)
		return
	}
	for _,user := range users {
		fmt.Printf("user:%#v\n",user)
	}

	fmt.Println("-----------------")
	users,err = QueryAndOrderByIDs([]int{4,5,6,1})
	if err != nil {
		fmt.Printf("QueryByIDs failed, err:%v\n",err)
		return
	}
	for _,user := range users {
		fmt.Printf("user:%#v\n",user)
	}
}
