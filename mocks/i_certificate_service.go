// Code generated by mockery v2.46.2. DO NOT EDIT.

package mocks

import (
	net "net"

	mock "github.com/stretchr/testify/mock"

	tls "crypto/tls"
)

// ICertificateService is an autogenerated mock type for the ICertificateService type
type ICertificateService struct {
	mock.Mock
}

// ExportCACert provides a mock function with given fields: caCert, filePath
func (_m *ICertificateService) ExportCACert(caCert []byte, filePath string) error {
	ret := _m.Called(caCert, filePath)

	if len(ret) == 0 {
		panic("no return value specified for ExportCACert")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func([]byte, string) error); ok {
		r0 = rf(caCert, filePath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GenerateSelfSignedCert provides a mock function with given fields: organization, commonName, dnsNames, ipAddresses
func (_m *ICertificateService) GenerateSelfSignedCert(organization string, commonName string, dnsNames []string, ipAddresses []net.IP) (tls.Certificate, error) {
	ret := _m.Called(organization, commonName, dnsNames, ipAddresses)

	if len(ret) == 0 {
		panic("no return value specified for GenerateSelfSignedCert")
	}

	var r0 tls.Certificate
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, []string, []net.IP) (tls.Certificate, error)); ok {
		return rf(organization, commonName, dnsNames, ipAddresses)
	}
	if rf, ok := ret.Get(0).(func(string, string, []string, []net.IP) tls.Certificate); ok {
		r0 = rf(organization, commonName, dnsNames, ipAddresses)
	} else {
		r0 = ret.Get(0).(tls.Certificate)
	}

	if rf, ok := ret.Get(1).(func(string, string, []string, []net.IP) error); ok {
		r1 = rf(organization, commonName, dnsNames, ipAddresses)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewICertificateService creates a new instance of ICertificateService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewICertificateService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ICertificateService {
	mock := &ICertificateService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
