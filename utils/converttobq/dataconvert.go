// gets the filepath of the results file as an argument to the program
// gets the toolname as an argument to the program
// checks if the results file exists and the size of it,
// if it exists, for the appropriate tool, follows its workflow and
// converts the results file into BQ ingest-able format
// continues to upload the results to a BQ dataset

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"

	"cloud.google.com/go/bigquery"

	"github.com/astaxie/flatmap"
)

var ks []string //for repo-supervisor

type error interface {
	Error() string
}

var (
	projectID   string
	rsDS        string
	rsT         string
	wfDS        string
	wfT         string
	toolName    = flag.String("toolName", "", "Tool name for which the data will be converted")
	filePath    = flag.String("filePath", "", "File path of the results file to convert")
	datasetName string
	tableName   string
	schema      bigquery.Schema
	f           *os.File
	fp          string
)

func exists(path string) (bool, int64, error) {
	fi, err := os.Stat(path)
	if err == nil {
		return true, fi.Size(), nil
	}
	if os.IsNotExist(err) {
		return false, int64(0), nil
	}
	return false, int64(0), err
}

//CheckIfError function to check for errors
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

//Info function to print pretty output
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

//flatteneachRes function to flatten ecah result JSON array for the tool repo-supervisor
func flatteneachRes(result string, wg *sync.WaitGroup) {
	var mp map[string]interface{}

	//reading the result string into a map of string interface
	err := json.Unmarshal([]byte(result), &mp)
	CheckIfError(err)

	//flattening that map
	fm, err := flatmap.Flatten(mp)
	CheckIfError(err)

	//Ranging through the flattened map and creating a string of File|Secret and adding to the ks string array
	for k := range fm {
		ks = append(ks, k+"|"+fm[k])
	}

	wg.Done()
}

func main() {

	//Loading the environment variables
	// err := gotenv.Load(".env")
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	flag.Parse()

	projectID = os.Getenv("PROJECT_ID")
	rsDS = os.Getenv("RS_DATASET_NAME")
	rsT = os.Getenv("RS_TABLE_NAME")
	wfDS = os.Getenv("WF_DATASET_NAME")
	wfT = os.Getenv("WF_TABLE_NAME")

	// instantiating the BQ client
	ctx := context.Background()
	bqclient, err := bigquery.NewClient(ctx, projectID)
	CheckIfError(err)

	//getting the filepath as a command line argument
	Info("Filepath: " + *filePath + "\n")

	value := false
	fsize := int64(0)

	//checking for file existence and file size
	for (value == false) || (fsize == int64(0)) {
		i, s, err := exists(*filePath)
		CheckIfError(err)
		value = i
		fsize = s
	}

	fmt.Println("File exists:", value)
	fmt.Println("File size:", fsize)
	fmt.Println("")
	//now, we know that the file exists and that its size>0

	//Opening the results file to read each line
	resultsFile, err := os.Open(*filePath)
	CheckIfError(err)
	defer resultsFile.Close()

	switch *toolName {
	case "repo-supervisor":
		//Instantiating the bufio scanner to read the results file
		resScanner := bufio.NewScanner(resultsFile)

		//For each result object being read, pass it to the flatteneachRes function
		var wg sync.WaitGroup
		for resScanner.Scan() {
			wg.Add(1)
			res := resScanner.Text()
			go flatteneachRes(res, &wg)
		}
		wg.Wait()

		//Creating a temp csv file to store the final results
		outputFile, err := os.Create("/tmp/final.csv")
		CheckIfError(err)
		defer outputFile.Close()

		//Iterating through the string array and writing each one of them in the above CSV file
		for _, el := range ks {
			outputFile.WriteString(el + "\n")
		}

		datasetName = rsDS
		tableName = rsT

		//defining the schema
		schema = bigquery.Schema{
			&bigquery.FieldSchema{Name: "File", Required: false, Type: bigquery.StringFieldType},
			&bigquery.FieldSchema{Name: "Secret", Repeated: false, Type: bigquery.StringFieldType},
		}

		fp = "/tmp/final.csv"

	case "wfuzz":
		datasetName = wfDS
		tableName = wfT

		//defining the schema
		schema = bigquery.Schema{
			&bigquery.FieldSchema{Name: "ID", Required: false, Type: bigquery.StringFieldType},
			&bigquery.FieldSchema{Name: "Response", Repeated: false, Type: bigquery.StringFieldType},
			&bigquery.FieldSchema{Name: "Lines", Repeated: false, Type: bigquery.StringFieldType},
			&bigquery.FieldSchema{Name: "Word", Repeated: false, Type: bigquery.StringFieldType},
			&bigquery.FieldSchema{Name: "Chars", Repeated: false, Type: bigquery.StringFieldType},
			&bigquery.FieldSchema{Name: "Request", Repeated: false, Type: bigquery.StringFieldType},
			&bigquery.FieldSchema{Name: "Success", Repeated: false, Type: bigquery.StringFieldType},
		}

		fp = *filePath

	}

	Info("Now, uploading the results to BigQuery...")
	f, err = os.Open(fp) //Need to open the file before uploading
	CheckIfError(err)
	defer f.Close()

	//reading the processed file into BQ reader
	rs := bigquery.NewReaderSource(f)
	rs.AllowJaggedRows = true
	rs.AllowQuotedNewlines = true
	switch *toolName {
	case "repo-supervisor":
		rs.FieldDelimiter = "|"
	case "wfuzz":
		rs.FieldDelimiter = ","
		rs.SkipLeadingRows = 1
	}
	rs.IgnoreUnknownValues = true
	rs.Schema = schema

	//instantiating the dataset
	ds := bqclient.Dataset(datasetName)
	loader := ds.Table(tableName).LoaderFrom(rs) //loading the results
	loader.CreateDisposition = bigquery.CreateNever

	//checking the job status
	job, err := loader.Run(ctx)
	CheckIfError(err)
	status, err := job.Wait(ctx)
	CheckIfError(err)
	err = status.Err()
	CheckIfError(err)
	fmt.Println("Done")

}
