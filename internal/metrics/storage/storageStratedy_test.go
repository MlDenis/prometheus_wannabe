package storage

import (
	"context"
	"testing"

	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
	"github.com/MlDenis/prometheus_wannabe/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type configMock struct {
	mock.Mock
}

type metricStorageMock struct {
	mock.Mock
}

const (
	metricType          = "metricType"
	metricName          = "metricName"
	metricValue float64 = 100
)

func TestStorageStrategy_AddGaugeMetricValue(t *testing.T) {
	tests := []struct {
		name                    string
		syncMode                bool
		inMemoryStorageError    error
		backupStorageErrorError error
		expectedResult          []metrics.Metric
		expectedError           error
	}{
		{
			name:                 "noSync_inMemoryStorage_error",
			syncMode:             false,
			inMemoryStorageError: test.ErrTest,
			expectedError:        test.ErrTest,
		},
		{
			name:                 "sync_inMemoryStorage_error",
			syncMode:             true,
			inMemoryStorageError: test.ErrTest,
			expectedError:        test.ErrTest,
		},
		{
			name:                    "noSync_backupStorage_error",
			syncMode:                false,
			backupStorageErrorError: test.ErrTest,
			expectedResult:          []metrics.Metric{test.CreateGaugeMetric("resultMetric", 100)},
		},
		{
			name:                    "sync_backupStorage_error",
			syncMode:                true,
			backupStorageErrorError: test.ErrTest,
			expectedError:           test.ErrTest,
		},
		{
			name:           "noSync_success",
			syncMode:       false,
			expectedResult: []metrics.Metric{test.CreateGaugeMetric("resultMetric", 100)},
		},
		{
			name:           "sync_success",
			syncMode:       true,
			expectedResult: []metrics.Metric{test.CreateGaugeMetric("resultMetric", 100)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			confMock := new(configMock)
			inMemoryStorageMock := new(metricStorageMock)
			backupStorageMock := new(metricStorageMock)

			metricsList := []metrics.Metric{test.CreateGaugeMetric(metricName, metricValue)}

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("AddMetricValues", ctx, metricsList).Return(tt.expectedResult, tt.inMemoryStorageError)
			backupStorageMock.On("AddMetricValues", ctx, tt.expectedResult).Return(tt.expectedResult, tt.backupStorageErrorError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, backupStorageMock)
			actualResult, actualError := strategy.AddMetricValues(ctx, metricsList)

			assert.Equal(t, tt.expectedResult, actualResult)
			assert.ErrorIs(t, actualError, tt.expectedError)

			inMemoryStorageMock.AssertCalled(t, "AddMetricValues", ctx, metricsList)

			if tt.inMemoryStorageError == nil {
				if tt.syncMode {
					backupStorageMock.AssertCalled(t, "AddMetricValues", ctx, tt.expectedResult)
				} else {
					backupStorageMock.AssertNotCalled(t, "AddMetricValues", mock.Anything, mock.Anything)
				}
			} else {
				backupStorageMock.AssertNotCalled(t, "AddMetricValues", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestStorageStrategy_AddCounterMetricValue(t *testing.T) {
	tests := []struct {
		name                    string
		syncMode                bool
		inMemoryStorageError    error
		backupStorageErrorError error
		expectedResult          []metrics.Metric
		expectedError           error
	}{
		{
			name:                 "noSync_inMemoryStorage_error",
			syncMode:             false,
			inMemoryStorageError: test.ErrTest,
			expectedError:        test.ErrTest,
		},
		{
			name:                 "sync_inMemoryStorage_error",
			syncMode:             true,
			inMemoryStorageError: test.ErrTest,
			expectedError:        test.ErrTest,
		},
		{
			name:                    "noSync_backupStorage_error",
			syncMode:                false,
			backupStorageErrorError: test.ErrTest,
			expectedResult:          []metrics.Metric{test.CreateCounterMetric("resultMetric", 100)},
		},
		{
			name:                    "sync_backupStorage_error",
			syncMode:                true,
			backupStorageErrorError: test.ErrTest,
			expectedError:           test.ErrTest,
		},
		{
			name:           "noSync_success",
			syncMode:       false,
			expectedResult: []metrics.Metric{test.CreateCounterMetric("resultMetric", 100)},
		},
		{
			name:           "sync_success",
			syncMode:       true,
			expectedResult: []metrics.Metric{test.CreateCounterMetric("resultMetric", 100)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			confMock := new(configMock)
			inMemoryStorageMock := new(metricStorageMock)
			backupStorageMock := new(metricStorageMock)

			metricsList := []metrics.Metric{test.CreateCounterMetric(metricName, metricValue)}

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("AddMetricValues", ctx, metricsList).Return(tt.expectedResult, tt.inMemoryStorageError)
			backupStorageMock.On("AddMetricValues", ctx, tt.expectedResult).Return(tt.expectedResult, tt.backupStorageErrorError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, backupStorageMock)
			actualResult, actualError := strategy.AddMetricValues(ctx, metricsList)

			assert.Equal(t, tt.expectedResult, actualResult)
			assert.ErrorIs(t, actualError, tt.expectedError)

			inMemoryStorageMock.AssertCalled(t, "AddMetricValues", ctx, metricsList)

			if tt.inMemoryStorageError == nil {
				if tt.syncMode {
					backupStorageMock.AssertCalled(t, "AddMetricValues", ctx, tt.expectedResult)
				} else {
					backupStorageMock.AssertNotCalled(t, "AddMetricValues", mock.Anything, mock.Anything)
				}
			} else {
				backupStorageMock.AssertNotCalled(t, "AddMetricValues", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestStorageStrategy_GetMetricValues(t *testing.T) {
	result := map[string]map[string]string{}

	tests := []struct {
		name           string
		syncMode       bool
		storageResult  map[string]map[string]string
		storageError   error
		expectedResult map[string]map[string]string
		expectedError  error
	}{
		{
			name:          "noSync_error",
			syncMode:      false,
			storageError:  test.ErrTest,
			expectedError: test.ErrTest,
		},
		{
			name:          "sync_error",
			syncMode:      true,
			storageError:  test.ErrTest,
			expectedError: test.ErrTest,
		},
		{
			name:           "noSync_success",
			syncMode:       true,
			storageResult:  result,
			expectedResult: result,
		},
		{
			name:           "sync_success",
			syncMode:       true,
			storageResult:  result,
			expectedResult: result,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			confMock := new(configMock)
			inMemoryStorageMock := new(metricStorageMock)
			backupStorageMock := new(metricStorageMock)

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("GetMetricValues", ctx).Return(tt.storageResult, tt.storageError)
			backupStorageMock.On("GetMetricValues", ctx).Return(tt.storageResult, tt.storageError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, backupStorageMock)
			actualResult, actualError := strategy.GetMetricValues(ctx)

			assert.Equal(t, tt.expectedResult, actualResult)
			assert.Equal(t, tt.expectedError, actualError)

			inMemoryStorageMock.AssertCalled(t, "GetMetricValues", ctx)
			backupStorageMock.AssertNotCalled(t, "GetMetricValues", mock.Anything)
		})
	}
}

func TestStorageStrategy_GetMetric(t *testing.T) {
	resultMetric := test.CreateGaugeMetric(metricName, metricValue)
	tests := []struct {
		name           string
		syncMode       bool
		storageResult  metrics.Metric
		storageError   error
		expectedResult metrics.Metric
		expectedError  error
	}{
		{
			name:          "noSync_error",
			syncMode:      false,
			storageError:  test.ErrTest,
			expectedError: test.ErrTest,
		},
		{
			name:          "sync_error",
			syncMode:      true,
			storageError:  test.ErrTest,
			expectedError: test.ErrTest,
		},
		{
			name:           "noSync_success",
			syncMode:       true,
			storageResult:  resultMetric,
			expectedResult: resultMetric,
		},
		{
			name:           "sync_success",
			syncMode:       true,
			storageResult:  resultMetric,
			expectedResult: resultMetric,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			confMock := new(configMock)
			inMemoryStorageMock := new(metricStorageMock)
			backupStorageMock := new(metricStorageMock)

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("GetMetric", ctx, metricType, metricName).Return(tt.storageResult, tt.storageError)
			backupStorageMock.On("GetMetric", ctx, metricType, metricName).Return(tt.storageResult, tt.storageError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, backupStorageMock)
			actualResult, actualError := strategy.GetMetric(ctx, metricType, metricName)

			assert.Equal(t, tt.expectedResult, actualResult)
			assert.Equal(t, tt.expectedError, actualError)

			inMemoryStorageMock.AssertCalled(t, "GetMetric", ctx, metricType, metricName)
			backupStorageMock.AssertNotCalled(t, "GetMetric", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestStorageStrategy_Restore(t *testing.T) {
	values := map[string]map[string]string{}

	tests := []struct {
		name          string
		syncMode      bool
		storageError  error
		expectedError error
	}{
		{
			name:          "noSync_error",
			syncMode:      false,
			storageError:  test.ErrTest,
			expectedError: test.ErrTest,
		},
		{
			name:          "sync_error",
			syncMode:      true,
			storageError:  test.ErrTest,
			expectedError: test.ErrTest,
		},
		{
			name:     "noSync_success",
			syncMode: true,
		},
		{
			name:     "sync_success",
			syncMode: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			confMock := new(configMock)
			inMemoryStorageMock := new(metricStorageMock)
			backupStorageMock := new(metricStorageMock)

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("Restore", ctx, values).Return(tt.storageError)
			backupStorageMock.On("Restore", ctx, values).Return(tt.storageError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, backupStorageMock)
			actualError := strategy.Restore(ctx, values)

			assert.Equal(t, tt.expectedError, actualError)

			inMemoryStorageMock.AssertCalled(t, "Restore", ctx, values)
			backupStorageMock.AssertNotCalled(t, "Restore", mock.Anything, mock.Anything)
		})
	}
}

func TestStorageStrategy_CreateBackup(t *testing.T) {
	values := map[string]map[string]string{}

	tests := []struct {
		name               string
		syncMode           bool
		currentStateValues map[string]map[string]string
		currentStateError  error
		restoreError       error
		expectedError      error
	}{
		{
			name:              "noSync_currentState_error",
			syncMode:          false,
			currentStateError: test.ErrTest,
			expectedError:     test.ErrTest,
		},
		{
			name:              "sync_currentState_error",
			syncMode:          true,
			currentStateError: test.ErrTest,
			expectedError:     test.ErrTest,
		},
		{
			name:               "noSync_restore_error",
			syncMode:           false,
			currentStateValues: values,
			restoreError:       test.ErrTest,
			expectedError:      test.ErrTest,
		},
		{
			name:               "sync_restore_error",
			syncMode:           true,
			currentStateValues: values,
			restoreError:       test.ErrTest,
			expectedError:      test.ErrTest,
		},
		{
			name:               "noSync_success",
			syncMode:           true,
			currentStateValues: values,
		},
		{
			name:               "sync_success",
			syncMode:           true,
			currentStateValues: values,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			confMock := new(configMock)
			inMemoryStorageMock := new(metricStorageMock)
			backupStorageMock := new(metricStorageMock)

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("GetMetricValues", ctx).Return(tt.currentStateValues, tt.currentStateError)
			backupStorageMock.On("Restore", ctx, tt.currentStateValues).Return(tt.restoreError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, backupStorageMock)
			actualError := strategy.CreateBackup(ctx)

			assert.ErrorIs(t, actualError, tt.expectedError)

			inMemoryStorageMock.AssertCalled(t, "GetMetricValues", ctx)

			if tt.currentStateError == nil {
				backupStorageMock.AssertCalled(t, "Restore", ctx, tt.currentStateValues)
			} else {
				backupStorageMock.AssertNotCalled(t, "Restore", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestStorageStrategy_RestoreFromBackup(t *testing.T) {
	values := map[string]map[string]string{}

	tests := []struct {
		name               string
		syncMode           bool
		currentStateValues map[string]map[string]string
		currentStateError  error
		restoreError       error
		expectedError      error
	}{
		{
			name:              "noSync_currentState_error",
			syncMode:          false,
			currentStateError: test.ErrTest,
			expectedError:     test.ErrTest,
		},
		{
			name:              "sync_currentState_error",
			syncMode:          true,
			currentStateError: test.ErrTest,
			expectedError:     test.ErrTest,
		},
		{
			name:               "noSync_restore_error",
			syncMode:           false,
			currentStateValues: values,
			restoreError:       test.ErrTest,
			expectedError:      test.ErrTest,
		},
		{
			name:               "sync_restore_error",
			syncMode:           true,
			currentStateValues: values,
			restoreError:       test.ErrTest,
			expectedError:      test.ErrTest,
		},
		{
			name:               "noSync_success",
			syncMode:           true,
			currentStateValues: values,
		},
		{
			name:               "sync_success",
			syncMode:           true,
			currentStateValues: values,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			confMock := new(configMock)
			inMemoryStorageMock := new(metricStorageMock)
			backupStorageMock := new(metricStorageMock)

			confMock.On("SyncMode").Return(tt.syncMode)
			backupStorageMock.On("GetMetricValues", ctx).Return(tt.currentStateValues, tt.currentStateError)
			inMemoryStorageMock.On("Restore", ctx, tt.currentStateValues).Return(tt.restoreError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, backupStorageMock)
			actualError := strategy.RestoreFromBackup(ctx)

			assert.ErrorIs(t, actualError, tt.expectedError)

			backupStorageMock.AssertCalled(t, "GetMetricValues", ctx)

			if tt.currentStateError == nil {
				inMemoryStorageMock.AssertCalled(t, "Restore", ctx, tt.currentStateValues)
			} else {
				inMemoryStorageMock.AssertNotCalled(t, "Restore", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestStorageStrategy_Close(t *testing.T) {
	values := map[string]map[string]string{}

	tests := []struct {
		name               string
		syncMode           bool
		currentStateValues map[string]map[string]string
		currentStateError  error
		restoreError       error
		expectedError      error
	}{
		{
			name:              "noSync_currentState_error",
			syncMode:          false,
			currentStateError: test.ErrTest,
			expectedError:     test.ErrTest,
		},
		{
			name:              "sync_currentState_error",
			syncMode:          true,
			currentStateError: test.ErrTest,
			expectedError:     test.ErrTest,
		},
		{
			name:               "noSync_restore_error",
			syncMode:           false,
			currentStateValues: values,
			restoreError:       test.ErrTest,
			expectedError:      test.ErrTest,
		},
		{
			name:               "sync_restore_error",
			syncMode:           true,
			currentStateValues: values,
			restoreError:       test.ErrTest,
			expectedError:      test.ErrTest,
		},
		{
			name:               "noSync_success",
			syncMode:           true,
			currentStateValues: values,
		},
		{
			name:               "sync_success",
			syncMode:           true,
			currentStateValues: values,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			confMock := new(configMock)
			inMemoryStorageMock := new(metricStorageMock)
			backupStorageMock := new(metricStorageMock)

			confMock.On("SyncMode").Return(tt.syncMode)
			inMemoryStorageMock.On("GetMetricValues", ctx).Return(tt.currentStateValues, tt.currentStateError)
			backupStorageMock.On("Restore", ctx, tt.currentStateValues).Return(tt.restoreError)

			strategy := NewStorageStrategy(confMock, inMemoryStorageMock, backupStorageMock)
			actualError := strategy.Close()

			assert.ErrorIs(t, actualError, tt.expectedError)

			inMemoryStorageMock.AssertCalled(t, "GetMetricValues", ctx)

			if tt.currentStateError == nil {
				backupStorageMock.AssertCalled(t, "Restore", ctx, tt.currentStateValues)
			} else {
				backupStorageMock.AssertNotCalled(t, "Restore", mock.Anything, mock.Anything)
			}
		})
	}
}

func (c *configMock) SyncMode() bool {
	args := c.Called()
	return args.Bool(0)
}

func (s *metricStorageMock) GetMetric(ctx context.Context, metricType string, metricName string) (metrics.Metric, error) {
	args := s.Called(ctx, metricType, metricName)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(metrics.Metric), args.Error(1)
}

func (s *metricStorageMock) AddMetricValues(ctx context.Context, metric []metrics.Metric) ([]metrics.Metric, error) {
	args := s.Called(ctx, metric)
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}

	return result.([]metrics.Metric), args.Error(1)
}

func (s *metricStorageMock) GetMetricValues(ctx context.Context) (map[string]map[string]string, error) {
	args := s.Called(ctx)
	return args.Get(0).(map[string]map[string]string), args.Error(1)
}

func (s *metricStorageMock) GetMetricValue(ctx context.Context, metricType string, metricName string) (float64, error) {
	args := s.Called(ctx, metricType, metricName)
	return args.Get(0).(float64), args.Error(1)
}

func (s *metricStorageMock) Restore(ctx context.Context, metricValues map[string]map[string]string) error {
	args := s.Called(ctx, metricValues)
	return args.Error(0)
}
