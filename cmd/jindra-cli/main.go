package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	jindra "github.com/kesselborn/jindra/api/v1alpha1"
)

func configMap(p jindra.Pipeline, buildNo int) {
	cm, err := p.PipelineRunConfigMap(buildNo)
	if err != nil {
		log.Fatalf("error converting jindra pipeline config to config map for pipeline run: %s", err)
	}

	fmt.Println("---")
	fmt.Println(interface2yaml(cm))
}

func defaulter(p jindra.Pipeline) {
	p.SetDefaults()

	fmt.Println("---")
	fmt.Println(interface2yaml(p))
}

func job(p jindra.Pipeline, buildNo int) {
	job, err := p.PipelineRunJob(buildNo)
	if err != nil {
		log.Fatalf("error converting jindra pipeline config to job for pipeline run: %s", err)
	}

	fmt.Println("---")
	fmt.Println(interface2yaml(job))
}

func secret(p jindra.Pipeline, buildNo int) {
	secret, err := p.NewRsyncSSHSecret(buildNo)
	if err != nil {
		log.Fatalf("error converting jindra pipeline config to secret for pipeline run: %s", err)
	}

	fmt.Println("---")
	fmt.Println(interface2yaml(secret))
}

func stage(p jindra.Pipeline, buildNo int, stageKey string) {
	cm, err := p.PipelineRunConfigMap(buildNo)
	if err != nil {
		log.Fatalf("error converting jindra pipeline config to config map for pipeline run: %s", err)
	}

	podSrc, ok := cm.Data[stageKey]
	if !ok {
		log.Fatalf("no stage config with name %s found", strings.TrimSuffix(stageKey, ".yaml"))
	}

	fmt.Println("---")
	fmt.Println(podSrc)
}

func stageNames(p jindra.Pipeline, buildNo int) {
	cm, err := p.PipelineRunConfigMap(buildNo)
	if err != nil {
		log.Fatalf("error converting jindra pipeline config to config map for pipeline run: %s", err)
	}

	for k := range cm.Data {
		fmt.Println(strings.TrimSuffix(k, ".yaml"))
	}
}

func validate(p jindra.Pipeline) {
	err := p.Validate()

	if err != nil {
		fmt.Fprintf(os.Stderr, "pipeline config is invalid: \n%s\n", err)
		os.Exit(1)
	}

	fmt.Println("pipeline config is valid")
}

func main() {
	config := flag.String("c", "", "jindra pipeline config (use '-c -' for reading from stdin)")
	help := flag.Bool("h", false, "show help text")
	buildNo := flag.Int("build-no", 42, "build number")

	flag.Usage = func() {
		fmt.Printf(`
Usage: %s [OPTIONS] <command>

Commands:
  all         : print all configs necessary to run the pipeline (can be piped into 'kubectl apply -f-')
  stage STAGE : print stage configuration
  stagenames  : print stage names (which can be used with the stage sub command)
  configmap   : print configmap
  job         : print runner job
  secret      : print secret

  defaulter   : print config with default values
  validate    : validate pipeline file

Options: 
`, os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	usage := func(success bool) {
		exitCode := 0
		if !success {
			exitCode = 1
			flag.CommandLine.SetOutput(os.Stderr)
		}
		flag.Usage()
		os.Exit(exitCode)
	}

	if config == nil || *config == "" || *help {
		usage(*help)
	}

	var err error
	var yamlData []byte

	if *config == "-" {
		yamlData, err = ioutil.ReadAll(os.Stdin)
	} else {
		yamlData, err = ioutil.ReadFile(*config)
	}

	if err != nil {
		log.Fatalf("error reading file %s: %s", *config, err)
	}

	p, err := jindra.NewPipelineFromYaml(yamlData)
	if err != nil {
		log.Fatalf("cannot convert yaml to jindra pipeline: %s", err)
	}

	switch flag.Arg(0) {
	case "all":
		secret(p, *buildNo)
		configMap(p, *buildNo)
		job(p, *buildNo)
	case "configmap":
		configMap(p, *buildNo)
	case "defaulter":
		defaulter(p)
	case "job":
		job(p, *buildNo)
	case "secret":
		secret(p, *buildNo)
	case "stage":
		if flag.Arg(1) == "" {
			usage(false)
		}
		stage(p, *buildNo, flag.Arg(1)+".yaml")
	case "stagenames":
		stageNames(p, *buildNo)
	case "validate":
		validate(p)
	case "":
		fmt.Println("no sub command given")
		usage(false)
	default:
		fmt.Printf("unknown sub command %s\n", flag.Arg(0))
		usage(false)
	}
}

func interface2yaml(x interface{}) string {
	jsonTxt, err := json.Marshal(x)
	if err != nil {
		panic("error marshalling to json: " + err.Error())
	}

	yamlTxt, err := yaml.JSONToYAML(jsonTxt)
	if err != nil {
		panic("error running JSONToYAML: " + err.Error())
	}

	return string(yamlTxt)
}
