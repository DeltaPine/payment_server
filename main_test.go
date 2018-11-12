// main_test.go

package main

import (
	"bytes"
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

var server Server

// Internal testsuite utility functions

func clearTable() {
	server.DB.C(COLLECTION).RemoveAll(nil)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	server.Dispatch.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n",
			expected, actual)
	}
}

func compareResponseCode(t *testing.T, expected, actual int) bool {
	if expected != actual {
		return false
	}
	return true
}

// Test entry point
func TestMain(m *testing.M) {
	server = Server{}
	server.InitializeDB("localhost:27017", "test_v1", "payments")
	code := m.Run()
	clearTable()
	os.Exit(code)
}

// BDD GoConvey tests.

// Initial test to determine platform availability. Request a
// collection of payments (there should be none), the server should
// respond with StatusOK (lack of payment records is not a
// failure). Test that the empty array is properly formatted.
func TestNewTestServerStart(t *testing.T) {
	clearTable()
	Convey("As a new database of payments", t, func() {
		req, _ := http.NewRequest("GET", "/payments", nil)
		response := executeRequest(req)

		Convey("The web server routing should be active", func() {
			So(compareResponseCode(t, http.StatusOK, response.Code),
				ShouldEqual, true)
		})
		Convey("And that the basic test of getting all payments", func() {
			Convey("Should return an empty JSON formatted array", func() {
				So(response.Body.String(),
					ShouldEqual,
					`{"data":[],"links":{"self":"https://api.test.form3.tech/v1/payments"}}`)

			})
		})
	})
}

// Test the basic function of fetching a record that does not exist.
// The server should return a StatusNotFound with an error message
// "Payment not found".
func TestNoPaymentRecord(t *testing.T) {
	Convey("Now that the database is confirmed to be empty", t, func() {
		Convey("This is a good time to test the no payment found server response code", func() {
			req, _ := http.NewRequest("GET", "/payment/11", nil)
			response := executeRequest(req)
			So(compareResponseCode(t, http.StatusNotFound, response.Code),
				ShouldEqual, true)
			Convey("And also to test the 'Payment not found' error", func() {
				var m map[string]string

				json.Unmarshal(response.Body.Bytes(), &m)
				So(m["error"], ShouldEqual, "Payment not found")
			})
		})
	})
}

// Test attempting to add a payment record without a Payment ID. If a
// client attempts to add a payment record without an ID, the payment
// addition request should be rejected, StatusBadRequest returned, and
// an error message delivered.
func TestNoPaymentID(t *testing.T) {
	Convey("Testing payment addition without a Payment ID", t, func() {
		payload := []byte(`{"type":"Payment","id":""}`)
		Convey("If a client attempts to add a payment record without an id", func() {
			req, _ := http.NewRequest("POST", "/payment", bytes.NewBuffer(payload))
			response := executeRequest(req)
			Convey("The payment addition request should be rejected", func() {
				So(compareResponseCode(t, http.StatusBadRequest, response.Code),
					ShouldEqual, true)
			})
			Convey("And an appropriate error is delivered", func() {
				var m map[string]string

				json.Unmarshal(response.Body.Bytes(), &m)
				So(m["error"], ShouldEqual,
					"Cannot add a payment without a Payment ID specified")
			})

		})
	})
}

// Test creating a valid payment record. The payment record payload
// should be assembled, sent to the server, and a StatusOK should be
// returned. Fetch the successfully added payment from the server and
// compare to the payload to ensure accuracy of information.
func TestCreateValidPayment(t *testing.T) {
	clearTable()
	Convey("Create successful payment record with a correct server status code return", t, func() {
		req, _ := http.NewRequest("POST", "/payment", bytes.NewBuffer(payload))
		response := executeRequest(req)
		So(compareResponseCode(t, http.StatusCreated,
			response.Code), ShouldEqual, true)
		Convey("Payment has been created. Fetch the added payment back from the server", func() {
			req, _ = http.NewRequest("GET", "/payment/4ee3a8d8-ca7b-4290-a52c-dd5b6165ec43", nil)
			response = executeRequest(req)
			So(compareResponseCode(t, http.StatusOK, response.Code),
				ShouldEqual, true)
			Convey("And check that payload payment and the fetched payment are equal", func() {
				var fpayment Payment
				var payload_payment Payment

				json.Unmarshal(payload, &payload_payment)
				json.Unmarshal(response.Body.Bytes(), &fpayment)
				So(reflect.DeepEqual(payload_payment,
					fpayment), ShouldEqual, true)

			})
		})
	})
}

