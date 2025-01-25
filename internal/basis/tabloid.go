package basis

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"

	_ "database/sql"

	_ "github.com/lib/pq"

	"gorono/internal/models"
)

type Metrics = models.Metrics

type DBstruct struct {
	DB *pgx.Conn
}

func InitDBStorage(ctx context.Context, dbEndPoint string) (*DBstruct, error) {
	dbStorage := &DBstruct{}
	baza, err := pgx.Connect(ctx, dbEndPoint)
	if err != nil {
		return nil, fmt.Errorf("can't connect to DB %s err %w", dbEndPoint, err)
	}
	err = TableCreation(ctx, baza)
	if err != nil {
		return nil, fmt.Errorf("can't create tables in DB %s err %w", dbEndPoint, err)
	}
	dbStorage.DB = baza
	return dbStorage, nil
}

func TableCreation(ctx context.Context, db *pgx.Conn) error {
	crea := "DROP TABLE Counter;"
	crea += "DROP TABLE Gauge;"
	crea += "CREATE TABLE IF NOT EXISTS Gauge(metricname VARCHAR(50) PRIMARY KEY, value FLOAT8);"
	tag, err := db.Exec(ctx, crea)
	if err != nil {
		return fmt.Errorf("error create Gauge table. Tag is \"%s\" error is %w", tag.String(), err)
	}
	crea = "CREATE TABLE IF NOT EXISTS Counter(metricname VARCHAR(50) PRIMARY KEY, value BIGINT);"
	tag, err = db.Exec(ctx, crea)
	if err != nil {
		return fmt.Errorf("error create Counter table. Tag is \"%s\" error is %w", tag.String(), err)
	}
	return nil
}

// -------------- put ONE metric to the table
func (dataBase *DBstruct) PutMetric(ctx context.Context, metr *Metrics) error {
	if !models.IsMetricsOK(*metr) {
		return fmt.Errorf("bad metric %+v", metr)
	}
	db := dataBase.DB
	var order string
	switch metr.MType {
	case "gauge":
		order = fmt.Sprintf("INSERT INTO Gauge AS args(metricname, value) VALUES ('%[1]s',%[2]g) ", metr.ID, *metr.Value)
		order += "ON CONFLICT (metricname) DO UPDATE SET metricname=args.metricname, value=EXCLUDED.value;"
	case "counter":
		order = fmt.Sprintf("INSERT INTO Counter AS args(metricname, value) VALUES ('%[1]s',%[2]d) ", metr.ID, *metr.Delta)
		order += "ON CONFLICT (metricname) DO UPDATE SET metricname=args.metricname, value=args.value+EXCLUDED.value;"
		// args.value - старое значение. EXCLUDED.value - новое, переданное для вставки или обновления
	default:
		return fmt.Errorf("wrong type %s", metr.MType)
	}
	_, err := db.Exec(ctx, order)
	if err != nil {
		return fmt.Errorf("error insert/update %+v error is %w", metr, err)
	}
	return nil
}

// ------ get ONE metric from the table
func (dataBase *DBstruct) GetMetric(ctx context.Context, metr *Metrics) (Metrics, error) {
	db := dataBase.DB
	metrix := Metrics{ID: metr.ID, MType: metr.MType} // new pure Metrics to return, nil Delta & Value ptrs
	switch metr.MType {
	case "gauge":
		var flo float64 // here we scan Value
		order := "SELECT value FROM gauge WHERE metricname = $1;"
		row := db.QueryRow(ctx, order, metr.ID)
		err := row.Scan(&flo)
		if err != nil {
			return *metr, fmt.Errorf("unknown metric %+v", metr)
		}
		metrix.Value = &flo
	case "counter":
		var inta int64 // here we scan Delta
		order := "SELECT value FROM counter WHERE metricname = $1;"
		row := db.QueryRow(ctx, order, metr.ID)
		err := row.Scan(&inta)
		if err != nil {
			return *metr, fmt.Errorf("unknown metric %+v", metr)
		}
		metrix.Delta = &inta
	default:
		return *metr, fmt.Errorf("wrong type %s", metr.MType)
	}
	return metrix, nil
}

// ----------- transaction. PUT ALL metrics to the tables ----------------------
func (dataBase *DBstruct) PutAllMetrics(ctx context.Context, metras *[]Metrics) error {
	db := dataBase.DB
	//func TableBuncher(ctx context.Context, db *pgx.Conn, metras *[]Metrics) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error db.Begin  %[1]w", err)
	}
	var order string
	for _, metr := range *metras {
		if !models.IsMetricsOK(metr) {
			log.Printf("wrong metric %+v", metr)
			continue
		}
		switch metr.MType {
		case "gauge":
			order = fmt.Sprintf("INSERT INTO Gauge AS args(metricname, value) VALUES ('%[1]s',%[2]g) ", metr.ID, *metr.Value)
			order += "ON CONFLICT (metricname) DO UPDATE SET metricname=args.metricname, value=EXCLUDED.value;"
		case "counter":
			order = fmt.Sprintf("INSERT INTO Counter AS args(metricname, value) VALUES ('%[1]s',%[2]d) ", metr.ID, *metr.Delta)
			order += "ON CONFLICT (metricname) DO UPDATE SET metricname=args.metricname, value=args.value+EXCLUDED.value;"
			// args.value - старое значение. EXCLUDED.value - новое, переданное для вставки или обновления
		default:
			log.Printf("wrong metric type \"%s\"\n", metr.MType)
			continue
		}
		_, err := tx.Exec(ctx, order)
		if err != nil {
			defer tx.Rollback(ctx)
			log.Printf("error put %+v. error is %v", metr, err)
			return err
		}
	}
	return tx.Commit(ctx)
}

// ------- get ALL metrics from the tables
func (dataBase *DBstruct) GetAllMetrics(ctx context.Context) (*[]Metrics, error) {
	db := dataBase.DB
	//func TableGetAllTables(ctx context.Context, db *pgx.Conn, metras *[]Metrics) error {
	zapros := `select 'counter' AS metrictype, metricname AS name, null AS value, value AS delta from counter
		UNION
	select 'gauge' AS metrictype, metricname as name, value as value, null as delta from gauge`

	var inta int64
	var flo float64
	metr := Metrics{ID: "", MType: "", Value: &flo, Delta: &inta}

	rows, err := db.Query(ctx, zapros)
	if err != nil {
		return nil, fmt.Errorf("error Query %[2]s:%[3]d  %[1]w", err, db.Config().Host, db.Config().Port)
	}
	metras := []Metrics{}
	for rows.Next() {
		err = rows.Scan(&metr.MType, &metr.ID, &metr.Value, &metr.Delta)
		if err != nil {
			return nil, fmt.Errorf("error table Scan %[2]s:%[3]d  %[1]w", err, db.Config().Host, db.Config().Port)
		}
		metras = append(metras, metr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("err := rows.Err()  %w", err)
	}
	return &metras, nil
}
func (dataBase *DBstruct) LoadMS(fnam string) error {
	return nil
}
func (dataBase *DBstruct) SaveMS(fnam string) error {
	return nil
}
func (dataBase *DBstruct) Saver(fnam string, i int) error {
	return nil
}
func (dataBase *DBstruct) Ping(ctx context.Context) error {
	err := dataBase.DB.Ping(ctx)
	if err != nil {
		log.Printf("No PING err %+v\n", err)
		return fmt.Errorf("no ping %w", err)
	}
	return nil
}
