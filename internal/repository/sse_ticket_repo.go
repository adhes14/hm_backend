package repository

import (
	"context"
	"errors"
	"time"

	"github.com/hospital_management/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type sseTicketRepository struct {
	db *pgxpool.Pool
}

func NewSSETicketRepository(db *pgxpool.Pool) SSETicketRepository {
	return &sseTicketRepository{db: db}
}

func (r *sseTicketRepository) Create(ctx context.Context, ticket *domain.SSETicket) error {
	query := `
		INSERT INTO sse_tickets (ticket, username, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, ticket.Ticket, ticket.Username, ticket.ExpiresAt)
	return err
}

func (r *sseTicketRepository) ValidateAndConsume(ctx context.Context, ticketStr string) (*domain.SSETicket, error) {
	query := `
		DELETE FROM sse_tickets
		WHERE ticket = $1 AND expires_at > $2
		RETURNING ticket, username, expires_at
	`
	var ticket domain.SSETicket
	err := r.db.QueryRow(ctx, query, ticketStr, time.Now()).Scan(
		&ticket.Ticket,
		&ticket.Username,
		&ticket.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invalid or expired ticket")
		}
		return nil, err
	}
	return &ticket, nil
}

func (r *sseTicketRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM sse_tickets
		WHERE expires_at <= $1
	`
	_, err := r.db.Exec(ctx, query, time.Now())
	return err
}
