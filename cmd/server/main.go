package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/MlDenis/prometheus_wannabe/internal/converter"
	"github.com/MlDenis/prometheus_wannabe/internal/database"
	"github.com/MlDenis/prometheus_wannabe/internal/database/postgre"
	"github.com/MlDenis/prometheus_wannabe/internal/database/stub"
	"github.com/MlDenis/prometheus_wannabe/internal/hash"
	"github.com/MlDenis/prometheus_wannabe/internal/logger"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/html"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/model"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/storage"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/storage/db"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/storage/file"
	"github.com/MlDenis/prometheus_wannabe/internal/metrics/storage/memory"
	"github.com/MlDenis/prometheus_wannabe/internal/worker"

	"github.com/caarlos0/env/v7"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "net/http/pprof"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

// Constants
const (
	counterMetricName = "counter"
	gaugeMetricName   = "gauge"
)

var compressContentTypes = []string{
	"application/javascript",
	"application/json",
	"text/css",
	"text/html",
	"text/plain",
	"text/xml",
}

// Configuration struct for holding server configuration.
type config struct {
	Key           string          `env:"KEY"`
	ServerURL     string          `env:"ADDRESS"`
	StoreInterval int             `env:"STORE_INTERVAL"`
	StoreFile     string          `env:"STORE_FILE"`
	Restore       bool            `env:"RESTORE"`
	DB            string          `env:"DATABASE_DSN"`
	LogLevel      zap.AtomicLevel `env:"LOG_LEVEL"`
}

// Struct for handling context keys related to metrics.
type metricInfoContextKey struct {
	key string
}

// Struct for handling metrics in the context of HTTP requests.
type metricsRequestContext struct {
	requestMetrics []*model.Metrics
	resultMetrics  []*model.Metrics
}

// main is the main entry point for the Prometheus Wannabe server.
// It initializes the server, parses configuration, sets up logging, database, storage, and starts the HTTP server.
func main() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", buildVersion, buildDate, buildCommit)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := createConfig()
	if err != nil {
		panic(logger.WrapError("create config file", err))
	}

	logger.InitLogger(fmt.Sprint(conf.LogLevel))

	logger.SugarLogger.Infof("Starting server with the following configuration:%v", conf)

	var base database.DataBase
	var backupStorage storage.MetricsStorage
	if conf.DB == "" {
		base = &stub.StubDataBase{}
		backupStorage = file.NewFileStorage(conf)
	} else {
		base, err = postgre.NewPostgresDataBase(ctx, conf)
		if err != nil {
			panic(logger.WrapError("create database", err))
		}

		backupStorage = db.NewDBStorage(base)
	}
	defer base.Close()

	inMemoryStorage := memory.NewInMemoryStorage()
	storageStrategy := storage.NewStorageStrategy(conf, inMemoryStorage, backupStorage)
	defer storageStrategy.Close()

	signer := hash.NewSigner(conf)
	converter := model.NewMetricsConverter(conf, signer)
	htmlPageBuilder := html.NewSimplePageBuilder()
	router := initRouter(storageStrategy, converter, htmlPageBuilder, base)

	if conf.Restore {
		logger.SugarLogger.Error("Restore metrics from backup")
		err = storageStrategy.RestoreFromBackup(ctx)
		if err != nil {
			logger.SugarLogger.Errorf("failed to restore state from backup: %v", err)
		}
	}

	if !conf.SyncMode() {
		logger.SugarLogger.Infof("Start periodic backup serice")
		backgroundStore := worker.NewHardWorker(func(ctx context.Context) error { return storageStrategy.CreateBackup(ctx) })
		go backgroundStore.StartWork(ctx, conf.StoreInterval)
	}

	logger.SugarLogger.Infof("Start listen " + conf.ServerURL)
	err = http.ListenAndServe(conf.ServerURL, router)
	if err != nil {
		logger.SugarLogger.Error(err)
	}

	logger.SugarLogger.Sync()
}

// createConfig parses command line flags and environment variables to create a configuration object.
func createConfig() (*config, error) {
	conf := &config{}

	flag.StringVar(&conf.Key, "k", "", "Signer secret key")
	flag.BoolVar(&conf.Restore, "r", true, "Restore metric values from the server backup file")
	flag.IntVar(&conf.StoreInterval, "i", 300, "Store backup interval")
	flag.StringVar(&conf.ServerURL, "a", "localhost:8080", "Server listen URL")
	flag.StringVar(&conf.StoreFile, "f", "/tmp/metrics-db.json", "Backup storage file path")
	flag.StringVar(&conf.DB, "d", "", "Database connection stirng")
	flag.Parse()

	err := env.Parse(conf)
	return conf, err
}

