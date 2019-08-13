package ticket

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func (r *repoSQLite) startTaskRunner() {
	var err error
	for {
		select {
		case task := <-r.scheduler:
			time.AfterFunc(task.runAfter, func() {
				err = task.taskFn(r.db)
				if err != nil {
					log.Println("task runner: " + err.Error())
				}
			})
		}
	}
}

func (r *repoSQLite) scoutTasks() {
	rows, err := r.db.Queryx("SELECT id, updated_at FROM ticket WHERE state = 'reserved' AND updated_at <= $1;", time.Now())
	if err != nil {
		log.Fatalln("failed to query tasks from database:", err.Error())
	}

	for rows.Next() {
		var ticketID int
		var updatedAt time.Time
		err := rows.Scan(&ticketID, &updatedAt)
		if err != nil {
			log.Fatalln("failed to parse ticket while scouting tasks:", err.Error())
		}

		updatedAt = timeAtLocation(updatedAt, "Europe/Bucharest")
		timePassed := time.Now().Sub(updatedAt)
		expireAfter := maxReservedTime - timePassed

		log.Printf("initial scout: ticket='%d' scheduled to be unreserved in %s", ticketID, expireAfter.Round(time.Second))
		r.addTask(taskCleanUnreservedAfter(expireAfter, ticketID))
	}
}

func (r *repoSQLite) addTask(runAfter time.Duration, taskFn taskFunction) {
	r.scheduler <- task{
		runAfter: runAfter,
		taskFn:   taskFn,
	}
}

func taskCleanUnreservedAfter(duration time.Duration, ticketID int) (time.Duration, taskFunction) {
	return duration, func(db sqlx.Ext) error {
		_, err := db.Exec("UPDATE ticket SET state = 'unreserved', user_id = NULL, session_id = NULL WHERE id = $1 AND state != $2;", ticketID, Booked)
		if err != nil {
			return errors.Wrapf(err, "couldn't unreserve ticket '%d'", ticketID)
		}

		return nil
	}
}

func timeAtLocation(tm time.Time, location string) time.Time {
	loc, err := time.LoadLocation(location)
	if err != nil {
		log.Fatalln("failed to load time location:", err.Error())
	}

	return tm.In(loc)
}
