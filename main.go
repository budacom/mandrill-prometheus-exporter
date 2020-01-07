package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	sent         = prometheus.NewDesc("mandrill_sent_total", "Total number of sent mails.", []string{"tag"}, nil)
	hardBounces  = prometheus.NewDesc("mandrill_hard_bounces", "Number of mails bounced hard", []string{"tag"}, nil)
	softBounces  = prometheus.NewDesc("mandrill_soft_bounces", "Number of mails bounced soft", []string{"tag"}, nil)
	rejects      = prometheus.NewDesc("mandrill_rejects", "Number of mails rejected", []string{"tag"}, nil)
	complaints   = prometheus.NewDesc("mandrill_complaints", "Number of complaints", []string{"tag"}, nil)
	unsubs       = prometheus.NewDesc("mandrill_unsubs", "Number of unsubscribes", []string{"tag"}, nil)
	opens        = prometheus.NewDesc("mandrill_opens", "Number of mails opened", []string{"tag"}, nil)
	clicks       = prometheus.NewDesc("mandrill_clicks", "Number of clicks inside mails", []string{"tag"}, nil)
	uniqueOpens  = prometheus.NewDesc("mandrill_unique_opens", "Unique number of mails opened", []string{"tag"}, nil)
	uniqueClicks = prometheus.NewDesc("mandrill_unique_clicks", "Unique number of clicks", []string{"tag"}, nil)
	reputation   = prometheus.NewDesc("mandrill_reputation", "Mandrill reputation", []string{"tag"}, nil)
)

type mandrillCollector struct {
	apiKey string
}

func (m mandrillCollector) Describe(ch chan<- *prometheus.Desc) {

	ch <- sent
	ch <- hardBounces
	ch <- softBounces
	ch <- rejects
	ch <- complaints
	ch <- unsubs
	ch <- opens
	ch <- clicks
	ch <- uniqueOpens
	ch <- uniqueClicks
	ch <- reputation
}

type mandrillTagData struct {
	Tag          string
	Sent         int
	HardBounces  int `json:"hard_bounces"`
	SoftBounces  int `json:"soft_bounces"`
	Rejects      int
	Complaints   int
	Unsubs       int
	Opens        int
	Clicks       int
	UniqueOpens  int `json:"unique_opens"`
	UniqueClicks int `json:"unique_clicks"`
	Reputation   int
}

func (m mandrillCollector) Collect(ch chan<- prometheus.Metric) {

	log.Print("getting data")
	//get Tags from Mandrill
	tagData, err := m.getTags()
	if err != nil {
		log.Print(err)
	}

	//iterate over tags and get stats
	for _, tag := range tagData {
		ch <- prometheus.MustNewConstMetric(sent, prometheus.CounterValue, float64(tag.Sent), tag.Tag)
		ch <- prometheus.MustNewConstMetric(hardBounces, prometheus.CounterValue, float64(tag.HardBounces), tag.Tag)
		ch <- prometheus.MustNewConstMetric(softBounces, prometheus.CounterValue, float64(tag.SoftBounces), tag.Tag)
		ch <- prometheus.MustNewConstMetric(rejects, prometheus.CounterValue, float64(tag.Rejects), tag.Tag)
		ch <- prometheus.MustNewConstMetric(complaints, prometheus.CounterValue, float64(tag.Complaints), tag.Tag)
		ch <- prometheus.MustNewConstMetric(unsubs, prometheus.CounterValue, float64(tag.Unsubs), tag.Tag)
		ch <- prometheus.MustNewConstMetric(opens, prometheus.CounterValue, float64(tag.Opens), tag.Tag)
		ch <- prometheus.MustNewConstMetric(clicks, prometheus.CounterValue, float64(tag.Clicks), tag.Tag)
		ch <- prometheus.MustNewConstMetric(uniqueOpens, prometheus.CounterValue, float64(tag.UniqueOpens), tag.Tag)
		ch <- prometheus.MustNewConstMetric(uniqueClicks, prometheus.CounterValue, float64(tag.UniqueClicks), tag.Tag)
		ch <- prometheus.MustNewConstMetric(reputation, prometheus.CounterValue, float64(tag.Reputation), tag.Tag)
	}
}

func (m mandrillCollector) getTags() ([]mandrillTagData, error) {

	body := bytes.Buffer{}
	body.WriteString("{\"key\": \"")
	body.WriteString(m.apiKey)
	body.WriteString("\"}")

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://mandrillapp.com/api/1.0/tags/list.json", &body)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	result := []mandrillTagData{}
	err = json.NewDecoder(resp.Body).Decode(&result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func main() {

	mc := mandrillCollector{
		apiKey: os.Getenv("MANDRILL_API_KEY"),
	}

	reg := prometheus.NewPedanticRegistry()

	reg.MustRegister(mc)

	//default Seite
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		             <head><title>Mandrill statistics Exporter</title></head>
		             <body>
		             <h1>Madrill statistics Exporter</h1>
		             <p><a href='metrics'>Metrics</a></p>
		             </body>
		             </html>`))
	})
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	//port 9153 https://github.com/prometheus/prometheus/wiki/Default-port-allocations
	http.ListenAndServe(":9153", nil)
}
