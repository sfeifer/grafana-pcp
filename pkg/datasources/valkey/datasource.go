package valkey

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/performancecopilot/grafana-pcp/pkg/datasources/valkey/api/pmseries"
	"github.com/performancecopilot/grafana-pcp/pkg/datasources/valkey/resource"
	"github.com/performancecopilot/grafana-pcp/pkg/datasources/valkey/series"
)

// NewDatasource returns datasource.ServeOpts.
func NewDatasource() datasource.ServeOpts {
	im := datasource.NewInstanceManager(newDataSourceInstance)
	ds := &valkeyDatasource{
		im: im,
	}

	mux := datasource.NewQueryTypeMux()
	mux.HandleFunc("", func(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
		var response *backend.QueryDataResponse
		return response, ds.im.Do(ctx, req.PluginContext, func(dsInst *valkeyDatasourceInstance) error {
			var err error
			response, err = dsInst.handleTimeSeriesQueries(ctx, req)
			return err
		})
	})

	return datasource.ServeOpts{
		QueryDataHandler:    mux,
		CheckHealthHandler:  ds,
		CallResourceHandler: ds,
	}
}

type valkeyDatasource struct {
	im instancemgmt.InstanceManager
}

type valkeyDatasourceInstance struct {
	pmseriesAPI     pmseries.API
	resourceService *resource.Service
	seriesService   *series.Service
}

func newDataSourceInstance(ctx context.Context, setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	var basicAuthSettings *pmseries.BasicAuthSettings
	// enable once pmproxy /series supports authentication
	/*if setting.BasicAuthEnabled {
		basicAuthSettings = &pmseries.BasicAuthSettings{
			Username: "",
			Password: "",
		}
	}*/

	pmseriesAPI, err := pmseries.NewPmseriesAPI(setting.URL, basicAuthSettings)
	if err != nil {
		return nil, err
	}

	seriesService, err := series.NewSeriesService(pmseriesAPI, 1024)
	if err != nil {
		return nil, err
	}

	return &valkeyDatasourceInstance{
		pmseriesAPI:     pmseriesAPI,
		resourceService: resource.NewResourceService(pmseriesAPI),
		seriesService:   seriesService,
	}, nil
}

func (ds *valkeyDatasourceInstance) Dispose() {
	// Called before creating a new instance
}

func (ds *valkeyDatasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	u, err := url.Parse(req.URL)
	if err != nil {
		return err
	}

	queryParams, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return err
	}

	method := u.Path
	return ds.im.Do(ctx, req.PluginContext, func(dsInst *valkeyDatasourceInstance) error {
		status := http.StatusOK
		result, err := dsInst.resourceService.CallResource(method, queryParams)
		if err != nil {
			status = http.StatusInternalServerError
			result = struct {
				Error string `json:"error"`
			}{err.Error()}
		}

		respBody, err := json.Marshal(result)
		if err != nil {
			return err
		}

		return sender.Send(&backend.CallResourceResponse{
			Status: status,
			Body:   respBody,
		})
	})
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (ds *valkeyDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	var result *backend.CheckHealthResult
	err := ds.im.Do(ctx, req.PluginContext, func(dsInst *valkeyDatasourceInstance) error {
		pingResponse, err := dsInst.pmseriesAPI.Ping()

		if err != nil {
			result = &backend.CheckHealthResult{
				Status:  backend.HealthStatusError,
				Message: err.Error(),
			}
		} else if !pingResponse.Success {
			result = &backend.CheckHealthResult{
				Status:  backend.HealthStatusError,
				Message: "Datasource is not working. Please check if Valkey is running and consult the pmproxy logs.",
			}
		} else {
			result = &backend.CheckHealthResult{
				Status:  backend.HealthStatusOk,
				Message: "Data source is working",
			}
		}
		return nil
	})

	if err != nil {
		result = &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: err.Error(),
		}
	}
	return result, nil
}
