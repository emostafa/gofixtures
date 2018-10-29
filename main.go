package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/ishehata/gofixtures/dal"
	"github.com/ishehata/gofixtures/dal/postgres"
	"github.com/ishehata/gofixtures/feed/cli"
	"github.com/ishehata/gofixtures/parser"
	"github.com/ishehata/gofixtures/parser/json"
	"github.com/ishehata/gofixtures/parser/yaml"
)

var queries []string

const version = "2.0.0"

func getParser(forType string) (parser.Parser, error) {
	switch forType {
	case ".json":
		return json.New(), nil
	case ".yaml":
		return yaml.New(), nil
	default:
		return nil, errors.New("unsupported input type, supported types are YAML and JSON")
	}
}

func main() {
	cmdArgs := os.Args
	if cmdArgs[1] == "version" {
		log.Printf("version: %s", version)
		return
	}
	// read input using CLI
	feeder := cli.New()
	dbConfInput, err := feeder.GetDBConf()
	if err != nil {
		feeder.Error(err, true)
	}

	dbConfParser, err := getParser(dbConfInput.Type)
	if err != nil {
		feeder.Print("failed to parse database configuration")
		feeder.Error(err, true)
	}
	dbConf, err := dbConfParser.ParseDBConf(dbConfInput.Data)
	if err != nil {
		feeder.Error(err, true)
	}

	// connect to database
	var datastore dal.Datastore
	switch dbConf.Driver {
	case "postgres":
		datastore = postgres.New(dbConf)
	default:
		feeder.Error(errors.New("unsupported database driver"), true)
	}
	feeder.Print("attempting to connect to datastore...")
	err = datastore.Connect()
	if err != nil {
		feeder.Error(err, true)
	}
	defer datastore.Close()
	feeder.Print("Connection to datastore has been established")

	// get the data that needs to be parsed
	feeder.Print("loading fixture files...")
	input, err := feeder.GetInput()
	if err != nil {
		feeder.Error(err, true)
	}

	// based on type of the data, determine which parser is going to be used
	for _, i := range input {
		p, err := getParser(i.Type)
		if err != nil {
			feeder.Error(err, true)
		}
		// parse the input
		data, err := p.Parse(i.Data)
		if err != nil {
			feeder.Print("Failed to parse input, proceeding to next input")
			feeder.Error(err, false)
			continue
		}
		err = datastore.Insert(data)
		if err != nil {
			feeder.Print("Failed to insert to datastore, " + err.Error())
			continue
		}
	}

	fmt.Println("Finished Parsing all the inputs...")
}
