package store

import (
	"github.com/jmoiron/sqlx"
)

// Store aggregates all store interfaces
type Store struct {
	User      UserStore
	Household HouseholdStore
	Task      TaskStore
	Shopping  ShoppingStore
	Bill      BillStore
	Timer     TimerStore
	EventLog  EventLogStore
}

// Stores is an alias for Store to maintain compatibility
type Stores struct {
	Users      UserStore
	Households HouseholdStore
	Tasks      TaskStore
	Shopping   ShoppingStore
	Bills      BillStore
	Timers     TimerStore
	EventLog   EventLogStore
}

// NewStore creates a new store instance with all sub-stores
func NewStore(db *sqlx.DB) *Store {
	return &Store{
		User:      NewUserStore(db),
		Household: NewHouseholdStore(db),
		Task:      NewTaskStore(db),
		Shopping:  NewShoppingStore(db),
		Bill:      NewBillStore(db),
		Timer:     NewTimerStore(db),
		EventLog:  NewEventLogStore(db),
	}
}
