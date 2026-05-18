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
	var count int
	err = pool.QueryRow(context.Background(), "SELECT count(*) FROM staff").Scan(&count)
	if err != nil {
		fmt.Println("Query error:", err)
		return
	}
	fmt.Println("Staff count:", count)
}
