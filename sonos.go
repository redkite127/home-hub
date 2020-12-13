package main

import (
	"net/http"
	"net/url"

	"github.com/spf13/viper"
)

func PlayTocToc() {
	//str := "toc toc toc"
	str := viper.GetString("toc_toc")
	go http.Get(viper.GetString("sonos_api_uri") + "/sayall/" + url.PathEscape(str) + "/fr/65")
}
