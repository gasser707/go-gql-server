package databases

import "time"

type Image struct {
	ID              int       `db:"id"`
	CreatedAt       time.Time `db:"created_at"`
	URL             string    `db:"url"`
	Description     string    `db:"description"`
	UserID          int       `db:"user_id"`
	Title           string    `db:"title"`
	Price           float64   `db:"price"`
	ForSale         bool      `db:"forSale"`
	Private         bool      `db:"private"`
	Archived        bool      `db:"archived"`
	DiscountPercent int       `db:"discountPercent"`
}

type Label struct {
	ID      int    `db:"id"`
	Tag     string `db:"tag"`
	ImageID int    `db:"image_id"`
}

// Sale is an object representing the database table.
type Sale struct {
	ID        int       `db:"id"`
	ImageID   int       `db:"image_id"`
	BuyerID   int       `db:"buyer_id"`
	SellerID  int       `db:"seller_id"`
	CreatedAt time.Time `db:"created_at"`
	Price     float64   `db:"price"`
}

type User struct {
	ID        int       `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	Username  string    `db:"username"`
	Role      string    `db:"role"`
	Bio       string    `db:"bio"`
	Avatar    string    `db:"avatar"`
	Email     string    `db:"email"`
	Password  string    `db:"password"`
	Verfied   bool      `db:"verified"`
}
