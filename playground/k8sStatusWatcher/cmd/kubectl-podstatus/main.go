package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kesselborn/jindra/k8sStatusWatcher"
)

func usage() {
	fmt.Printf(`
Utility to output the state of the current pods

%s [-n NAMESPACE] <pod>

OPTIONS:
`, os.Args[0])
	flag.PrintDefaults()

}

func main() {
	ns := flag.String("n", os.Getenv("KUBECTL_NAMESPACE"), "namespace")
	flag.Parse()

	pod := flag.Arg(0)

	if pod == "" {
		usage()
		log.Fatal("missing pod name argument")
	}

	jsonString, err := k8sStatusWatcher.PodJson(*ns, pod)

	podInfo := k8sStatusWatcher.NewPodInfoFromJson(jsonString)
	b, err := json.MarshalIndent(podInfo, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	os.Stdout.Write(b)
}
