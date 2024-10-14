# go-sql-orm


使用方法：
```go
package main

import (
	"fmt"
	"time"

	orm "github.com/FrankLeeC/go-sql-orm"
)

/*
create database test CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

use test;

create table if not exists person(
id bigint primary key auto_increment,
user_name varchar(20),
user_age int,
birth_date datetime,
create_time datetime,
update_time datetime
);
*/

func init() {
	dsName := "default"
	ds := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/test", "", "")
	orm.RegisterDatsource(orm.NewDatasourceConfig(dsName, ds).MaxConn(10).MaxIdleConn(3))
}

type Person struct {
	ID         int64     `column:"id"`
	Name       string    `column:"user_name"`
	Age        int       `column:"user_age"`
	BirthDate  time.Time `column:"birth_date"`
	CreateTime time.Time `column:"create_time"`
	UpdateTime time.Time `column:"update_time"`
}

type CountResult struct {
	Count int `column:"cnt"`
}

type AgeRangeResult struct {
	AgeRange string `column:"age_range"`
	Count    int    `column:"cnt"`
}

type TableHandler struct {
	tb            string
	fullCols      []string
	colsForInsert []string
}

func (a *TableHandler) Init() {
	a.tb = "person"
	demo := Person{}
	a.fullCols = orm.ColumnsExcept(demo)
	a.colsForInsert = orm.ColumnsExcept(demo, "id")
}

func (a *TableHandler) Insert(data ...*Person) (int64, error) {
	return orm.CreateContext().Insert(a.tb, a.colsForInsert, data).Exec()
}

func (a *TableHandler) SelectById(id int64) (Person, error) {
	var r Person
	err := orm.CreateContext().Select(a.tb, a.fullCols, "id=?", id).Result(&r)
	return r, err
}

func (a *TableHandler) UpdateById(cols []string, data *Person) (int64, error) {
	return orm.CreateContext().Update(a.tb, cols, "id=?").ReflectParamsFrom(data, []string{"id"}).Exec()
}

func (a *TableHandler) Delete(condition string, params ...interface{}) (int64, error) {
	return orm.CreateContext().Delete(a.tb, condition).Params(params...).Exec()
}

func (a *TableHandler) Page(condition string, params []interface{}, offset, limit int) (int, []Person, error) {
	var cnt CountResult
	var rs []Person
	err := orm.CreateContext().Select(a.tb, []string{"count(1) as cnt"}, condition, params...).Result(&cnt)
	if err != nil {
		return 0, nil, err
	}
	err = orm.CreateContext().Select(a.tb, a.fullCols, condition, params...).OrderByDesc("user_age").Limit(offset, limit).Result(&rs)
	if err != nil {
		return 0, nil, err
	}
	return cnt.Count, rs, nil
}

func (a *TableHandler) Select(condition string, params ...interface{}) ([]Person, error) {
	var ps []Person
	err := orm.CreateContext().Select(a.tb, a.fullCols, condition, params...).Result(&ps)
	if err != nil {
		return nil, err
	}
	return ps, nil
}

func (a *TableHandler) Search() ([]AgeRangeResult, error) {
	sql := "select age_range, count(1) as cnt from (select case when user_age <=20 then '0-20' when user_age <= 25 then '21-25' when user_age <= 30 then '26-30' else '30~' end as age_range, id from person) t group by age_range"
	var r []AgeRangeResult
	err := orm.CreateContext().Search(sql).Result(&r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (a *TableHandler) Transaction() error {
	tx, err := orm.CreateContext().Begin()
	if err != nil {
		return err
	}
	defer func() {
		err := recover()
		if err != nil {
		} else {
			tx.Commit()
		}
	}()
	_, err = tx.Update(a.tb, []string{"user_name", "update_time"}, "user_name = ?").Params("p33", time.Now(), "p3").Exec()
	if err != nil {
		tx.Rollback()
		return err
	}
	t6, _ := time.ParseInLocation("2006-01-02", "1996-03-19", time.Local)
	p := Person{Name: "p6", Age: 28, BirthDate: t6, CreateTime: time.Now(), UpdateTime: time.Now()}
	_, err = tx.Insert(a.tb, a.colsForInsert, p).Exec()
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	fmt.Println("commit success")
	return nil
}

func main() {
	h := TableHandler{}
	h.Init()
	now := time.Now()
	t1, _ := time.ParseInLocation("2006-01-02", "2006-05-06", time.Local)
	t2, _ := time.ParseInLocation("2006-01-02", "2004-02-16", time.Local)
	t3, _ := time.ParseInLocation("2006-01-02", "1999-10-21", time.Local)
	t4, _ := time.ParseInLocation("2006-01-02", "1998-06-21", time.Local)
	t5, _ := time.ParseInLocation("2006-01-02", "1997-03-05", time.Local)
	ps := []*Person{
		&Person{Name: "p1", Age: 18, BirthDate: t1, CreateTime: now, UpdateTime: now},
		&Person{Name: "p2", Age: 20, BirthDate: t2, CreateTime: now, UpdateTime: now},
		&Person{Name: "p3", Age: 25, BirthDate: t3, CreateTime: now, UpdateTime: now},
		&Person{Name: "p4", Age: 26, BirthDate: t4, CreateTime: now, UpdateTime: now},
		&Person{Name: "p5", Age: 27, BirthDate: t5, CreateTime: now, UpdateTime: now},
	}
	fmt.Println("----------- insert -----------")
	t, err := h.Insert(ps...)
	if err != nil {
		panic(err)
	}
	fmt.Println("affect rows:", t)
	fmt.Println("----------- insert -----------")

	fmt.Println("----------- select -----------")
	rs, err := h.Select("birth_date >= ?", "2005-01-01")
	if err != nil {
		panic(err)
	}
	for _, r := range rs {
		fmt.Println(r)
	}
	fmt.Println("----------- select -----------")

	fmt.Println("----------- update -----------")
	fmt.Println("sleep 3 seconds")
	time.Sleep(time.Second * 3)
	rs[0].Name = "p111"
	rs[0].UpdateTime = time.Now()
	affected, err := h.UpdateById([]string{"user_name", "update_time"}, &rs[0])
	if err != nil {
		panic(err)
	}
	fmt.Println("affected rows:", affected)
	fmt.Println("----------- update -----------")

	fmt.Println("----------- select by id -----------")
	id := rs[0].ID
	p, err := h.SelectById(id)
	if err != nil {
		panic(err)
	}
	fmt.Println(p)
	fmt.Println("----------- select by id -----------")

	fmt.Println("----------- delete -----------")
	affected, err = h.Delete("id = ?", rs[0].ID)
	if err != nil {
		panic(err)
	}
	fmt.Println("affected rows:", affected)
	fmt.Println("----------- delete -----------")

	fmt.Println("----------- page -----------")
	c, page, err := h.Page("birth_date >= ?", []interface{}{"1998-01-01"}, 2, 1)
	if err != nil {
		panic(err)
	}
	fmt.Println("total count:", c)
	fmt.Println("page:", page)
	fmt.Println("----------- page -----------")

	fmt.Println("----------- customize search -----------")
	r, err := h.Search()
	if err != nil {
		panic(err)
	}
	fmt.Println("age range result:", r)
	fmt.Println("----------- customize search -----------")

	fmt.Println("----------- transaction -----------")
	h.Transaction()
	fmt.Println("----------- transaction -----------")

	fmt.Println("----------- select -----------")
	rs, err = h.Select("")
	if err != nil {
		panic(err)
	}
	for _, r := range rs {
		fmt.Println(r)
	}
	fmt.Println("----------- select -----------")

}

```

