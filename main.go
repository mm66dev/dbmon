package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/dattran92/simplequeue/queue"
	_ "github.com/lib/pq"
	"github.com/robfig/cron"
)

type mp_info struct {
	epmg_id  int
	epg_id   int
	cmd_id   int
	cron_id  int
	ep_id    int
	url      string
	cmd      string
	start_ts string
	end_ts   string
	result   string
}

var (
	cronMap = make(map[cron.EntryID]int)
)
var q queue.Queue

func populate_ctq_schedule(c *cron.Cron, q *queue.Queue) {

	db, err := sql.Open("postgres", "postgres://postgres:pwd@localhost:5432/dbmon?sslmode=disable")
	if err != nil {
		fmt.Println(`Could not connect to db`)
		fmt.Println(err)
	}
	defer db.Close()
	rows, err := db.Query("SELECT expr,cron_id FROM cron")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	for rows.Next() {
		var cron_id int
		var expr string

		rows.Scan(&expr, &cron_id)
		fmt.Println(expr, cron_id)
		cid, _ := c.AddFunc(expr, func() { q.Enqueue(cron_id) })
		cronMap[cid] = cron_id
	}
	fmt.Println(cronMap)
}

/*
job-schedules and associated tasks
for each task run
*/
func manage_ctq_schedule(q *queue.Queue) {
	for {
		fmt.Println(q.Values)
		for len(q.Values) > 0 {
			go process_cron_id(q.Dequeue())
		}
		time.Sleep(10 * time.Second)
	}
}

func process_cron_id(cron_id *int) {

	fmt.Printf("cron_id: %d\n", cron_id)
	db, err := sql.Open("postgres", "postgres://postgres:pwd@localhost:5432/dbmon?sslmode=disable")
	if err != nil {
		fmt.Println(`Could not connect to db`)
		fmt.Println(err)
	}
	defer db.Close()
	rows, err := db.Query("select epmg_id,epg_id,cmd_id,cron_id,ep_id,url,cmd FROM mp_info where cron_id=$1", *cron_id)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var mpi mp_info

		rows.Scan(&mpi.epmg_id, &mpi.epg_id, &mpi.cmd_id, &mpi.cron_id, &mpi.ep_id, &mpi.url, &mpi.cmd)
		process_cmd(mpi)

	}
}

func process_cmd(mpi mp_info) {
	mpi.start_ts = time.Now().String()

	db, err := sql.Open("postgres", mpi.url)
	if err != nil {
		fmt.Println(`Could not connect to db`)
		fmt.Println(err)
	}
	defer db.Close()

	rows, _ := db.Query(mpi.cmd)
	defer rows.Close()
	//fmt.Println("json", Jsonify(rows))
	mpi.end_ts = time.Now().String()
	fmt.Println(mpi)

	db1, err := sql.Open("postgres", "postgres://postgres:pwd@localhost:5432/dbmon?sslmode=disable")
	if err != nil {
		log.Fatal("open error occurred:", err)
		fmt.Println(err)
	}
	defer db1.Close()
	st1, err := db1.Prepare("insert into job_log(epmg_id,epg_id,cmd_id,cron_id,ep_id,start_ts,end_ts,result) values ($1,$2,$3,$4,$5,$6,$7,$8)")
	if err != nil {
		log.Fatal("prepare error occurred:", err)
	}
	defer st1.Close()
	_, err = st1.Exec(mpi.epmg_id, mpi.epg_id, mpi.cmd_id, mpi.cron_id, mpi.ep_id, mpi.start_ts, mpi.end_ts, fmt.Sprintf("%s", Jsonify(rows)))
	if err != nil {
		log.Fatal("exec error occurred:", err)
	}
}

func process_rows2json_string(rows *sql.Rows) string {

	columns, _ := rows.Columns()

	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))

	for i := range values {
		scanArgs[i] = &values[i]
	}

	list := make([]map[string]interface{}, 0)
	for rows.Next() {
		_ = rows.Scan(scanArgs...)
		var m = map[string]interface{}{}
		for i, column := range columns {
			m[column] = values[i]
			fmt.Printf("%s, %T", column, values[i])
		}
		//obj, _ := json.Marshal(m)
		list = append(list, m)
		//ss += string(obj)
	}
	b, _ := json.MarshalIndent(list, "", "")
	return fmt.Sprintf("%s", b)
}
func Jsonify(rows *sql.Rows) []string {
	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}

	values := make([]interface{}, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	c := 0
	results := make(map[string]interface{})
	data := []string{}

	for rows.Next() {
		if c > 0 {
			data = append(data, ",")
		}

		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}

		for i, value := range values {
			switch value.(type) {
			case nil:
				results[columns[i]] = nil

			case []byte:
				s := string(value.([]byte))
				x, err := strconv.Atoi(s)

				if err != nil {
					results[columns[i]] = s
				} else {
					results[columns[i]] = x
				}

			default:
				results[columns[i]] = value
			}
		}

		//b, _ := json.Marshal(results)
		b, _ := json.MarshalIndent(results, "", "")
		data = append(data, strings.TrimSpace(string(b)))
		c++
	}

	return data
}
func main() {
	c := cron.New()
	q = queue.Init()
	populate_ctq_schedule(c, &q)

	c.Start()
	defer c.Stop()

	go manage_ctq_schedule(&q)

	go func() {
		for {
			for _, entry := range c.Entries() {
				fmt.Println(entry, cronMap[entry.ID])
			}
			time.Sleep(30 * time.Second)
		}
	}()

	time.Sleep(1000 * time.Second)
	fmt.Println(len(q.Values))
	fmt.Println(q.Values)
}
