// model.go - The modelled json payment transaction.

package main

import (
	"errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Payment is the main payment record structure with annotated bson
// and json tags.
type Payment struct {
	Type           string `bson:"type" json:"type"`
	ID             string `bson:"_id" json:"id"`
	Version        int    `bson:"version" json:"version"`
	OrganisationID string `bson:"organisation_id" json:"organisation_id"`
	Attributes     struct {
		Amount           string `bson:"amount" json:"amount"`
		BeneficiaryParty struct {
			AccountName       string `bson:"account_name" json:"account_name"`
			AccountNumber     string `bson:"account_number" json:"account_number"`
			AccountNumberCode string `bson:"account_number_code" json:"account_number_code"`
			AccountType       int    `bson:"account_type" json:"account_type"`
			Address           string `bson:"address" json:"address"`
			BankID            string `bson:"bank_id" json:"bank_id"`
			BankIDCode        string `bson:"bank_id_code" json:"bank_id_code"`
			Name              string `bson:"name" json:"name"`
		} `bson:"beneficiary_party" json:"beneficiary_party"`
		ChargesInformation struct {
			BearerCode    string `bson:"bearer_code" json:"bearer_code"`
			SenderCharges []struct {
				Amount   string `bson:"amount" json:"amount"`
				Currency string `bson:"currency" json:"currency"`
			} `bson:"sender_charges" json:"sender_charges"`
			ReceiverChargesAmount   string `bson:"receiver_charges_amount" json:"receiver_charges_amount"`
			ReceiverChargesCurrency string `bson:"receiver_charges_currency" json:"receiver_charges_currency"`
		} `bson:"charges_information" json:"charges_information"`
		Currency    string `bson:"currency" json:"currency"`
		DebtorParty struct {
			AccountName       string `bson:"account_name" json:"account_name"`
			AccountNumber     string `bson:"account_number" json:"account_number"`
			AccountNumberCode string `bson:"account_number_code" json:"account_number_code"`
			Address           string `bson:"address" json:"address"`
			BankID            string `bson:"bank_id" json:"bank_id"`
			BankIDCode        string `bson:"bank_id_code" json:"bank_id_code"`
			Name              string `bson:"name" json:"name"`
		} `bson:"debtor_party" json:"debtor_party"`
		EndToEndReference string `bson:"end_to_end_reference" json:"end_to_end_reference"`
		Fx                struct {
			ContractReference string `bson:"contract_reference" json:"contract_reference"`
			ExchangeRate      string `bson:"exchange_rate" json:"exchange_rate"`
			OriginalAmount    string `bson:"original_amount" json:"original_amount"`
			OriginalCurrency  string `bson:"original_currency" json:"original_currency"`
		} `bson:"fx" json:"fx"`
		NumericReference     string `bson:"numeric_reference" json:"numeric_reference"`
		PaymentID            string `bson:"payment_id" json:"payment_id"`
		PaymentPurpose       string `bson:"payment_purpose" json:"payment_purpose"`
		PaymentScheme        string `bson:"payment_scheme" json:"payment_scheme"`
		PaymentType          string `bson:"payment_type" json:"payment_type"`
		ProcessingDate       string `bson:"processing_date" json:"processing_date"`
		Reference            string `bson:"reference" json:"reference"`
		SchemePaymentSubType string `bson:"scheme_payment_sub_type" json:"scheme_payment_sub_type"`
		SchemePaymentType    string `bson:"scheme_payment_type" json:"scheme_payment_type"`
		SponsorParty         struct {
			AccountNumber string `bson:"account_number" json:"account_number"`
			BankID        string `bson:"bank_id" json:"bank_id"`
			BankIDCode    string `bson:"bank_id_code" json:"bank_id_code"`
		} `bson:"sponsor_party" json:"sponsor_party"`
	} `bson:"attributes" json:"attributes"`
}

// Payments is collection appropriate payment record structure.
type Payments struct {
	P     []Payment `json:"data"`
	Links struct {
		Self string `json:"self"`
	} `json:"links"`
}

// modelGetPayments will retrieve all payment records from the backing
// data store.
func (p *Payment) modelGetPayments(db *mgo.Database) ([]Payment, error) {
	payments := []Payment{}
	err := db.C(COLLECTION).Find(bson.M{}).All(&payments)
	return payments, err
}

// modelGetPayment, given the element ID in Payment, will retrieve
// the corresponding payment record from the backing
// data store.
func (p *Payment) modelGetPayment(db *mgo.Database) (int, Payment, error) {
	var payment Payment
	var count = 0

	if checkEmptyPaymentID(p) == true {
		return -1, payment, errors.New("No Payment ID specified")
	}
	query, count, err := returnPaymentCountAndQuery(db, p)
	if err != nil {
		return -1, payment, err
	} else if count == 0 {
		return count, payment, errors.New("Payment not found")
	} else if count > 1 {
		return -1, payment, errors.New("More than one payment returned per ID")
	} else {
		query.One(&payment)
	}

	return count, payment, err
}

// modelDeletePaymentValidCheck, given the element ID in Payment, will
// return the corresponding validity of whether a payment record can
// be deleted. If the payment record cannot be deleted, the function
// raises an error with a 'reason' string, otherwise it returns nil if
// a payment record can be deleted.
func (p *Payment) modelDeletePaymentValidCheck(db *mgo.Database) error {
	if checkEmptyPaymentID(p) == true {
		return errors.New("Cannot delete a payment without a Payment ID specified")
	}

	count, err := returnPaymentCount(db, p)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("A payment with this Payment ID doesn't exists")
	}
	return nil
}

