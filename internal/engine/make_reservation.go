package engine

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"github.com/stripe/stripe-go"
)

func (e *Engine) MakeReservation(userName string) (string, error) {
	if userName == "" {
		return "", errors.New("creation of user with empty name failed")
	}

	user, err := e.UserRepo.Create(userName)
	if err != nil {
		if !e.UserRepo.IsDuplicateErr(err) {
			return "", errors.Wrapf(err, "creation of user '%s' failed", userName)
		}
		user, err = e.UserRepo.GetByName(userName)
		if err != nil {
			return "", err
		}
	}

	if user.ID == -1 {
		return "", errors.New("can't reserve ticket, can't retrieve user")
	}

	ticket, err := e.TicketRepo.Reserve(user.ID)
	if err != nil {
		if e.TicketRepo.IsAlreadyReservedErr(err) {
			return user.NameSlug, fmt.Errorf("user '%s' already has a reserved ticket", userName)
		}
		return "", errors.Wrapf(err, "failed to reserve ticket for user '%s'", userName)
	}

	session := e.PaymentProcessor.CreateCheckoutSession(
		fmt.Sprintf("booked?ticket_id=%d", ticket.ID),
		fmt.Sprintf("canceled?ticket_id=%d", ticket.ID),
	)

	elementNames := []string{
		"Rhodium",
		"Platinum",
		"Gold",
		"Ruthenium",
		"Iridium",
	}

	rand.Seed(time.Now().Unix())
	element := elementNames[rand.Intn(len(elementNames))]

	session.AddItem(
		fmt.Sprintf("%s Ticket #%d", element, ticket.ID),
		"One of a kind "+element+" ticket",
		int64(ticket.Cost),
		1,
		stripe.CurrencyGBP,
	)

	err = session.Start()
	if err != nil {
		return "", errors.Wrap(err, "failed to create checkout session for ticket booking")
	}

	sessionID, err := session.GetID()
	if err != nil {
		return "", errors.Wrap(err, "failed to retrieve checkout session id for ticket booking")
	}

	err = e.TicketRepo.AttachCheckoutSessionID(ticket.ID, sessionID)
	if err != nil {
		return "", err
	}

	return user.NameSlug, nil
}
