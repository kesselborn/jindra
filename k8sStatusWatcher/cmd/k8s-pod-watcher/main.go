package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kesselborn/jindra/k8sStatusWatcher"
)

var debug *log.Logger

type NullWriter struct{}

func init() {
	debug = log.New(NullWriter{}, "[debug] ", log.LUTC)
}

func (_ NullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func handler(ns string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		pathParts := strings.Split(req.URL.Path, "/")
		debug.Printf("request path: %s\n", req.URL.Path)
		if len(pathParts) != 3 {
			http.Error(w, `unknown path ... needs to be:
/pod/<pod>?containers=<container1>,<container2>,...
/pod/<pod>?failed
`, 404)
			return
		}

		pod := pathParts[2]
		debug.Printf("pod: %s\n", pod)

		jsonString, err := k8sStatusWatcher.PodJson(ns, pod)
		if err != nil {
			log.Println("error:", err)
		}

		podInfo := k8sStatusWatcher.NewPodInfoFromJson(jsonString)
		containers := req.FormValue("containers")
		state := podInfo.State(strings.Split(containers, ",")...)
		//b, err := json.MarshalIndent(podInfo, "", "  ")
		//if err != nil {
		//	fmt.Println("error:", err)
		//}
		//w.Header().Set("Conent-Type", "application/json")
		//io.WriteString(w, string(b))
		io.WriteString(w, state)
	}
}

func main() {
	ns := flag.String("ns", os.Getenv("KUBECTL_NAMESPACE"), "namespace")
	debugFlag := flag.Bool("debug", false, "debug")
	semaphoreFile := flag.String("semaphore-file", "", "stop server once this file goes away")
	addr := flag.String("addr", "0.0.0.0:8080", "address where to listen to")
	flag.Parse()

	if *semaphoreFile == "" {
		flag.PrintDefaults()
		panic("semaphore file not set!")
	}

	if *debugFlag {
		debug.SetOutput(os.Stdout)
	}

	go func() {
		for {
			_, err := os.Stat(*semaphoreFile)
			if err != nil {
				if os.IsNotExist(err) {
					break
				}
				fmt.Fprintf(os.Stderr, "error stating %s: %s, continuing anyways", *semaphoreFile, err)
			}
			time.Sleep(1 * time.Second)
		}
		fmt.Printf("semaphore file %s went away -- exiting\n", *semaphoreFile)
		os.Exit(0)
	}()

	http.HandleFunc("/pod/", handler(*ns))
	log.Printf("listening on %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))

}
