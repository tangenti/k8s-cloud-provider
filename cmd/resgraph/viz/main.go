/*
Copyright 2023 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	coreexec "os/exec"
	"strings"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph/algo/graphviz"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph/exec"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph/workflow/plan"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph/workflow/testcase"
	"github.com/kr/pretty"
	"k8s.io/klog/v2"

	_ "github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/resgraph/workflow/testcase/lb"
)

var (
	flags = struct {
		http    string
		dotExec string
	}{
		http:    "localhost:8080",
		dotExec: "/usr/bin/dot",
	}
)

func main() {
	flag.Parse()

	http.HandleFunc("/", mainPage)
	http.HandleFunc("/show", showPage)

	klog.Error(http.ListenAndServe(flags.http, nil))
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<!DOCTYPE html>\n"))
	w.Write([]byte("<ol>\n"))

	for _, c := range testcase.Cases() {
		w.Write([]byte(fmt.Sprintf("<li><a href=\"/show?q=%s\">%s</a></li>\n", c.Name, c.Name)))
	}
	w.Write([]byte("</ol>\n"))
}

func showPage(w http.ResponseWriter, r *http.Request) {
	outln := func(s string) { w.Write([]byte(s + "\n")) }
	outf := func(f string, args ...any) {
		w.Write([]byte(fmt.Sprintf(f, args...) + "\n"))
	}

	name := r.URL.Query().Get("q")
	tc := testcase.Case(name)
	if tc == nil {
		w.WriteHeader(404)
		outf("%q is not a valid test case", name)
		return
	}

	outln("<!DOCTYPE html>")
	outf("<h1>%s</h1>", name)

	mock := cloud.NewMockGCE(&cloud.SingleProjectRouter{ID: "proj"})
	result, err := plan.Do(context.Background(), mock, tc.Steps[0].Graph)

	outln("<pre>")
	outf("plan.Do() = _, %v", err)
	outln("</pre>")

	outln("<h2>Got graph</h2>")
	svg, err := dotSVG(graphviz.Do(result.Got))
	if err == nil {
		outf("%s", svg)
	} else {
		outf("dotSVG() = %v", err)
	}

	outln("<h2>Want graph</h2>")
	svg, err = dotSVG(graphviz.Do(result.Want))
	if err == nil {
		outln(svg)
	} else {
		outf("dotSVG() = %v", err)
	}

	var viz exec.VizTracer
	ex := exec.NewSerialExecutor(result.Actions, exec.DryRunOption(true), exec.TracerOption(&viz))
	pending, err := ex.Run(nil, nil)

	outln("<h2>Plan dry-run</h2>")
	svg, err = dotSVG(viz.String())
	if err == nil {
		outln(svg)
	} else {
		outf("<pre>dotSVG() = %v</pre>", err)
	}

	outln("<h2>Plan pending items</h2>\n")

	if len(pending) == 0 {
		outln("all actions were executable")
	} else {

		outln("<ol>")
		for _, item := range pending {
			outf("<li>%s</li>", item.Summary())
		}
		outln("</ol>")
	}

	outln("<h2>Test case</h2>\n")
	outln("<pre>")
	outln(pretty.Sprint(tc))
	outln("</pre>")
}

func dotSVG(text string) (string, error) {
	cmd := coreexec.Command(flags.dotExec, "-Tsvg")

	inPipe, err := cmd.StdinPipe()
	if err != nil {
		klog.Errorf("StdinPipe = %v", err)
		return "", err
	}

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		klog.Errorf("StdoutPipe = %v", err)
		return "", err
	}

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		klog.Errorf("StderrPipe = %v", err)
		return "", err
	}

	var cmdErr error
	cmdDone := make(chan struct{})

	go func() {
		cmdErr = cmd.Run()
		if err != nil {
			klog.Errorf("cmd.Run() = %v", cmdErr)
		}
		close(cmdDone)
	}()

	n, err := inPipe.Write([]byte(text))
	inPipe.Close()

	klog.Infof("Write() = %d, %v", n, err)

	bytes, err := io.ReadAll(outPipe)
	if err != nil {
		klog.Errorf("ReadAll(outPipe) = %v", err)
	}

	errBytes, err := io.ReadAll(errPipe)
	klog.Infof("ReadAll(errPipe) = %v, %s", err, errBytes)

	<-cmdDone
	if cmdErr != nil {
		return "", cmdErr
	}

	// Trim off some of the extra stuff in the output of dot.
	svgText := string(bytes)
	start := strings.Index(svgText, "<svg")
	if start != -1 {
		svgText = svgText[start:]
	}

	return svgText, nil
}
