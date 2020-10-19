package main

import (
	"database/sql"
	"log"
	"net/http"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

type Image struct {
	Id         int
	Name       string
	ImageBytes string
	Flag       int
}

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "root"
	dbName := "images"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

var tmpl = template.Must(template.ParseGlob("form/*"))

func Index(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	selDB, err := db.Query("SELECT * FROM image ORDER BY id DESC")
	if err != nil {
		panic(err.Error())
	}
	image := Image{}
	res := []Image{}
	for selDB.Next() {
		var id int
		var name string
		buf := make([]byte, 65535)
		var flag int
		err = selDB.Scan(&id, &name, &buf, &flag)
		if err != nil {
			panic(err.Error())
		}
		image.Id = id
		image.Name = name
		image.ImageBytes = string(buf)
		image.Flag = flag
		res = append(res, image)
	}
	tmpl.ExecuteTemplate(w, "Index", res)
	defer db.Close()
}

func Show(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	nId := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT * FROM image WHERE id=?", nId)
	if err != nil {
		panic(err.Error())
	}
	image := Image{}
	for selDB.Next() {
		var id int
		var name string
		buf := make([]byte, 65535)
		var flag int
		err = selDB.Scan(&id, &name, &buf, &flag)
		if err != nil {
			panic(err.Error())
		}
		image.Id = id
		image.Name = name
		image.ImageBytes = string(buf)
		image.Flag = flag
	}
	tmpl.ExecuteTemplate(w, "Show", image)
	defer db.Close()
}

func New(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "New", nil)
}

func Edit(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	nId := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT * FROM image WHERE id=?", nId)
	if err != nil {
		log.Println(err)
		panic(err.Error())
	}
	image := Image{}
	for selDB.Next() {
		var id int
		var name string
		buf := make([]byte, 65535)
		var flag int
		err = selDB.Scan(&id, &name, &buf, &flag)
		if err != nil {
			panic(err.Error())
		}
		image.Id = id
		image.Name = name
		image.ImageBytes = string(buf)
		image.Flag = flag
	}
	tmpl.ExecuteTemplate(w, "Edit", image)
	defer db.Close()
}

func Insert(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		id := 0
		name := r.FormValue("name")
		image := r.FormValue("image")
		flag := r.FormValue("flag")
		insForm, err := db.Prepare("INSERT INTO image(id, name, image, flag) VALUES(?,?,?,?)")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(id, name, image, flag)
		log.Println("INSERT: Name: " + name + " | image: " + image)
	}
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

func Update(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		name := r.FormValue("name")
		image := r.FormValue("image")
		id := r.FormValue("uid")
		flag := r.FormValue("flag")
		insForm, err := db.Prepare("UPDATE image SET name=?, image=? , flag=? WHERE id=?")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(name, image, flag, id)
		log.Println("UPDATE: Name: " + name + " | image: " + image)
	}
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	image := r.URL.Query().Get("id")
	delForm, err := db.Prepare("DELETE FROM image WHERE id=?")
	if err != nil {
		panic(err.Error())
	}
	delForm.Exec(image)
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

func main() {
	log.Println("Server started on: http://localhost:8444")
	http.HandleFunc("/", Index)
	http.HandleFunc("/show", Show)
	http.HandleFunc("/new", New)
	http.HandleFunc("/edit", Edit)
	http.HandleFunc("/insert", Insert)
	http.HandleFunc("/update", Update)
	http.HandleFunc("/delete", Delete)
	http.ListenAndServe(":8444", nil)
	log.Println("END")
}