日志：

```log
----------- insert -----------
affect rows: 5
----------- insert -----------
----------- select -----------
{1 p1 18 2006-05-06 00:00:00 +0800 CST 2024-10-14 13:47:51 +0800 CST 2024-10-14 13:47:51 +0800 CST}
----------- select -----------
----------- update -----------
sleep 3 seconds
affected rows: 1
----------- update -----------
----------- select by id -----------
{1 p111 18 2006-05-06 00:00:00 +0800 CST 2024-10-14 13:47:51 +0800 CST 2024-10-14 13:47:54 +0800 CST}
----------- select by id -----------
----------- delete -----------
affected rows: 1
----------- delete -----------
----------- page -----------
total count: 3
page: [{2 p2 20 2004-02-16 00:00:00 +0800 CST 2024-10-14 13:47:51 +0800 CST 2024-10-14 13:47:51 +0800 CST}]
----------- page -----------
----------- customize search -----------
age range result: [{0-20 1} {21-25 1} {26-30 2}]
----------- customize search -----------
----------- transaction -----------
commit success
----------- transaction -----------
----------- select -----------
{2 p2 20 2004-02-16 00:00:00 +0800 CST 2024-10-14 13:47:51 +0800 CST 2024-10-14 13:47:51 +0800 CST}
{3 p33 25 1999-10-21 00:00:00 +0800 CST 2024-10-14 13:47:51 +0800 CST 2024-10-14 13:47:54 +0800 CST}
{4 p4 26 1998-06-21 00:00:00 +0800 CST 2024-10-14 13:47:51 +0800 CST 2024-10-14 13:47:51 +0800 CST}
{5 p5 27 1997-03-05 00:00:00 +0800 CST 2024-10-14 13:47:51 +0800 CST 2024-10-14 13:47:51 +0800 CST}
{6 p6 28 1996-03-19 00:00:00 +0800 CST 2024-10-14 13:47:54 +0800 CST 2024-10-14 13:47:54 +0800 CST}
----------- select -----------

```