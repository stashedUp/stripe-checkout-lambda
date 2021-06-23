package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/checkout/session"
)

var (
	HTTPMethodNotSupported = errors.New("no name was provided in the HTTP body")
)

// ErrorResponseMessage represents the structure of the error
// object sent in failed responses.
type ErrorResponseMessage struct {
	Message string `json:"message"`
}

// ErrorResponse represents the structure of the error object sent
// in failed responses.
type ErrorResponse struct {
	Error *ErrorResponseMessage `json:"error"`
}

type Message struct {
	OwnerEmail     string `json:"owner_email"`
	ContactEmail   string `json:"contact_email,omitempty"`
	ContactName    string `json:"contact_name,omitempty"`
	ContactPhone   string `json:"contact_phone,omitempty"`
	MessageContent string `json:"message_content"`
}

const (
	DEFAULT = "http://example.com"
)

var (
	statusCode int
)

func init() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// For sample support and debugging, not required for production:
	stripe.SetAppInfo(&stripe.AppInfo{
		Name:    "stripe-samples/checkout-one-time-payments",
		Version: "0.0.1",
		URL:     "https://github.com/stripe-samples/checkout-one-time-payments",
	})
}

//HandleRequest incoming request for checkout
func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	redirect := make(map[string]string)
	statusCode = 200

	redirect["Location"] = DEFAULT
	redirect["Access-Control-Allow-Origin"] = "*"
	redirect["Access-Control-Allow-Headers"] = "*"

	if request.HTTPMethod == "GET" {

		statusCode = 200
		body := handleCheckoutSession(&request)

		return events.APIGatewayProxyResponse{Headers: redirect, StatusCode: statusCode, Body: body}, nil
	} else if request.HTTPMethod == "POST" {
		fmt.Println("Post method")
		body := handleCreateCheckoutSession(&request)
		return events.APIGatewayProxyResponse{Headers: redirect, StatusCode: statusCode, Body: body}, nil
	} else {
		fmt.Printf("NEITHER\n")
		return events.APIGatewayProxyResponse{}, HTTPMethodNotSupported
	}
}

func main() {
	lambda.Start(HandleRequest)
}

func writeJSON(v interface{}) string {
	fmt.Println("Attempting to write to JSON")

	var buf bytes.Buffer
	var res bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		log.Printf("json.NewEncoder.Encode: %v", err)
		return res.String()
	}

	if _, err := io.Copy(&res, &buf); err != nil {
		log.Printf("io.Copy: %v", err)
		return res.String()
	}
	return res.String()

}

type checkoutSessionCreateReq struct {
	Price  string `json:"price"`
	Email  string `json:"email"`
	Domain string `json:"domain"`
}

func handleCreateCheckoutSession(r *events.APIGatewayProxyRequest) string {

	req := checkoutSessionCreateReq{}
	json.Unmarshal([]byte(r.Body), &req)

	domainURL := os.Getenv("DOMAIN")
	//domainURL := os.Getenv("DOMAIN")
	custEmail := req.Email
	prodPriceIdent := req.Price

	// Pulls the list of payment method types from environment variables (`.env`).
	// In practice, users often hard code the list of strings.
	paymentMethodTypes := strings.Split(os.Getenv("PAYMENT_METHOD_TYPES"), ",")

	// Create new Checkout Session for the order
	// Other optional params include:
	// [billing_address_collection] - to display billing address details on the page
	// [customer] - if you have an existing Stripe Customer ID
	// [payment_intent_data] - lets capture the payment later
	// [customer_email] - lets you prefill the email input in the form
	// For full details see https://stripe.com/docs/api/checkout/sessions/create

	// ?session_id={CHECKOUT_SESSION_ID} means the redirect will have the session ID
	// set as a query param
	params := &stripe.CheckoutSessionParams{
		SuccessURL:         stripe.String(domainURL + "/success.html?email=" + custEmail + "&session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:          stripe.String(domainURL + "/canceled.html"),
		PaymentMethodTypes: stripe.StringSlice(paymentMethodTypes),
		Mode:               stripe.String(string(stripe.CheckoutSessionModePayment)),
		CustomerEmail:      stripe.String(custEmail),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Quantity: stripe.Int64(1),
				Price:    stripe.String(prodPriceIdent),
			},
		},
	}
	s, err := session.New(params)
	if err != nil {
		//http.Error(w, fmt.Sprintf("error while creating session %v", err.Error()), http.StatusInternalServerError)
		fmt.Println("Something went wrong")
		return ""
	}

	res := writeJSON(struct {
		SessionID string `json:"sessionId"`
	}{
		SessionID: s.ID,
	})

	return res
}

//NOT USED
func handleCheckoutSession(r *events.APIGatewayProxyRequest) string {
	fmt.Println("handleCheckoutSession")
	if r.HTTPMethod != "GET" {
		fmt.Println("Cannot GET for handleCheckoutSession")
		return ""
	}
	sessionID := r.QueryStringParameters

	s, _ := session.Get(sessionID["sessionId"], nil)
	fmt.Println(s)
	return writeJSON(s)
}
