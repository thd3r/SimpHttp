package runner

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/thd3r/SimpHttp/pkg/net"
	"github.com/thd3r/SimpHttp/pkg/net/client"
	"github.com/thd3r/SimpHttp/pkg/report"
	"github.com/thd3r/SimpHttp/pkg/utils"
)

type SimpHttpBase struct {
	Targets    []string
	Ports      string
	Threads    int
	Timeout    int
	Verbose    bool
	Client     *client.Client
}

func NewSimpHttp(target, ports string, threads, timeout int, verbose bool) *SimpHttpBase {
	var targets []string

	if target != "" {
		if utils.IsFile(target) {
			file, err := os.Open(target)
			if err != nil {
				utils.VerbosePrint(verbose, fmt.Sprintf("%s\t%s\n", utils.ColoredText("red", "ERR"), strings.ToLower(err.Error())))
			}
			defer file.Close()

			lines, err := utils.ReadLines(file)
			if err != nil {
				utils.VerbosePrint(verbose, fmt.Sprintf("%s\t%s\n", utils.ColoredText("red", "ERR"), strings.ToLower(err.Error())))
			}

			targets = append(targets, lines...)
		} else {
			targets = append(targets, target)
		}
	} else {
		lines, err := utils.ReadLines(os.Stdin)
		if err != nil {
			utils.VerbosePrint(verbose, fmt.Sprintf("%s\t%s\n", utils.ColoredText("red", "ERR"), strings.ToLower(err.Error())))
		}
		targets = append(targets, lines...)
	}

	if ports != "" {
		ports = "80,443," + ports
	} else {
		ports = "80,443"
	}

	clients := client.NewClient(timeout)

	return &SimpHttpBase{
		Targets:    targets,
		Ports:      ports,
		Threads:    threads,
		Timeout:    timeout,
		Verbose:    verbose,
		Client:     clients,
	}
}

func (base *SimpHttpBase) SimpHttpRun() {
	if len(base.Targets) == 0 {
		utils.VerbosePrint(base.Verbose, fmt.Sprintf("%s\tno targets provided\n", utils.ColoredText("red", "ERR")))
		return
	}

	targets := make(chan string)
	validHost := make(chan string)
	dataOutput := make(chan report.DataOutput)

	var validateHostWG sync.WaitGroup
	for i := 0; i < base.Threads; i++ {
		validateHostWG.Add(1)

		go func() {
			defer validateHostWG.Done()
			defer func() {
				if r := recover(); r != nil {
					utils.VerbosePrint(base.Verbose, fmt.Sprintf("%s\t%s\n", utils.ColoredText("red", "ERR"), strings.ToLower(r.(string))))
				}
			}()

			for host := range targets {
				for _, port := range strings.Split(base.Ports, ",") {
					if net.IsReachableHost(host, port, time.Duration(3*time.Second)) {
						validHost <- fmt.Sprintf("%s:%s", host, port)
					}
				}
			}
		}()
	}

	var httpWG sync.WaitGroup
	for i := 0; i < base.Threads; i++ {
		httpWG.Add(1)

		go func() {
			defer httpWG.Done()
			defer func() {
				if r := recover(); r != nil {
					utils.VerbosePrint(base.Verbose, fmt.Sprintf("%s\t%s\n", utils.ColoredText("red", "ERR"), strings.ToLower(r.(string))))
				}
			}()

			for host := range validHost {
				for _, proto := range []string{"http", "https"} {
					utils.VerbosePrint(base.Verbose, fmt.Sprintf("%s\tGET\t%dw\tprocessing %s for %s\n", utils.ColoredText("bblue", "INF"), len(host), strings.ToUpper(proto), host))

					resp, err := base.Client.Do("GET", fmt.Sprintf("%s://%s", proto, host))
					if err != nil {
						errMsg := fmt.Sprintf("fetching %s — %v", host, err)

						utils.VerbosePrint(base.Verbose, fmt.Sprintf("%s\t%s\n", utils.ColoredText("red", "ERR"), errMsg))
						dataOutput <- report.DataOutput{
							Proto:    strings.ToUpper(proto),
							Host:     host,
							ErrorMsg: &errMsg,
						}
						continue
					}

					size, err := io.Copy(io.Discard, resp.Body)
					if err != nil {
						errMsg := fmt.Sprintf("reading response body for %s — %v", host, err)

						utils.VerbosePrint(base.Verbose, fmt.Sprintf("%s\t%s\n", utils.ColoredText("red", "ERR"), errMsg))
						dataOutput <- report.DataOutput{
							Proto:    strings.ToUpper(proto),
							Host:     host,
							Status:   &resp.Status,
							ErrorMsg: &errMsg,
						}
						continue
					}

					sizeBody := fmt.Sprintf("%dw", size)
					base.ProcessResponse(resp, host, strings.ToUpper(proto), sizeBody, dataOutput)

					resp.Body.Close()
				}
			}
		}()
	}

	go func() {
		for _, host := range base.Targets {
			targets <- host
		}
		close(targets)
	}()

	go func() {
		validateHostWG.Wait()
		close(validHost)
	}()

	go func() {
		httpWG.Wait()
		close(dataOutput)
	}()

	report.JsonReport(base.Verbose, dataOutput)
}

func (base *SimpHttpBase) ProcessResponse(resp *http.Response, host, proto, sizeBody string, dataOutput chan<- report.DataOutput) {
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		fmt.Printf("%s\t%s\t%s\t%s\n", utils.ColoredText("green", strconv.Itoa(resp.StatusCode)), resp.Request.Method, sizeBody, resp.Request.URL)
		dataOutput <- report.DataOutput{
			Proto:    proto,
			Host:     host,
			Status:   &resp.Status,
			SizeBody: &sizeBody,
		}
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		location := resp.Header.Get("Location")

		if !strings.HasPrefix(location, "http") {
			fmt.Printf("%s\t%s\t%s\t%s => %s\n", utils.ColoredText("blue", strconv.Itoa(resp.StatusCode)), resp.Request.Method, sizeBody, resp.Request.URL, utils.ColoredText("cyan", fmt.Sprintf("%s%s", resp.Request.URL, location)))
			dataOutput <- report.DataOutput{
				Proto:    proto,
				Host:     host,
				Status:   &resp.Status,
				SizeBody: &sizeBody,
				Redirect: &location,
			}
		} else {
			fmt.Printf("%s\t%s\t%s\t%s => %s\n", utils.ColoredText("blue", strconv.Itoa(resp.StatusCode)), resp.Request.Method, sizeBody, resp.Request.URL, utils.ColoredText("cyan", location))
			dataOutput <- report.DataOutput{
				Proto:    proto,
				Host:     host,
				Status:   &resp.Status,
				SizeBody: &sizeBody,
				Redirect: &location,
			}
		}
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		fmt.Printf("%s\t%s\t%s\t%s\n", utils.ColoredText("magenta", strconv.Itoa(resp.StatusCode)), resp.Request.Method, sizeBody, resp.Request.URL)
		dataOutput <- report.DataOutput{
			Proto:    proto,
			Host:     host,
			Status:   &resp.Status,
			SizeBody: &sizeBody,
		}
	case resp.StatusCode >= 500 && resp.StatusCode < 600:
		fmt.Printf("%s\t%s\t%s\t%s\n", utils.ColoredText("yellow", strconv.Itoa(resp.StatusCode)), resp.Request.Method, sizeBody, resp.Request.URL)
		dataOutput <- report.DataOutput{
			Proto:    proto,
			Host:     host,
			Status:   &resp.Status,
			SizeBody: &sizeBody,
		}
	}
}
