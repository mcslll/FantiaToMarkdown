package fantia

import "time"

type Cookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	Expires  float64 `json:"expires"`
	HttpOnly bool   `json:"httpOnly"`
	Secure   bool   `json:"secure"`
}

type Post struct {
	Title       string
	Url         string
	PublishTime time.Time
}

type Fanclub struct {
	ID   string
	Name string
}
