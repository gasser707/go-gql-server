// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	"strconv"
)

type Image struct {
	ID          string   `json:"id"`
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	User        *User    `json:"user"`
	Labels      []string `json:"labels"`
	URL         string   `json:"url"`
}

type NewImage struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Labels      []string `json:"labels"`
	URL         string   `json:"url"`
}

type NewUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Images   []*Image `json:"images"`
	Role     Role     `json:"role"`
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
