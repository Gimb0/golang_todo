package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

// Task Data Structure
type Task struct {
	ID          int
	Name        string
	Description string
}

var templates = template.Must(template.ParseGlob("./templates/*"))

func initialiseDB() {
	database, err := sql.Open("sqlite3", "./todotasks.db")
	if os.IsExist(err) {
		return
	}
	statement, err := database.Prepare("CREATE TABLE IF NOT EXISTS tasks (id INTEGER PRIMARY KEY, name TEXT, description TEXT)")
	if err != nil {
		log.Fatal(err)
	}
	statement.Exec()
}

func getAllTasks() (*sql.Rows, error) {
	database, err := sql.Open("sqlite3", "./todotasks.db")
	if err != nil {
		return nil, err
	}
	rows, err := database.Query("SELECT id, name, description FROM tasks")
	if err != nil {
		return rows, err
	}
	return rows, nil
}

func getTaskFromDB(taskID int) (*sql.Rows, error) {
	database, err := sql.Open("sqlite3", "./todotasks.db")
	if err != nil {
		log.Fatal(err)
	}
	rows, err := database.Query("SELECT id, name, description FROM tasks WHERE id=?", taskID)
	if err != nil {
		return rows, err
	}
	return rows, nil
}

func renderToDoTemplate(w http.ResponseWriter, tmpl string, t Task) {
	err := templates.ExecuteTemplate(w, tmpl+".html", t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	operation := r.URL.Path[1:5]
	taskID := r.URL.Path[6:]

	taskIDInt, err := strconv.Atoi(taskID)
	if err != nil {
		log.Fatal(err)
	}
	row, err := getTaskFromDB(taskIDInt)
	if err != nil {
		log.Fatal(err)
	}

	var results Task
	for row.Next() {
		err = row.Scan(&results.ID, &results.Name, &results.Description)
		if err != nil {
			log.Fatal(err)
		}
	}
	if operation == "view" {
		renderToDoTemplate(w, "view", results)
	} else if operation == "edit" {
		renderToDoTemplate(w, "edit", results)
	} else {
		http.Redirect(w, r, "/", 500)
	}

}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	database, err := sql.Open("sqlite3", "./todotasks.db")
	if err != nil {
		log.Print(err)
	}

	defer database.Close()

	var output string

	statement, err := database.Prepare("SELECT COUNT(*) FROM tasks WHERE id=?")
	if err != nil {
		log.Fatal(err)
	}

	defer statement.Close()

	taskID := r.URL.Path[6:]
	taskName := r.FormValue("name")
	taskDesc := r.FormValue("description")

	err = statement.QueryRow(taskID).Scan(&output)
	switch {
	case err == sql.ErrNoRows:
		statement, err := database.Prepare("INSERT INTO tasks (id, name, description) VALUES (?, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		statement.Exec(taskID, taskName, taskDesc)
	case err != nil:
		log.Print("%s\n", err)
	default:
		statement, err := database.Prepare("UPDATE tasks SET id=?, name=?, description=? WHERE id=?")
		if err != nil {
			log.Fatal(err)
		}
		statement.Exec(taskID, taskName, taskDesc, taskID)
	}

	http.Redirect(w, r, "/view/"+taskID, 302)
}

func styleHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/styles.css")
}

func mainIndexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/index.html")
}

func main() {
	initialiseDB()
	http.HandleFunc("/", mainIndexHandler)
	http.HandleFunc("/view/", taskHandler)
	http.HandleFunc("/edit/", taskHandler)
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/styles/", styleHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
