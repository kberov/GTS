package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Page struct {
	Message string
}

var port int

var wordsMap = make(map[string]string)
var wordsList []string
var wordsListMapOrdered []map[string]string

//Prepare regexps to match the provided word for translation
const vowels = `^(?:a|e|i|o|u)\w+?`
const consonants = `^(b|c|d|f|g|h|j|k|l|m|n|p|q|r|s|t|v|w|x|y|z)(\w+?)$`
const xr = "(?i)^xr"
const cons_qu = "^(" + consonants + "qu" + ")"

var re_vowels = regexp.MustCompile(`(?i)` + vowels)
var re_xr = regexp.MustCompile(xr)
var re_cons = regexp.MustCompile("(?i)" + consonants)
var re_cons_qu = regexp.MustCompile(cons_qu)
var re_last = regexp.MustCompile(`^(\w+)([\.\?\!])$`)

// A template instance prepared, once and executed many times.
var t = template.Must(template.ParseFiles("templates/index.html"))

func main() {
	parseFlags()
	serve()
}

func parseFlags() {
	flag.IntVar(&port, "port", 8080, "The port on which this server listens. Defaults to 8080.")
	// More flags if needed here
	flag.Parse()
}

func serve() {
	http.HandleFunc("/", indexPage)
	http.HandleFunc("/word", addWord)
	http.HandleFunc("/sentence", addSentence)
	http.HandleFunc("/history", showHistory)
	// Prepare to run
	port_string := fmt.Sprintf(":%s", strconv.Itoa(port))
	log.Printf("Serving at http://localhost%s\n", port_string)
	// Run.
	log.Fatal(http.ListenAndServe(port_string, nil))
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	pathNotFound := ""
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		pathNotFound = `The page "` + r.URL.Path + ` was not found!`
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	p := &Page{Message: pathNotFound}
	t.Execute(w, p)
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

	return wordsMap[word]
}

func translate(word string) string {
	if re_vowels.MatchString(word) {
		log.Printf(`Word matched "%s".`, vowels)
		return "g" + word
	}
	if re_xr.MatchString(word) {
		log.Printf(`Word matched "%s".`, xr)
		return "ge" + word
	}
	if re_cons.MatchString(word) {
		log.Printf(`Word matched "%s".`, consonants)
		return re_cons.ReplaceAllString(word, "${2}${1}ogo")
	}
	if re_cons_qu.MatchString(word) {
		log.Printf(`Word matched "%s".`, cons_qu)
		return re_cons.ReplaceAllString(word, "$1ogo")
	}
	return word
}

func addSentence(res http.ResponseWriter, req *http.Request) {
	englishSentence := make(map[string]string)
	if req.Method == http.MethodGet {
		http.Redirect(res, req, "/", http.StatusFound)
		return
	}
	res.Header().Set("Content-Type", "application/json; charset=utf-8")

	body, _ := ioutil.ReadAll(req.Body)
	json.Unmarshal([]byte(body), &englishSentence)
	log.Printf(`english-sentence: %s\n`, englishSentence["english-sentence"])
	gopherSentense :=
		regexp.MustCompile(`\s+`).Split(englishSentence["english-sentence"], -1)
	lenght := len(gopherSentense)
	log.Printf("split: %v\n lenght: %d", gopherSentense, lenght)

	if re_last.MatchString(gopherSentense[lenght-1]) {
		last_two := re_last.FindAllStringSubmatch(gopherSentense[lenght-1], -1)
		log.Printf("popped: %v\nlast_two:'%v'", gopherSentense, last_two)
		gopherSentense[lenght-1] = last_two[0][1]
		gopherSentense = append(gopherSentense, last_two[0][2])
	}
	for i := 0; i < lenght; i++ {
		gopherSentense[i] = translate(gopherSentense[i])
	}

	fmt.Fprint(res,
		fmt.Sprintf(`{"gopher-sentense":"%s"}`,
			strings.Join(gopherSentense, " ")))
}

func showHistory(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json; charset=utf-8")

	for i := 0; i < len(wordsList); i++ {
		translation := make(map[string]string)
		translation[wordsList[i]] = wordsMap[wordsList[i]]
		wordsListMapOrdered = append(wordsListMapOrdered, translation)
	}

	listOfTranslations, _ := json.Marshal(wordsListMapOrdered)
	fmt.Fprint(res, string(listOfTranslations)+"\n")
}
