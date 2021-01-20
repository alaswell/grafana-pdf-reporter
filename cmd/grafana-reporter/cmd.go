package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type responseWriter struct {
	buf bytes.Buffer
}

func (responseWriter) Header() http.Header {
	return http.Header{}
}

func (responseWriter) WriteHeader(statusCode int) {}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.buf.Write(b)
}

func buildOutputPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.SetOutput(os.Stdout)
		log.Fatal(err)
	}

	// create output directory, if it doesn't exist already
	outputPath := cwd + "/" + *outputDir
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		_ = os.Mkdir(outputPath, os.ModePerm)
	}
	return outputPath + "/" + *outputFile
}

func buildVariableList(varList string) string {
	retStr := ""
	if varList != "" {
		keyPairs := strings.Split(varList, ",")
		for _, kp := range keyPairs {
			retStr = retStr + "&var-" + kp
		}
	}
	return retStr
}

func cmdHandler(router *mux.Router) error {
	rqStr := "/api/v5/report/%s?apitoken=%s&%s%s"
	if *apiVersion == "v4" {
		rqStr = "/api/report/%s?apitoken=%s&%s%s"
	}

	if template != nil && *template != "" {
		rqStr += "&template=" + *template
	}

	varList := buildVariableList(*variableList)
	rq, err := http.NewRequest("GET", fmt.Sprintf(rqStr, *dashboard, *apiKey, *timeSpan, varList), nil)
	if err != nil {
		return err
	}
	rw := responseWriter{}
	router.ServeHTTP(&rw, rq)

	/* create the file here so handler can set the filename first, when appropriate */
	outputPath := buildOutputPath()
	fp, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer fp.Close()

	_, err = io.Copy(fp, &rw.buf)
	return err
}
