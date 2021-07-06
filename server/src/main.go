package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

/**

	 "package main" is like a hat for my code, so, that's nice, I hope that hat is black.

		Author of this fucking website is fucking motherfucker who loves to broke things... idk if this shit will ever run.
	Compile errors - 1425, seems to there's no problems.
		I'm Naru Koshin, be afraid from me and my code because your eyes will burn when you will read my code and of course, you will lost your virginity.
**/

var db *sql.DB

// creating some struct for final output
type User struct {
	Id int `json:"id"`
	Name string `json:"name"`
	Username string `json:"username"`
}

type Response struct {
	Error bool `json:"error"`
	Message string `json:"message"`
	Data User `json:"data"`
}

func getUsers(w http.ResponseWriter, r *http.Request){
	// starting my fucking sql statement
	stmt, err := db.Query("SELECT `id`, `name`, `username` FROM `users`")
	if err != nil {
		log.Fatal(err)
	}
	// ready to get out the fucking data
	var users []User
	for stmt.Next() {
		var u User 
		if err := stmt.Scan(&u.Id, &u.Name, &u.Username); err != nil {
			log.Fatal(err)
		}
		users = append(users, u)
	}
	stmt.Close()

	// setting beautiful header
	w.Header().Set("Content-Type", "application/json")

	// printing json output
	json.NewEncoder(w).Encode(users)
}

