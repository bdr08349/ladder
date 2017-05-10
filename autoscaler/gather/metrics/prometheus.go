package metrics

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"time"

	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"github.com/themotion/ladder/autoscaler/gather"
	"github.com/themotion/ladder/log"
	"github.com/themotion/ladder/types"
	utilmath "github.com/themotion/ladder/util/math"
)

const (
	// Opts
	pmAddresses = "addresses"
	pmQuery     = "query"

	// the name
	pmRegName = "prometheus_metric"
	pmType    = ""
)

// PrometheusMetric represents an object for gathering metrics from Prometheus
type PrometheusMetric struct {
	addresses []string
	qry       string

	apiCs []v1.API // api clients one per endpoint
	log   *log.Log // custom logger

}

func init() {
	gather.Register(pmRegName, gather.CreatorFunc(func(ctx context.Context, opts map[string]interface{}) (gather.Gatherer, error) {
		return NewPrometheusMetric(ctx, opts)
	}))
}

// NewPrometheusMetric creates an Prometheus gatherer
func NewPrometheusMetric(ctx context.Context, opts map[string]interface{}) (p *PrometheusMetric, err error) {
	// Recover from wrong type assertions
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	p = &PrometheusMetric{}

	var ok bool

	// interfaces and arrays stuff conversion
	addrsI := opts[pmAddresses].([]interface{})
	addrs := make([]string, len(addrsI))

	for i, a := range addrsI {
		addrs[i] = a.(string)
	}
	p.addresses = addrs

	// Check address
	if len(p.addresses) == 0 {
		return nil, fmt.Errorf("%s configuration opt is required", pmAddresses)
	}

	// Check query
	if p.qry, ok = opts[pmQuery].(string); !ok || p.qry == "" {
		return nil, fmt.Errorf("%s configuration opt is required", pmQuery)
	}

	p.apiCs = make([]v1.API, len(p.addresses))

	for i, a := range p.addresses {
		// Create the client
		apiC, err := api.NewClient(api.Config{Address: a})
		if err != nil {
			return nil, err
		}
		p.apiCs[i] = v1.NewAPI(apiC)
	}

	// Logger
	asName, ok := ctx.Value("autoscaler").(string)
	if !ok {
		asName = "unknown"
	}
	p.log = log.WithFields(log.Fields{
		"autoscaler": asName,
		"kind":       "gatherer",
		"name":       pmRegName,
	})

	return
}

// query wraps the query to Prometheus handling the retry of the queries on
// different Prometheus endpoints
func (p *PrometheusMetric) query(q string, ts time.Time) (model.Value, error) {
	errs := []error{}

	// Make the query in order on each enpoint until one success
	for i, c := range p.apiCs {
		res, err := c.Query(context.TODO(), q, ts)
		// if ok then return the response
		if err == nil {
			return res, nil
		}
		// Add the error to the list of errors
		errs = append(errs, err)
		p.log.Warningf("prometheus '%d' endpoint failed: %v", i, err)
	}

	b := bytes.Buffer{}
	for _, e := range errs {
		b.WriteString(e.Error())
		b.WriteString("; ")
	}

	return nil, fmt.Errorf(b.String())
}

// Gather will gather metrics from prometheus
func (p *PrometheusMetric) Gather(_ context.Context) (types.Quantity, error) {
	q := types.Quantity{}
	// Make query request
	resp, err := p.query(p.qry, time.Now().UTC())
	if err != nil {
		return q, err
	}

	// Only vectors are valid metrics
	if resp.Type() != model.ValVector {
		return q, fmt.Errorf("received metric needs to be a vector, received: %s", resp.Type())
	}
	m := resp.(model.Vector)

	// Only one sample is valid
	if len(m) != 1 {
		return q, fmt.Errorf("wrong samples length, should be one, current is: %d", len(m))
	}

	// Get the value (round the value) and if there is no metric error
	v := float64((*model.Sample)(m[0]).Value)
	if math.IsNaN(v) {
		return q, fmt.Errorf("prometheus returned a metric is NaN, this means no metric")
	}
	q.Q = utilmath.RoundInt64(v)

	p.log.Debugf("Got prometheus metric:\n  - %s", resp)

	return q, nil
}
