package VK

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

func New(token string, versionAPI string) *VK {
	query := query{URL: url.URL{Scheme: "https", Host: "api.vk.com"}}

	values := query.Query()
	values.Add("access_token", token)
	values.Add("v", versionAPI)

	query.RawQuery = values.Encode()

	std := log.New(os.Stderr, "", log.LstdFlags)

	return &VK{query, false, std}
}

type VK struct {
	url   query
	Debug bool
	log   *log.Logger
}

func (vk *VK) SetLogger(log *log.Logger) {
	vk.log = log
}

type query struct {
	url.URL
	url.Values
}

type ErrorResponse struct {
	Code    int    `json:"error_code"`
	Message string `json:"error_msg"`
}

func (e ErrorResponse) String() string {
	return fmt.Sprintf("VK: code: %d; messages: %s", e.Code, e.Message)
}

func (e ErrorResponse) error() error {
	if e.Message == "" {
		return nil
	} else {
		return errors.New(fmt.Sprint(e))
	}
}

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

type Attachment struct {
	Type  string `json:"type"`
	Photo Photo  `json:"photo"`
}

type Photo struct {
	Id      int    `json:"id"`
	AlbumID int    `json:"album_id"`
	OwnerID int    `json:"owner_id"`
	UserID  int    `json:"user_id"`
	Text    string `json:"text"`
	Date    int    `json:"date"`
	Sizes   []struct {
		Type   string `json:"type"`
		Url    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"sizes"`
}
type LongPollServer struct {
	Response struct {
		Key    string `json:"key"`
		Server string `json:"server"`
		Ts     string `json:"ts"`
	} `json:"response"`
	Error ErrorResponse
}

type ResponseLongPollServer struct {
	Ts      string              `json:"ts"`
	Updates []LongPollingUpdate `json:"updates"`
}

func (vk VK) GetMessages() (Message, error) {
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

	return *messages, nil //messages.Error.error()
}

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

	error := struct {
		Error ErrorResponse `json:"error"`
	}{}

	if err := json.NewDecoder(resp).Decode(&error); err != nil {
		return err
	}

	return error.Error.error()
}

type User struct {
	ID              int    `json:"id"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	IsClosed        bool   `json:"is_closed"`
	CanAccessClosed bool   `json:"can_access_closed"`
}

type UserList struct {
	Response struct {
		Count int    `json:"count"`
		Items []User `json:"items"`
	} `json:"response"`
}

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
		Error ErrorResponse `json:"error"`
	})

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return user.User, err
	}
	return user.User, user.Error.error()

}

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
		Error    ErrorResponse `json:"error"`
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

type GetMessages struct {
	Conversation Conversation `json:"conversation"`
	Messages     Message      `json:"last_message"`
	Profiles     User       `json:"profiles"`
}

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
		Error ErrorResponse `json:"error"`
	})

	if err := json.NewDecoder(resp).Decode(&user); err != nil {
		return user.User, err
	}
	return user.User, user.Error.error()
}

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
		Error ErrorResponse `json:"error"`
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
