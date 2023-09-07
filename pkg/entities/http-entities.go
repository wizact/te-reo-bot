package entities

// AppError as app error container
type AppError struct {
	Error   error  `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// PostResponse is the tweet/mastodon Id after a successful update operation
type PostResponse struct {
	TwitterId string `json:"tweetId"`
	TootId    string `json:"tootId"`
	Message   string `json:"message"`
}

// FriendlyError is sanitised error message sent back to the user
type FriendlyError struct {
	Message string `json:"message"`
}
