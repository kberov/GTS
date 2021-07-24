package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
)

var port int
var wordsMap = make(map[string]string, 100)
var wordsList = make([]string, 100)
var wordsListMapOrdered = make([]map[string]string, 100)

//Prepare regexps to match the provided word for translation
const vowels = "^(?:a|e|i|o|u)"
const consonants = "^(b|c|d|f|g|h|j|k|l|m|n|p|q|r|s|t|v|w|x|y|z)"
const xr = "^xr"
const cons_qu = "^(" + consonants + "qu" + ")"

var re_vowels = regexp.MustCompile(`(?i)` + vowels)
var re_xr = regexp.MustCompile(xr)
var re_const_qu = regexp.MustCompile(cons_qu)

func main() {
	parseFlags()
	serve()
	//fmt.Println("vim-go")
}

func parseFlags() {
	flag.IntVar(&port, "port", 8080, "The port on which this server listens. Defaults to 8080.")
	// More flags if needed here
	flag.Parse()
}

func serve() {
	// We register our handlers on server routes using the http.HandleFunc
	// convenience function.
	http.HandleFunc("/", indexPage)
	http.HandleFunc("/word", addWord)
	// Prepare to run
	port_string := fmt.Sprintf(":%s", strconv.Itoa(port))
	log.Printf("Serving at http://localhost%s\n", port_string)
	// Run.
	log.Fatal(http.ListenAndServe(port_string, nil))
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
<html>
  <head>
	<title>Go Translator Service</title>
  </head>
  <body>
    <h1>Go Translator Service</h1>
    <p>To translate a word, please execute a POST request to <a
    href="/word">/word</a> with JSON body like
    <code>{“english-word”:”&lt;a single English word&gt;”}</code>,
    which will be translated into Gophers' language and appended to a
    list of translated words.</p> <p>To retrieve the list of words
    added in JSON format, make a GET request to <a
    href="/history">/history</a>.</p>
  </body>
<html>
`)
}

func addWord(res http.ResponseWriter, req *http.Request) {
	englishWord := make(map[string]string)
	if req.Method == http.MethodGet {
		http.Redirect(res, req, "/", http.StatusFound)
		return
	}
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	body, _ := ioutil.ReadAll(req.Body)
	//TODO: implement  some validation
	json.Unmarshal([]byte(body), &englishWord)
	log.Printf(`english-word: %S\n`, englishWord["english-word"])
	//TODO: write the translator function
	fmt.Fprint(res, fmt.Sprintf(`{"gopher-word":"%s"}`, translateAndAdd(englishWord["english-word"])))
	fmt.Fprint(res, "\n")
}

func translateAndAdd(word string) string {
	wordsList = append(wordsList, word)
	sort.Slice(wordsList, func(i, j int) bool {
		return wordsList[i] < wordsList[j]
	})
	wordsMap[word] = translate(word)
	log.Printf(`
	The words now are: %v
	The list now is:   %v
`, wordsMap, wordsList)
	return word
}

func orderByEnglish() {

}

func translate(word string) string {
	if re_vowels.MatchString(word) {
		return "g" + word
	}
	if re_xr.MatchString(word) {
		return "ge" + word
	}
	//if(yes, err = re_xr)
	return fmt.Sprintf("tr%s", word)
}
