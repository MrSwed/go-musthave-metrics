// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/MrSwed/go-musthave-metrics/internal/repository (interfaces: Repository)

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	domain "github.com/MrSwed/go-musthave-metrics/internal/domain"
	repository "github.com/MrSwed/go-musthave-metrics/internal/repository"
	gomock "github.com/golang/mock/gomock"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// GetAllCounters mocks base method.
func (m *MockRepository) GetAllCounters() (domain.Counters, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllCounters")
	ret0, _ := ret[0].(domain.Counters)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllCounters indicates an expected call of GetAllCounters.
func (mr *MockRepositoryMockRecorder) GetAllCounters() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllCounters", reflect.TypeOf((*MockRepository)(nil).GetAllCounters))
}

// GetAllGauges mocks base method.
func (m *MockRepository) GetAllGauges() (domain.Gauges, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllGauges")
	ret0, _ := ret[0].(domain.Gauges)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllGauges indicates an expected call of GetAllGauges.
func (mr *MockRepositoryMockRecorder) GetAllGauges() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllGauges", reflect.TypeOf((*MockRepository)(nil).GetAllGauges))
}

// GetCounter mocks base method.
func (m *MockRepository) GetCounter(arg0 string) (domain.Counter, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCounter", arg0)
	ret0, _ := ret[0].(domain.Counter)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCounter indicates an expected call of GetCounter.
func (mr *MockRepositoryMockRecorder) GetCounter(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCounter", reflect.TypeOf((*MockRepository)(nil).GetCounter), arg0)
}

// GetGauge mocks base method.
func (m *MockRepository) GetGauge(arg0 string) (domain.Gauge, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGauge", arg0)
	ret0, _ := ret[0].(domain.Gauge)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetGauge indicates an expected call of GetGauge.
func (mr *MockRepositoryMockRecorder) GetGauge(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGauge", reflect.TypeOf((*MockRepository)(nil).GetGauge), arg0)
}

// RestoreFromFile mocks base method.
func (m *MockRepository) RestoreFromFile(arg0 *repository.MemStorageRepo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RestoreFromFile", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RestoreFromFile indicates an expected call of RestoreFromFile.
func (mr *MockRepositoryMockRecorder) RestoreFromFile(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RestoreFromFile", reflect.TypeOf((*MockRepository)(nil).RestoreFromFile), arg0)
}

// SaveToFile mocks base method.
func (m *MockRepository) SaveToFile(arg0 *repository.MemStorageRepo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveToFile", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveToFile indicates an expected call of SaveToFile.
func (mr *MockRepositoryMockRecorder) SaveToFile(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveToFile", reflect.TypeOf((*MockRepository)(nil).SaveToFile), arg0)
}

// SetCounter mocks base method.
func (m *MockRepository) SetCounter(arg0 string, arg1 domain.Counter) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetCounter", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetCounter indicates an expected call of SetCounter.
func (mr *MockRepositoryMockRecorder) SetCounter(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetCounter", reflect.TypeOf((*MockRepository)(nil).SetCounter), arg0, arg1)
}

// SetGauge mocks base method.
func (m *MockRepository) SetGauge(arg0 string, arg1 domain.Gauge) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetGauge", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetGauge indicates an expected call of SetGauge.
func (mr *MockRepositoryMockRecorder) SetGauge(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetGauge", reflect.TypeOf((*MockRepository)(nil).SetGauge), arg0, arg1)
}

// SetMetrics mocks base method.
func (m *MockRepository) SetMetrics(arg0 []domain.Metric) ([]domain.Metric, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetMetrics", arg0)
	ret0, _ := ret[0].([]domain.Metric)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SetMetrics indicates an expected call of SetMetrics.
func (mr *MockRepositoryMockRecorder) SetMetrics(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetMetrics", reflect.TypeOf((*MockRepository)(nil).SetMetrics), arg0)
}
