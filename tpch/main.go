package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func runQuery(tidbAddr string, tidbPort int, tidbDatabase string, queryFile string, engine string) (time.Duration, error) {
	cur := time.Now()
	var stderr bytes.Buffer

	queryContext, err := ioutil.ReadFile(queryFile)
	if err != nil {
		return time.Duration(0), err
	}

	query := string(queryContext)

	variables := "set @@session.tidb_allow_batch_cop = 1;set @@session.tidb_opt_distinct_agg_push_down = 1;set @@session.tidb_distsql_scan_concurrency = 30;set @@session.tidb_opt_agg_push_down = 0;"

	isolation := "set @@session.tidb_isolation_read_engines=\"" + engine + "\";"

	cmd := exec.Command("mysql",
		fmt.Sprintf("-h%v", tidbAddr),
		"-uroot",
		fmt.Sprintf("-P%v", tidbPort),
		fmt.Sprintf("-D%v", tidbDatabase),
		fmt.Sprintf("--local_infile"),
		"--comments",
		"-e", variables+isolation+query,
	)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("error occurred while running commmand: %v\nstderr: %+v\n", err, stderr.String())
		os.Exit(-1)
	}
	dur := time.Since(cur)
	return dur, nil
}

func main() {
	tidbAddrFlag := flag.String("addr", "127.0.0.1", "TiDB's address")
	tidbPortFlag := flag.Int("port", 4000, "TiDB's listening port")
	tpchDatabaseFlag := flag.String("database", "test", "The database of TPC-H dataset")
	tpchCountFlag := flag.Int("count", 3, "The count times factor of bench")
	queryDirFlag := flag.String("dir", "./queries", "The directory where the query SQLs are")

	flag.Parse()
	files, err := ioutil.ReadDir(*queryDirFlag)
	if err != nil {
		fmt.Printf("error occurred while reading directory: %v\n", err)
		os.Exit(-1)
	}

	// executing all queries.
	allEngines := []string{"tikv,tiflash", "tikv", "tiflash"}
	for _, engine := range allEngines {
		fmt.Printf("Running with engine=%s\n", engine)
		for _, file := range files {
			f := filepath.Join(*queryDirFlag, file.Name())
			fmt.Printf("%v\n", f)
			totCost := time.Duration(0)
			// each query run for tpchCountFlag times.
			for i := 0; i < *tpchCountFlag; i++ {
				dur, err := runQuery(*tidbAddrFlag, *tidbPortFlag, *tpchDatabaseFlag, f, engine)
				if err != nil {
					fmt.Printf("error occurred while executing query from %s, error: %s\n", f, err)
				}
				totCost += dur
				fmt.Printf("%v's %vth run finished in %v\n", file.Name(), i, dur)
			}
			fmt.Printf("%v costs: %v\n", file.Name(), 1.0*totCost.Milliseconds()/int64(*tpchCountFlag))
		}
	}
}