// initRouter initializes the HTTP router for the server, including middleware and route handlers.
func initRouter(metricsStorage storage.MetricsStorage, converter *model.MetricsConverter, htmlPageBuilder html.HTMLPageBuilder, dbStorage database.DataBase) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Compress(gzip.BestSpeed, compressContentTypes...))
	router.Mount("/debug", middleware.Profiler())
	router.Route("/update", func(r chi.Router) {
		r.With(fillSingleJSONContext, updateMetrics(metricsStorage, converter)).
			Post("/", successSingleJSONResponse())
		r.With(fillCommonURLContext, fillGaugeURLContext, updateMetrics(metricsStorage, converter)).
			Post("/gauge/{metricName}/{metricValue}", successURLResponse())
		r.With(fillCommonURLContext, fillCounterURLContext, updateMetrics(metricsStorage, converter)).
			Post("/counter/{metricName}/{metricValue}", successURLResponse())
		r.Post("/{metricType}/{metricName}/{metricValue}", func(w http.ResponseWriter, r *http.Request) {
			message := fmt.Sprintf("unknown metric type: %s", chi.URLParam(r, "metricType"))
			logger.SugarLogger.Error("failed to update metric: " + message)
			http.Error(w, message, http.StatusNotImplemented)
		})
	})

	router.Route("/updates", func(r chi.Router) {
		r.With(fillMultiJSONContext, updateMetrics(metricsStorage, converter)).
			Post("/", successMultiJSONResponse())
	})

	router.Route("/value", func(r chi.Router) {
		r.With(fillSingleJSONContext, fillMetricValues(metricsStorage, converter)).
			Post("/", successSingleJSONResponse())

		r.With(fillCommonURLContext, fillMetricValues(metricsStorage, converter)).
			Get("/{metricType}/{metricName}", successURLValueResponse(converter))
	})

	router.Route("/ping", func(r chi.Router) {
		r.Get("/", handleDBPing(dbStorage))
	})

	router.Route("/", func(r chi.Router) {
		r.Get("/", handleMetricsPage(htmlPageBuilder, metricsStorage))
		r.Get("/metrics", handleMetricsPage(htmlPageBuilder, metricsStorage))
	})

	return router
}

