package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"bufio"
	"time"
	"os"
	"net/smtp"
	"reflect"
//	"encoding/base64"

	"github.com/dattran92/simplequeue/queue"
	_ "github.com/lib/pq"
	"github.com/robfig/cron"
)
type AppProperties struct {
	Repo_go_driver_name string
	Repo_go_data_source_name string
	Target_monitor_driver_name string
	Target_monitor_data_source_name string
	Smtp_mail_server_name string
	Smtp_mail_server_port string
	Smtp_login_method string
	Smtp_login_username string
	Smtp_login_password string
	Smtp_email_from string
	Smtp_email_to string
	Smtp_email_cc string
}
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
type job_log_vw struct {
	cmd_name string
	ep_id string
	start_ts string
	end_ts string
	result string
}
var (
	cronMap = make(map[cron.EntryID]int)
)
var q queue.Queue
var aProps AppProperties

func populate_ctq_schedule(c *cron.Cron, q *queue.Queue) {

	db, err := sql.Open(aProps.Repo_go_driver_name, aProps.Repo_go_data_source_name)
	if err != nil {
		fmt.Println("Error in populate_ctq_schedule X001", err )
	}
	defer db.Close()
	rows, err := db.Query("SELECT expr,cron_id FROM cron")
	if err != nil {
		fmt.Println("Error in populate_ctq_schedule X001", err )
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
	db, err := sql.Open(aProps.Repo_go_driver_name, aProps.Repo_go_data_source_name)
	if err != nil {
		fmt.Println("Error in process_cron_id X001", err )
	}
	defer db.Close()
	rows, err := db.Query("select epmg_id,epg_id,cmd_id,cron_id,ep_id,url,cmd FROM mp_info where cron_id=$1", *cron_id)
	if err != nil {
		fmt.Println("Error in process_cron_id X002", err )
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
	fmt.Println(mpi.url)
	db, err := sql.Open("postgres", mpi.url)
	err=db.Ping()
	if err != nil {
		fmt.Println("Error in process_cmd X001", err )
		return
	}
	defer db.Close()

	rows, err := db.Query(mpi.cmd)
	if err != nil {
		fmt.Println("Error in process_cmd X002", err )
		return
	}
	defer rows.Close()
	//fmt.Println("json", Jsonify(rows))
	mpi.end_ts = time.Now().String()
	fmt.Println(mpi)

	db1, err := sql.Open(aProps.Repo_go_driver_name, aProps.Repo_go_data_source_name)
	if err != nil {
		fmt.Println("Error in process_cmd X003", err )
		return
	}
	defer db1.Close()
	st1, err := db1.Prepare("insert into job_log(epmg_id,epg_id,cmd_id,cron_id,ep_id,start_ts,end_ts,result) values ($1,$2,$3,$4,$5,$6,$7,$8)")
	if err != nil {
		fmt.Println("Error in process_cmd X004", err )
		return
	}
	defer st1.Close()

	_, err = st1.Exec(mpi.epmg_id, mpi.epg_id, mpi.cmd_id, mpi.cron_id, mpi.ep_id, mpi.start_ts, mpi.end_ts, fmt.Sprintf("%s", Jsonify(rows)))
	if err != nil {
		fmt.Println("Error in process_cmd X005", err )
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
		fmt.Println("Error in Jsonify X001", err )
		return []string{err.Error()}
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
			fmt.Println("Error in Jsonify X002", err )
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
func Alert_job_log() {
	var alv job_log_vw
	db, err := sql.Open(aProps.Repo_go_driver_name, aProps.Repo_go_data_source_name)
	if err != nil {
		fmt.Println("Error in Alert_job_log X001", err )
	}
	defer db.Close()
	for {
		fmt.Println("Alert_job_log - running")
		rows, err := db.Query("select * from job_log_vw")
		if err != nil {
			fmt.Println("Error in Alert_job_log X002", err )
		}
		defer rows.Close()
		for rows.Next() {
			rows.Scan(&alv.cmd_name,&alv.ep_id,&alv.start_ts,&alv.end_ts,&alv.result )
			fmt.Println("calling SendMail... ")
			err:=SendMail(aProps.Smtp_email_to,alv.cmd_name + " on " + alv.ep_id,alv.result)
			if err != nil {
				fmt.Println("Error in Alert_job_log X003", err )
			}
		}
		time.Sleep(10 * time.Second)		
	}
}
func SendMail(to string, subject string, body string) error {
	fmt.Println("SendMail - " + aProps.Smtp_email_from + "," + to + "," + subject)
	c, err := smtp.Dial(aProps.Smtp_mail_server_name + ":" + aProps.Smtp_mail_server_port)
	if err != nil {fmt.Println("Error in SendMail X001", err )}
	defer c.Close()
	if err := c.Mail(aProps.Smtp_email_from); err != nil {fmt.Println("Error in SendMail X002", err )}
    if err := c.Rcpt(to); err != nil {fmt.Println("Error in SendMail X002", err )}

	w, err := c.Data()
	if err != nil {fmt.Println("Error in SendMail X003", err )}

	msg := "From: " + aProps.Smtp_email_from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n" +
		"Content-Transfer-Encoding: base64\r\n" +
		"\r\n" + string([]byte(body))
		//"\r\n" + base64.StdEncoding.EncodeToString([]byte(body))		

	_, err = w.Write([]byte(msg))
    if err != nil {fmt.Println("Error in SendMail X004", err )}
	if err := w.Close(); err != nil {fmt.Println("Error in SendMail X005", err )} 
	fmt.Println("SendMain successful...", err )
    return c.Quit()
}
//Load App.properties from file using reflection
func LoadPropsFromFile( filename string,  props *AppProperties) {
	file, err:= os.Open(filename)
	if (err != nil ) {
		fmt.Println("Error in LoadPropsFromFile X001", err )
		os.Exit(1)
	}
	defer file.Close()
	var prop map[string]string
	prop = make(map[string]string)
	s := bufio.NewScanner(file)

	for s.Scan() {
		read_line := strings.TrimSpace(s.Text())
		if ( len(read_line) == 0   ) { continue }
		if ( read_line[0:1] == "#" ) { continue }
		str:=strings.Split(read_line,"=")
		prop[str[0]]=strings.Join(str[1:],"=")
	}
	v_props := reflect.ValueOf(props).Elem()
	t_props := v_props.Type()
	for i := 0; i < v_props.NumField(); i++ {
		f := v_props.Field(i)
		if (f.CanSet()) { f.SetString(prop[strings.ToLower(t_props.Field(i).Name)]) }
		fmt.Printf("%d: %s %s = %v\n", i,t_props.Field(i).Name, f.Type(), f.Interface())
	}
}
func main() {
		aProps=AppProperties{}
		LoadPropsFromFile("App.properties",&aProps)
		fmt.Println(aProps.Target_monitor_driver_name, aProps.Target_monitor_data_source_name)
		//os.Exit(0)
		c := cron.New()
		q = queue.Init()
		populate_ctq_schedule(c, &q)
	
		c.Start()
		defer c.Stop()
	
		go manage_ctq_schedule(&q)
		go Alert_job_log()
	
		go func() {
			for {
				for _, entry := range c.Entries() {
					fmt.Println(entry, cronMap[entry.ID])
				}
				time.Sleep(30 * time.Second)
			}
		}()
		fmt.Println("it will run for 1000 seconds")
		time.Sleep(1000 * time.Second)
		fmt.Println(len(q.Values))
		fmt.Println(q.Values)
}

