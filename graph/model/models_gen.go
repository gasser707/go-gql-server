// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

type BuyImageInput struct {
	ImageID string `json:"imageId"`
}

type DeleteImageInput struct {
	ID string `json:"id"`
}

type Image struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	User        *User      `json:"user"`
	Labels      []string   `json:"labels"`
	URL         string     `json:"url"`
	Private     bool       `json:"private"`
	ForSale     bool       `json:"forSale"`
	Created     *time.Time `json:"created"`
}

type ImageFilterInput struct {
	ID          *string  `json:"id"`
	UserID      *string  `json:"userId"`
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Labels      []string `json:"labels"`
	Private     *bool    `json:"private"`
	ForSale     *bool    `json:"forSale"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type NewImageInput struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Labels      []string `json:"labels"`
	URL         string   `json:"url"`
	Private     bool     `json:"private"`
	ForSale     bool     `json:"forSale"`
}

type NewUserInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
	Avatar   string `json:"avatar"`
}

type Sale struct {
	ID     string     `json:"id"`
	Image  *Image     `json:"image"`
	Buyer  *User      `json:"buyer"`
	Seller *User      `json:"seller"`
	Time   *time.Time `json:"time"`
}

type UpdateImageInput struct {
	ID          string   `json:"id"`
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Labels      []string `json:"labels"`
	URL         *string  `json:"url"`
	Private     *bool    `json:"private"`
	ForSale     *bool    `json:"forSale"`
}

type UpdateUserInput struct {
	ID       string  `json:"id"`
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Bio      *string `json:"bio"`
	Avatar   *string `json:"avatar"`
}

type User struct {
	ID       string     `json:"id"`
	Username string     `json:"username"`
	Email    string     `json:"email"`
	Role     Role       `json:"role"`
	Bio      string     `json:"bio"`
	Avatar   string     `json:"avatar"`
	Joined   *time.Time `json:"joined"`
	Images   []*Image   `json:"images"`
}

type UserFilterInput struct {
	ID       *string `json:"id"`
	Username *string `json:"username"`
	Email    *string `json:"email"`
	Role     *Role   `json:"role"`
	Bio      *string `json:"bio"`
}

type Role string

const (
	RoleAdmin     Role = "ADMIN"
	RoleUser      Role = "USER"
	RoleModerator Role = "MODERATOR"
)

var AllRole = []Role{
	RoleAdmin,
	RoleUser,
	RoleModerator,
}

func (e Role) IsValid() bool {
	switch e {
	case RoleAdmin, RoleUser, RoleModerator:
		return true
	}
	return false
}

func (e Role) String() string {
	return string(e)
}

func (e *Role) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = Role(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid Role", str)
	}
	return nil
}

func (e Role) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
