package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type clientRequest struct {
	RequestUrl  string `json:"requestUrl"`
	RequestType string `json:"requestType"`
	RequestJson string `json:"requestJson"`
	RequestKey  string `json:"requestKey"`
}

func writeError(w http.ResponseWriter, errorCode int, errorMessage string) {
	log.Println("ERROR:", errorMessage)
	w.WriteHeader(errorCode)
	fmt.Fprintf(w, "{\"error\":\""+errorMessage+"\"}")
}

func isJson(s string) bool {
	var js map[string]interface{}
	    var jsArray []interface{}
	    return json.Unmarshal([]byte(s), &js) == nil || json.Unmarshal([]byte(s), &jsArray) == nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, X-Requested-With, Session")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		writeError(w, http.StatusBadRequest, "Request must be POST")
		return
	}

	bits, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var cr clientRequest
	err = json.Unmarshal(bits, &cr)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	client := &http.Client{}
	jsonStr := []byte(cr.RequestJson)
	req, err := http.NewRequest(cr.RequestType, cr.RequestUrl, bytes.NewBuffer(jsonStr))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Header.Add("Authorization", cr.RequestKey)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	stringBody := string(body)
	bodyIsJson := isJson(stringBody)
	log.Println(resp.StatusCode, "API RESPONSE:", stringBody)
        if resp.StatusCode > 299 {
		if bodyIsJson {
                	w.WriteHeader(resp.StatusCode)
			w.Write(body)
		} else {
			writeError(w, resp.StatusCode, stringBody)
		}
		return
        }

	if bodyIsJson {
		w.Write(body)
	} else {
		fmt.Fprintf(w, "{\"message\":\"" + stringBody + "\"}")
	}

}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
