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

	//variables := "set @@session.tidb_allow_batch_cop = 1;set @@session.tidb_opt_distinct_agg_push_down = 1;set @@session.tidb_distsql_scan_concurrency = 30;set @@session.tidb_opt_agg_push_down = 0;"
    variables := "set @@session.tidb_allow_batch_cop = 1;set @@session.tidb_opt_distinct_agg_push_down = 1;set @@session.tidb_distsql_scan_concurrency = 30;set @@session.tidb_projection_concurrency = 16;set @@session.tidb_hashagg_partial_concurrency = 16;set @@session.tidb_hashagg_final_concurrency = 16;set @@session.tidb_hash_join_concurrency = 16;set @@session.tidb_index_lookup_concurrency = 16;set @@session.tidb_index_lookup_join_concurrency = 16;"

	isolation := "set @@session.tidb_isolation_read_engines=\"" + engine + "\";"

    //sql := variables + isolation + query;
    //fmt.Printf("Query: %s\n", sql)
    fmt.Printf("Variables: %s\n", variables + isolation)

	cmd := exec.Command("mysql",
		fmt.Sprintf("-h%v", tidbAddr),
		"-uroot",
		fmt.Sprintf("-P%v", tidbPort),
		fmt.Sprintf("-D%v", tidbDatabase),
		fmt.Sprintf("--local_infile"),
		"--comments",
		"-e", sql,
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

	allEngines := []string{"tikv,tiflash", "tiflash", "tikv"}
	for _, engine := range allEngines {
		fmt.Printf("Running with engine=%s\n", engine)
		// executing all queries.
		for iQuery := 1; iQuery <= 22; iQuery++ {
			filename := fmt.Sprintf("%d.sql", iQuery)
			f := filepath.Join(*queryDirFlag, filename)
			fmt.Printf("%v\n", f)
			var totCost, minCost, maxCost time.Duration
			// each query run for tpchCountFlag times.
			for i := 0; i < *tpchCountFlag; i++ {
				dur, err := runQuery(*tidbAddrFlag, *tidbPortFlag, *tpchDatabaseFlag, f, engine)
				if err != nil {
					fmt.Printf("error occurred while executing query from %s, error: %s\n", f, err)
				}
				fmt.Printf("%v's %vth run finished in %.2f\n", filename, i, dur.Seconds())
				if i == 0 {
					totCost = dur
					minCost = dur
					maxCost = dur
				} else {
					if minCost > dur {
						minCost = dur
					}
					if maxCost < dur {
						maxCost = dur
					}
					totCost += dur
				}
			}
			avgCost := float64(totCost.Milliseconds()/int64(*tpchCountFlag)) / 1000.0
			fmt.Printf("%v avg: %.2f, min: %.2f, max: %.2f\n", filename, avgCost, minCost.Seconds(), maxCost.Seconds())
		}
		fmt.Println("=========")
	}
}
