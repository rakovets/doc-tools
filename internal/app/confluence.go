package app

type ConfluenceContent struct {
	Id      string  `json:"id"`
	Title   string  `json:"title"`
	Type    string  `json:"type"`
	Version Version `json:"version"`
	Body    Body    `json:"body"`
}

type ConfluenceContentRequest struct {
	Id      string      `json:"id"`
	Title   string      `json:"title"`
	Type    string      `json:"type"`
	Version Version     `json:"version"`
	Body    RequestBody `json:"body"`
}

type Body struct {
	Storage Content `json:"storage"`
}

type RequestBody struct {
	Wiki Content `json:"wiki"`
}

type Content struct {
	Value string `json:"value"`
}

// Version defines the content version number
// the version number is used for updating content
type Version struct {
	Number    int    `json:"number"`
	MinorEdit bool   `json:"minorEdit"`
	Message   string `json:"message,omitempty"`
	By        *User  `json:"by,omitempty"`
	When      string `json:"when,omitempty"`
}

// User defines user information
type User struct {
	Type        string `json:"type"`
	Username    string `json:"username"`
	UserKey     string `json:"userKey"`
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName"`
}
