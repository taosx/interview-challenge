package ticket

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/taosx/interview-challenge/internal/domain"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/taosx/interview-challenge/internal/storage"
)

const (
	maxReservedTime time.Duration = 5 * time.Minute
)

type taskFunction func(r sqlx.Ext) error
type task struct {
	runAfter time.Duration
	taskFn   taskFunction
}

type repoSQLite struct {
	db        *storage.SQLiteStorage
	scheduler chan task
	mu        sync.Mutex
}

func NewSQLiteRepo(db *storage.SQLiteStorage) TicketRepository {
	repo := &repoSQLite{
		db:        db,
		scheduler: make(chan task),
		mu:        sync.Mutex{},
	}

	repo.autoMigrate().initialPopulate()

	go repo.startTaskRunner()
	repo.scoutTasks()

	return repo
}

func (r *repoSQLite) CountUnreserved() (int, error) {
	var count int

	err := r.db.QueryRowx("SELECT count(id) FROM ticket WHERE state = $1;", Unreserved).Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}

// Reserve reserves a ticket by changing the state to Reserved and adding a user_id
func (r *repoSQLite) Reserve(userID int) (*domain.Ticket, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ticket, err := unreservedTicket(r.db)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving unreserved ticket failed")
	}

	err = reservingTicket(r.db, ticket.ID, userID)
	if err != nil {
		return nil, err
	}

	r.addTask(taskCleanUnreservedAfter(maxReservedTime, ticket.ID))

	domainTicket := ticket.toDomain()
	return &domainTicket, nil
}

func (r *repoSQLite) GetReservedByID(id int) (*domain.Ticket, error) {
	ticket := new(Ticket)
	err := r.db.QueryRowx("SELECT * FROM ticket WHERE state = $1 LIMIT 1;", Reserved).StructScan(ticket)
	if err != nil {
		return nil, err
	}

	domainTicket := ticket.toDomain()
	return &domainTicket, nil
}

func (r *repoSQLite) GetReservedBySessionID(sessionID string) (*domain.Ticket, error) {
	ticket := new(Ticket)
	err := r.db.QueryRowx("SELECT * FROM ticket WHERE state = $1 AND session_id = $2 LIMIT 1;", Reserved, sessionID).StructScan(ticket)
	if err != nil {
		return nil, err
	}

	domainTicket := ticket.toDomain()
	return &domainTicket, nil
}

func (r *repoSQLite) GetReservedByUserID(userID int) (*domain.Ticket, error) {
	ticket := new(Ticket)

	err := r.db.QueryRowx(`
	SELECT t.* FROM ticket AS t
	LEFT OUTER JOIN user AS u
		ON t.user_id = u.id
	WHERE t.state = $1 AND u.id = $2;
	`, Reserved, userID).StructScan(ticket)
	if err != nil {
		return nil, err
	}

	domainTicket := ticket.toDomain()
	return &domainTicket, nil
}

func (r *repoSQLite) AttachCheckoutSessionID(ticketID int, sessionID string) error {
	result, err := r.db.Exec(`
	UPDATE ticket
	SET session_id = $1
	WHERE id = $2;
	`, sessionID, ticketID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "failed to attach checkout session to ticket '%d'", ticketID)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("failed to attach checkout session ticket %d, no rows affected", ticketID)
	}

	return nil
}

// Book books a ticket to a user by changing the state to Booked
func (r *repoSQLite) Book(ticketID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	result, err := r.db.Exec(`
	UPDATE ticket
	SET state = $1
	WHERE id = $2;
	`, Booked, ticketID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "failed to book ticket '%d'", ticketID)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("failed to book ticket %d, no rows affected", ticketID)
	}

	return nil
}

func ticketByID(db sqlx.Ext, id int) (*Ticket, error) {
	ticket := new(Ticket)

	err := db.QueryRowx("SELECT * FROM ticket WHERE id = $1 LIMIT 1;", id).Scan(ticket)
	if err != nil {
		return nil, err
	}

	return ticket, nil
}

func unreservedTicket(db sqlx.Ext) (*Ticket, error) {
	ticket := new(Ticket)
	err := db.QueryRowx("SELECT * FROM ticket WHERE state = $1 LIMIT 1;", Unreserved).StructScan(ticket)
	if err != nil {
		return nil, err
	}

	return ticket, nil
}

func reservingTicket(db sqlx.Ext, ticketID, userID int) error {
	result, err := db.Exec(`
	UPDATE ticket
	SET user_id=$1,
		state=$2
	WHERE id=$3;
	`, userID, string(Reserved), ticketID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrapf(err, "failed to reserve ticket '%d'", ticketID)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("failed to reserve ticket %d, no rows affected: probably all tickets have been acquired", ticketID)
	}

	return nil
}

func (r *repoSQLite) IsAlreadyReservedErr(err error) bool {
	if err.Error() == "UNIQUE constraint failed: ticket.user_id" {
		return true
	}
	return false
}

func (r *repoSQLite) autoMigrate() *repoSQLite {
	q := `
	CREATE TABLE IF NOT EXISTS "ticket" (
		"id"	     INTEGER PRIMARY KEY AUTOINCREMENT,
		"cost"		 INTEGER NOT NULL DEFAULT 3000,
		"state"		 TEXT NOT NULL DEFAULT 'unreserved',
		"user_id"	 INTEGER DEFAULT NULL UNIQUE,
		"session_id" TEXT DEFAULT NULL,
		"created_at" TIMESTAMP NOT NULL DEFAULT current_timestamp,
		"updated_at" TIMESTAMP NOT NULL DEFAULT current_timestamp,
		FOREIGN KEY("user_id") REFERENCES user("id"),
		CHECK (("state" = 'unreserved' AND "user_id" IS NULL) OR ("state" IN ('reserved', 'booked') AND "user_id" IS NOT NULL))
	);

	CREATE TRIGGER IF NOT EXISTS ticket_updated_at
	AFTER UPDATE
	ON ticket FOR EACH ROW
	BEGIN
	  UPDATE ticket SET updated_at = current_timestamp
		WHERE id = old.id;
	END;
	`

	_, err := r.db.Exec(q)
	if err != nil {
		log.Fatalln("migration failed: " + err.Error())
	}

	return r
}

func (r *repoSQLite) initialPopulate() *repoSQLite {
	q := `
	DELETE FROM ticket;
	INSERT INTO ticket DEFAULT VALUES;
	INSERT INTO ticket DEFAULT VALUES;
	INSERT INTO ticket DEFAULT VALUES;
	INSERT INTO ticket DEFAULT VALUES;
	INSERT INTO ticket DEFAULT VALUES;
	`

	_, err := r.db.Exec(q)
	if err != nil {
		log.Fatalln("tickets population failed: " + err.Error())
	}

	return r
}
