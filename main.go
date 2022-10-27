package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	osquery "github.com/osquery/osquery-go"
)

var socketPath string = "/var/osquery/osquery.em"

var queryResults = struct {
	Results map[string]interface{} `json:"query_results"`
}{
	Results: map[string]interface{}{},
}

func main() {
	e := echo.New()

	// osquery client
	client, err := osquery.NewClient(socketPath, 5*time.Second)
	if err != nil {
		fmt.Println("error creating osquery client %d", err)
		return
	}

	defer client.Close()
	// routes
	anonRunQueryHandler := func(c echo.Context) error {
		return runQueryHandler(c, client)
	}
	e.POST("/run_query", anonRunQueryHandler)
	e.Logger.Fatal(e.Start(":1323"))
}

func postQueryResults(queryJSON []byte) {
	client := http.Client{}
	req, err := http.NewRequest("POST", "http://localhost:3000/query_results", bytes.NewBuffer(queryJSON))
	if err != nil {
		//Handle Error
	}

	req.Header = http.Header{
		"Content-Type": {"application/json"},
	}

	res, err := client.Do(req)
	if err != nil {
		//Handle Error
	}
	fmt.Println(res)
}

var Queries = map[string]string{
	"uptime":                                "SELECT * FROM uptime;",
	"device_os_information":                 "SELECT * FROM os_version;",
	"device_hardware_info":                  "SELECT * FROM system_info;",
	"user_chrome_extensions":                "SELECT * FROM users CROSS JOIN chrome_extensions USING (uid);",
	"user_chrome_extension_content_scripts": "SELECT * FROM users CROSS JOIN chrome_extension_content_scripts USING (uid);",
	"process_not_on_disk":                   "SELECT name, path, pid FROM processes WHERE on_disk = 0;",
	"running_applications":                  "SELECT * FROM running_apps;",
	"device_crashes":                        "SELECT type,pid, path, exception_type, exception_notes, datetime FROM crashes;",
}

var bestPracticesSimpleColumns = map[string]string{
	"sip_enabled":        "SELECT enabled AS compliant FROM sip_config WHERE config_flag='sip'",
	"gatekeeper_enabled": "SELECT assessments_enabled AS compliant FROM gatekeeper",
	"filevault_enabled":  "SELECT de.encrypted AS compliant FROM mounts m join disk_encryption de ON m.device_alias = de.name WHERE m.path = '/'",
	"firewall_enabled":   "SELECT global_state AS compliant FROM alf",
	// Sharing prefs
	"screen_sharing_disabled":      "SELECT screen_sharing = 0 AS compliant FROM sharing_preferences",
	"file_sharing_disabled":        "SELECT file_sharing = 0 AS compliant FROM sharing_preferences",
	"printer_sharing_disabled":     "SELECT printer_sharing = 0 AS compliant FROM sharing_preferences",
	"remote_login_disabled":        "SELECT remote_login = 0 AS compliant FROM sharing_preferences",
	"remote_management_disabled":   "SELECT remote_management = 0 AS compliant FROM sharing_preferences",
	"remote_apple_events_disabled": "SELECT remote_apple_events = 0 AS compliant FROM sharing_preferences",
	"internet_sharing_disabled":    "SELECT internet_sharing = 0 AS compliant FROM sharing_preferences",
	"bluetooth_sharing_disabled":   "SELECT bluetooth_sharing = 0 AS compliant FROM sharing_preferences",
	"disc_sharing_disabled":        "SELECT disc_sharing = 0 AS compliant FROM sharing_preferences",
}

func runQueryHandler(c echo.Context, client *osquery.ExtensionManagerClient) error {
	queryToRun := new(RunQueryPayload)
	if err := c.Bind(queryToRun); err != nil {
		fmt.Println("messed up binding the query to run")
		return err
	}

	runQuery(client, queryToRun)
	return c.JSON(http.StatusCreated, queryToRun)
}

func runQuery(client *osquery.ExtensionManagerClient, query *RunQueryPayload) {
	resp, err := client.Query(query.Query)
	if err != nil {
		fmt.Println("bad query")
		return
	}
	queryResults.Results["endpoint_id"] = query.EndpointId
	queryResults.Results["query"] = query.Query
	queryResults.Results["result"] = resp.Response

	fmt.Println(queryResults)
	queryJSON, _ := json.Marshal(&queryResults.Results)
	postQueryResults(queryJSON)
}

type RunQueryPayload struct {
	Query      string `json:"query"`
	EndpointId string `json:"endpoint_id"`
}