// Test trying to post a payment record with an ID that already exists
// in the server. Post a payment record, check the server code, and
// try to post the same payment record with the same ID. Check the
// server returns an appropriate status code and error message.
func TestDuplicateIDPayment(t *testing.T) {
	clearTable()
	Convey("Post a successful payment record with correct server status code return", t, func() {
		req, _ := http.NewRequest("POST", "/payment", bytes.NewBuffer(payload))
		response := executeRequest(req)
		So(compareResponseCode(t, http.StatusCreated,
			response.Code), ShouldEqual, true)
		Convey("Try to create another payment with the same Payment ID and check server status", func() {
			req, _ := http.NewRequest("POST", "/payment", bytes.NewBuffer(payload))
			response := executeRequest(req)
			So(compareResponseCode(t, http.StatusBadRequest, response.Code),
				ShouldEqual, true)
			Convey("Ensure a payment exists error is delivered", func() {
				var m map[string]string

				json.Unmarshal(response.Body.Bytes(), &m)
				So(m["error"], ShouldEqual,
					"A payment with this Payment ID already exists")
			})
		})
	})
}

// Test a delete of a valid payment record. Post a payment ID, ensure that the server
// returns the correct status code, then delete the payment and once
// again ensure that the server returns the correct status code.
func TestDeleteValidPayment(t *testing.T) {
	clearTable()
	Convey("Post a successful payment creation with correct server status code return", t, func() {
		req, _ := http.NewRequest("POST", "/payment", bytes.NewBuffer(payload))
		response := executeRequest(req)
		So(compareResponseCode(t, http.StatusCreated,
			response.Code), ShouldEqual, true)
		Convey("Payment added. Delete payment", func() {
			req, _ = http.NewRequest("DELETE",
				"/payment/4ee3a8d8-ca7b-4290-a52c-dd5b6165ec43", nil)
			response = executeRequest(req)
			So(compareResponseCode(t, http.StatusOK, response.Code),
				ShouldEqual, true)
		})
	})
}

// Test a delete of a payment record that does not exist on the
// server. Ensure that the server returns the appropriate status, and
// that the error message is correctly returned.
func TestDeleteNoRecord(t *testing.T) {
	clearTable()
	Convey("Attempt to delete a non-existing payment", t, func() {
		req, _ := http.NewRequest("DELETE", "/payment/12", nil)
		response := executeRequest(req)
		So(compareResponseCode(t, http.StatusNotFound, response.Code),
			ShouldEqual, true)
		Convey("And a payment does not exist  error is delivered", func() {
			var m map[string]string

			json.Unmarshal(response.Body.Bytes(), &m)
			So(m["error"], ShouldEqual, "A payment with this Payment ID doesn't exists")
		})
	})
}