// fillCommonURLContext is a middleware to fill common metric information in the context for HTTP requests.
func fillCommonURLContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metricsContext := ensureMetricsContext(r)
		metricsContext.requestMetrics = append(metricsContext.requestMetrics, &model.Metrics{
			ID:    chi.URLParam(r, "metricName"),
			MType: chi.URLParam(r, "metricType"),
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// fillGaugeURLContext is a middleware to fill gauge metric information in the context for HTTP requests.
func fillGaugeURLContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metricsContext := ensureMetricsContext(r)
		if len(metricsContext.requestMetrics) != 1 {
			logger.SugarLogger.Error("fillGaugeURLContext: wrong context")
			http.Error(w, "fillGaugeURLContext: wrong context", http.StatusInternalServerError)
			return
		}

		strValue := chi.URLParam(r, "metricValue")
		value, err := converter.ToFloat64(strValue)
		if err != nil {
			http.Error(w, logger.WrapError(fmt.Sprintf("parse value: %v", strValue), err).Error(), http.StatusBadRequest)
			return
		}

		metricsContext.requestMetrics[0].MType = gaugeMetricName
		metricsContext.requestMetrics[0].Value = &value
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// fillCounterURLContext is a middleware to fill counter metric information in the context for HTTP requests.
func fillCounterURLContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metricsContext := ensureMetricsContext(r)
		if len(metricsContext.requestMetrics) != 1 {
			logger.SugarLogger.Error("fillCounterURLContext: wrong context")
			http.Error(w, "fillCounterURLContext: wrong context", http.StatusInternalServerError)
			return
		}

		strValue := chi.URLParam(r, "metricValue")
		value, err := converter.ToInt64(strValue)
		if err != nil {
			http.Error(w, logger.WrapError(fmt.Sprintf("parse value: %v", strValue), err).Error(), http.StatusBadRequest)
			return
		}

		metricsContext.requestMetrics[0].MType = counterMetricName
		metricsContext.requestMetrics[0].Delta = &value
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// fillSingleJSONContext is a middleware to fill single JSON metric information in the context for HTTP requests.
func fillSingleJSONContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metricsContext := ensureMetricsContext(r)
		var reader io.Reader
		if r.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, logger.WrapError("create gzip reader", err).Error(), http.StatusInternalServerError)
				return
			}
			reader = gz
			defer gz.Close()
		} else {
			reader = r.Body
		}

		metricContext := &model.Metrics{}
		metricsContext.requestMetrics = append(metricsContext.requestMetrics, metricContext)

		err := json.NewDecoder(reader).Decode(metricContext)
		if err != nil {
			http.Error(w, logger.WrapError("unmarhsal json context", err).Error(), http.StatusBadRequest)
			return
		}

		if metricContext.ID == "" {
			logger.SugarLogger.Error("Fail to collect json context: metric name is missed")
			http.Error(w, "metric name is missed", http.StatusBadRequest)
			return
		}

		if metricContext.MType == "" {
			logger.SugarLogger.Error("Fail to collect json context: metric type is missed")
			http.Error(w, "metric types is missed", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// fillMultiJSONContext is a middleware to fill multiple JSON metric information in the context for HTTP requests.
func fillMultiJSONContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, metricsContext := ensureMetricsContext(r)
		var reader io.Reader
		if r.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, logger.WrapError("create gzip reader", err).Error(), http.StatusInternalServerError)
				return
			}
			reader = gz
			defer gz.Close()
		} else {
			reader = r.Body
		}

		metricsContext.requestMetrics = []*model.Metrics{}
		err := json.NewDecoder(reader).Decode(&metricsContext.requestMetrics)
		if err != nil {
			http.Error(w, logger.WrapError("unmarshal request metrics", err).Error(), http.StatusBadRequest)
			return
		}

		for _, requestMetric := range metricsContext.requestMetrics {
			if requestMetric.ID == "" {
				logger.SugarLogger.Error("Fail to collect json context: metric name is missed")
				http.Error(w, "metric name is missed", http.StatusBadRequest)
				return
			}

			if requestMetric.MType == "" {
				logger.SugarLogger.Error("Fail to collect json context: metric type is missed")
				http.Error(w, "metric types is missed", http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// updateMetrics is a middleware to handle metric updates in the storage.
func updateMetrics(storage storage.MetricsStorage, converter *model.MetricsConverter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, metricsContext := ensureMetricsContext(r)
			metricsList := make([]metrics.Metric, len(metricsContext.requestMetrics))
			for i, metricContext := range metricsContext.requestMetrics {
				metric, err := converter.FromModelMetric(metricContext)
				if err != nil {
					logger.SugarLogger.Errorf("Fail to parse metric: %v", err)

					var errUnknownMetricType *model.UnknownMetricTypeError
					if errors.As(err, &errUnknownMetricType) {
						http.Error(w, fmt.Sprintf("unknown metric type: %s", errUnknownMetricType.UnknownType), http.StatusNotImplemented)
					} else {
						http.Error(w, err.Error(), http.StatusBadRequest)
					}
					return
				}

				metricsList[i] = metric
			}

			resultMetrics, err := storage.AddMetricValues(ctx, metricsList)
			if err != nil {
				http.Error(w, logger.WrapError("update metric", err).Error(), http.StatusInternalServerError)
				return
			}

			metricsContext.resultMetrics = make([]*model.Metrics, len(resultMetrics))
			for i, resultMetric := range resultMetrics {
				newValue, err := converter.ToModelMetric(resultMetric)
				if err != nil {
					http.Error(w, logger.WrapError("convert metric", err).Error(), http.StatusInternalServerError)
					return
				}

				logger.SugarLogger.Errorf("Updated metric: %v. newValue: %v", resultMetric.GetName(), newValue)
				metricsContext.resultMetrics[i] = newValue
			}

			next.ServeHTTP(w, r)
		})
	}
}

// fillMetricValues is a middleware to fill metric values from the storage in the context for HTTP requests.
func fillMetricValues(storage storage.MetricsStorage, converter *model.MetricsConverter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, metricsContext := ensureMetricsContext(r)
			metricsContext.resultMetrics = make([]*model.Metrics, len(metricsContext.requestMetrics))
			for i, metricContext := range metricsContext.requestMetrics {
				metric, err := storage.GetMetric(ctx, metricContext.MType, metricContext.ID)
				if err != nil {
					logger.SugarLogger.Errorf("Fail to get metric value: %v", err)
					http.Error(w, "Metric not found", http.StatusNotFound)
					return
				}

				resultValue, err := converter.ToModelMetric(metric)
				if err != nil {
					http.Error(w, logger.WrapError("get metric value", err).Error(), http.StatusInternalServerError)
					return
				}

				metricsContext.resultMetrics[i] = resultValue
			}

			next.ServeHTTP(w, r)
		})
	}
}

