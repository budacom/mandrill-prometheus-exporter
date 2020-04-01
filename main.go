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

	accountSentTotal   = prometheus.NewDesc("mandrill_account_sent_total", "Accounts total number of sent mails", []string{"subaccount"}, nil)
	accountReputation  = prometheus.NewDesc("mandrill_account_reputation", "Accounts Mandrill reputation", []string{"subaccount"}, nil)
	accountCustomQuota = prometheus.NewDesc("mandrill_account_custom_quota", "Accounts Mandrill custom quota", []string{"subaccount"}, nil)

	userReputation  = prometheus.NewDesc("mandrill_user_reputation", "User Mandrill reputation", []string{"username"}, nil)
	userHourlyQuota = prometheus.NewDesc("mandrill_user_hourly_quota", "User hourly quota", []string{"username"}, nil)
	userBacklog     = prometheus.NewDesc("mandrill_user_backlog", "User mail backlog", []string{"username"}, nil)

	senderSent = prometheus.NewDesc("mandrill_sender_sent_total", "Sender total number of sent mails.", []string{"address"}, nil)
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

	ch <- accountSentTotal
	ch <- accountReputation
	ch <- accountCustomQuota

	ch <- userReputation
	ch <- userHourlyQuota
	ch <- userBacklog

	ch <- senderSent
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

type mandrillSubtaccountData struct {
	ID          string
	CustomQuota int `json:"custom_quota"`
	Reputation  int
	SendTotal   int `json:"sent_total"`
}

type mandrillUserData struct {
	Username    string
	Reputation  int
	HourlyQuota int `json:"hourly_quota"`
	Backlog     int
}

type mandrillSenderData struct {
	Address string
	Sent    int
}

func (m mandrillCollector) Collect(ch chan<- prometheus.Metric) {

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

	//get Subaccounts from Mandrill
	subaccountsData, err := m.getSubaccounts()
	if err != nil {
		log.Print(err)
	}

	//iterate over tags and get stats
	for _, subaccount := range subaccountsData {
		ch <- prometheus.MustNewConstMetric(accountSentTotal, prometheus.GaugeValue, float64(subaccount.SendTotal), subaccount.ID)
		ch <- prometheus.MustNewConstMetric(accountReputation, prometheus.GaugeValue, float64(subaccount.Reputation), subaccount.ID)
		ch <- prometheus.MustNewConstMetric(accountCustomQuota, prometheus.GaugeValue, float64(subaccount.CustomQuota), subaccount.ID)
	}

	//get User from Mandrill
	userData, err := m.getUser()
	if err != nil {
		log.Print(err)
	}

	// get user stats
	ch <- prometheus.MustNewConstMetric(userReputation, prometheus.GaugeValue, float64(userData.Reputation), userData.Username)
	ch <- prometheus.MustNewConstMetric(userHourlyQuota, prometheus.GaugeValue, float64(userData.HourlyQuota), userData.Username)
	ch <- prometheus.MustNewConstMetric(userBacklog, prometheus.GaugeValue, float64(userData.Backlog), userData.Username)

	//get senders from Mandrill
	sendersData, err := m.getSenders()
	if err != nil {
		log.Print(err)
	}

	//iterate over tags and get stats
	for _, sender := range sendersData {
		ch <- prometheus.MustNewConstMetric(senderSent, prometheus.GaugeValue, float64(sender.Sent), sender.Address)
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

func (m mandrillCollector) getSubaccounts() ([]mandrillSubtaccountData, error) {

	body := bytes.Buffer{}
	body.WriteString("{\"key\": \"")
	body.WriteString(m.apiKey)
	body.WriteString("\"}")

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://mandrillapp.com/api/1.0/subaccounts/list.json", &body)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	result := []mandrillSubtaccountData{}
	err = json.NewDecoder(resp.Body).Decode(&result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (m mandrillCollector) getUser() (mandrillUserData, error) {

	body := bytes.Buffer{}
	body.WriteString("{\"key\": \"")
	body.WriteString(m.apiKey)
	body.WriteString("\"}")

	result := mandrillUserData{}
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://mandrillapp.com/api/1.0/users/info.json", &body)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&result)

	if err != nil {
		return result, err
	}

	return result, nil
}

func (m mandrillCollector) getSenders() ([]mandrillSenderData, error) {

	body := bytes.Buffer{}
	body.WriteString("{\"key\": \"")
	body.WriteString(m.apiKey)
	body.WriteString("\"}")

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://mandrillapp.com/api/1.0/users/senders.json", &body)
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	result := []mandrillSenderData{}
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

	//health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	//port 9153 https://github.com/prometheus/prometheus/wiki/Default-port-allocations
	http.ListenAndServe(":9153", nil)
}