// Test an update of a payment record. First create a payment in the
// database, check the server status, and then fetch that payment
// record, again checking correct server status codes. With that
// payment record, perform a consistency check to ensure that the to-be
// updated record is different from the updated record. Write the
// modification record to the server and check the server status
// code. Retrieve the now modified record from the server, check
// status codes, and ensure that the modification is correct.
func TestValidUpdate(t *testing.T) {
	clearTable()
	Convey("Create a successful payment with correct server status code returned", t, func() {
		req, _ := http.NewRequest("POST", "/payment", bytes.NewBuffer(payload))
		response := executeRequest(req)
		So(compareResponseCode(t, http.StatusCreated, response.Code),
			ShouldEqual, true)
	})
	Convey("Fetch the created payment from the server", t, func() {
		var before_payment Payment
		var after_payment Payment
		var payload_payment Payment

		json.Unmarshal(payload2, &payload_payment)
		req, _ := http.NewRequest("GET", "/payment/4ee3a8d8-ca7b-4290-a52c-dd5b6165ec43", nil)
		response := executeRequest(req)
		So(compareResponseCode(t, http.StatusOK, response.Code), ShouldEqual, true)
		json.Unmarshal(response.Body.Bytes(), &before_payment)
		Convey("Check the retrieved payment is different from the proposed modification",
			func() {
				isEqual := reflect.DeepEqual(before_payment, payload_payment)
				So(isEqual, ShouldNotEqual, true)
			})
		Convey("Write the modification to the server",
			func() {
				req, _ = http.NewRequest("PUT",
					"/payment/4ee3a8d8-ca7b-4290-a52c-dd5b6165ec43",
					bytes.NewBuffer(payload2))
				response = executeRequest(req)
				So(compareResponseCode(t, http.StatusOK, response.Code),
					ShouldEqual, true)
			})
		Convey("Fetch the newly modified payment from the server",
			func() {
				req, _ = http.NewRequest("GET",
					"/payment/4ee3a8d8-ca7b-4290-a52c-dd5b6165ec43", nil)
				response = executeRequest(req)
				So(compareResponseCode(t, http.StatusOK, response.Code),
					ShouldEqual, true)
			})
		json.Unmarshal(response.Body.Bytes(), &after_payment)
		Convey("Check the retrieved modified payment is the same as the modification requested",
			func() {
				So(reflect.DeepEqual(after_payment,
					payload_payment), ShouldEqual, true)
			})
	})
}

// Perform an updated on a payment record where the payment record
// requested to be modified is not found on the server.  Attempt a
// modify with a known non-existing ID, check the server status code
// is correctly returned and an error message is returned.
func TestUpdatePaymentNotFound(t *testing.T) {
	clearTable()
	Convey("Attempt to update a non-existent payment", t, func() {
		var payload_payment Payment

		req, _ := http.NewRequest("PUT", "/payment/123", bytes.NewBuffer(payload2))
		response := executeRequest(req)
		json.Unmarshal(payload2, &payload_payment)
		Convey("Write the modification to the server with a non-existent payment ID", func() {
			So(compareResponseCode(t, http.StatusNotFound, response.Code),
				ShouldEqual, true)
		})
		Convey("Check a payment does not exist error is delivered", func() {
			var m map[string]string
			json.Unmarshal(response.Body.Bytes(), &m)
			So(m["error"], ShouldEqual,
				"A payment with this Payment ID does not exist")
		})
	})
}

// Test the return of a collection of payment records, returned in a
// sequence. First populate the database with four payment
// records and check the server status codes for each addition. Then,
// retrieve the collection of payment records, check the server
// status, and then check the IDs are the same as the IDs that
// initially populated the server.
func TestGetMultiplePayments(t *testing.T) {
	clearTable()
	paymentIDs := []string{"4ee3a8d8-ca7b-4290-a52c-dd5b6165ec43",
		"5ee3a8d8-ca7b-4290-a52c-dd5b6165ec43",
		"6ee3a8d8-ca7b-4290-a52c-dd5b6165ec43",
		"7ee3a8d8-ca7b-4290-a52c-dd5b6165ec43"}
	Convey("Create four successful payments with correct server status code returned", t, func() {
		var payload_payment Payment

		json.Unmarshal(payload2, &payload_payment)
		for index, _ := range paymentIDs {
			payload_payment.ID = paymentIDs[index]
			json_payload, _ := json.Marshal(payload_payment)
			req, _ := http.NewRequest("POST",
				"/payment", bytes.NewBuffer(json_payload))
			response := executeRequest(req)
			So(compareResponseCode(t, http.StatusCreated, response.Code),
				ShouldEqual, true)
		}
		Convey("Retrieve payments with correct server status code returned", func() {
			var result Payments
			req, _ := http.NewRequest("GET", "/payments", nil)
			response := executeRequest(req)
			So(compareResponseCode(t, http.StatusOK, response.Code),
				ShouldEqual, true)
			Convey("Iterate and compare payment IDs", func() {
				json.Unmarshal(response.Body.Bytes(), &result)
				for index, element := range paymentIDs {
					So(result.P[index].ID,
						ShouldEqual, element)
				}
			})
		})
	})
}

