package agent

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/sofuture/kubernotes/cluster"
)

func (a *Agent) SpawnAPI() error {
	listener, err := net.Listen("tcp", a.Bind)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Println("api listening on", a.Bind)

	// /logs endpoint to display logs over http
	http.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {

		// accept jobid via querystring ?job=id
		jobID := r.FormValue("job")

		// accept line count via querystring ?count=#
		countRaw := r.FormValue("count")

		// if we get an invalid count parameter, just default to 20 lines
		count, err := strconv.Atoi(countRaw)
		if err != nil {
			count = 20
		}

		respond := func(code int, message string) {
			w.WriteHeader(code)
			fmt.Fprint(w, message)
		}

		// if we got a job ID
		if jobID != "" {
			// grab logs
			logs, err := a.Local.GetLogs(&cluster.Job{ID: jobID}, count)
			if err != nil {
				respond(404, fmt.Sprintf("job not found: %s", jobID))
			} else {
				// write logs out
				respond(200, logs)
			}
		} else {
			// fail out
			respond(400, "no job specified")
		}
	})

	return http.Serve(listener, nil)
}
