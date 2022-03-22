package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db, _ = gorm.Open("mysql", "root:root@/todolist?charset=utf8&parseTime=True&loc=Local")

type TodoItemModel struct {
	Id          int `gorm:"primary_key`
	Description string
	Completed   bool
}

func UpdateTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	err := GetItemByID(id)
	if err == false {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": false, "error": "Record Not Found"}`)
		return
	}

	completed, _ := strconv.ParseBool(r.FormValue("completed"))

	todo := &TodoItemModel{}
	db.First(&todo, id)
	todo.Completed = completed
	db.Save(&todo)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"updated": true}`)
	return
}

func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	err := GetItemByID(id)
	if err == false {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted": false, "error":"Record Not Found"}`)
		return
	}

	todo := &TodoItemModel{}
	db.First(&todo, id)
	db.Delete(&todo)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"deleted": true}`)

}

func GetCompletedTodos(w http.ResponseWriter, r *http.Request) {
	completedTodos := GetTodoItems(true)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(completedTodos)
}

func CreateTodo(w http.ResponseWriter, r *http.Request) {
	description := r.FormValue("description")
	todo := &TodoItemModel{Description: description, Completed: false}
	db.Create(&todo)
	result := db.Last(&todo)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result.Value)
}

func GetTodoItems(completed bool) interface{} {
	var todos []TodoItemModel
	TodoItems := db.Where("completed = ?", completed).Find(&todos).Value
	return TodoItems
}

func Health(w http.ResponseWriter, r *http.Request) {
	fmt.Println("API seems to be ok")
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}

func GetItemByID(Id int) bool {
	todo := &TodoItemModel{}
	result := db.First(&todo, Id)
	if result.Error != nil {
		fmt.Println("Todo not found")
		return false
	}

	return true
}

func GetIncompleteTodos(w http.ResponseWriter, r *http.Request) {
	IncompleteTodoItems := GetTodoItems(false)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IncompleteTodoItems)
}

func main() {
	defer db.Close()

	// db.Debug().DropTableIfExists(&TodoItemModel{})
	// db.Debug().AutoMigrate(&TodoItemModel{})

	fmt.Println("App running on port 8000")
	router := mux.NewRouter()
	router.HandleFunc("/health", Health).Methods("GET")
	router.HandleFunc("/todo-completed", GetCompletedTodos).Methods("GET")
	router.HandleFunc("/todo-incomplete", GetIncompleteTodos).Methods("GET")
	router.HandleFunc("/todo", CreateTodo).Methods("POST")
	router.HandleFunc("/todo/{id}", UpdateTodo).Methods("POST")
	router.HandleFunc("/todo/{id}", DeleteTodo).Methods("DELETE")
	http.ListenAndServe(":8000", router)
}
