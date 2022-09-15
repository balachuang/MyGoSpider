package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/magiconair/properties"
)

type queryCondition struct {
    queryString string
    queryTitle  string
    queryTarget string
    queryDetail string
    queryOutput bool
    queryOutParnt bool
    queryDrilldwn bool
    subQuery []queryCondition
}

var propFileName string = "spider.conf"
var fetchCount int = 0
var queryConditions []queryCondition
var includeEmpty bool = false
var outputUrl bool = false
var checkstopDepth bool = false
var checkstopCount bool = false
var stopDepth int = 3
var stopCount int = 100
var outputFile bool = true
var outputFileName string = "result.txt"
var logDocument bool = false

func main() {

    prop := properties.MustLoadFile(propFileName, properties.UTF8)

    startMethod := prop.MustGetString("start-method")
    startUrl := prop.MustGetString("start-url")
    startFile := prop.MustGetString("start-file")
    checkstopDepth = prop.MustGetBool("check-stop-by-depth")
    checkstopCount = prop.MustGetBool("check-stop-by-count")
    stopDepth = prop.MustGetInt("stop-depth")
    stopCount = prop.MustGetInt("stop-count")

    includeEmpty = prop.MustGetBool("include-empty-value")
    outputUrl = prop.MustGetBool("output-with-url")

    outputFile = prop.MustGetBool("output-to-file")
    outputFileName = prop.MustGetString("output-file-name")

    logDocument = prop.MustGetBool("log-document")

        // 預先準備要抓的 query string
    prepareQueryConditions(prop);

    // start crawler
    if strings.ToLower(startMethod) == "url" {
        // start from one url
        crawl(startUrl, 0)
    } else if strings.ToLower(startMethod) == "file" {
        // open file to get all urls
        file, err := os.Open(startFile)
        if err != nil {
            log.Fatalf("[ERROR] Error in open URL file: %s", err)
        }
        fileScanner := bufio.NewScanner(file)
        for fileScanner.Scan() { crawl(fileScanner.Text(), 0) }
        if err := fileScanner.Err(); err != nil {
            log.Fatalf("[ERROR] Error in read URL file: %s", err)
        }
        file.Close()
    }
}

func crawl(url string, depth int) {

    log.Println("target: ", url)

    // load dom object from url
    doc, success := loadUrl(url)
    if !success {
        log.Println("[ERROR] Not a valid document.")
        return
    }
    fetchCount = fetchCount + 1

    // show all document for debug
    if logDocument { log.Println(doc.Html()) }

    // Find target DOM items
    for _, thisQry := range queryConditions {
        doc.Find(thisQry.queryString).Each(func(idx int, s *goquery.Selection) {

            var value string = ""
            switch thisQry.queryTarget {
            case "TEXT":
                value = strings.TrimSpace(s.Text())
            case "ATTR":
                val, exist := s.Attr(thisQry.queryDetail)
                if exist { value = strings.TrimSpace(val) }
            }

            // prepare output string and output
            queryResult := ""
            if outputUrl {
                queryResult = fmt.Sprintf("[%s] %s: %s", url, thisQry.queryTitle, value)
            } else {
                queryResult = fmt.Sprintf("%s: %s", thisQry.queryTitle, value)
            }
            if thisQry.queryOutput && (includeEmpty || (value != "")) {
                // output to log
                log.Println("[OUTPUT]", queryResult)
                // output to file
                outputToFile(queryResult + "\r\n")
            }

            if thisQry.subQuery != nil {
                for _, thisSubQry := range thisQry.subQuery {
                    s.Find(thisSubQry.queryString).Each(func(idx int, ss *goquery.Selection) {
                        subVal := ""
                        switch thisSubQry.queryTarget {
                        case "TEXT":
                            subVal = strings.TrimSpace(ss.Text())
                        case "ATTR":
                            val, exist := ss.Attr(thisSubQry.queryDetail)
                            if exist { subVal = strings.TrimSpace(val) }
                        }
            
                        // prepare output string and output
                        querySubResult := ""
                        if outputUrl {
                            if thisSubQry.queryOutParnt {
                                querySubResult = fmt.Sprintf("[%s][%s] %s: %s", url, value, thisSubQry.queryTitle, subVal)
                            } else {
                                querySubResult = fmt.Sprintf("[%s] %s: %s", url, thisSubQry.queryTitle, subVal)
                            }
                        } else {
                            if thisSubQry.queryOutParnt {
                                querySubResult = fmt.Sprintf("[%s] %s: %s", value, thisSubQry.queryTitle, subVal)
                            } else {
                                querySubResult = fmt.Sprintf("%s: %s", thisSubQry.queryTitle, subVal)
                            }
                        }
                        if thisSubQry.queryOutput && (includeEmpty || (subVal != "")) {
                            // output to log
                            log.Println("[OUTPUT]", querySubResult)
                            // output to file
                            outputToFile(querySubResult + "\r\n")
                        }
                    })

                    // go drill-down url
                    if meetStopCritiron(depth) && thisSubQry.queryDrilldwn { crawl(value, depth+1) }
                }
            }

            // go drill-down url
            if meetStopCritiron(depth) && thisQry.queryDrilldwn { crawl(value, depth+1) }
        })
    }
}