// API based unit tests.

// Test the request of a sequence of payment IDs when the server has
// no records. Check server status codes and ensure the return is
// formatted, and empt, JSON.
func TestEmptyTable(t *testing.T) {
	clearTable()
	req, _ := http.NewRequest("GET", "/payments", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	body := response.Body.String()
	if body != `{"data":[],"links":{"self":"https://api.test.form3.tech/v1/payments"}}` {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

// Test API entry-point for the retrieval of a payment record and
// where that payment record does not exist. Test to make sure the
// correct server status code is returned and that a suitable error
// message is produced.
func TestGetNonExistentPayment(t *testing.T) {
	clearTable()
	req, _ := http.NewRequest("GET", "/payment/11", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Payment not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Payment not found'. Got '%s'", m["error"])
	}
}

// Test API entry-point for create a payment record. Post the payment
// record to the server and check the status code to indicate success.
func TestCreatePayment(t *testing.T) {
	clearTable()
	req, _ := http.NewRequest("POST", "/payment", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)
}

// Test API entry-point for retrieving a payment record. Retrieve the
// payment record created by TestCreatePayment, check the server
// status code, and compare the record to the payload.
func TestGetProduct(t *testing.T) {
	var cpayment Payment
	var fpayment Payment

	json.Unmarshal(payload, &cpayment)
	// Payment should have been created and persisted to
	// storage. Fetch it and compare.
	req, _ := http.NewRequest("GET", "/payment/4ee3a8d8-ca7b-4290-a52c-dd5b6165ec43", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	json.Unmarshal(response.Body.Bytes(), &fpayment)
	if reflect.DeepEqual(cpayment, fpayment) != true {
		t.Error("Payload and store payment not equal")
	}
}

// Test API entry-point for updating a payment record. Retrieve the
// payment record created by TestCreatePayment, check the server
// status code, write the modification and check the status
// code. Finally compare the record to the payload.
func TestUpdatePayment(t *testing.T) {
	var payload_payment Payment
	var before_payment Payment
	var after_payment Payment
	var response *httptest.ResponseRecorder

	json.Unmarshal(payload2, &payload_payment)
	// Get and marshal the payment to be modified into a
	// structure before modification
	req, _ := http.NewRequest("GET",
		"/payment/4ee3a8d8-ca7b-4290-a52c-dd5b6165ec43", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	json.Unmarshal(response.Body.Bytes(), &before_payment)

	// Check to make sure the modification and intended payments
	// are different
	if reflect.DeepEqual(payload2, before_payment) == true {
		t.Error("Modification payload payment and stored payments are equal")
	}

	// Write the modification to the server
	req, _ = http.NewRequest("PUT",
		"/payment/4ee3a8d8-ca7b-4290-a52c-dd5b6165ec43",
		bytes.NewBuffer(payload2))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// Marshall the now modified payment payload into a structure
	req, _ = http.NewRequest("GET",
		"/payment/4ee3a8d8-ca7b-4290-a52c-dd5b6165ec43", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	json.Unmarshal(response.Body.Bytes(), &after_payment)

	// Check to make sure the modified and before modification payments
	// are not equal
	if reflect.DeepEqual(before_payment, after_payment) == true {
		t.Error("Modification payload payment and stored payments are equal")
	}

	// Check to make sure the now modified and stored payment is
	// equal to the payload modification payments
	if reflect.DeepEqual(payload_payment, after_payment) != true {
		t.Error("Modification payload payment and and after modification payment are not equal")
	}

}

// Test Delete a payment API entry point. Get a payment record, check
// the server status code, deleted that record and check the server
// status code and finally, try to retrieve the payment record.
func TestDeletePayment(t *testing.T) {
	req, _ := http.NewRequest("GET",
		"/payment/4ee3a8d8-ca7b-4290-a52c-dd5b6165ec43", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	req, _ = http.NewRequest("DELETE",
		"/payment/4ee3a8d8-ca7b-4290-a52c-dd5b6165ec43", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	req, _ = http.NewRequest("GET",
		"/payment/4ee3a8d8-ca7b-4290-a52c-dd5b6165ec43", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

// Test payload
var payload = []byte(`{"type":"Payment","id":"4ee3a8d8-ca7b-4290-a52c-dd5b6165ec43","version":0,"organisation_id":"743d5b63-8e6f-432e-a8fa-c5d8d2ee5fcb","attributes":{"amount":"100.21","beneficiary_party":{"account_name":"W Owens","account_number":"31926819","account_number_code":"BBAN","account_type":0,"address":"1 The Beneficiary Localtown SE2","bank_id":"403000","bank_id_code":"GBDSC","name":"Wilfred Jeremiah Owens"},"charges_information":{"bearer_code":"SHAR","sender_charges":[{"amount":"5.00","currency":"GBP"},{"amount":"10.00","currency":"USD"}],"receiver_charges_amount":"1.00","receiver_charges_currency":"USD"},"currency":"GBP","debtor_party":{"account_name":"EJ Brown Black","account_number":"GB29XABC10161234567801","account_number_code":"IBAN","address":"10 Debtor Crescent Sourcetown NE1","bank_id":"203301","bank_id_code":"GBDSC","name":"Emelia Jane Brown"},"end_to_end_reference":"Wil piano Jan","fx":{"contract_reference":"FX123","exchange_rate":"2.00000","original_amount":"200.42","original_currency":"USD"},"numeric_reference":"1002001","payment_id":"123456789012345678","payment_purpose":"Paying for goods/services","payment_scheme":"FPS","payment_type":"Credit","processing_date":"2017-01-18","reference":"Payment for Em's piano lessons","scheme_payment_sub_type":"InternetBanking","scheme_payment_type":"ImmediatePayment","sponsor_party":{"account_number":"56781234","bank_id":"123123","bank_id_code":"GBDSC"}}}`)

// Modified Payload for update test
// Amount changed to 121.00
// Debtor Payment name changed to Brown Blue
var payload2 = []byte(`{"type":"Payment","id":"4ee3a8d8-ca7b-4290-a52c-dd5b6165ec43","version":0,"organisation_id":"743d5b63-8e6f-432e-a8fa-c5d8d2ee5fcb","attributes":{"amount":"121.00","beneficiary_party":{"account_name":"W Owens","account_number":"31926819","account_number_code":"BBAN","account_type":0,"address":"1 The Beneficiary Localtown SE2","bank_id":"403000","bank_id_code":"GBDSC","name":"Wilfred Jeremiah Owens"},"charges_information":{"bearer_code":"SHAR","sender_charges":[{"amount":"5.00","currency":"GBP"},{"amount":"10.00","currency":"USD"}],"receiver_charges_amount":"1.00","receiver_charges_currency":"USD"},"currency":"GBP","debtor_party":{"account_name":"EJ Brown Blue","account_number":"GB29XABC10161234567801","account_number_code":"IBAN","address":"10 Debtor Crescent Sourcetown NE1","bank_id":"203301","bank_id_code":"GBDSC","name":"Emelia Jane Brown"},"end_to_end_reference":"Wil piano Jan","fx":{"contract_reference":"FX123","exchange_rate":"2.00000","original_amount":"200.42","original_currency":"USD"},"numeric_reference":"1002001","payment_id":"123456789012345678","payment_purpose":"Paying for goods/services","payment_scheme":"FPS","payment_type":"Credit","processing_date":"2017-01-18","reference":"Payment for Em's piano lessons","scheme_payment_sub_type":"InternetBanking","scheme_payment_type":"ImmediatePayment","sponsor_party":{"account_number":"56781234","bank_id":"123123","bank_id_code":"GBDSC"}}}`)
