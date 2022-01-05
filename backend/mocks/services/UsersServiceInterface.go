// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	context "context"

	custom "github.com/gasser707/go-gql-server/graphql/custom"
	mock "github.com/stretchr/testify/mock"

	model "github.com/gasser707/go-gql-server/graphql/model"
)

// UsersServiceInterface is an autogenerated mock type for the UsersServiceInterface type
type UsersServiceInterface struct {
	mock.Mock
}

// GetUserById provides a mock function with given fields: ctx, ID
func (_m *UsersServiceInterface) GetUserById(ctx context.Context, ID string) (*custom.User, error) {
	ret := _m.Called(ctx, ID)

	var r0 *custom.User
	if rf, ok := ret.Get(0).(func(context.Context, string) *custom.User); ok {
		r0 = rf(ctx, ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custom.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUsers provides a mock function with given fields: ctx, input
func (_m *UsersServiceInterface) GetUsers(ctx context.Context, input *model.UserFilterInput) ([]*custom.User, error) {
	ret := _m.Called(ctx, input)

	var r0 []*custom.User
	if rf, ok := ret.Get(0).(func(context.Context, *model.UserFilterInput) []*custom.User); ok {
		r0 = rf(ctx, input)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*custom.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *model.UserFilterInput) error); ok {
		r1 = rf(ctx, input)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RegisterUser provides a mock function with given fields: ctx, input
func (_m *UsersServiceInterface) RegisterUser(ctx context.Context, input model.NewUserInput) (*custom.User, error) {
	ret := _m.Called(ctx, input)

	var r0 *custom.User
	if rf, ok := ret.Get(0).(func(context.Context, model.NewUserInput) *custom.User); ok {
		r0 = rf(ctx, input)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custom.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.NewUserInput) error); ok {
		r1 = rf(ctx, input)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateUser provides a mock function with given fields: ctx, input
func (_m *UsersServiceInterface) UpdateUser(ctx context.Context, input model.UpdateUserInput) (*custom.User, error) {
	ret := _m.Called(ctx, input)

	var r0 *custom.User
	if rf, ok := ret.Get(0).(func(context.Context, model.UpdateUserInput) *custom.User); ok {
		r0 = rf(ctx, input)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*custom.User)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.UpdateUserInput) error); ok {
		r1 = rf(ctx, input)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