func loadUrl(url string)(doc *goquery.Document, success bool) {

    // get HTML content
    res, err := http.Get(url)
    if err != nil {
        log.Println("[ERROR] Not a valid url")
        return nil, false
    }
    defer res.Body.Close()
    if res.StatusCode != 200 {
        log.Println("[ERROR] status code error: ", res.StatusCode, res.Status)
        return nil, false
    }
  
    // Load the HTML document
    doctmp, err := goquery.NewDocumentFromReader(res.Body)
    if err != nil {
        log.Println("[ERROR] Not a valid document.")
        return nil, false
    }

    return doctmp, true
}

// read query-conditions in properties
func prepareQueryConditions(prop *properties.Properties) {

    // get other conditions
    queryCount := prop.MustGetInt("query-string-count")
    for i := 1; i <= queryCount; i++ {
        // query parameters
        queryString   := prop.MustGetString(fmt.Sprintf("query-string-%d", i))
        queryTitle    := prop.MustGetString(fmt.Sprintf("query-string-%d-title", i))
        queryTarget   := prop.MustGetString(fmt.Sprintf("query-string-%d-target", i))
        queryDrilldwn := prop.MustGetBool(  fmt.Sprintf("query-string-%d-drilldown", i))
        queryOutput   := prop.MustGetBool(  fmt.Sprintf("query-string-%d-output", i))
        querySub      := prop.MustGetBool(  fmt.Sprintf("query-string-%d-sub-query", i))

        thisQueryCond := parseSingleCondition(
                queryString, queryTitle, queryTarget,
                queryDrilldwn, queryOutput, false)
        if querySub {
            subQryCount := prop.MustGetInt(fmt.Sprintf("query-string-%d-sub-count", i))
            for j := 1; j <= subQryCount; j++ {
                // query parameters
                subQryString   := prop.MustGetString(fmt.Sprintf("query-string-%d-sub-%d-string", i, j))
                subQryTitle    := prop.MustGetString(fmt.Sprintf("query-string-%d-sub-%d-title",  i, j))
                subQryTarget   := prop.MustGetString(fmt.Sprintf("query-string-%d-sub-%d-target", i, j))
                subQryDrilldwn := prop.MustGetBool(  fmt.Sprintf("query-string-%d-sub-%d-drilldown", i, j))
                subQryOutput   := prop.MustGetBool(  fmt.Sprintf("query-string-%d-sub-%d-output", i, j))
                subQryOutParnt := prop.MustGetBool(  fmt.Sprintf("query-string-%d-sub-%d-output-with-parent", i, j))

                thisSubQueryCond := parseSingleCondition(
                        subQryString, subQryTitle, subQryTarget,
                        subQryDrilldwn, subQryOutput, subQryOutParnt)
                thisQueryCond.subQuery = append(thisQueryCond.subQuery, thisSubQueryCond)
            }
        }

        queryConditions = append(queryConditions, thisQueryCond)
    }
}

func parseSingleCondition(qStr string, qTtl string, qTar string, qExp bool, qOut bool, qOutParnt bool)(qry queryCondition) {

    //thisQueryCond := new(queryCondition)
    var thisQueryCond queryCondition
    thisQueryCond.queryString = strings.TrimSpace(qStr)
    thisQueryCond.queryTitle = strings.TrimSpace(qTtl)

    queryTargetParts := strings.Split(strings.TrimSpace(qTar), ":")
    switch len(queryTargetParts) {
    case 1:
        thisQueryCond.queryTarget = strings.ToUpper(queryTargetParts[0])
        thisQueryCond.queryDetail = ""
    case 2:
        thisQueryCond.queryTarget = strings.ToUpper(queryTargetParts[0])
        thisQueryCond.queryDetail = queryTargetParts[1]
    }

    thisQueryCond.queryDrilldwn = qExp
    thisQueryCond.queryOutput = qOut
    thisQueryCond.queryOutParnt = qOutParnt

    return thisQueryCond
}

// check if stop spider
func meetStopCritiron(currDepth int)(isStop bool) {

    stop := false
    stop = stop || (checkstopDepth && (currDepth < stopDepth))
    stop = stop || (checkstopCount && (fetchCount < stopCount))

    return stop
}

// output to result file
func outputToFile(outputString string) {
    if outputFile {
        f, err := os.OpenFile(outputFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil { log.Println("[ERROR] Error in open output file: ", err) }
        defer f.Close()
        if _, err := f.WriteString(outputString); err != nil {
            log.Println("[ERROR] Error in write to output file: ", err)
        }
    }
}

/*
func queryHandler(url string, thisQry queryCondition, s *goquery.Selection)(json string) {
    // output of this query
    resultJson := fmt.Sprintf("{\r\n\turl: %s\r\n", url)

    thisVal := ""
    switch thisQry.queryTarget {
    case "TEXT":
        thisVal = s.Text()
    case "ATTR":
        val, exist := s.Attr(thisQry.queryDetail)
        if exist { thisVal = val }
    }

    if includeEmpty || (thisVal != "") {
        resultJson += fmt.Sprintf("\t%s: %s\r\n", thisQry.queryTitle, thisVal)
    }

    return resultJson
}
*/
