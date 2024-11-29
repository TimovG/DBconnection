package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xuri/excelize/v2"
)

type CPUMetric struct {
	Timestamp string
	Queries   int64
}

func CPUMetricVanDB(db *sql.DB) ([]CPUMetric, error) {
	query := `
		SHOW GLOBAL STATUS LIKE 'Queries';
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []CPUMetric
	for rows.Next() {
		var metric CPUMetric
		var name string
		if err := rows.Scan(&name, &metric.Queries); err != nil {
			return nil, err
		}
		metric.Timestamp = time.Now().Format(time.RFC3339) // Huidige tijd als timestamp
		metrics = append(metrics, metric)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}

func exportToExcel(metrics []CPUMetric, filePath string) error {
	f := excelize.NewFile()
	sheet := "Sheet1"

	// Headers toevoegen
	f.SetCellValue(sheet, "A1", "Timestamp")
	f.SetCellValue(sheet, "B1", "Queries")

	// Data toevoegen
	for i, metric := range metrics {
		f.SetCellValue(sheet, fmt.Sprintf("A%d", i+2), metric.Timestamp)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", i+2), fmt.Sprintf("%d", metric.Queries))
	}

	// het bestand op te slaan
	err := f.SaveAs(filePath)
	if err != nil {
		return err
	}

	log.Println("Excel file saved successfully!")
	return nil
}

func main() {
	// Verbinding maken met de database
	username := "dbadmin"
	password := "test12345!"
	hostname := "newyork.mysql.database.azure.com"
	port := "3306"
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?tls=true", username, password, hostname, port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}
	defer db.Close()

	// Haal CPU-statistieken op (gebruik de status van de server als indicator)
	metrics, err := CPUMetricVanDB(db)
	if err != nil {
		log.Fatalf("Error fetching CPU metrics: %v", err)
	}

	// Exporteer statistieken naar Excel
	outputFile := "./cpu_metrics.xlsx"
	err = exportToExcel(metrics, outputFile)
	if err != nil {
		log.Fatalf("Error exporting to Excel: %v", err)
	}

	fmt.Printf("Metrics exported to: %s\n", outputFile)
}
