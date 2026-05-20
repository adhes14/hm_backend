package service

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hospital_management/backend/internal/domain"
	"github.com/hospital_management/backend/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SSEService interface {
	CreateTicket(ctx context.Context, username string) (string, error)
	ValidateTicket(ctx context.Context, ticketStr string) (*domain.SSETicket, error)
	RegisterClient(username string) (chan domain.SSEEvent, func())
	Broadcast(event domain.SSEEvent)
	StartWorker(ctx context.Context)
	GetSettings(ctx context.Context) (map[string]string, error)
	UpdateSettings(ctx context.Context, settings map[string]string) error
}

type sseService struct {
	pool        *pgxpool.Pool
	ticketRepo  repository.SSETicketRepository
	settingRepo repository.SettingsRepository

	mu      sync.RWMutex
	clients map[chan domain.SSEEvent]string // maps client channel to username
}

func NewSSEService(pool *pgxpool.Pool, ticketRepo repository.SSETicketRepository, settingRepo repository.SettingsRepository) SSEService {
	return &sseService{
		pool:        pool,
		ticketRepo:  ticketRepo,
		settingRepo: settingRepo,
		clients:     make(map[chan domain.SSEEvent]string),
	}
}

func (s *sseService) CreateTicket(ctx context.Context, username string) (string, error) {
	ticketStr := uuid.New().String()
	ticket := &domain.SSETicket{
		Ticket:    ticketStr,
		Username:  username,
		ExpiresAt: time.Now().Add(30 * time.Second),
	}
	err := s.ticketRepo.Create(ctx, ticket)
	if err != nil {
		return "", err
	}
	return ticketStr, nil
}

func (s *sseService) ValidateTicket(ctx context.Context, ticketStr string) (*domain.SSETicket, error) {
	return s.ticketRepo.ValidateAndConsume(ctx, ticketStr)
}

func (s *sseService) RegisterClient(username string) (chan domain.SSEEvent, func()) {
	ch := make(chan domain.SSEEvent, 100)
	s.mu.Lock()
	s.clients[ch] = username
	s.mu.Unlock()

	cleanup := func() {
		s.mu.Lock()
		delete(s.clients, ch)
		close(ch)
		s.mu.Unlock()
	}
	return ch, cleanup
}

func (s *sseService) Broadcast(event domain.SSEEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for ch := range s.clients {
		select {
		case ch <- event:
		default:
			// Channel full, drop event for this slow reader
		}
	}
}

func (s *sseService) GetSettings(ctx context.Context) (map[string]string, error) {
	settings, err := s.settingRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	res := make(map[string]string)
	for _, setting := range settings {
		res[setting.Key] = setting.Value
	}
	return res, nil
}

func (s *sseService) UpdateSettings(ctx context.Context, settings map[string]string) error {
	for k, v := range settings {
		err := s.settingRepo.Update(ctx, k, v)
		if err != nil {
			return err
		}
	}
	s.Broadcast(domain.SSEEvent{
		Type: "settings_updated",
		Data: settings,
	})
	return nil
}

func (s *sseService) StartWorker(ctx context.Context) {
	// Periodic clean of expired tickets
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = s.ticketRepo.DeleteExpired(context.Background())
			}
		}
	}()

	// Periodic check of active surveillance times
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Run initial check immediately
	s.runCheck(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runCheck(ctx)
		}
	}
}

func (s *sseService) runCheck(ctx context.Context) {
	soundControl := "true"
	soundDischarge := "false"

	settings, err := s.GetSettings(ctx)
	if err == nil {
		if val, ok := settings["sound_alert_control_overdue"]; ok {
			soundControl = val
		}
		if val, ok := settings["sound_alert_discharge_ready"]; ok {
			soundDischarge = val
		}
	}

	query := `
		SELECT 
			a.id, 
			a.patient_id, 
			p.full_name, 
			a.bed_id, 
			b.number, 
			bt.prefix, 
			a.next_control_at, 
			a.estimated_discharge_at
		FROM admissions a
		JOIN patients p ON a.patient_id = p.id
		JOIN beds b ON a.bed_id = b.id
		JOIN bed_types bt ON b.bed_type_id = bt.id
		WHERE a.status = 'active'
	`
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		log.Printf("SSE Worker error: %v", err)
		return
	}
	defer rows.Close()

	now := time.Now().UTC()
	for rows.Next() {
		var (
			admissionID          uuid.UUID
			patientID            uuid.UUID
			patientName          string
			bedID                int
			bedNumber            int
			bedPrefix            string
			nextControlAt        *time.Time
			estimatedDischargeAt *time.Time
		)
		err := rows.Scan(
			&admissionID,
			&patientID,
			&patientName,
			&bedID,
			&bedNumber,
			&bedPrefix,
			&nextControlAt,
			&estimatedDischargeAt,
		)
		if err != nil {
			log.Printf("SSE Scan error: %v", err)
			continue
		}

		if nextControlAt != nil && nextControlAt.Before(now) {
			s.Broadcast(domain.SSEEvent{
				Type: "control_overdue",
				Data: map[string]interface{}{
					"admission_id":    admissionID.String(),
					"patient_name":    patientName,
					"bed_id":          bedID,
					"bed_number":      bedNumber,
					"bed_prefix":      bedPrefix,
					"next_control_at": nextControlAt.Format(time.RFC3339),
					"sound":           soundControl == "true",
				},
			})
		}

		if estimatedDischargeAt != nil && estimatedDischargeAt.Before(now) {
			s.Broadcast(domain.SSEEvent{
				Type: "discharge_ready",
				Data: map[string]interface{}{
					"admission_id":           admissionID.String(),
					"patient_name":           patientName,
					"bed_id":                 bedID,
					"bed_number":             bedNumber,
					"bed_prefix":             bedPrefix,
					"estimated_discharge_at": estimatedDischargeAt.Format(time.RFC3339),
					"sound":                  soundDischarge == "true",
				},
			})
		}
	}
}
