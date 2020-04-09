package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func PlayTocToc() {

	str := "Hi"
	dat, err := ioutil.ReadFile("./toc_toc.txt")
	if err == nil {
		str = strings.TrimSuffix(string(dat), "\n")
	}
	go http.Get("http://10.161.0.111:5005/Kitchen/say/" + url.PathEscape(str))
	go http.Get("http://10.161.0.111:5005/Dressing/say/" + url.PathEscape(str))
	go http.Get("http://10.161.0.111:5005/Dining%20Room/say/" + url.PathEscape(str))
}
