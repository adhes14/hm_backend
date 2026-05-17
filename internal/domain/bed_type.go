package domain

type BedType struct {
	ID                          int    `json:"id"`
	Name                        string `json:"name"`
	Prefix                      string `json:"prefix"`
	RequiresPostpartumFollowup  bool   `json:"requires_postpartum_followup"`
}
