package dbaser

import (
	"context"
	"fmt"

	//	"log"

	"github.com/jackc/pgx/v5"
)

func TableCreation(ctx context.Context, db *pgx.Conn) error {

	//log.Printf("TABLE CREATION %v\n", db)

	crea := "CREATE TABLE IF NOT EXISTS Gauge(metricname VARCHAR(30) PRIMARY KEY, value FLOAT8);"
	tag, err := db.Exec(ctx, crea)
	if err != nil {
		return fmt.Errorf("error create Gauge table. Tag is \"%s\" error is %w", tag.String(), err)
	}
	crea = "CREATE TABLE IF NOT EXISTS Counter(metricname VARCHAR(30) PRIMARY KEY, value BIGINT);"
	tag, err = db.Exec(ctx, crea)
	if err != nil {
		return fmt.Errorf("error create Counter table. Tag is \"%s\" error is %w", tag.String(), err)
	}
	return nil
}

func TablePutGauge(ctx context.Context, db *pgx.Conn, mname string, value float64) error {

	order := fmt.Sprintf("INSERT INTO Gauge(metricname, value) VALUES ('%[1]s',%[2]f);", mname, value)
	tag1, err := db.Exec(ctx, order)

	//	log.Printf("TableInsertGauge err %v\n db %v\n\n", err, db)
	if err == nil {
		return nil
	}
	order = fmt.Sprintf("UPDATE Gauge SET value=%[2]f WHERE metricname='%[1]s'", mname, value)
	tag2, err := db.Exec(ctx, order)
	//	log.Printf("TableUpdateGauge err %v\n db %v\n\n", err, db)
	if err == nil {
		return nil
	}
	return fmt.Errorf("error UPDATE Gauge %s with %f value. TagInsert is \"%s\" TagUpdate is \"%s\" error is %w",
		mname, value, tag1.String(), tag2.String(), err)
}

func TableGetGauge(ctx context.Context, db *pgx.Conn, mname string) (float64, error) {
	var flo float64
	str := "SELECT value FROM gauge WHERE metricname = $1;"
	row := db.QueryRow(ctx, str, mname)
	err := row.Scan(&flo)
	if err != nil {
		return 0.0, fmt.Errorf("error get %s gauge metric.  %w", mname, err)
	}
	return flo, nil
}

func TablePutCounter(ctx context.Context, db *pgx.Conn, mname string, value int64) error {
	order := fmt.Sprintf("INSERT INTO Counter(metricname, value) VALUES ('%[1]s',%[2]d);", mname, value)
	tag1, err := db.Exec(ctx, order)
	if err == nil {
		return nil
	}
	order = fmt.Sprintf("UPDATE Counter SET value=%[2]d WHERE metricname='%[1]s'", mname, value)
	tag2, err := db.Exec(ctx, order)
	if err == nil {
		return nil
	}
	return fmt.Errorf("error UPDATE Counter %s with %d value. TagInsert is \"%s\" TagUpdate is \"%s\" error is %w",
		mname, value, tag1.String(), tag2.String(), err)
}

func TableGetCounter(ctx context.Context, db *pgx.Conn, mname string) (int64, error) {
	var inta int64
	str := "SELECT value FROM counter WHERE metricname = $1;"
	row := db.QueryRow(ctx, str, mname)
	err := row.Scan(&inta)
	if err != nil {
		return 0, fmt.Errorf("error get %s counter metric.  %w", mname, err)
	}
	return inta, nil
}
