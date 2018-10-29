package postgres

import (
	"fmt"
	"log"
	"strings"

	"github.com/ishehata/gofixtures/dal"
	"github.com/ishehata/gofixtures/entity"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgresql driver
)

// New Creates
func New(config entity.DBConfig) dal.Datastore {
	return &postgresDatastore{
		config: config,
	}
}

type postgresDatastore struct {
	db     *sqlx.DB
	config entity.DBConfig
}

// Connect connects to postgresql db based on parameters of DBConfig object
func (datastore *postgresDatastore) Connect() error {
	connString := buildConnectionString(datastore.config)
	// open connection and ping
	db, err := sqlx.Connect("postgres", connString)
	if err != nil {
		return err
	}
	datastore.db = db
	return nil
}

func (datastore *postgresDatastore) createTable(tableName string, columns []string) error {
	columnsDef := strings.Join(columns, " VARCHAR, ")
	columnsDef += " VARCHAR"
	q := fmt.Sprintf("CREATE TABLE IF NOT EXISTS public.%s (%s)", tableName, columnsDef)
	_, err := datastore.db.Exec(q)
	return err
}

func keys(m map[string]interface{}) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

func (datastore *postgresDatastore) Insert(fixture entity.Fixture) error {
	tx, err := datastore.db.Begin()
	if err != nil {
		return err
	}
	if datastore.config.AutoCreateTables {
		columnsList := keys(fixture.Records[0])
		if err := datastore.createTable(fixture.Table, columnsList); err != nil {
			log.Fatal(err)
		}
	}
	for _, record := range fixture.Records {
		query := buildNamedQuery(fixture.Table, record)
		_, err := datastore.db.NamedExec(query, record)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (datastore *postgresDatastore) Close() {
	datastore.db.Close()
}

func buildConnectionString(conf entity.DBConfig) string {
	return fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable",
		conf.Host, conf.Database, conf.User, conf.Password)
}

func buildNamedQuery(table string, record map[string]interface{}) string {
	cols := getColumns(record)
	colsStr := ""
	valuesStr := ""
	for i, c := range cols {
		colsStr += c
		valuesStr += ":" + c
		if i != len(cols)-1 {
			colsStr += ","
			valuesStr += ","
		}
	}
	q := fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)", table, colsStr, valuesStr)
	return q
}

func getColumns(record map[string]interface{}) []string {
	cols := []string{}
	for key := range record {
		cols = append(cols, key)
	}
	return cols
}
