package VK

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func (vk VK) GetMessagesUploadServer(peerID int) (MessagesUploadServer, error) {
	vk.url.Path = "/method/docs.getMessagesUploadServer"

	query := vk.url.Query()
	query.Add("peer_id", strconv.Itoa(peerID))
	vk.url.RawQuery = query.Encode()

	resp, err := http.Get(vk.url.String())
	if err != nil {
		return MessagesUploadServer{}, err
	}

	response := struct {
		Response struct {
			UploadURL string `json:"upload_url"`
		} `json:"response"`
		Error ErrorResponse `json:"error"`
	}{}

	m := new(MessagesUploadServer)

	buf := bytes.Buffer{}
	buf.ReadFrom(resp.Body)
	fmt.Println(buf.String())

	if err := json.NewDecoder(&buf).Decode(&response); err != nil {
		return MessagesUploadServer{}, err
	}

	m.URL, err = url.Parse(response.Response.UploadURL)
	if err != nil {
		return MessagesUploadServer{}, err
	}

	return *m, response.Error.error()
}

type MessagesUploadServer struct {
	URL *url.URL `json:"upload_url"`
}

type File struct {
	File string `json:"file"`
}

func (m MessagesUploadServer) SendFile(reader io.Reader) (File, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("test", "test")
	if err != nil {
		return File{}, err
	}

	fh, err := os.Open("test")
	if err != nil {
		return File{}, err
	}
	defer fh.Close()

	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		return File{}, err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	fmt.Println(contentType)
	fmt.Println(bodyBuf.String())

	fmt.Println(m.URL.String())
	resp, err := http.Post(m.URL.String(), contentType, bodyBuf)
	if err != nil {
		return File{}, err
	}

	io.Copy(os.Stdout, resp.Body)
	return File{}, nil
}
