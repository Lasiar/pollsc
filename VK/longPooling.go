package VK

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

type LongPollingUpdate struct {
	Type   string `json:"type"`
	Object struct {
		Date        int          `json:"date"`
		Text        string       `json:"text"`
		FromID      int          `json:"from_id"`
		Attachments []Attachment `json:"attachments"`
	} `json:"object"`
}

func (vk VK) GetLongPoolServer(GroupID int) (LongPollServer, error) {
	vk.url.Path = "/method/groups.getLongPollServer"

	query := vk.url.Query()
	query.Add("group_id", strconv.Itoa(GroupID))
	vk.url.RawQuery = query.Encode()

	resp, err := vk.exec()
	if err != nil {
		return LongPollServer{}, err
	}

	longPollServer := new(LongPollServer)

	if err := json.NewDecoder(resp).Decode(&longPollServer); err != nil {
		return LongPollServer{}, err
	}

	return *longPollServer, longPollServer.Error.error()
}

func (lps LongPollServer) Listen() chan LongPollingUpdate {
	c := make(chan LongPollingUpdate, 1024)

	u, err := new(url.URL).Parse(lps.Response.Server)
	if err != nil {
		log.Fatal(err)
	}
	query := u.Query()

	query.Add("act", "a_check")
	query.Add("key", lps.Response.Key)
	query.Add("ts", lps.Response.Ts)
	query.Add("wait", "25")
	u.RawQuery = query.Encode()

	responseLongPollServer := new(ResponseLongPollServer)

	go func(u *url.URL, c chan LongPollingUpdate, response ResponseLongPollServer) {

		for {

			buf := bytes.Buffer{}

			resp, err := http.Get(u.String())
			if err != nil {
				log.Fatal(err)
			}

			buf.ReadFrom(resp.Body)

			//	log.Println(buf.String())

			if err := json.NewDecoder(&buf).Decode(&responseLongPollServer); err != nil {
				log.Println(err)
			}

			for _, msg := range responseLongPollServer.Updates {
				c <- msg
			}

			responseLongPollServer.Updates = nil
			query := u.Query()
			query.Set("ts", responseLongPollServer.Ts)
			u.RawQuery = query.Encode()
		}

	}(u, c, *responseLongPollServer)

	return c
}
