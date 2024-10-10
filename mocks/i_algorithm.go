// Code generated by mockery v2.46.2. DO NOT EDIT.

package mocks

import (
	pkg "github.com/be-heroes/ultron/pkg"
	mock "github.com/stretchr/testify/mock"
)

// IAlgorithm is an autogenerated mock type for the IAlgorithm type
type IAlgorithm struct {
	mock.Mock
}

// NetworkScore provides a mock function with given fields: node, pod
func (_m *IAlgorithm) NetworkScore(node *pkg.WeightedNode, pod *pkg.WeightedPod) float64 {
	ret := _m.Called(node, pod)

	if len(ret) == 0 {
		panic("no return value specified for NetworkScore")
	}

	var r0 float64
	if rf, ok := ret.Get(0).(func(*pkg.WeightedNode, *pkg.WeightedPod) float64); ok {
		r0 = rf(node, pod)
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}

// NodeScore provides a mock function with given fields: node
func (_m *IAlgorithm) NodeScore(node *pkg.WeightedNode) float64 {
	ret := _m.Called(node)

	if len(ret) == 0 {
		panic("no return value specified for NodeScore")
	}

	var r0 float64
	if rf, ok := ret.Get(0).(func(*pkg.WeightedNode) float64); ok {
		r0 = rf(node)
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}

// PodScore provides a mock function with given fields: pod
func (_m *IAlgorithm) PodScore(pod *pkg.WeightedPod) float64 {
	ret := _m.Called(pod)

	if len(ret) == 0 {
		panic("no return value specified for PodScore")
	}

	var r0 float64
	if rf, ok := ret.Get(0).(func(*pkg.WeightedPod) float64); ok {
		r0 = rf(pod)
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}

// PriceScore provides a mock function with given fields: node
func (_m *IAlgorithm) PriceScore(node *pkg.WeightedNode) float64 {
	ret := _m.Called(node)

	if len(ret) == 0 {
		panic("no return value specified for PriceScore")
	}

	var r0 float64
	if rf, ok := ret.Get(0).(func(*pkg.WeightedNode) float64); ok {
		r0 = rf(node)
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}

// ResourceScore provides a mock function with given fields: node, pod
func (_m *IAlgorithm) ResourceScore(node *pkg.WeightedNode, pod *pkg.WeightedPod) float64 {
	ret := _m.Called(node, pod)

	if len(ret) == 0 {
		panic("no return value specified for ResourceScore")
	}

	var r0 float64
	if rf, ok := ret.Get(0).(func(*pkg.WeightedNode, *pkg.WeightedPod) float64); ok {
		r0 = rf(node, pod)
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}

// StorageScore provides a mock function with given fields: node, pod
func (_m *IAlgorithm) StorageScore(node *pkg.WeightedNode, pod *pkg.WeightedPod) float64 {
	ret := _m.Called(node, pod)

	if len(ret) == 0 {
		panic("no return value specified for StorageScore")
	}

	var r0 float64
	if rf, ok := ret.Get(0).(func(*pkg.WeightedNode, *pkg.WeightedPod) float64); ok {
		r0 = rf(node, pod)
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}

// TotalScore provides a mock function with given fields: node, pod
func (_m *IAlgorithm) TotalScore(node *pkg.WeightedNode, pod *pkg.WeightedPod) float64 {
	ret := _m.Called(node, pod)

	if len(ret) == 0 {
		panic("no return value specified for TotalScore")
	}

	var r0 float64
	if rf, ok := ret.Get(0).(func(*pkg.WeightedNode, *pkg.WeightedPod) float64); ok {
		r0 = rf(node, pod)
	} else {
		r0 = ret.Get(0).(float64)
	}

	return r0
}

// NewIAlgorithm creates a new instance of IAlgorithm. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIAlgorithm(t interface {
	mock.TestingT
	Cleanup(func())
}) *IAlgorithm {
	mock := &IAlgorithm{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
