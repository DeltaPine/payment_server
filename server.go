// Server.go a marshalling interface for Gorilla router, dispatcher
// and integrator for the backing MongoDB database.
package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"log"
	"net/http"
)

// Server consists of a Dispatcher, a database session and a database
// object.
type Server struct {
	Dispatch *mux.Router
	Session  *mgo.Session
	DB       *mgo.Database
}

// COLLECTION the name of the document
var COLLECTION string

// InitializeDB takes three parameters: host, dbname and
// collection. It initializes the database driver and starts the web
// server and dispatcher. Please note that the backing database should
// be already started outside of this program, The host string is
// defined in the standard format of address:port
// (i.e. localhost:8080) and is where the web server will listen for
// incoming connections.
func (server *Server) InitializeDB(host string, dbname string, collection string) {
	if host == "" || dbname == "" || collection == "" {
		log.Fatal("You must specify a valid host, database name and collection")
	}

	session, err := mgo.Dial(host)
	if err != nil {
		log.Fatal(err)
	}

	session.SetMode(mgo.Monotonic, true)
	COLLECTION = collection
	server.Session = session
	server.DB = session.DB(dbname)
	server.Dispatch = mux.NewRouter()
	server.initializeRoutes()
}

// initializeRoutes is a dispatcher for the various RESTFUL methods of
// input and output for the web server. It sets up the
// payment/payments URL and defines GET, POST, PUT and DELETE for the
// payment URL and a GET for the payments URL.
func (server *Server) initializeRoutes() {
	server.Dispatch.HandleFunc("/payments",
		server.getPayments).Methods("GET")
	server.Dispatch.HandleFunc("/payment",
		server.createPayment).Methods("POST")
	server.Dispatch.HandleFunc("/payment/{id}",
		server.getPayment).Methods("GET")
	server.Dispatch.HandleFunc("/payment/{id}",
		server.updatePayment).Methods("PUT")
	server.Dispatch.HandleFunc("/payment/{id}",
		server.deletePayment).Methods("DELETE")
}

// Run is the main event loop and starts the web server to listening on
// the defined port for input.
func (server *Server) Run(addr string) {
	defer server.Session.Close()
	log.Fatal(http.ListenAndServe(addr, server.Dispatch))
}

// getPayments is the entry-point dispatcher for the collection of
// returned payment records. It responds to the URL payments and an
// appropriate GET request.
func (server *Server) getPayments(w http.ResponseWriter, r *http.Request) {
	var p Payment
	var payment []Payment
	var paymentScope Payments

	payment, err := p.modelGetPayments(server.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	paymentScope.P = payment
	paymentScope.Links.Self = "https://api.test.form3.tech/v1/payments"
	respondWithJSON(w, http.StatusOK, paymentScope)
}

// createPayment is the entry-point dispatcher for the creation of
// payment records to the backing store. It responds to the URL payment and an
// appropriate POST request.
func (server *Server) createPayment(w http.ResponseWriter, r *http.Request) {
	var p Payment
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid payload request")
		return
	}

	if err := p.modelCreatePaymentValidCheck(server.DB); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := p.modelCreatePayment(server.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, p)
}

// getPayment is the entry-point dispatcher for the retrieval of
// single payment records from the backing store. It responds to the URL
// payment/{id} and an appropriate GET request.
func (server *Server) getPayment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	p := Payment{ID: id}

	count, payment, err := p.modelGetPayment(server.DB)
	if err != nil && count < 0 {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	} else if err != nil && count == 0 {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, payment)
}

// updatePayment is the entry-point dispatcher for the retrieval and
// update of single payment records from the backing store. It
// responds to the URL payment/{id} and an appropriate PUT request.
func (server *Server) updatePayment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p := Payment{ID: vars["id"]}
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&p); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	defer r.Body.Close()

	if err := p.modelUpdatePaymentValidCheck(server.DB); err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	if err := p.modelUpdatePayment(server.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, p)
}

// deletePayment is the entry-point dispatcher for the deletion of
// a single payment record from the backing store. It responds to the URL
// payment/{id} and an appropriate DELETE request.
func (server *Server) deletePayment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p := Payment{ID: vars["id"]}

	if err := p.modelDeletePaymentValidCheck(server.DB); err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	if err := p.modelDeletePayment(server.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

// respondWithError is a convenience function that emits the status
// specified in code with an error defined in message to the
// http.ResponseWriter contained in w.
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON is a convenience function that emits, in JSON,
// whatever payload is in the payload interface. It sets the status
// defined in the code parameter, composes the JSON headers and emits
// the content to the http.ResponseWriter contained in w.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
