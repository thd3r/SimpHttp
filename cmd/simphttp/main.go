package main 

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/thd3r/SimpHttp/pkg/utils"
	"github.com/thd3r/SimpHttp/pkg/report"
	"github.com/thd3r/SimpHttp/pkg/runner"
)

func init() {
	var banner = fmt.Sprintf(`
	╔═╗┬┌┬┐┌─┐╦ ╦┌┬┐┌┬┐┌─┐
	╚═╗││││├─┘╠═╣ │  │ ├─┘
	╚═╝┴┴ ┴┴  ╩ ╩ ┴  ┴ ┴  
		  %s
	`, utils.Version())

	fmt.Println(banner)

	flag.Usage = func() {
		usage := []string{
			"Usage of SimpHttp:",
			"",
			"  -target	Single target or file containing multiple targets",
			"  -ports	Ports to probe (default: 80,443)",
			"  -threads	number of concurrent threads (default: 40)",
			"  -timeout	HTTP request timeout in seconds (default: 10)",
			"  -verbose	Show verbose output",
			"  -version 	Show SimpHttp version",
			"",
		}
		fmt.Printf("%s\n", strings.Join(usage, "\n"))
	}
}

func main() {
	var target string
	flag.StringVar(&target, "target", "", "")

	var ports string
	flag.StringVar(&ports, "ports", "", "")

	var threads int
	flag.IntVar(&threads, "threads", 40, "")

	var timeout int
	flag.IntVar(&timeout, "timeout", 10, "")

	var verbose bool
	flag.BoolVar(&verbose, "verbose", false, "")

	var version bool
	flag.BoolVar(&version, "version", false, "")

	flag.Parse()

	if version {
		fmt.Printf("SimpHttp current version: %s\n", strings.ReplaceAll(utils.CurrentVersion, "v", ""))
		os.Exit(0)
	}

	fmt.Printf(":: SimpHttp — A minimalist HTTP/HTTPS-aware domain probe\n")
	fmt.Printf(":: Generating report at %s\n\n", report.FilePath)

	simp := runner.NewSimpHttp(target, ports, threads, timeout, verbose)
	simp.SimpHttpRun()
}
