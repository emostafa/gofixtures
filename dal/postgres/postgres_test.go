package postgres

import (
	"fmt"
	"os"
	"testing"

	"github.com/ishehata/gofixtures/entity"
)

func prepareTestData(numberOfRecords int) entity.Fixture {
	records := make([]map[string]interface{}, numberOfRecords)
	for i := 0; i < numberOfRecords; i++ {
		records[i] = map[string]interface{}{
			"name": fmt.Sprintf("Product %d", i),
			"slug": fmt.Sprintf("product_%d", i),
		}
	}
	fixture := entity.Fixture{
		Table:   "products",
		Records: records,
	}

	return fixture
}

func TestInsertion(t *testing.T) {
	fixture := prepareTestData(100)

	dbConfig := entity.DBConfig{
		Driver:           "postgres",
		Database:         os.Getenv("GOFIXTURES_TEST_DB_NAME"),
		User:             os.Getenv("GOFIXTURES_TEST_DB_USER"),
		Password:         os.Getenv("GOFIXTURES_TEST_DB_PASSWORD"),
		Host:             os.Getenv("GOFIXTURES_TEST_DB_HOST"),
		AutoCreateTables: true,
	}
	datastore := New(dbConfig)
	err := datastore.Connect()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	err = datastore.Insert(fixture)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
}

func BenchmarkInsertion(b *testing.B) {

	dbConfig := entity.DBConfig{
		Driver:   "postgres",
		Database: os.Getenv("GOFIXTURES_TEST_DB_NAME"),
		User:     os.Getenv("GOFIXTURES_TEST_DB_USER"),
		Password: os.Getenv("GOFIXTURES_TEST_DB_PASSWORD"),
		Host:     os.Getenv("GOFIXTURES_TEST_DB_HOST"),
	}
	datastore := New(dbConfig)
	err := datastore.Connect()
	if err != nil {
		b.Error(err)
		b.Fail()
	}
	for i := 0; i < b.N; i++ {
		fixture := prepareTestData(b.N)
		err := datastore.Insert(fixture)
		if err != nil {
			b.Error(err)
			b.Fail()
		}
	}
}
