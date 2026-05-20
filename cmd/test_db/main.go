package main

import (
	"context"
	"fmt"

	"github.com/hospital_management/backend/internal/database"
)

func main() {
	pool, err := database.NewPool(context.Background())
	if err != nil {
		fmt.Println("DB error:", err)
		return
	}
	defer pool.Close()

	keys := []string{
		"sound_alert_control_overdue",
		"sound_alert_discharge_ready",
		"sound_alert_patient_admitted",
		"sound_alert_patient_discharged",
	}

	query := `
		INSERT INTO system_settings (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = $2
	`

	for _, k := range keys {
		_, err := pool.Exec(context.Background(), query, k, "false")
		if err != nil {
			fmt.Printf("Error inserting key %q: %v\n", k, err)
		} else {
			fmt.Printf("Successfully upserted key %q\n", k)
		}
	}
}
