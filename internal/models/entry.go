package models

import (
	"time"

	"github.com/jinzhu/gorm"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/rickb777/date/period"

	"github.com/matoous/mailback/internal/when"
)

// Entry is single mailing entry in mailback. Entry encapsulates all that is needed for the service to work.
// Entry contains information such as when should the email be send back and what the content should be.
type Entry struct {
	// ID is unique ID for the entry. ID can be used to unsubscribe for periodic entries.
	ID string `gorm:"primary_key"`
	// Data is the data that should be send back to the user.
	Data string
	// Title is the subject received in the initial request that will be send back to the user in subject.
	Title string
	// Mail is target mail (the initial sender of the email) to send the data to when the time comes.
	Mail string
	// ScheduledFor holds when the email is supposed to be send back to the user.
	// It is necessary to create the ScheduledFor timestamp at the API side instead of database because of
	// the periodical emails that would be really hard to process only on DB side.
	ScheduledFor time.Time
	// CreateAt is the time this entry was created at.
	CreatedAt time.Time
	// Period is optional period that the entry should be send at.
	Period *period.Period `gorm:"-"`
	// PeriodString is used to save the marshaled period into database.
	PeriodString *string
	// Fails counts the number of fails sending the email back to the user.
	Fails uint8
}

// NewEntry creates new entry from received data.
func NewEntry(from, content, title string, r *when.Result) (*Entry, error) {
	id, err := gonanoid.Nanoid()
	if err != nil {
		return nil, err
	}
	scheduledFor := r.Time
	if r.Period != nil {
		scheduledFor, _ = r.Period.AddTo(scheduledFor)
	}
	return &Entry{
		ID:           id,
		Data:         content,
		Mail:         from,
		Title:        title,
		ScheduledFor: scheduledFor,
		Period:       r.Period,
	}, nil
}

// BeforeSave converts period to PeriodString in order to save it into database.
func (e *Entry) BeforeSave() (err error) {
	if e.Period != nil {
		m := e.Period.String()
		e.PeriodString = &m
	}
	return
}

// BeforeUpdate converts period to PeriodString in order to save it into database.
func (e *Entry) BeforeUpdate(_ *gorm.Scope) (err error) {
	if e.Period != nil {
		m := e.Period.String()
		e.PeriodString = &m
	}
	return
}

// AfterFind tries to parse period (if entry has one) and sets it into the Period field of the Entry.
func (e *Entry) AfterFind() (err error) {
	if e.PeriodString != nil {
		p, err := period.Parse(*e.PeriodString)
		if err != nil {
			return err
		}
		e.Period = &p
	}
	return
}
