package feedcast_crawler

import (
	"encoding/json"
	"feedcast.crawler/crawler"
	_ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"net/http"
)

func init() {
	functions.HTTP("crawl", crawl)
}

func crawl(w http.ResponseWriter, r *http.Request) {
	var req crawler.Config

	if "POST" == r.Method {
		if err := json.NewDecoder(r.Body).Decode(&req); nil != err {
			w.WriteHeader(http.StatusBadRequest)
			res, _ := json.Marshal(map[string]string{"error": err.Error()})
			w.Write(res)
			return
		}

		req.Touch()
		if err := req.Validate(); nil != err {
			w.WriteHeader(http.StatusBadRequest)
			res, _ := json.Marshal(map[string]string{"error": err.Error()})
			w.Write(res)
			return
		}

		cr := crawler.NewCrawler(req)
		result := cr.Run()
		var buf []byte

		if req.WithPageContent {
			buf, _ = json.Marshal(result)
		} else {
			links := []string{}
			for k, _ := range result {
				links = append(links, k)
			}
			buf, _ = json.Marshal(links)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(buf)

		return
	}

	w.WriteHeader(http.StatusBadRequest)
}