// successURLValueResponse is a handler function to respond with the value of a single metric in plain text.
func successURLValueResponse(converter *model.MetricsConverter) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, metricsContext := ensureMetricsContext(r)

		if len(metricsContext.resultMetrics) != 1 {
			logger.SugarLogger.Error("successURLValueResponse: wrong context")
			http.Error(w, "successURLValueResponse: wrong context", http.StatusInternalServerError)
			return
		}

		metric, err := converter.FromModelMetric(metricsContext.resultMetrics[0])
		if err != nil {
			http.Error(w, logger.WrapError("convert result metric", err).Error(), http.StatusInternalServerError)
			return
		}

		successResponse(w, "text/plain", metric.GetStringValue())
	}
}

// handleMetricsPage is a handler function to respond with an HTML page containing metric values.
func handleMetricsPage(builder html.HTMLPageBuilder, storage storage.MetricsStorage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		values, err := storage.GetMetricValues(r.Context())
		if err != nil {
			http.Error(w, logger.WrapError("get metric values", err).Error(), http.StatusInternalServerError)
			return
		}
		successResponse(w, "text/html", builder.BuildMetricsPage(values))
	}
}

// successURLResponse is a handler function to respond with a success message in plain text.
func successURLResponse() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		successResponse(w, "text/plain", "ok")
	}
}

// successSingleJSONResponse is a handler function to respond with a single JSON metric in JSON format.
func successSingleJSONResponse() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, metricsContext := ensureMetricsContext(r)

		if len(metricsContext.resultMetrics) != 1 {
			logger.SugarLogger.Error("successSingleJSONResponse: wrong context")
			http.Error(w, "successSingleJSONResponse: wrong context", http.StatusInternalServerError)
			return
		}

		result, err := json.Marshal(metricsContext.resultMetrics[0])
		if err != nil {
			http.Error(w, logger.WrapError("serialise result", err).Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(result)
		if err != nil {
			logger.SugarLogger.Errorf("failed to write response: %v", err)
		}
	}
}

// successMultiJSONResponse is a handler function to respond with a stub JSON metric in JSON format.
func successMultiJSONResponse() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		stubResult := &model.Metrics{}
		result, err := json.Marshal(stubResult)
		if err != nil {
			http.Error(w, logger.WrapError("serialise result", err).Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(result)
		if err != nil {
			logger.SugarLogger.Errorf("failed to write response: %v", err)
		}
	}
}

// successResponse is a generic function to respond with a success message in the specified content type.
func successResponse(w http.ResponseWriter, contentType string, message string) {
	w.Header().Add("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(message))
	if err != nil {
		logger.SugarLogger.Errorf("failed to write response: %v", err)
	}
}

// handleDBPing is a handler function to respond with a success message if the database is pingable.
func handleDBPing(dbStorage database.DataBase) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := dbStorage.Ping(r.Context())
		if err == nil {
			successResponse(w, "text/plain", "ok")
		} else {
			http.Error(w, logger.WrapError("ping database", err).Error(), http.StatusInternalServerError)
		}
	}
}

// ensureMetricsContext ensures that the metrics context is present in the HTTP request context.
func ensureMetricsContext(r *http.Request) (context.Context, *metricsRequestContext) {
	const metricsContextKey = "metricsContextKey"
	ctx := r.Context()
	metricsContext, ok := ctx.Value(metricInfoContextKey{key: metricsContextKey}).(*metricsRequestContext)
	if !ok {
		metricsContext = &metricsRequestContext{}
		ctx = context.WithValue(r.Context(), metricInfoContextKey{key: metricsContextKey}, metricsContext)
	}

	return ctx, metricsContext
}

// StoreFilePath returns the configured file path for storing backups.
func (c *config) StoreFilePath() string {
	return c.StoreFile
}

// SyncMode returns true if the server is running in sync mode, i.e., using a database or with zero store interval.
func (c *config) SyncMode() bool {
	return c.DB != "" || c.StoreInterval == 0
}

// String returns a formatted string representation of the server configuration.
func (c *config) String() string {
	return fmt.Sprintf("\nServerURL:\t%v\nStoreInterval:\t%v\nStoreFile:\t%v\nRestore:\t%v\nDb:\t%v",
		c.ServerURL, c.StoreInterval, c.StoreFile, c.Restore, c.DB)
}

// GetKey returns the secret key used for signing metrics.
func (c *config) GetKey() []byte {
	return []byte(c.Key)
}

// SignMetrics returns true if metrics should be signed.
func (c *config) SignMetrics() bool {
	return c.Key != ""
}

// GetConnectionString returns the configured database connection string.
func (c *config) GetConnectionString() string {
	return c.DB
}