func getUserById(w http.ResponseWriter, r *http.Request){
	// id from request
	userid := mux.Vars(r)["id"]

	// preparing statement
	stmt, err := db.Prepare("SELECT `id`, `name`, `username` FROM `users` WHERE `id` = ? LIMIT 1")
	if err != nil {
		log.Fatal(err)
	}

	var user User

	// executing fucking statement
	stmt.QueryRow(userid).Scan(&user.Id, &user.Name, &user.Username)

	// setting fucking header
	w.Header().Set("Content-Type", "application/json")

	// checking if the fucking user has in database or somewhere fucked up and drinking vodka..bitch
	if user.Id != 0 {
		json.NewEncoder(w).Encode(user)
	} else {
		e := Response {
			Error: true,
			Message: "The user with a specific id not found",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(e)
	}
}

func insertUsers(w http.ResponseWriter, r *http.Request){
	// reading fucking output from json
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	// storing output to struct
	var user User
	json.Unmarshal(body, &user)
	
	// verifying fucking user data
	if user.Name != "" && user.Username != "" {
		// checking if the fucking user is already in the fucking database
		stmt, err := db.Prepare("SELECT `id` FROM `users` WHERE `name` = ? OR `username` = ? LIMIT 1")
		if err != nil {
			log.Fatal(err)
		}
		// executing statement
		var id int
		stmt.QueryRow(user.Name, user.Username).Scan(&id)
		
		// validating user
		if id != 0 {
			e := Response {
				Error: true,
				Message: "User already exists in the database, please use another name and username",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(e)
		} else {
			// preparing to insert our fucking user
			stmt, err := db.Prepare("INSERT INTO users (name, username) VALUES(?, ?)")
			if err != nil {
				log.Fatal(err)
			}
			// executing statement
			stmt.Exec(user.Name, user.Username)
			stmt.Close()

			// getting the new user id number
			stmt, err = db.Prepare("SELECT id FROM users WHERE username = ? and name = ? LIMIT 1")
			if err != nil {
				log.Fatal(err)
			}
			var uid int
			stmt.QueryRow(user.Username, user.Name).Scan(&uid)
			stmt.Close()

			// sending fucking response
			w.Header().Set("Content-Type", "application/json")
			re := Response {
				Error: false,
				Message: "User successfuly inserted in the database",
				Data: User {
					Id: uid,
					Name: user.Name,
					Username: user.Username,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(re)
		}
	} else {
		e := Response {
			Error: true,
			Message: "Please fill all the fields, missing name or username fields",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(e)
	}
}

func deleteUser(w http.ResponseWriter, r *http.Request){
	id := mux.Vars(r)["id"]
	stmt, err := db.Prepare("DELETE FROM users WHERE id = ?")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := stmt.Exec(id); err != nil {
		log.Fatal(err)
	} else {
		id, _ := strconv.Atoi(id)
		w.Header().Set("Content-Type", "application/json")
		re := Response {
			Error: false,
			Message: "User was successfuly deleted from the database",
			Data: User {
				Id: id,
			},
		}
		json.NewEncoder(w).Encode(re)
	}
}

func updateUser(w http.ResponseWriter, r *http.Request){
	// id from the request
	id := mux.Vars(r)["id"]

	// starting the sql statement and fetching data from the database
	stmt, err := db.Prepare("SELECT id, name, username FROM users WHERE id = ? LIMIT 1")
	if err != nil {
		log.Fatal(err)
	}


	// creating variables where we will store our data from the database
	var (
		Id int
		Name string
		Username string
	)

	// fetching data from the database
	stmt.QueryRow(id).Scan(&Id, &Name, &Username)

	// closing the database connection
	stmt.Close()

	// checking, If user exists
	if Id != 0 {
		// reading the raw json body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		// creating variable where we will store data from the json body
		var user User

		// storing all data from the json body
		json.Unmarshal(body, &user)

		// preparing new sql statement to update user data
		stmt, err := db.Prepare("UPDATE users SET name = ?, username = ? WHERE id = ?")
		if err != nil {
			log.Fatal(err)
		}

		// creating new variables to where we will store updated data of the user
		var (
			newName string
			newUsername string
		)

		// validating name, if name field in json is empty , then we will stay the same name what is already in the database
		if user.Name == "" {
			newName = Name
		} else {
			newName = user.Name
		}

		// validating username, if username field in json is empty , then we will stay the same username what is already in the database
		if user.Username == "" {
			newUsername = Username
		} else {
			newUsername = user.Username
		}

		// executing the sql statement
		stmt.Exec(newName, newUsername, id)

		// closing statement
		stmt.Close()

		// writing response about successfuly user update
		w.Header().Set("Content-Type", "application/json")
		res := Response {
			Error: false,
			Message: "User was successfully updated",
			Data: User {
				Id: Id,
				Name: newName,
				Username: newUsername,
			},
		}
		json.NewEncoder(w).Encode(res)
	} else {
		// user not found error
		err := Response {
			Error: true,
			Message: "User with requested id not found, please check your request and try again",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)

	}
}

func main(){
	// my fucking awesome website.

	// creating my fucking fucked up mux router
	router := mux.NewRouter()

	// getting all fucking users from the database
	router.HandleFunc("/users", getUsers).Methods("GET")

	// getting user by id
	router.HandleFunc("/user/{id}", getUserById).Methods("GET")

	// inserting new users
	router.HandleFunc("/users", insertUsers).Methods("PUT")

	// deleting user
	router.HandleFunc("/user/{id}", deleteUser).Methods("DELETE")

	// updating user
	router.HandleFunc("/user/{id}", updateUser).Methods("PATCH")

	// serving my fucking web server
	fmt.Println("Fucking server is running on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

/**
 * my little fucking web server installation
**/
func init(){
	// creating fucking sqlite3 file if that motherfucker doesn't exists
	database := "server/database/database.db"
	// checking if that fucking file exists or not
	if _, err := os.Stat(database); err != nil {
		// so... that motherfucker doesn't exist, that's why we are creating new one
		file, err := os.Create(database)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
	}
	// connection time... let's relive this motherfucker
	var err error
	db, err = sql.Open("sqlite3", database)
	if err != nil {
		log.Fatal(err)
	}
	// hora, hora, creating some tables...right?
	db.Exec("CREATE TABLE IF NOT EXISTS `users` (`id` INTEGER PRIMARY KEY AUTOINCREMENT, `name` TEXT, `username` TEXT)")
}