// modelDeletePayment, given the element ID in Payment, will
// delete the corresponding payment record in the backing store. If an
// error occurs, an error will be returned.
func (p *Payment) modelDeletePayment(db *mgo.Database) error {
	err := db.C(COLLECTION).Remove(bson.M{"_id": p.ID})
	return err
}

// modelCreatePaymentValidCheck, given the element ID in Payment, will
// return the corresponding validity of whether a payment record can
// be created in the backing store. If the payment record cannot be
// created, the function raises an error with a 'reason' string,
// otherwise it returns nil if a payment record can be created.
func (p *Payment) modelCreatePaymentValidCheck(db *mgo.Database) error {
	if checkEmptyPaymentID(p) == true {
		return errors.New("Cannot add a payment without a Payment ID specified")
	}

	count, err := returnPaymentCount(db, p)
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("A payment with this Payment ID already exists")
	}
	return nil
}

// modelCreatePayment, given the full population of Payment, will
// create the corresponding payment record in the backing store. If an
// error occurs, an error will be returned.
func (p *Payment) modelCreatePayment(db *mgo.Database) error {
	err := db.C(COLLECTION).Insert(&p)
	return err
}

// modelUpdatePaymentValidCheck, given the element ID in Payment, will
// return the corresponding validity of whether a payment record can
// be modified in the backing store. If the payment record cannot be
// modified, the function raises an error with a 'reason' string,
// otherwise it returns nil if a payment record can be modified.
func (p *Payment) modelUpdatePaymentValidCheck(db *mgo.Database) error {
	if checkEmptyPaymentID(p) == true {
		return errors.New("Cannot update a payment without a Payment ID specified")
	}

	count, err := returnPaymentCount(db, p)

	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("A payment with this Payment ID does not exist")
	}
	return nil
}

// modelUpdatePayment, given the full population of Payment, will
// update the corresponding payment record in the backing store. If an
// error occurs, an error will be returned.
func (p *Payment) modelUpdatePayment(db *mgo.Database) error {
	err := db.C(COLLECTION).UpdateId(p.ID, &p)
	return err
}

// checkEmptyPaymentID is a convenience function to ascertain whether
// the ID field is populated. Currently the only check performed is
// whether the ID = "" which the function defines as empty.
func checkEmptyPaymentID(p *Payment) bool {
	if p.ID == "" {
		return true
	}
	return false
}

// returnPaymentCount is a convenience function to ascertain the number
// of payment records defined by the Payment ID field. This function
// should only return 0 or 1 in valid cases (though it makes no
// distinction on validity). If -1 is returned an error occurred in
// the query and the error is returned.
func returnPaymentCount(db *mgo.Database, p *Payment) (int, error) {
	query := db.C(COLLECTION).Find(bson.M{"_id": p.ID})
	count, err := query.Count()
	if err != nil {
		return -1, err
	}
	return count, nil
}

// returnPaymentCountAndQuery is a convenience function to ascertain the
// number of payment records defined by the Payment ID field. This
// function should only return 0 or 1 in valid cases (though it makes
// no distinction on validity). If -1 is returned an error occurred in
// the query and the error is returned. An additional object, if no
// errors occur is returned: the Query object created by the function.
func returnPaymentCountAndQuery(db *mgo.Database, p *Payment) (*mgo.Query, int, error) {
	query := db.C(COLLECTION).Find(bson.M{"_id": p.ID})
	count, err := query.Count()
	if err != nil {
		return nil, -1, err
	}
	return query, count, nil
}
