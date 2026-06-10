package domain

type Ward struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	BedCount    int    `json:"bed_count,omitempty"`
}
