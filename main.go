package main

import (
	"encoding/json"
	"flag"
	"github.com/dustin/go-humanize"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var templates = []string{"templates/icinga.html"}

const (
	BASEURL = "http://localhost/"
)

var (
	baseURLFlag       = flag.String("baseurl", BASEURL, "Base URL to icingaweb2")
	customHeaderName  = flag.String("custom-header-name", "", "Custom request header name")
	customHeaderValue = flag.String("custom-header-value", "", "Custom request header value")
)

type ServiceProblem struct {
	HostDisplayName             string `json:"host_display_name"`
	HostName                    string `json:"host_name"`
	HostState                   int    `json:"host_state"`
	ServiceAcknowledged         int    `json:"service_acknowledged"`
	ServiceActiveChecksEnabled  int    `json:"service_active_checks_enabled"`
	ServiceAttempt              string `json:"service_attempt"`
	ServiceDescription          string `json:"service_description"`
	ServiceDisplayName          string `json:"service_display_name"`
	ServiceHandled              int    `json:"service_handled"`
	ServiceIconImage            string `json:"service_icon_image"`
	ServiceIconImageAlt         string `json:"service_icon_image_alt"`
	ServiceInDowntime           int    `json:"service_in_downtime"`
	ServiceIsFlapping           int    `json:"service_is_flapping"`
	ServiceLastStateChange      int    `json:"service_last_state_change"`
	ServiceNotificationsEnabled int    `json:"service_notifications_enabled"`
	ServiceOutput               string `json:"service_output"`
	ServicePassiveChecksEnabled int    `json:"service_passive_checks_enabled"`
	ServicePerfdata             string `json:"service_perfdata"`
	ServiceServerity            int    `json:"service_severity"`
	ServiceState                int    `json:"service_state"`
	ServiceStateType            int    `json:"service_state_type"`
}

func (s ServiceProblem) HumanDuration() string {
	tm := time.Unix(int64(s.ServiceLastStateChange), 0)

	return humanize.Time(tm)
}

type Out struct {
	ServiceProblems   []ServiceProblem
	AcknowledgedCount int
	CriticalCount     int
	WarningCount      int
	UnknownCount      int
	TZ                string
}

func (o Out) CurrentTime() string {
	now := time.Now()
	if location, err := time.LoadLocation(o.TZ); err == nil {
		now = time.Now().In(location)
	}
	return now.Format("15:04:05 MST")
}

func getJson(req *http.Request, target interface{}) error {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	log.Printf("Icinga status: %d\n", resp.StatusCode)
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(target)
	if err != nil {
		log.Printf("Error occurred: %s\n", err.Error())
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			log.Printf("Body:\n%s\n", body)
		}
	}
	return err
}

func icinga(w http.ResponseWriter, r *http.Request) {
	url := *baseURLFlag + "/monitoring/list/services?service_problem=1&sort=service_severity&dir=desc"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Unable to create http request: %s\n", err)
	}
	req.Header.Add("accept", "application/json")

	if *customHeaderName != "" {
		req.Header.Add(*customHeaderName, *customHeaderValue)
	}

	problems := []ServiceProblem{}
	if err := getJson(req, &problems); err != nil {
		log.Printf("Unable to execute http request: %s\n", err)
		http.Error(w, "Unable to fetch from upstream: "+err.Error(), 500)
		return
	}

	var out Out
	for _, p := range problems {
		if p.ServiceAcknowledged == 1 {
			out.AcknowledgedCount += 1
			continue
		}
		if p.ServiceState == 1 {
			out.WarningCount += 1
		}
		if p.ServiceState == 2 {
			out.CriticalCount += 1
		}
		if p.ServiceState == 3 {
			out.UnknownCount += 1
		}
		out.ServiceProblems = append(out.ServiceProblems, p)

	}

	out.TZ = r.FormValue("tz")

	data, err := Asset("templates/icinga.html")
	if err != nil {
		log.Fatalf("Asset not found: %s\n", err)
	}

	t := template.New("icinga")
	t, err = t.Parse(string(data))
	if err != nil {
		log.Fatalf("Unable to parse template: %s\n", err)
	}
	err = t.Execute(w, out)
	if err != nil {
		log.Fatalf("Unable to parse template: %s\n", err)
	}
}

func main() {
	flag.Parse()
	log := log.New(os.Stdout, "- ", log.LstdFlags)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(assetFS())))

	http.HandleFunc("/icinga", icinga)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
