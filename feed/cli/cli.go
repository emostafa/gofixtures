package cli

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/ishehata/gofixtures/entity"
	"github.com/ishehata/gofixtures/feed"
)

type feeder struct {
	filename   string
	dbConfFile string
	directory  string
	currentDir string
	AutoTables bool
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

// New creates a new instace of the CLI feeder
func New() feed.Feeder {
	feeder := &feeder{}
	feeder.readCommandLineFlags()
	feeder.currentDir, _ = filepath.Abs("./")
	return feeder
}

const defaultFixturesDirectory = "fixtures"
const defaultDBConfigFile = "db/dbconf.yaml"

// readCommandLineFlags reads the options supplied in the comamnd line
func (cli *feeder) readCommandLineFlags() {
	flag.StringVar(&cli.directory, "dir", defaultFixturesDirectory, "The path of the fixtures directory")
	flag.StringVar(&cli.filename, "file", "", "The path of a fixture file to load")
	flag.StringVar(&cli.dbConfFile, "dbconf", defaultDBConfigFile, "The path of dbconf file to load database configuration")
	flag.BoolVar(&cli.AutoTables, "autoTables", false, "Automatically create tables if they doesn't exists, false by default")

	flag.Parse()
}

// DatabaseConf generate the db configuration string. Make sure
// to call ReadCommandLineFlags() first, Because DatabaseConf()
// checks first if command line arguments has been passed
// if not it looks for dbconf.yaml in ./db folder.
func (cli *feeder) GetDBConf() (entity.DBConfigInput, error) {
	// determine the file type, e.g: yaml or json
	ext := filepath.Ext(cli.dbConfFile)

	input := entity.DBConfigInput{
		Type: ext,
	}

	cli.Print("reading database configuration file...")
	f, err := os.Open(cli.dbConfFile)
	if err != nil {
		return input, err
	}
	cli.Print("database configuration file has been loaded successfully")

	input.Data = f

	return input, nil
}

// GetInput reads the files selected by the user
// and return them as inputs
func (cli *feeder) GetInput() ([]entity.Input, error) {
	var inputs []entity.Input
	// get list of files that will be parsed
	files, err := cli.filesToParse()
	if err != nil {
		return inputs, err
	}
	inputs = make([]entity.Input, len(files))
	cli.Print(fmt.Sprintf("found %d fixture files", len(files)))
	for i, file := range files {
		cli.Print(fmt.Sprintf("reading fixture: %s", file))
		f, err := os.Open(file)
		if err != nil {
			return inputs, err
		}
		ext := filepath.Ext(file)
		input := entity.Input{
			Type: ext,
			Data: f,
		}
		inputs[i] = input
	}
	return inputs, nil
}

// FilesToParse checks if there is a filename is passed in the command line, If not,
// Check if a directory is passed, Else, Expect to find a dir named "fixtures" to load
// files form.
// Returns a list of string of filenames
func (cli *feeder) filesToParse() ([]string, error) {
	files := []string{}
	if cli.filename != "" {
		filename := path.Join(cli.currentDir, cli.filename)
		files = append(files, filename)
	} else {
		// if no file is selected then read all files in the fixtures directory
		// if the user didn't specify a directly, the default one will be used
		// which was set in flag parsing section
		fileinfos, err := ioutil.ReadDir(cli.directory)
		if err != nil {
			return files, err
		}
		for _, f := range fileinfos {
			filename := path.Join(cli.directory, f.Name())
			files = append(files, filename)
		}
	}
	return files, nil
}

// Print logs text to the end user
func (cli *feeder) Print(text string) {
	log.Println(text)
}

// Error prints and error to the user, exists if its a fatal error
func (cli *feeder) Error(err error, fatal bool) {
	txt := fmt.Sprintf("%#v", err)
	log.Println(err)
	if fatal {
		log.Fatal(txt)
		return
	}
	log.Println(txt)
}
