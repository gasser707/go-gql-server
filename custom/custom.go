package custom

import (
	"time"
	"github.com/gasser707/go-gql-server/graph/model"
)


type Image struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	UserID       string      `json:"userId"`
	Labels      []string   `json:"labels"`
	URL         string     `json:"url"`
	Private     bool       `json:"private"`
	ForSale     bool       `json:"forSale"`
	Created     *time.Time `json:"created"`
	Price       float64    `json:"price"`
}


type Sale struct {
	ID     string     `json:"id"`
	Image  *Image     `json:"image"`
	BuyerID  string      `json:"buyerId"`
	SellerID string      `json:"sellerId"`
	Time   *time.Time `json:"time"`
	Price  float64    `json:"price"`
}


type User struct {
	ID       string     `json:"id"`
	Username string     `json:"username"`
	Email    string     `json:"email"`
	Role     model.Role       `json:"role"`
	Bio      string     `json:"bio"`
	Avatar   string     `json:"avatar"`
	Joined   *time.Time `json:"joined"`
	Images   []*Image   `json:"images"`
}
