package dbaser

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	//	"log"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Struct4db struct {
	Ctx        context.Context
	IsBase     bool
	MetricBase *pgx.Conn
}

func TableCreation(MetricBaseStruct *Struct4db) error {
	crea := "CREATE TABLE IF NOT EXISTS Gauge(metricname VARCHAR(30) PRIMARY KEY, value FLOAT8);"
	tag, err := MetricBaseStruct.MetricBase.Exec(MetricBaseStruct.Ctx, crea)
	if err != nil {
		return fmt.Errorf("error create Gauge table. Tag is \"%s\" error is %w", tag.String(), err)
	}
	crea = "CREATE TABLE IF NOT EXISTS Counter(metricname VARCHAR(30) PRIMARY KEY, value BIGINT);"
	tag, err = MetricBaseStruct.MetricBase.Exec(MetricBaseStruct.Ctx, crea)
	if err != nil {
		return fmt.Errorf("error create Counter table. Tag is \"%s\" error is %w", tag.String(), err)
	}
	return nil
}

func TableGetAllMetrics[MV int64 | float64](MetricBaseStruct *Struct4db, mappa *map[string]MV) error {
	tip := reflect.TypeOf(mappa).String()
	var zapros string
	var value MV
	var str string
	switch tip {
	case "*map[string]int64":
		zapros = "SELECT * FROM counter;"
	case "*map[string]float64":
		zapros = "SELECT * FROM gauge;"
	default:
		return fmt.Errorf("wrong type %s", tip)
	}
	rows, err := MetricBaseStruct.MetricBase.Query(MetricBaseStruct.Ctx, zapros)
	if err != nil {
		return fmt.Errorf("error Query %[2]s:%[3]d database  %[1]w", err,
			MetricBaseStruct.MetricBase.Config().Host, MetricBaseStruct.MetricBase.Config().Port)
	}
	for rows.Next() {
		err = rows.Scan(&str, &value)
		if err != nil {
			return fmt.Errorf("error counter table Scan %[2]s:%[3]d database\n%[1]w", err,
				MetricBaseStruct.MetricBase.Config().Host, MetricBaseStruct.MetricBase.Config().Port)
		}
		(*mappa)[str] = value
	}
	return nil
}

func TablePutMetric(MetricBaseStruct *Struct4db, metr *Metrics) error {
	//	(X || Y) && !(X && Y)
	XoR := (metr.Value == nil && metr.Delta == nil) || (metr.Value != nil && metr.Delta != nil)
	if XoR {
		return fmt.Errorf("metr.Value %v & metr.Delta %v have both nil/notnil values - impossible %v", metr.Value, metr.Delta)
	}
	var orderInsert, orderUpdate string
	if metr.Value != nil { // GAUGE case
		orderInsert = fmt.Sprintf("INSERT INTO Gauge(metricname, value) VALUES ('%[1]s',%[2]g);", metr.ID, *metr.Value)
		orderUpdate = fmt.Sprintf("UPDATE Gauge SET value=%[2]g WHERE metricname='%[1]s'", metr.ID, *metr.Value)
	} else { // COUNTER case
		orderInsert = fmt.Sprintf("INSERT INTO counter(metricname, value) VALUES ('%[1]s',%[2]d);", metr.ID, *metr.Delta)
		// for COUNTERs value=value+%[2]d !!!!
		orderUpdate = fmt.Sprintf("UPDATE counter SET value=value+%[2]d WHERE metricname='%[1]s'", metr.ID, *metr.Delta)
	}
	tag1, err := MetricBaseStruct.MetricBase.Exec(MetricBaseStruct.Ctx, orderInsert)
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code != pgerrcode.UniqueViolation {
		return fmt.Errorf("error Insert %+v TagInsert is \"%s\" *pgconn.PgError %+v error is %w",
			metr, tag1.String(), pgErr, err)
	}
	tag2, err := MetricBaseStruct.MetricBase.Exec(MetricBaseStruct.Ctx, orderUpdate)
	if err == nil {
		return nil
	}
	return fmt.Errorf("error UPDATE %+v TagInsert is \"%s\" TagUpdate is \"%s\" error is %w",
		metr, tag1.String(), tag2.String(), err)
}

func TableGetMetric(MetricBaseStruct *Struct4db, metr *Metrics) error {
	str := fmt.Sprintf("SELECT value FROM %s WHERE metricname = $1;", metr.MType)
	row := MetricBaseStruct.MetricBase.QueryRow(MetricBaseStruct.Ctx, str, metr.ID)
	switch metr.MType {
	case "counter":
		err := row.Scan(metr.Delta)
		if err != nil {
			return fmt.Errorf("error get %s %s metric.  %w", metr.ID, metr.MType, err)
		}
	case "gauge":
		err := row.Scan(metr.Value)
		if err != nil {
			return fmt.Errorf("error get %s %s metric.  %w", metr.ID, metr.MType, err)
		}
	default:
		return fmt.Errorf("error get %s gauge metric.", metr.ID)
	}
	return nil
}

type Gauge float64
type Counter int64
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func TableBuncher(MetricBaseStruct *Struct4db, metrArray []Metrics) error {
	tx, err := MetricBaseStruct.MetricBase.Begin(MetricBaseStruct.Ctx)
	if err != nil {
		return fmt.Errorf("error db.Begin  %[1]w", err)
	}
	var order string
	for _, metrica := range metrArray {
		if metrica.MType == "gauge" {
			order = fmt.Sprintf("UPDATE gauge SET value=%[2]g WHERE metricname='%[1]s'", metrica.ID, *metrica.Value)
		} else {
			order = fmt.Sprintf("UPDATE counter SET value=value+%[2]d WHERE metricname='%[1]s'", metrica.ID, *metrica.Delta)
		}
		tagUpdate, _ := tx.Exec(MetricBaseStruct.Ctx, order)
		tu := tagUpdate.RowsAffected()
		if tu != 0 { // если удалось записать - уже существует и INSERT не нужен
			continue
		}
		if metrica.MType == "gauge" {
			order = fmt.Sprintf("INSERT INTO Gauge(metricname, value) VALUES ('%[1]s',%[2]g);", metrica.ID, *metrica.Value)
		} else {
			order = fmt.Sprintf("INSERT INTO counter(metricname, value) VALUES ('%[1]s',%[2]d);", metrica.ID, *metrica.Delta)
		}
		tagInsert, err := tx.Exec(MetricBaseStruct.Ctx, order)
		if err != nil {
			return fmt.Errorf("TableBuncher error UPDATE Metric %-v TagInsert is \"%s\" TagUpdate is \"%s\" error is %w",
				metrica, tagInsert.String(), tagUpdate.String(), err)
		}
	}
	return tx.Commit(MetricBaseStruct.Ctx)
}
