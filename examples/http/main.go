package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-redis/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID   int    `gorm:"primary_key"`
	Name string `gorm:"not_null"`
}

var db *gorm.DB
var rdb *redis.Client

func init() {
	var err error
	db, err = gorm.Open(mysql.Open("root:123456@/test_mb?charset=utf8&parseTime=True&loc=Local"))
	if err != nil {
		fmt.Println(err)
		return
	}
	// Migrate the schema
	db.AutoMigrate(&User{})

	rdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "123456",
		DB:       0,
	})
}

func main() {
	fmt.Println("--------------------------------")
	fmt.Println(RunServer())
	fmt.Println("--------------------------------")
}

func RunServer() error {
	http.HandleFunc("/test", CallTest)
	http.HandleFunc("/example", CallExternalRequest)
	http.HandleFunc("/sql", ExtensionSQL)
	http.HandleFunc("/redis", ExtensionRedis)
	http.HandleFunc("/add", ExtensionAdd)
	http.HandleFunc("/single-file", ExtensionGetFile)
	return http.ListenAndServe(":8011", nil) //TODO: add flag...
}

func CallTest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}

func CallExternalRequest(w http.ResponseWriter, r *http.Request) {
	// call external request.
	resp, err := http.Get("https://example.org/get?name=John&age=55") // the url need mock.
	if err != nil {
		fmt.Println("err01: ", err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("err02: ", err)
	}

	fmt.Println(string(body))

	w.Write(body)
}

func ExtensionSQL(w http.ResponseWriter, r *http.Request) {
	var user User
	db.First(&user, 1) //Mock MySQL is required here
	user.Name = "John"
	db.Save(user)
}

func ExtensionRedis(w http.ResponseWriter, r *http.Request) {
	rdb.Set("Name", "John", -1)
}

// parameter test
func ExtensionAdd(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	a, _ := strconv.Atoi(query.Get("a"))
	b, _ := strconv.Atoi(query.Get("b"))
	w.Write([]byte(fmt.Sprint("sum: ", a+b)))
}

func ExtensionGetFile(w http.ResponseWriter, r *http.Request) {
	// get file
	resp, err := http.Get("https://example.org/getfile?filename=test.txt") // the url need mock.
	if err != nil {
		fmt.Println("err01: ", err)
		return
	}

	bd, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("err02: ", err)
		return
	}

	fmt.Printf("%s\n", bd)

	w.Write(bd)
}
