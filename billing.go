package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type (
	//то что пошлется на URL
	BillingResp struct {
		Destination string `json:"destination,omitempty"`
		Amount      string `json:"amount,omitempty"`
		From        string `json:"from,omitempty"`
		To          string `json:to,omitempty`
		SessionID   string `json:sessionId,omitEmpty`
	}
	ErrorResponse struct {
		Code    int64  `json:code`
		Message string `json:message`
	}
	BillingReq struct {
		Amount  string        `bson:"amount" json:"amount,omitempty"`
		From    string        `bson:"from" json:from,omitempty`
		To      string        `bson:"to" json:to,omitempty`
		CardNum string        `bson:"cardnum" json:cardNum,onitempty`
		ID      bson.ObjectId `bson:"id" json:"id,omitempty"`
	}
)

var session = initMongo()

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/register", CreatePayment).Methods("POST")
	router.HandleFunc("/register/{id:[0-9]+}", ShowPaymentResults).Methods("POST")
	log.Fatal(http.ListenAndServe(":9090", router))
}
func CreatePayment(w http.ResponseWriter, r *http.Request) {
	var payment BillingReq
	err := json.NewDecoder(r.Body).Decode(&payment)
	if err != nil {
		ErrorWithJSON(w, ErrorResponse{Code: 400, Message: "Invalid parameters"}, http.StatusBadRequest)
		return
	}
	if !LouneVerification(payment.CardNum) {
		ErrorWithJSON(w, ErrorResponse{Code: 400, Message: "Invalid CardNum"}, http.StatusBadRequest)
	}
	//creating DB session and putting our payment in
	newSession := session.DB("glofox").C("payments")
	payment.ID = bson.NewObjectId()
	err = newSession.Insert(payment)
	if err != nil {
		if mgo.IsDup(err) {
			ErrorWithJSON(w, ErrorResponse{Code: 400, Message: "Payment ID collision"}, http.StatusBadRequest)
			return
		}
		ErrorWithJSON(w, ErrorResponse{Code: 500, Message: "Database error"}, http.StatusInternalServerError)
		log.Println("Creating new payment error: ", err)
	}
	err = newSession.Find(bson.M{"id": payment.ID}).One(&payment)
	if err != nil {
		log.Fatal(err)
	}
	response, _ := json.MarshalIndent(payment, " ", " ")
	ResponseWithJSON(w, response, http.StatusOK)
}

func ShowPaymentResults(w http.ResponseWriter, r *http.Request) {
	var payment BillingResp
	err := json.NewDecoder(r.Body).Decode(&payment)
	if err != nil {
		ErrorWithJSON(w, ErrorResponse{Code: 400, Message: "Invalid parameters"}, http.StatusBadRequest)
		return
	}
	responce, _ := json.MarshalIndent(payment, " ", " ")
	ResponseWithJSON(w, responce, http.StatusOK)
}
func ErrorWithJSON(w http.ResponseWriter, message ErrorResponse, code int) {
	resp, _ := json.Marshal(message)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write([]byte(resp))
}

func ResponseWithJSON(w http.ResponseWriter, json []byte, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(json)
}

func initMongo() *mgo.Session {
	session, err := mgo.Dial("db")
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	return session
}

func LouneVerification(input string) bool {
	sum := 0
	for i := 0; i < len(input); i++ {
		digit, _ := strconv.Atoi(string(input[len(input)-i-1]))
		if i%2 == 1 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0
}
