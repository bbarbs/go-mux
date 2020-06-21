package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) InitializeDBConnection() {
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	port, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	connectionString := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), port, os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("error", err)
	}

	fmt.Println("Successfully init db connection!!!")
}

func (a *App) InitializeRoutes() {
	a.Router = mux.NewRouter()
	a.Router.HandleFunc("/", HomeHandler)
	a.Router.HandleFunc("/products", a.GetProductsHandler).Methods("GET")
	a.Router.HandleFunc("/product", a.CreateProductHandler).Methods("POST")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.GetProductHandler).Methods("GET")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.UpdateProductHandler).Methods("PUT")
	a.Router.HandleFunc("/product/{id:[0-9]+}", a.DeleteProductHandler).Methods("DELETE")
	http.ListenAndServe(":8080", a.Router)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Running!!!")
}

func (a *App) GetProductHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product id")
		return
	}

	p := Product{ID: id}
	if err := p.getProduct(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Product not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	respondWithJSON(w, http.StatusOK, p)
}

func (a *App) GetProductsHandler(w http.ResponseWriter, r *http.Request) {
	products, err := getProducts(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, products)
}

func (a *App) CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	var p Product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if err := p.createProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, p)
}

func (a *App) UpdateProductHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product Id")
		return
	}

	var p Product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	p.ID = id
	if err := p.updateProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, p)
}

func (a *App) DeleteProductHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid product id")
		return
	}

	var p Product
	p.ID = id
	if err := p.deleteProduct(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}
