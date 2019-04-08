package vk

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// VK main vk instance
type VK struct {
	url   query
	Debug bool
	log   *log.Logger
}

type query struct {
	url.URL
	url.Values
}

// Conversation vk chats
type Conversation struct {
	Peer struct {
		ID      int    `json:"id"`
		Type    string `json:"type"`
		LocalID int    `json:"local_id"`
	} `json:"peer"`
	InRead        int `json:"in_read"`
	OutRead       int `json:"out_read"`
	LastMessageID int `json:"last_message_id"`
	CanWrite      struct {
		Allowed bool `json:"allowed"`
	} `json:"can_write"`
}

// GetMessages get message info
type GetMessages struct {
	Conversation Conversation `json:"conversation"`
	Messages     Message      `json:"last_message"`
	Profiles     User         `json:"profiles"`
}

// Message vk message
type Message struct {
	ID          int          `json:"id"`
	Date        int          `json:"date"`
	Out         int          `json:"out"`
	UserID      int          `json:"user_id"`
	ReadState   int          `json:"read_state"`
	Title       string       `json:"title"`
	Body        string       `json:"body"`
	Emoji       int          `json:"emoji"`
	Text        string       `json:"text"`
	FromID      int          `json:"from_id"`
	Attachments []Attachment `json:"attachments"`
}

// Attachment vk attachment
type Attachment struct {
	Type  string `json:"type"`
	Photo Photo  `json:"photo"`
}

// Photo photo vk
type Photo struct {
	ID      int    `json:"id"`
	AlbumID int    `json:"album_id"`
	OwnerID int    `json:"owner_id"`
	UserID  int    `json:"user_id"`
	Text    string `json:"text"`
	Date    int    `json:"date"`
	Sizes   []struct {
		Type   string `json:"type"`
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"sizes"`
}

type errorResponse struct {
	Code    int    `json:"error_code"`
	Message string `json:"error_msg"`
}

// User vk account info
type User struct {
	ID              int    `json:"id"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	IsClosed        bool   `json:"is_closed"`
	CanAccessClosed bool   `json:"can_access_closed"`
}

// New new vk instance
func New(token, versionAPI string) *VK {
	query := query{URL: url.URL{Scheme: "https", Host: "api.vk.com"}}

	values := query.Query()
	values.Add("access_token", token)
	values.Add("v", versionAPI)

	query.RawQuery = values.Encode()

	std := log.New(os.Stderr, "", log.LstdFlags)

	return &VK{query, false, std}
}

// SetLogger set specific logger for vk bot
func (vk *VK) SetLogger(logger *log.Logger) {
	vk.log = logger
}

// GetConversations get conversations
func (vk VK) GetConversations() (Message, error) {
	query := vk.url.Query()

	query.Add("count", "2")

	vk.url.Path = "/method/messages.getConversations"
	vk.url.RawQuery = query.Encode()

	resp, err := vk.exec()
	if err != nil {
		return Message{}, err
	}

	messages := new(Message)

	if err := json.NewDecoder(resp).Decode(&messages); err != nil {
		fmt.Println(err)
	}

	return *messages, nil
}

// MessagesSend  send messages for specific user by id
func (vk VK) MessagesSend(message string, userID int) error {
	vk.url.Path = "/method/messages.send"

	query := vk.url.Query()
	query.Add("user_id", strconv.Itoa(userID))
	query.Add("message", message)
	query.Add("random_id", strconv.Itoa(int(rand.Int31n(math.MaxInt32))))

	vk.url.RawQuery = query.Encode()

	resp, err := vk.exec()
	if err != nil {
		return err
	}

	errorReq := struct {
		Error errorResponse `json:"errorReq"`
	}{}

	if err := json.NewDecoder(resp).Decode(&errorReq); err != nil {
		return err
	}

	return errorReq.Error.error()
}

// SearchUser get user by specific query
func (vk VK) SearchUser(searchQuery string) (User, error) {
	vk.url.Path = "/method/users.search"

	query := vk.url.Query()
	query.Add("q", searchQuery)
	vk.url.RawQuery = query.Encode()

	resp, err := http.Get(vk.url.String())
	if err != nil {
		return User{}, err
	}

	user := new(struct {
		User
		Error errorResponse `json:"error"`
	})

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return user.User, err
	}
	return user.User, user.Error.error()

}

// UsersGet get user info by []id
func (vk VK) UsersGet(ids ...int) ([]User, error) {
	vk.url.Path = "/method/users.get"

	var idString string

	log.Println(len(ids))

	for _, id := range ids {

		switch len(ids) {
		case 0:
			return nil, errors.New("id length is zero")
		case 1:
			idString += strconv.Itoa(id)
		}

		idString += strconv.Itoa(id) + ","

	}

	query := vk.url.Query()
	query.Add("user_ids", idString)

	vk.url.RawQuery = query.Encode()

	response := struct {
		Response []User        `json:"response"`
		Error    errorResponse `json:"error"`
	}{}

	resp, err := vk.exec()
	if err != nil {
		return nil, err
	}
	if err := json.NewDecoder(resp).Decode(&response); err != nil {
		return nil, err
	}

	return response.Response, response.Error.error()

}

// SearchFriends search friends by specific query
func (vk VK) SearchFriends(searchQuery string) (User, error) {
	vk.url.Path = "/method/friends.search"

	query := vk.url.Query()
	query.Add("q", searchQuery)
	vk.url.RawQuery = query.Encode()

	resp, err := vk.exec()
	if err != nil {
		return User{}, err
	}

	user := new(struct {
		User
		Error errorResponse `json:"error"`
	})

	if err := json.NewDecoder(resp).Decode(&user); err != nil {
		return user.User, err
	}
	return user.User, user.Error.error()
}

// MessagesGetConversations todo rewrite all vk bot
func (vk VK) MessagesGetConversations() ([]GetMessages, error) {
	vk.url.Path = "/method/messages.getConversations"

	query := vk.url.Query()
	query.Add("count", "20")
	query.Add("extended", "true")
	vk.url.RawQuery = query.Encode()

	resp, err := vk.exec()
	if err != nil {
		return []GetMessages{}, err
	}

	response := struct {
		Response struct {
			Items    []GetMessages `json:"items"`
			Profiles []User        `json:"profiles"`
		} `json:"response"`
		Error errorResponse `json:"error"`
	}{}

	if err := json.NewDecoder(resp).Decode(&response); err != nil {
		return []GetMessages{}, err
	}

	return response.Response.Items, response.Error.error()

}

func (vk VK) exec() (io.Reader, error) {
	resp, err := http.Get(vk.url.URL.String())
	if err != nil {
		return nil, err
	}

	if vk.Debug {
		buf := bytes.Buffer{}

		if _, err := buf.ReadFrom(resp.Body); err != nil {
			return nil, err
		}

		vk.log.Printf("[debug] : %v", buf.String())

		return &buf, nil
	}
	return resp.Body, nil
}

func (e errorResponse) String() string {
	return fmt.Sprintf("vk: code: %d; messages: %s", e.Code, e.Message)
}

func (e errorResponse) error() error {
	if e.Message == "" {
		return nil
	}
	return errors.New(fmt.Sprint(e))
}
