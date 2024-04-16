// Code generated by mockery v2.42.2. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	models "github.com/karmaplush/simple-diet-tracker/internal/domain/models"
)

// AccountProvider is an autogenerated mock type for the AccountProvider type
type AccountProvider struct {
	mock.Mock
}

// GetAccountByContextJWT provides a mock function with given fields: ctx
func (_m *AccountProvider) GetAccountByContextJWT(ctx context.Context) (models.Account, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetAccountByContextJWT")
	}

	var r0 models.Account
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (models.Account, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) models.Account); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(models.Account)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewAccountProvider creates a new instance of AccountProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAccountProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *AccountProvider {
	mock := &AccountProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
