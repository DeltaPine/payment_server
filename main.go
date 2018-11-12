// main.go

package main

// Main entry point for the payment server. Initialze the DB, call
// the dispatcher and wait.
func main() {
	paymentServer := Server{}
	paymentServer.InitializeDB("localhost:27017", "payments_v1", "payments")
	paymentServer.Run("localhost:8080")
}
