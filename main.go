package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type User struct {
	ID       int
	Username string
	Password string
}

type UserRequest struct {
	Username string
	Password string
}

type Album struct {
	ID     int
	Title  string
	Artist string
	Price  float32
}

type AlbumRequest struct {
	Title  string
	Artist string
	Price  float32
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome"))
}

func handleAlbums(w http.ResponseWriter, r *http.Request) {
	var albums []Album
	result := db.Table("album").Find(&albums)

	if result.Error != nil {
		w.Write([]byte("Failed to get data"))
	}

	json.NewEncoder(w).Encode(albums)
}

func handleAlbumDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var formattedId, err = strconv.Atoi(id)

	if err != nil {
		w.Write([]byte("Failed to get data"))
	}

	var album = Album{ID: formattedId}
	result := db.Table("album").First(&album)

	if result.Error != nil {
		w.Write([]byte("Failed to get data"))
	}

	json.NewEncoder(w).Encode(album)
}

func handleAlbumCreate(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)

	var albumRequest AlbumRequest
	json.Unmarshal(reqBody, &albumRequest)

	result := db.Table("album").Create(&albumRequest)

	if result.Error != nil {
		w.Write([]byte("Failed to create data"))
	}

	w.Write([]byte("Success to create data"))
}

func handleAlbumUpdate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	reqBody, _ := ioutil.ReadAll(r.Body)

	var albumRequest AlbumRequest
	json.Unmarshal(reqBody, &albumRequest)

	var formattedId, err = strconv.Atoi(id)

	if err != nil {
		w.Write([]byte("Failed to get data"))
	}

	album := Album{ID: formattedId, Title: albumRequest.Title, Artist: albumRequest.Artist, Price: albumRequest.Price}

	result := db.Table("album").Save(&album)

	if result.Error != nil {
		w.Write([]byte("Failed to update data"))
	}

	w.Write([]byte("Success to update data"))
}

func handleAlbumDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var formattedId, err = strconv.Atoi(id)

	if err != nil {
		w.Write([]byte("Failed to get data"))
	}

	album := Album{ID: formattedId}

	result := db.Table("album").Delete(&album)

	if result.Error != nil {
		w.Write([]byte("Failed to delete data"))
	}

	w.Write([]byte("Success to delete data"))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)

	var userRequest UserRequest
	json.Unmarshal(reqBody, &userRequest)

	var users []User
	result := db.Table("user").Where(&User{Username: userRequest.Username, Password: userRequest.Password}).Find(&users)

	if result.Error != nil {
		w.Write([]byte("Login failed"))
	}

	if len(users) > 0 {
		tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": userRequest.Username,
		}).SignedString([]byte("my secret"))

		if err != nil {
			w.Write([]byte("Login failed"))
		}

		w.Write([]byte(tokenString))
	} else {
		w.Write([]byte("Login failed"))
	}
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)

	var userRequest UserRequest
	json.Unmarshal(reqBody, &userRequest)

	result := db.Table("user").Create(&userRequest)

	if result.Error != nil {
		w.Write([]byte("Failed to register"))
	}

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": userRequest.Username,
	}).SignedString([]byte("my secret"))

	if err != nil {
		w.Write([]byte("Failed to register"))
	}

	w.Write([]byte(tokenString))
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Authorization")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			w.Write([]byte("Failed to get data"))
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte("my secret"), nil
	})

	if err != nil {
		w.Write([]byte("Failed to get data"))
		return
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var users []User
		result := db.Table("user").Find(&users)

		if result.Error != nil {
			w.Write([]byte("Failed to get data"))
		}

		json.NewEncoder(w).Encode(users)
	} else {
		w.Write([]byte("Failed to get data"))
	}
}

var db *gorm.DB

func main() {
	dsn := "root:roottoor@tcp(127.0.0.1:3306)/recordings?charset=utf8mb4&parseTime=True&loc=Local"

	var err error
	db, err = gorm.Open(mysql.Open(dsn))

	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", handleHome)

	router.HandleFunc("/login", handleLogin).Methods("POST")
	router.HandleFunc("/register", handleRegister).Methods("POST")
	router.HandleFunc("/users", handleUsers)

	router.HandleFunc("/albums", handleAlbumCreate).Methods("POST")
	router.HandleFunc("/albums", handleAlbums)
	router.HandleFunc("/albums/{id}", handleAlbumUpdate).Methods("PUT")
	router.HandleFunc("/albums/{id}", handleAlbumDelete).Methods("DELETE")
	router.HandleFunc("/albums/{id}", handleAlbumDetail)

	http.ListenAndServe(":3000", router)
}
