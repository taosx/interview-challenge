package processorStripe

import (
	"errors"
	"fmt"

	"github.com/stripe/stripe-go/checkout/session"

	"github.com/stripe/stripe-go"
)

type Processor struct {
	publicURL string
}

func New(publicURL, stripeKey string) *Processor {
	stripe.Key = stripeKey

	return &Processor{
		publicURL: publicURL,
	}
}

type Session struct {
	checkoutSession *stripe.CheckoutSession
	checkoutParams  *stripe.CheckoutSessionParams
	processor       *Processor
}

func (p Processor) CreateCheckoutSession(successPath, cancelPath string) *Session {
	successURL := fmt.Sprintf("%s/%s", p.publicURL, successPath)
	cancelURL := fmt.Sprintf("%s/%s", p.publicURL, cancelPath)

	return &Session{
		processor: &p,
		checkoutParams: &stripe.CheckoutSessionParams{
			PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
			LineItems:          []*stripe.CheckoutSessionLineItemParams{},
			SuccessURL:         &successURL,
			CancelURL:          &cancelURL,
		},
		checkoutSession: nil,
	}
}

func (s *Session) AddItem(name, description string, amount, quantity int64, currency stripe.Currency) {
	item := &stripe.CheckoutSessionLineItemParams{
		Name:        stripe.String(name),
		Description: stripe.String(description),
		Amount:      stripe.Int64(amount),
		Currency:    stripe.String(string(currency)),
		Quantity:    stripe.Int64(quantity),
	}

	s.checkoutParams.LineItems = append(s.checkoutParams.LineItems, item)
}

func (s *Session) Start() error {
	if s.checkoutParams == nil {
		return errors.New("missing stripe checkout params")
	}

	if s.checkoutParams.SuccessURL == nil {
		return errors.New("stripe session without successURL")
	}

	if s.checkoutParams.CancelURL == nil {
		return errors.New("stripe session without cancelURL")
	}

	if len(s.checkoutParams.LineItems) < 1 {
		return errors.New("stripe session without items")
	}

	session, err := session.New(s.checkoutParams)
	if err != nil {
		return err
	}

	s.checkoutSession = session
	return nil
}

func (s *Session) GetID() (string, error) {
	if s.checkoutSession == nil {
		return "", errors.New("stripe session id not found, try to start session")
	}
	return s.checkoutSession.ID, nil
}
