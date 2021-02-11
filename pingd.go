package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
        "time"
        "strconv"

	"github.com/takama/daemon"
        "github.com/go-ping/ping"
)

const (
	name        = "pingd"
	description = "pingd: Perform ping to host and return response time in seconds"
)

var stdlog, errlog *log.Logger

type Service struct {
	daemon.Daemon
}

func handleIndex(w http.ResponseWriter, req *http.Request) {
	fmt.Println("use /ping4?ip=<ipv4>&timeout=<seconds>")
}

func handlePing4Request(w http.ResponseWriter, req *http.Request) {
	var result float64
        var timeout int = 10

	// Obtain the ip address from the requestor. As this routine likely sits behind a reverse proxy,
	// first test for an ip in the header. Only if not there, use the RemoteAddr method.

	ip := req.Header.Get("X-Real-Ip")
	if ip == "" {
		ip = req.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = req.RemoteAddr
	}

        // Fetch the request ip address, check it and process accordingly

	ipv4, _ := req.URL.Query()["ip"]
        strTimeout, ok := req.URL.Query()["timeout"]

        if ok {
          timeout, _ = strconv.Atoi(strTimeout[0])
        }

        // Perform ping

	pinger, err := ping.NewPinger(ipv4[0])

        if err != nil {
                result = -3.0
		errlog.Println("Error (allocating pinger): ", err)
        } else {
		pinger.Count = 1
                pinger.Timeout = time.Duration(timeout)*time.Second
		pinger.SetPrivileged(true)

		pinger.Run()

		stats := pinger.Statistics()
               	if stats.PacketsRecv == 0 {
			result = -1.0
		} else {
			result = float64(stats.MinRtt.Nanoseconds())/1000000
		}
        }

	fmt.Fprintf(w, "%.5f", result)
	stdlog.Printf("Ping to %s took %.5f ms (timeout %d sec)",ipv4[0],result,timeout)
}

func (service *Service) Manage() (string, error) {
	// The deamon control section has been copied from an example of the Takama/daemon library. It's
	// straightforward and install/using/removing the daemon is a breeze ...

	usage := "Usage: pingd install | remove | start | stop | status"

	if len(os.Args) > 1 {
		command := os.Args[1]
		switch command {
		case "install":
			return service.Install()
		case "remove":
			return service.Remove()
		case "start":
			return service.Start()
		case "stop":
			return service.Stop()
		case "status":
			return service.Status()
		default:
			return usage, nil
		}
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/ping4", handlePing4Request)

	// Start the daemon in a child process. We can handle multiple requests in parallel from here.

	go func() {
		http.ListenAndServe(":7008", nil)
	}()

	// Log that our service is ready and listening on port 7008.

	stdlog.Println("Service started, listening on port 7008")

	// Wait for signals in an infinite loop. Note that we only accept a kill signal; no other signals are caught.

	for {
		select {
		case killSignal := <-interrupt:
			stdlog.Println("Got signal:", killSignal)
			if killSignal == os.Interrupt {
				return "Daemon was interrupted by system signal", nil
			}
			return "Daemon was killed", nil
		}
	}
}

func init() {
	// As we are a daemon we need to divert our standard and error output.

	stdlog = log.New(os.Stdout, "", 0)
	errlog = log.New(os.Stderr, "", 0)
}

func main() {
	// Create a new daemon process.

	srv, err := daemon.New(name, description)

	// If we failed, report the error and halt.

	if err != nil {
		errlog.Println("Error: ", err)
		os.Exit(1)
	}

	// Otherwise initialize and run the daemon service.

	service := &Service{srv}
	status, err := service.Manage()

	// If initialization failed (short of memory for examples) report the error and halt.

	if err != nil {
		errlog.Println(status, "\nError: ", err)
		os.Exit(1)
	}

	// Report our daemon status.
	fmt.Println(status)
}
