// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// EmailAdaptorInterface is an autogenerated mock type for the EmailAdaptorInterface type
type EmailAdaptorInterface struct {
	mock.Mock
}

// SendReceiptEmail provides a mock function with given fields: ctx, sender, to, sellerName, buyerName, imageID, imageTitle, paymentMethod
func (_m *EmailAdaptorInterface) SendReceiptEmail(ctx context.Context, sender string, to []string, sellerName string, buyerName string, imageID string, imageTitle string, paymentMethod string) {
	_m.Called(ctx, sender, to, sellerName, buyerName, imageID, imageTitle, paymentMethod)
}

// SendResetPassEmail provides a mock function with given fields: ctx, sender, to, name, resetLink
func (_m *EmailAdaptorInterface) SendResetPassEmail(ctx context.Context, sender string, to []string, name string, resetLink string) {
	_m.Called(ctx, sender, to, name, resetLink)
}

// SendWelcomeEmail provides a mock function with given fields: ctx, sender, to, name
func (_m *EmailAdaptorInterface) SendWelcomeEmail(ctx context.Context, sender string, to []string, name string) {
	_m.Called(ctx, sender, to, name)
}