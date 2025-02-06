package model

type Rule struct {
	ID       string    `json:"id"`
	Strategy *Strategy `json:"strategy"`
	Clauses  []Clause  `json:"clauses"`
}
